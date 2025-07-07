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
	"sort"
	"time"
)

var config *rest.Config
var clientset *kubernetes.Clientset

type MessageRequest struct {
	Message  string   `json:"message"`
	Trace    []string `json:"trace"`
	SenderIp string   `json:"sender_ip"`
}

type CustomMessageRequest struct {
	Message string `json:"message"`
}

func messageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var message_request MessageRequest
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&message_request)

		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid JSON: %e", err), http.StatusBadRequest)
			return
		}

		response_message := fmt.Sprintf("Received JSON from pod '%+v': %+v\n", getCurrentPodName(), message_request)
		fmt.Fprintf(w, "%s", response_message)
		fmt.Println(response_message)

		fmt.Println("Sending message to next pod")

		time.Sleep(1000 * time.Millisecond)

		sendMessage(message_request.Message, message_request.Trace, message_request.SenderIp)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

func createMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var message_request CustomMessageRequest
		decoder := json.NewDecoder(r.Body)

		err := decoder.Decode(&message_request)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid JSON: %e", err), http.StatusBadRequest)
			return
		}

		sendMessage(message_request.Message, []string{}, "null")
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

func createHealthCheckMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		fmt.Fprintf(w, "Pod is running")
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

func getLabelSelector() string {
	label_selector := os.Getenv("APP_NAME")
	if label_selector == "" {
		log.Fatal("APP_NAME is not set")
	}

	return label_selector
}

func hostServer() {
	http.HandleFunc("/message", messageHandler)
	http.HandleFunc("/custom-message", createMessageHandler)
	http.HandleFunc("/healthcheck", createHealthCheckMessageHandler)
	http.ListenAndServe(":8080", nil)
}

// func randomlySendMessagesAround(clientset *kubernetes.Clientset) {
// 	for {
// 		time.Sleep(5000)
// 		sendMessage(clientset)
// 	}
// }

func sendMessage(message string, trace []string, previous_sender_ip string) error {
	pod_ip := os.Getenv("POD_IP")

	new_trace := append(trace, getCurrentPodName())
	// Do not store more than 10 pod names in the trace and remove excess names from the start of the trace
	new_trace = new_trace[len(trace)-10:]

	message_request := MessageRequest{
		Message:  message,
		Trace:    new_trace,
		SenderIp: pod_ip,
	}

	json_data, err := json.Marshal(message_request)
	if err != nil {
		return fmt.Errorf("error serializing to JSON: %w", err)
	}

	// ips, err := retrieveAllPodIPs()
	ips, err := retrieveAllPodIPsWithK3sApi()
	if err != nil {
		return fmt.Errorf("error when looking up msg-app pod IPs: %w", err)
	}

	ips = filter(ips, pod_ip)

	if len(ips) == 0 {
		return nil
	}

	// Ensure that IPs are sorted
	sort.Slice(ips, func(i, j int) bool {
		ip1 := net.ParseIP(ips[i])
		ip2 := net.ParseIP(ips[j])
		return bytesCompare(ip1, ip2) < 0
	})

	// pod_ip := os.Getenv("POD_IP")
	// for _, ip := range ips {
	// 	if ip != pod_ip {
	// 		url := fmt.Sprintf("http://%s:8080/message", ip)
	// 		_, err = http.Post(url, "application/json", bytes.NewBuffer(json_data))
	// 		if err != nil {
	// 			fmt.Println("Error sending messaging over HTTP:", err)
	// 			return
	// 		}

	// 		fmt.Printf("Sent message over HTTP to pod with IP '%+v'\n", ip)
	// 	} else {
	// 		fmt.Printf("Skipping sender pod with IP '%+v'\n", pod_ip)
	// 	}
	// }

	var next_ip string
	if previous_sender_ip != "null" {
		last_ip_index, err := indexOf(ips, previous_sender_ip)
		if err != nil {
			return fmt.Errorf("failed to find the next IP address to send message to")
		}

		next_ip = ips[(last_ip_index+1)%len(ips)]
	} else {
		next_ip = ips[0]
	}

	url := fmt.Sprintf("http://%s:8080/message", next_ip)
	_, err = http.Post(url, "application/json", bytes.NewBuffer(json_data))
	if err != nil {
		return fmt.Errorf("error sending messaging over HTTP: %w", err)
	}

	fmt.Printf("Sent message over HTTP to pod with IP '%+v'\n", next_ip)

	return nil
}

// Compare two net.IPs as bytes
func bytesCompare(a, b net.IP) int {
	a = a.To16()
	b = b.To16()
	for i := 0; i < len(a); i++ {
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return 1
		}
	}
	return 0
}

func indexOf(collection []string, value string) (int, error) {
	for i, v := range collection {
		if v == value {
			return i, nil
		}
	}

	return -1, fmt.Errorf("failed to find index of value %v in collection %v", value, collection)
}

func filter(collection []string, value string) []string {
	var result = []string{}
	for _, v := range collection {
		if v != value {
			result = append(result, v)
		}
	}

	return result
}

func retrieveAllPodIPs() ([]string, error) {
	return net.LookupHost("msg-app-headless")
}

func retrieveAllPodIPsWithK3sApi() ([]string, error) {
	// get pods in all the namespaces by omitting namespace
	// Or specify namespace to get pods in particular namespace
	pods, err := clientset.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", getLabelSelector()),
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
	var err error
	// creates the in-cluster config
	config, err = rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	go hostServer()

	time.Sleep(8000 * time.Millisecond)

	// go randomlySendMessagesAround(clientset)
	err = sendMessage("Hello from pod "+getCurrentPodName(), []string{}, "null")
	if err != nil {
		fmt.Println("Something went wrong when attempting to send a message: ", err)
	}
}
