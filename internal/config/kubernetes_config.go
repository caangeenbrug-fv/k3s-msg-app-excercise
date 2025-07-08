package config

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var clientset *kubernetes.Clientset

func CreateKubernetesClient() error {
	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}

	clientset, err = kubernetes.NewForConfig(config)
	return err
}

func GetKubernetesClientSet() *kubernetes.Clientset {
	return clientset
}
