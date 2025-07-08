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
		log.Printf("%s\n", response_message)
		log.Println(response_message)

		log.Println("Sending message to next pod")

		time.Sleep(1000 * time.Millisecond)

		err = sendMessage(message_request.Message, message_request.Trace, message_request.SenderIp)
		if err == nil {
			log.Printf("failed to send message: %e", err)
		}
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

		err = sendMessage(message_request.Message, []string{}, "null")
		if err != nil {
			log.Printf("failed to send message: %e", err)
		}
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

func createHealthCheckMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		log.Println("Pod is running")
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

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Failed to host HTTP server: %e\n", err)
	}

	log.Println("Hosting server...")
}

func randomlySendMessagesAround() {
	for i := 0; i < 10; i++ {
		err := sendMessage(fmt.Sprintf("Sending message from pod %s\n", getCurrentPodName()), []string {}, "null")
		if err != nil {
			log.Printf("failed to send message: %e\n", err)
		}

		time.Sleep(5000 * time.Millisecond)
	}
}

func sendMessage(message string, trace []string, previous_sender_ip string) error {
	pod_ip := os.Getenv("POD_IP")

	new_trace := append(trace, getCurrentPodName())
	max_trace_len := 10

	if (len(new_trace) > max_trace_len) {
		// Do not store more than 10 pod names in the trace and remove excess names from the start of the trace
		new_trace = new_trace[len(trace)-max_trace_len:]
	}

	message_request := MessageRequest{
		Message:  message,
		Trace:    new_trace,
		SenderIp: pod_ip,
	}

	json_data, err := json.Marshal(message_request)
	if err != nil {
		return fmt.Errorf("error serializing to JSON: %w", err)
	}

	var ips []string
	// Send to service if previous sender is not known
	if previous_sender_ip == "null" {
		ips, err = getServiceIps()
	} else {
		ips, err = getAllPodIPs()
		ips = filter(ips, pod_ip)
	}
	
	if err != nil {
		return fmt.Errorf("error when looking up msg-app pod IPs: %w", err)
	}

	if len(ips) == 0 {
		return fmt.Errorf("loop is broken for some reason")
	}

	// Ensure that IPs are sorted
	sort.Slice(ips, func(i, j int) bool {
		ip1 := net.ParseIP(ips[i])
		ip2 := net.ParseIP(ips[j])
		return bytesCompare(ip1, ip2) < 0
	})

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

	url := fmt.Sprintf("http://%s/message", next_ip)
	_, err = http.Post(url, "application/json", bytes.NewBuffer(json_data))
	if err != nil {
		return fmt.Errorf("error sending messaging over HTTP: %w", err)
	}

	log.Printf("Sent message over HTTP to pod with IP '%+v'\n", next_ip)

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

func getAllPodIPs() ([]string, error) {
	return net.LookupHost("msg-app-software-service-headless")
}

func getServiceIps() ([]string, error) {
	// get pods in all the namespaces by omitting namespace
	// Or specify namespace to get pods in particular namespace
	service, err := clientset.CoreV1().Services("default").Get(context.TODO(), fmt.Sprintf("%s-software-service", getLabelSelector()), metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return service.Spec.ClusterIPs, nil
}

func main() {
	log.Println("Started main thread")

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

	log.Println("Created Kubernetes client set")

	go randomlySendMessagesAround()
	hostServer()

	log.Println("Exiting main thread")
}
