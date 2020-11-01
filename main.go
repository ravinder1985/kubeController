package main

import (
	"net/http"
	"os"
)

// GetConf would override values in case set in the environment variable
func GetConf(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// In the future it could be comming from config files.
var (
	numberOfWorkers    int    = 3
	url                string = GetConf("CAT_FACTS_URL", "http://cat-fact.herokuapp.com/facts")
	kubeConfigLocation string = GetConf("KUBE_CONFIGS", "~/.kube/config")
)

// Create kubernetesOps object in global space
var kubernetesOps KubernetesOps

func init() {
	_, err := kubernetesOps.SystemInitialize()
	if err != nil {
		panic(err)
	}
	_, err = kubernetesOps.StartEventWatcher()
	if err != nil {
		panic(err)
	}
}

func main() {
	// Following line is to block the code to exit while its running. There are other ways of doing it too.
	http.ListenAndServe(":8080", nil)
}
