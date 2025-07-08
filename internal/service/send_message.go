package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sort"

	"example.com/messaging-app/internal/model"
	"example.com/messaging-app/internal/util"
    "example.com/messaging-app/internal/config"
)

func SendMessage(message string, trace []string, previous_sender_ip string) error {
	if previous_sender_ip == "null" {
		return sendNewMessage(message)
	}

	return sendMessageToNextPod(message, trace, previous_sender_ip)
}

func sendNewMessage(message string) error {
	message_request := model.MessageRequest{
		Message:     message,
		Trace:       []string{},
		SenderPodIP: util.GetPodIP(),
	}

	json_data, err := json.Marshal(message_request)
	if err != nil {
		return fmt.Errorf("error serializing to JSON: %w", err)
	}

	ips, err := util.GetServiceIps()
	if err != nil {
		return err
	}

	for _, ip := range ips {
        // This sends to a service, which is bound to port 80
		url := fmt.Sprintf("http://%s/message", ip)
		_, err = http.Post(url, "application/json", bytes.NewBuffer(json_data))
		if err != nil {
			return fmt.Errorf("error sending messaging over HTTP: %w", err)
		}

		fmt.Printf("sent new message to IP '%s'\n", ip)
	}

	return nil
}

func sendMessageToNextPod(message string, trace []string, previous_sender_ip string) error {
	// Prevent payloads from infinitely growing in size
	new_trace := shiftSliceForward(append(trace, util.GetPodName()), 10)

	message_request := model.MessageRequest{
		Message:     message,
		Trace:       new_trace,
		SenderPodIP: util.GetPodIP(),
	}

	json_data, err := json.Marshal(message_request)
	if err != nil {
		return fmt.Errorf("error serializing to JSON: %w", err)
	}

	next_ip, err := getNextPodIP(previous_sender_ip)
	if err != nil {
		return fmt.Errorf("failed to get next pod IP: %w", err)
	}

    // TODO: FIX THE FUCKING TIMEOUT
    // THis sends directly to a pod, which is listening to port 8080
	url := fmt.Sprintf("http://%s:%s/message", next_ip, config.GetPort())
	_, err = http.Post(url, "application/json", bytes.NewBuffer(json_data))
	if err != nil {
		return fmt.Errorf("error sending messaging over HTTP: %w", err)
	}

	fmt.Printf("sent message to next pod with IP '%+v'\n", next_ip)

	return nil
}

func getNextPodIP(previous_sender_ip string) (string, error) {
	ips, err := util.GetAllPodIPs()
	if err != nil {
		return "", fmt.Errorf("error when looking up msg-app pod IPs: %w", err)
	}

	ips = util.Filter(ips, util.GetPodIP())
	if len(ips) == 0 {
		return "", fmt.Errorf("next pod IP was not found")
	}

    fmt.Printf("POD IPS: %s\n", ips)

	// Ensure that IPs are sorted
	sort.Slice(ips, func(i, j int) bool {
		ip1 := net.ParseIP(ips[i])
		ip2 := net.ParseIP(ips[j])
		return util.CompareIPs(ip1, ip2) < 0
	})

	last_ip_index, err := util.IndexOf(ips, previous_sender_ip)
	if err != nil {
		return "", fmt.Errorf("failed to find the next IP address to send message to")
	}

    next_pod_ip_index := (last_ip_index+1)%len(ips)
    if next_pod_ip_index > len(ips)-1 || next_pod_ip_index < 0 {
        return "", fmt.Errorf("next pod IP index %d is out of range for IPs slice %s", next_pod_ip_index, ips)
    }

	return ips[next_pod_ip_index], nil
}

func shiftSliceForward[T comparable](slice []T, max_len int) []T {
	if len(slice) > max_len {
		slice = slice[len(slice)-max_len:]
	}

	return slice
}
