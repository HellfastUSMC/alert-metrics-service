package main

import (
	"fmt"
	"github.com/HellfastUSMC/alert-metrics-service/internal/flags"
	"github.com/HellfastUSMC/alert-metrics-service/internal/handlers"
	"github.com/HellfastUSMC/alert-metrics-service/internal/storage"
	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi/v5"
	"net/http"
	"os"
)

func main() {
	conf := storage.SysConfig{}
	parseErr := env.Parse(&conf)
	fmt.Println(os.Getenv("ADDRESS"))
	if parseErr != nil {
		fmt.Println(parseErr)
	}
	flags.ParseServerAddr()
	if conf.ServerAddress == "" {
		conf.ServerAddress = flags.ServerAddr
	}
	fmt.Printf("Server address: %s\n", conf.ServerAddress)
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
	err := http.ListenAndServe(conf.ServerAddress, router)
	if err != nil {
		fmt.Printf("there's an error in server starting - %e", err)
	}
	fmt.Println("Server started at " + conf.ServerAddress)
}
