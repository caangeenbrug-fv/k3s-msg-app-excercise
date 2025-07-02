package main

import (
    "encoding/json"
    "fmt"
    "os"
    "log"
    "time"
    "bytes"
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

        fmt.Fprintf(w, "Received JSON from pod '%+v': %+v\n", getCurrentPodName(), message_request)
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

    http.Post("messaging-app/message", "application/json", bytes.NewBuffer(json_data))
}

func main() {
    go randomlySendMessagesAround()
    hostServer()
}
