package main

import (
    "encoding/json"
    "fmt"
    "os"
    "log"
    "time"
    "bytes"
    "net"
    "net/http"
)

type MessageRequest struct {
    Message string `json:"message"`
    Sender  string    `json:"sender"`
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

func randomlySendMessagesAround() {
    for {
        time.Sleep(500)
        sendMessage()
    }
}

func sendMessage() {
    message_request := MessageRequest {
        Message: "Hello",
        Sender: getCurrentPodName(),
    }
    
    json_data, err := json.Marshal(message_request)
    if err != nil {
        fmt.Println("Error serializing to JSON:", err)
        return
    }

    ips, err := net.LookupHost("msg-app-headless")
    if err != nil {
        fmt.Println("Error when looking up msg-app pod IPs")
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

func main() {
    go randomlySendMessagesAround()
    hostServer()
}
