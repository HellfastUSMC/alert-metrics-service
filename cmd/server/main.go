package main

import (
	"fmt"
	"github.com/HellfastUSMC/alert-metrics-service/internal/handlers"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/update/", handlers.GetMetrics)
	err := http.ListenAndServe("localhost:8088", mux)
	if err != nil {
		fmt.Println(fmt.Errorf("there's an error in server starting - %e", err))
	}
}
