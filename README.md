# kubeController
This Controller updates the annotation in running pods in kubernetes 


# Required Environment varibales
    * export KUBE_CONFIGS="~/.kube/config"                       // Path to you config file
    * export CAT_FACTS_URL="http://cat-fact.herokuapp.com/facts" // you can override facts ur

# Build Controller
    * Linux - GOOS=linux GOARCH=amd64 go build
    * Mac - go build

# Run Controller
    * ./kubeController



# Cat Facts Controller -

## Prompt
Write a kubernetes controller that annotates every pod on a cluster with a random cat fact.

Considerations:
* When the controller starts up, all existing pods should be annotated with cat facts.
* While the controller is running, newly-created pods should be annotated with cat facts.
* The controller should resync every 10 minutes (just in case).
* Cat facts should be downloaded from a REST API rather than being hard-coded.

Suggestions:
* Kubernetes has [clients](https://kubernetes.io/docs/tasks/administer-cluster/access-cluster-api/#programmatic-access-to-the-api) for many major languages including Go and Python.
* [Docker Desktop](https://www.docker.com/products/docker-desktop) ships with a simple Kubernetes environment that you can use for testing.
* The [controller pattern](https://kubernetes.io/docs/concepts/architecture/controller/#controller-pattern) is fairly well documented online. For performance reasons you should use an informer or a watch rather than polling the kubernetes API server.
* There is a nice [cat facts API here](https://alexwohlbruck.github.io/cat-facts/docs/).


## Requirements
- Use whichever programming language you want
- Include a README
- Do not spend more than 2-5 hours on this project
- Reach out to us with any clarifying questions you might have


## Submission
Put your submission in a git repository and email us a link to it.

If that is not possible, alternatively zip up your submission, and email it back to us.


## Examples

Initially, there are no cat facts:

```
❯ kubectl get pods -A -o=custom-columns=NAMESPACE:.metadata.namespace,NAME:.metadata.name,STATUS:.status.phase,FACT:.metadata.annotations.cat-fact
NAMESPACE     NAME                                     STATUS    FACT
kube-system   coredns-66bff467f8-2x29n                 Running   <none>
kube-system   coredns-66bff467f8-7xnkh                 Running   <none>
kube-system   etcd-docker-desktop                      Running   <none>
kube-system   kube-apiserver-docker-desktop            Running   <none>
kube-system   kube-controller-manager-docker-desktop   Running   <none>
kube-system   kube-proxy-hnmmn                         Running   <none>
kube-system   kube-scheduler-docker-desktop            Running   <none>
kube-system   storage-provisioner                      Running   <none>
kube-system   vpnkit-controller                        Running   <none>
```

When the controller starts up, cat fact annotations are applied to all existing pods:
```
❯ go run cat_fact_controller.go &
[1] 74749

❯ kubectl get pods -A -o=custom-columns=NAMESPACE:.metadata.namespace,NAME:.metadata.name,STATUS:.status.phase,FACT:.metadata.annotations.cat-fact
NAMESPACE     NAME                                     STATUS    FACT
kube-system   coredns-66bff467f8-2x29n                 Running   Bobtail cats owe their shortened tails to a natural genetic mutation that has appeared in cats across time and in various regions of the world. The American Bobtail can be traced back to Yodi, a cat with the mutation that was found in Arizona in the 1960s. Yodi passed the genetic quirk on to his kittens, thus creating a new breed.
kube-system   coredns-66bff467f8-7xnkh                 Running   Cat are cute.
kube-system   etcd-docker-desktop                      Running   Some animal shelters require kittens be adopted in pairs to ensure they came from the same litter.
kube-system   kube-apiserver-docker-desktop            Running   On October 24, 1963, a tuxedo cat named Félicette entered outer space aboard a French Véronique AG1 rocket and made feline history. Félicette returned from the 15-minute trip in once piece and earned the praise of French scientists, who said she made "a valuable contribution to research.".
kube-system   kube-controller-manager-docker-desktop   Running   Cats are among only a few animals that walk by moving their two right legs one after another and then their two left legs, rather than moving diagonal limbs simultaneously. Giraffes and camels also have this quality.
kube-system   kube-proxy-hnmmn                         Running   Cats were mythic symbols of divinity in ancient Egypt.
kube-system   kube-scheduler-docker-desktop            Running   Mymymymy cat.
kube-system   storage-provisioner                      Running   The cat skull is unusual among mammals in having very large eye sockets and a powerful and specialized jaw.
kube-system   vpnkit-controller                        Running   Andy Warhol amassed quite the collection of feline friends during his lifetime, beginning with a Siamese cat named Hester that was given to him by actress Gloria Swanson. By breeding Hester with a cat named Sam, Warhol ended up with multiple litters of kittens, at one point housing 25 cats in his Upper East Side townhouse in NYC.
```

And while the controller is running, annotations are automatically added to newly-created pods as well:
```
❯ kubectl delete pod -n kube-system coredns-66bff467f8-2x29n
❯ kubectl delete pod -n kube-system coredns-66bff467f8-7xnkh

❯ kubectl get pods -A -o=custom-columns=NAMESPACE:.metadata.namespace,NAME:.metadata.name,STATUS:.status.phase,FACT:.metadata.annotations.cat-fact
NAMESPACE     NAME                                     STATUS    FACT
kube-system   coredns-66bff467f8-l9prq                 Pending   One legend claims that cats were created when a lion on Noah's Ark sneezed and two kittens came out.
kube-system   coredns-66bff467f8-v5zhs                 Pending   The Bombay cat breed was developed to resemble a miniature panther.
kube-system   etcd-docker-desktop                      Running   Some animal shelters require kittens be adopted in pairs to ensure they came from the same litter.
kube-system   kube-apiserver-docker-desktop            Running   On October 24, 1963, a tuxedo cat named Félicette entered outer space aboard a French Véronique AG1 rocket and made feline history. Félicette returned from the 15-minute trip in once piece and earned the praise of French scientists, who said she made "a valuable contribution to research.".
kube-system   kube-controller-manager-docker-desktop   Running   Cats are among only a few animals that walk by moving their two right legs one after another and then their two left legs, rather than moving diagonal limbs simultaneously. Giraffes and camels also have this quality.
kube-system   kube-proxy-hnmmn                         Running   Cats were mythic symbols of divinity in ancient Egypt.
kube-system   kube-scheduler-docker-desktop            Running   Mymymymy cat.
kube-system   storage-provisioner                      Running   The cat skull is unusual among mammals in having very large eye sockets and a powerful and specialized jaw.
kube-system   vpnkit-controller                        Running   Andy Warhol amassed quite the collection of feline friends during his lifetime, beginning with a Siamese cat named Hester that was given to him by actress Gloria Swanson. By breeding Hester with a cat named Sam, Warhol ended up with multiple litters of kittens, at one point housing 25 cats in his Upper East Side townhouse in NYC.
```
(note the two `coredns` pods are new but have cat facts assigned)
README.md
Displaying README.md.