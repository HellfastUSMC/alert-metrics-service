package main

import (
	"fmt"
	"github.com/HellfastUSMC/alert-metrics-service/internal/flags"
	"github.com/HellfastUSMC/alert-metrics-service/internal/handlers"
	"github.com/HellfastUSMC/alert-metrics-service/internal/storage"
	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func main() {
	conf := storage.SysConfig{}
	parseErr := env.Parse(&conf)
	if parseErr != nil {
		fmt.Println(parseErr)
	}
	if conf.ServerAddress == "" {
		flags.ParseServerAddr(&conf)
	}
	var Store = storage.MemStorage{Gauge: map[string]storage.Gauge{}, Counter: map[string]storage.Counter{}}
	router := chi.NewRouter()
	router.Route("/", func(router chi.Router) {
		router.Get("/", handlers.GetAllStats(&Store))
		router.Route("/value", func(router chi.Router) {
			router.Get("/{metricType}/{metricName}", handlers.ReturnMetric(&Store))
		})
		router.Route("/update", func(router chi.Router) {
			router.Post("/{metricType}/{metricName}/{metricValue}", handlers.GetMetrics(&Store))
		})
	})
	fmt.Println("Starting server at " + conf.ServerAddress)
	err := http.ListenAndServe(conf.ServerAddress, router)
	if err != nil {
		fmt.Printf("there's an error in server starting - %e", err)
	}
}
