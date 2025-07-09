package handler

import (
	"fmt"
	"net/http"
)

func CreateHealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
        fmt.Fprintln(w, "Pod is running")
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}
