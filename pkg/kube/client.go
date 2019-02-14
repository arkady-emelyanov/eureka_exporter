package kube

import (
	"flag"
	"io/ioutil"
	"k8s.io/client-go/kubernetes"
	"log"
	"os"
	"path/filepath"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var kubeconfig *rest.Config

func init() {
	config, err := getConfig()
	if err != nil {
		panic(err)
	}
	kubeconfig = config
}

func GetClient() (*kubernetes.Clientset, error) {
	return kubernetes.NewForConfig(kubeconfig)
}

func InCluster() bool {
	_, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	return err == nil
}

func GetNamespace() string {
	if InCluster() == false {
		return ""
	}

	b, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		log.Printf("Error getting namespace from /var/run: %s", err)
		return ""
	}

	return string(b)
}

func getConfig() (*rest.Config, error) {
	// first, try to get in-cluster configuration
	clusterConfig, err := rest.InClusterConfig()
	if err == nil {
		return clusterConfig, nil
	}

	// try next one
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String(
			"kubeconfig",
			filepath.Join(home, ".kube", "config"),
			"(optional) absolute path to the kubeconfig file",
		)
	} else {
		kubeconfig = flag.String(
			"kubeconfig",
			"",
			"absolute path to the kubeconfig file",
		)
	}
	flag.Parse()

	// use the current context in kubeconfig
	localConfig, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return nil, err
	}
	return localConfig, nil
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
