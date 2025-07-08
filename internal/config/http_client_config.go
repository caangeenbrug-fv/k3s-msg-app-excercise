package config

import (
	"net/http"
	"time"
)

var client http.Client

func CreateHttpClient() {
	client = http.Client{
		Timeout: 5 * time.Second,
	}
}

func GetHttpClient() http.Client {
	return client
}
