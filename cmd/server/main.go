package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/HellfastUSMC/alert-metrics-service/internal/config"
	"github.com/HellfastUSMC/alert-metrics-service/internal/controllers"
	"github.com/HellfastUSMC/alert-metrics-service/internal/server-storage"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

func main() {
	log := zerolog.New(os.Stdout).Level(zerolog.TraceLevel)
	conf, err := config.GetServerConfigData()
	if err != nil {
		log.Error().Err(err)
	}
	controller := controllers.NewServerController(&log, conf, serverstorage.NewMemStorage())
	router := chi.NewRouter()
	router.Mount("/", controller.Route())
	controller.Info().Msg("Starting server at " + controller.Config.ServerAddress)
	err = http.ListenAndServe(controller.Config.ServerAddress, router)
	if err != nil {
		fmt.Println(err)
		controller.Error().Err(err)
	}
}
