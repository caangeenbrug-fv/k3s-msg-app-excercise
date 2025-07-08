package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
    "example.com/messaging-app/internal/model"
    "example.com/messaging-app/internal/service"
)

func CreateNewMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var message_request model.CustomMessageRequest
		decoder := json.NewDecoder(r.Body)

		err := decoder.Decode(&message_request)
		if err != nil {
			http.Error(w, fmt.Sprintf("invalid JSON: %e", err), http.StatusBadRequest)
			return
		}

		err = service.SendMessage(message_request.Message, []string{}, "null")
		if err != nil {
			fmt.Printf("failed to send message: %e\n", err)
		}
	} else {
		http.Error(w, "invalid request method", http.StatusMethodNotAllowed)
	}
}
