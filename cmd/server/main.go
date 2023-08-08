package main

import (
	"fmt"
	"github.com/HellfastUSMC/alert-metrics-service/internal/controllers"
	"github.com/sirupsen/logrus"
	"net/http"

	"github.com/HellfastUSMC/alert-metrics-service/internal/flags"
	"github.com/HellfastUSMC/alert-metrics-service/internal/handlers"
	"github.com/HellfastUSMC/alert-metrics-service/internal/storage"

	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi/v5"
)

func main() {
	controllers.ServerController{
		Logger:  logrus.New(),
		Config:  flags.NewConfig(),
		Handler: storage.NewMemStorage(),
	}
	logger := logrus.New()
	conf := flags.NewConfig()
	store := storage.NewMemStorage()
	router := chi.NewRouter()
	parseErr := env.Parse(&conf)
	if parseErr != nil {
		fmt.Println(parseErr)
	}
	if conf.ServerAddress == "" {
		conf.ParseServerAddr()
	}

	router.Route("/", func(router chi.Router) {
		router.Get("/", handlers.GetAllStats(store))
		router.Route("/value", func(router chi.Router) {
			router.Get("/{metricType}/{metricName}", handlers.ReturnMetric(store))
		})
		router.Route("/update", func(router chi.Router) {
			router.Post("/{metricType}/{metricName}/{metricValue}", handlers.GetMetrics(store))
		})
	})
	fmt.Println("Starting server at " + conf.ServerAddress)
	err := http.ListenAndServe(conf.ServerAddress, router)
	if err != nil {
		fmt.Printf("there's an error in server starting - %e", err)
	}
}
