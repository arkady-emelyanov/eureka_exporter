package kube

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	flag "github.com/spf13/pflag"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	KubernetesConfigFlag = "config"
	k8sNsFile            = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
)

var (
	kubeconfig *string
	client     *kubernetes.Clientset
)

func init() {
	if home := homeDir(); home != "" {
		kubeconfig = flag.StringP(
			KubernetesConfigFlag,
			"c",
			filepath.Join(homeDir(), ".kube", "config"),
			"Kubernetes config file path (when running outside of cluster)",
		)
	} else {
		kubeconfig = flag.StringP(
			KubernetesConfigFlag,
			"c",
			"",
			"Kubernetes config file path (when running outside of cluster)",
		)
	}
}

func GetClient() *kubernetes.Clientset {
	if client == nil {
		config, err := getConfig()
		if err != nil {
			panic(err)
		}
		client, err = kubernetes.NewForConfig(config)
		if err != nil {
			panic(err)
		}
	}

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
