package main

import (
	"example.com/messaging-app/internal/config"
	"example.com/messaging-app/internal/server"
	"example.com/messaging-app/internal/service"
	"example.com/messaging-app/internal/util"
	"fmt"
	"time"
)

func randomlySendMessagesAround() {
	for i := 0; i < 10; i++ {
		err := service.SendMessage(fmt.Sprintf("Sending message from pod %s\n", util.GetPodIP()), []string{}, "null")
		if err != nil {
			fmt.Printf("failed to send message: %e\n", err)
		}

		time.Sleep(5000 * time.Millisecond)
	}
}

func main() {
	err := config.CreateKubernetesClient()
	if err != nil {
		panic(fmt.Sprintf("failed to load Kubernetes client: %e", err))
	}

	config.CreateHttpClient()

	go randomlySendMessagesAround()
	server.Host()
}
