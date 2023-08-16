package main

import (
	"net/http"
	"os"

	"github.com/HellfastUSMC/alert-metrics-service/internal/config"
	"github.com/HellfastUSMC/alert-metrics-service/internal/controllers"
	"github.com/HellfastUSMC/alert-metrics-service/internal/server-storage"
	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

func main() {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	log.SetOutput(os.Stdout)
	conf, err := config.NewConfig()
	if err != nil {
		log.Warning(err)
	}
	if conf.ServerAddress == "" {
		conf.ParseServerAddr()
	}
	controller := controllers.NewServerController(log, conf, serverstorage.NewMemStorage())
	router := chi.NewRouter()
	router.Mount("/", controller.Route())
	controller.Infof("Starting server at " + controller.Config.ServerAddress)
	err = http.ListenAndServe(controller.Config.ServerAddress, router)
	if err != nil {
		controller.Errorf("there's an error in server starting - %e", err)
	}
}
