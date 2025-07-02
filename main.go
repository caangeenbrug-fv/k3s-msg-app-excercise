package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

type MessageRequest struct {
	Message string `json:"message"`
	Sender  string `json:"sender"`
}

func messageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var message_request MessageRequest
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&message_request)

		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		response_message := fmt.Sprintf("Received JSON from pod '%+v': %+v\n", getCurrentPodName(), message_request)
		fmt.Fprintf(w, "%s", response_message)
		fmt.Println(response_message)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

func getCurrentPodName() string {
	pod_name := os.Getenv("HOSTNAME")
	if pod_name == "" {
		log.Fatal("HOSTNAME is not set")
	}

	return pod_name
}

func hostServer() {
	http.HandleFunc("/message", messageHandler)
	http.ListenAndServe(":8080", nil)
}

func randomlySendMessagesAround(clientset *kubernetes.Clientset) {
	for {
		time.Sleep(500)
		sendMessage(clientset)
	}
}

func sendMessage(clientset *kubernetes.Clientset) {
	message_request := MessageRequest{
		Message: "Hello",
		Sender:  getCurrentPodName(),
	}

	json_data, err := json.Marshal(message_request)
	if err != nil {
		fmt.Println("Error serializing to JSON:", err)
		return
	}

	// ips, err := retrieveAllPodIPs()
	ips, err := retrieveAllPodIPsWithK3sApi(clientset)
	if err != nil {
		fmt.Println("Error when looking up msg-app pod IPs:", err)
		return
	}

	pod_ip := os.Getenv("POD_IP")
	for _, ip := range ips {
		if ip != pod_ip {
			url := fmt.Sprintf("http://%s:8080/message", ip)
			_, err = http.Post(url, "application/json", bytes.NewBuffer(json_data))
			if err != nil {
				fmt.Println("Error sending messaging over HTTP:", err)
				return
			}

			fmt.Printf("Sent message over HTTP to pod with IP '%+v'\n", pod_ip)
		} else {
			fmt.Printf("Skipping sender pod with IP '%+v'\n", pod_ip)
		}
	}
}

func retrieveAllPodIPs() ([]string, error) {
	return net.LookupHost("msg-app-headless")
}

func retrieveAllPodIPsWithK3sApi(clientset *kubernetes.Clientset) ([]string, error) {
	// get pods in all the namespaces by omitting namespace
	// Or specify namespace to get pods in particular namespace
	pods, err := clientset.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{
		LabelSelector: "app=msg-app",
        
	})
	if err != nil {
		return nil, err
	}

	var ips []string = make([]string, len(pods.Items))
	for i, pod := range pods.Items {
		ips[i] = pod.Status.PodIP
	}

	return ips, nil
}

func main() {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	go randomlySendMessagesAround(clientset)
	hostServer()
}
