package main

import (
	"fmt"
	"github.com/HellfastUSMC/alert-metrics-service/internal/handlers"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func main() {

	router := chi.NewRouter()
	router.Route("/", func(router chi.Router) {
		router.Get("/", handlers.GetAllStats)
		router.Route("/value", func(router chi.Router) {
			router.Get("/{metricType}/{metricName}", handlers.ReturnMetric)
		})
		router.Route("/update", func(router chi.Router) {
			router.Post("/{metricType}/{metricName}/{metricValue}", handlers.GetMetrics)
		})
	})
	err := http.ListenAndServe("localhost:8080", router)
	if err != nil {
		fmt.Printf("there's an error in server starting - %e", err)
	}
}
