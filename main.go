package main

import (
	"flag"
	"fmt"
	"path/filepath"
	"time"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	informerFactory := informers.NewSharedInformerFactory(clientset, 30*time.Second)
	podInformer := informerFactory.Core().V1().Pods()

	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    addPod,
		DeleteFunc: deletePod,
	})

	stopCh := make(chan struct{})
	defer close(stopCh)

	// Start the informer
	go informerFactory.Start(stopCh)

	// Wait until the cache is synced
	if !cache.WaitForCacheSync(stopCh, podInformer.Informer().HasSynced) {
		panic("Timed out waiting for caches to sync")
	}

	<-stopCh
}

func addPod(obj interface{}) {
	fmt.Println("Add Pod")
}

func deletePod(obj interface{}) {
	fmt.Println("Delete Pod")
}
