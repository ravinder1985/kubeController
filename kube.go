package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	coretypev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	kuberestclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// In case you want to test your functions, you can create interface
// // KubernetesPlugin interface is designed to add/update running pods metadata
// type KubernetesPlugin interface {
// 	SystemInitialize() (bool, error)
// 	StartEventWatcher() (bool, error)
// 	UpdateFactsWorkers()
// }

// KubernetesOps ...
type KubernetesOps struct {
	Configs      *kuberestclient.Config
	ClientSet    *kubernetes.Clientset
	PodInterface coretypev1.PodInterface

	UpdateFactsChan chan *corev1.Pod

	Data Data
}

// Facts only going to hold Type and Text about cats
type Facts struct {
	Type string
	Text string
}

// Data object holds facts
type Data struct {
	All []Facts
}

// InitializeData would fetch the facts from the URL configured.
func (data *Data) InitializeData(url string) bool {

	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return false
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return false
	}

	json.Unmarshal(body, &data)
	defer resp.Body.Close()
	return true
}

// GetFacts would return facts about cats.
func (data *Data) GetFacts() string {
	length := len(data.All) + 1
	rand.Seed(time.Now().UnixNano())
	return data.All[rand.Intn(length)].Text
}

func (kubernetesOps *KubernetesOps) startUpdatePodsWorkers(numberOfWorkers int) {
	// Start workers
	log.Println("Start UpdateFactsWorkers")
	for i := 0; i < numberOfWorkers; i++ {
		go kubernetesOps.UpdateFactsWorkers(i)
	}
}

func (kubernetesOps *KubernetesOps) loadData(url string) (bool, error) {
	// NOTE: If endpoint return different types of facts then need to process the data to keep only facts about cats
	if !kubernetesOps.Data.InitializeData(url) {
		log.Println("Unable to initalize facts data")
		return false, errors.New("Failed to get data from http://cat-fact.herokuapp.com/facts")
	}
	return true, nil
}

// SystemInitialize would initilize the system.
func (kubernetesOps *KubernetesOps) SystemInitialize() (bool, error) {

	if kubernetesOps.UpdateFactsChan == nil {
		log.Println("Initialize memory for the UpdateFactsChan")
		kubernetesOps.UpdateFactsChan = make(chan *corev1.Pod, 10)
	}

	var err error

	kubernetesOps.startUpdatePodsWorkers(numberOfWorkers)
	// Local data to local memory for further use.
	_, err = kubernetesOps.loadData(url)
	if err != nil {
		log.Println("Failed to load facts", err)
		return false, err
	}

	kubernetesOps.Configs, err = clientcmd.BuildConfigFromFlags("", kubeConfigLocation)
	if err != nil {
		log.Println("Failed to get to Kubernetes Cluster configurations", err)
		return false, err
	}

	kubernetesOps.ClientSet, err = kubernetes.NewForConfig(kubernetesOps.Configs)
	if err != nil {
		log.Println("Failed to connect to Kubernetes Cluster", err)
		return false, err
	}
	// initilize podInterface to monitor all the names spaces.
	kubernetesOps.PodInterface = kubernetesOps.ClientSet.CoreV1().Pods("")
	return true, nil
}

// ValidateAndPushForUpdate will validate conditions before push to Queue for update
func (kubernetesOps *KubernetesOps) ValidateAndPushForUpdate(pod *corev1.Pod) {
	if pod.Status.Phase == "Running" {
		annotationsMap := pod.GetAnnotations()
		// Make sure there is not any cat-fact already exist
		if _, ok := annotationsMap["cat-fact"]; !ok {
			log.Println("add/update annotation in -", pod.GetNamespace(), pod.Name, pod.Status.PodIP, pod.Status.HostIP, pod.Status.Phase)
			// Update only if there is no cat-fact available already
			kubernetesOps.UpdateFactsChan <- pod
		}
	}
}

// Resync will trigger events every 10 mintue to make sure every pod has a annotations
func (kubernetesOps *KubernetesOps) Resync() {
	for {
		<-time.After(10 * time.Second)
		log.Println("----- Resync -----")
		podsList, _ := kubernetesOps.PodInterface.List(v1.ListOptions{})
		for _, pod := range podsList.Items {
			kubernetesOps.ValidateAndPushForUpdate(&pod)
		}
	}
}

// StartEventWatcher would start the thread to monitor pods activity in the cluster
func (kubernetesOps *KubernetesOps) StartEventWatcher() (bool, error) {

	watch, err := kubernetesOps.PodInterface.Watch(v1.ListOptions{})
	if err != nil {
		log.Println("Unable to get watch channel", err)
		return false, err
	}

	go func() {
		for {
			select {
			case event := <-watch.ResultChan():
				switch event.Type {
				case "ADDED", "MODIFIED":
					log.Println("Type: ", event.Type)
					pod, ok := event.Object.(*corev1.Pod)
					if !ok {
						log.Println("Unexpected object type")
					} else {
						kubernetesOps.ValidateAndPushForUpdate(pod)
					}
				default:
					log.Println("Type: ", event.Type)
				}
			}
		}
	}()
	// Resync to make sure all pods has cat-facts
	go kubernetesOps.Resync()
	return true, nil
}

// UpdateFactsWorkers is responsible to update running pods in kubernetes
func (kubernetesOps *KubernetesOps) UpdateFactsWorkers(id int) {
	fmt.Println("UpdateFactsWorkers thread started: ", id)
	for podDetails := range kubernetesOps.UpdateFactsChan {
		// Get annotations to validate whether facts already exist.
		annotationsMap := podDetails.GetAnnotations()
		if annotationsMap == nil {
			fmt.Println("Initialize map to store annotations")
			annotationsMap = make(map[string]string, 1)
		}

		// fmt.Println("---------", podDetails.ObjectMeta.Namespace, podDetails.Status.ContainerStatuses)
		// Double check in case messages are comming from the different system and there is not any validation for cat-facts
		if _, ok := annotationsMap["cat-fact"]; !ok {
			// Copy the object so we wont change anything to original object in case its getting used somewhere else.
			podDetailsCopy := podDetails.DeepCopy()
			annotationsMap["cat-fact"] = kubernetesOps.Data.GetFacts()
			podDetailsCopy.ObjectMeta.Annotations = annotationsMap

			_, err := kubernetesOps.ClientSet.CoreV1().Pods(podDetailsCopy.ObjectMeta.Namespace).Update(podDetailsCopy)
			if err != nil {
				fmt.Println("unable to update pod ", podDetailsCopy.ObjectMeta.Name, "in namespace", podDetailsCopy.ObjectMeta.Namespace, err)
			} else {
				fmt.Println("UpdateFactsWorkers Thread: ", id, "Updated cat-facts to ", annotationsMap["cat-fact"])
			}
		} else {
			log.Println("Pod already has a cat-facts annotation")
		}
	}
}
