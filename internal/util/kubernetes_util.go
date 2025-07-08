package util

import (
	"context"
	"fmt"
	"net"
	"os"
	"example.com/messaging-app/internal/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetAllPodIPs() ([]string, error) {
	return net.LookupHost(fmt.Sprintf("%s-software-service-headless", getAppName()))
}

func GetPodIP() string {
	return os.Getenv("POD_IP")
}

func GetServiceIps() ([]string, error) {
	// get pods in all the namespaces by omitting namespace
	// Or specify namespace to get pods in particular namespace
	clientset := config.GetKubernetesClientSet()
	service, err := clientset.CoreV1().Services("default").Get(context.TODO(), fmt.Sprintf("%s-software-service", getAppName()), metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return service.Spec.ClusterIPs, nil
}

func GetPodName() string {
	pod_name := os.Getenv("HOSTNAME")
	if pod_name == "" {
		panic("HOSTNAME is not set")
	}

	return pod_name
}

func getAppName() string {
	app_name := os.Getenv("APP_NAME")
	if app_name == "" {
		panic("APP_NAME is not set")
	}

	return app_name
}
