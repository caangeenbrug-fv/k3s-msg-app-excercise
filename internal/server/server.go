package server

import (
	"fmt"
	"net/http"

	"example.com/messaging-app/internal/handler"
    "example.com/messaging-app/internal/config"
)

func Host() error {
	http.HandleFunc("/message", handler.CreateIncomingMessageHandler)
	http.HandleFunc("/create-message", handler.CreateNewMessageHandler)
	http.HandleFunc("/healthcheck", handler.CreateHealthCheckHandler)
	return http.ListenAndServe(fmt.Sprintf(":%s", config.GetPort()), nil)
}
