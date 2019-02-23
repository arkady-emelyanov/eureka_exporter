package kube

import (
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	k8sNsFile = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
)

var (
	client *kubernetes.Clientset
)

func init() {
	config, err := getConfig()
	if err != nil {
		panic(err)
	}
	client, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
}

func GetClient() *kubernetes.Clientset {
	return client
}

func InCluster() bool {
	_, err := os.Stat(k8sNsFile)
	return err == nil
}

func GetNamespace() string {
	if InCluster() == false {
		return ""
	}

	b, err := ioutil.ReadFile(k8sNsFile)
	if err != nil {
		log.Error().
			Str("file", k8sNsFile).
			Err(err).
			Msg("Error getting namespace")

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
