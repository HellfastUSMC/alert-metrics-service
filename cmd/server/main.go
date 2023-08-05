package main

import (
	"flag"
	"fmt"
	"github.com/HellfastUSMC/alert-metrics-service/internal/flags"
	"github.com/HellfastUSMC/alert-metrics-service/internal/handlers"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func init() {
	flags.ParseServerAddr()
}

func main() {
	flag.Parse()
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
	err := http.ListenAndServe(flags.ServerAddr, router)
	if err != nil {
		fmt.Printf("there's an error in server starting - %e", err)
	}
	fmt.Println("Server started at " + flags.ServerAddr)
}
