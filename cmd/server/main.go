package main

import (
	"fmt"
	"github.com/HellfastUSMC/alert-metrics-service/internal/handlers"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/update/", handlers.GetMetrics)
	err := http.ListenAndServe("localhost:8080", mux)
	if err != nil {
		fmt.Printf("there's an error in server starting - %e", err)
	}
}
