package main

import (
	"net/http"
	"os"

	"github.com/HellfastUSMC/alert-metrics-service/internal/config"
	"github.com/HellfastUSMC/alert-metrics-service/internal/controllers"
	"github.com/HellfastUSMC/alert-metrics-service/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

func main() {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	log.SetOutput(os.Stdout)
	conf, err := config.NewConfig()
	controller := controllers.ServerController{
		Logger:   log,
		Config:   conf,
		MemStore: storage.NewMemStorage(),
		CRouter:  chi.NewRouter(),
	}
	if err != nil {
		controller.Logger.Warning(err)
	}
	if controller.Config.ServerAddress == "" {
		controller.Config.ParseServerAddr()
	}
	router := chi.NewRouter()
	router.Mount("/", controller.Router())
	controller.Logger.Infof("Starting server at " + controller.Config.ServerAddress)
	err = http.ListenAndServe(controller.Config.ServerAddress, router)
	if err != nil {
		controller.Logger.Errorf("there's an error in server starting - %e", err)
	}
}
