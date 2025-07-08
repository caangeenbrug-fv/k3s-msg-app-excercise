package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"example.com/messaging-app/internal/model"
	"example.com/messaging-app/internal/service"
	"example.com/messaging-app/internal/util"
)

func CreateIncomingMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var message_request model.MessageRequest
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&message_request)

		if err != nil {
			http.Error(w, fmt.Sprintf("invalid JSON: %e", err), http.StatusBadRequest)
			return
		}

		fmt.Printf("received JSON from pod '%+v': %+v\n", util.GetPodName(), message_request)

		time.Sleep(1000 * time.Millisecond)

		err = service.SendMessage(message_request.Message, message_request.Trace, message_request.SenderPodIP)
		if err == nil {
			fmt.Printf("failed to send message: %e\n", err)
		}
	} else {
		http.Error(w, "invalid request method", http.StatusMethodNotAllowed)
	}
}
