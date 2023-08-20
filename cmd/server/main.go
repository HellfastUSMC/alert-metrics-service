package main

import (
	"github.com/HellfastUSMC/alert-metrics-service/internal/config"
	"github.com/HellfastUSMC/alert-metrics-service/internal/controllers"
	"github.com/HellfastUSMC/alert-metrics-service/internal/server-storage"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"net/http"
	"os"
	"time"
)

func main() {
	log := zerolog.New(os.Stdout).Level(zerolog.TraceLevel)
	conf, err := config.NewConfig()
	if err != nil {
		log.Warn().Err(err)
	}
	if conf.ServerAddress == "" {
		if err := conf.ParseServerFlags(); err != nil {
			log.Warn().Err(err)
		}
	}
	controller := controllers.NewServerController(&log, conf, serverstorage.NewMemStorage())
	if err := controller.ReadDump(); err != nil {
		controller.Error().Err(err)
	}
	router := chi.NewRouter()
	router.Mount("/", controller.Route())
	controller.Info().Msg("Starting server at " + controller.Config.ServerAddress)
	go func() {
		err = http.ListenAndServe(controller.Config.ServerAddress, router)
		if err != nil {
			controller.Error().Err(err)
		}
	}()
	tickDump := time.NewTicker(time.Duration(controller.Config.StoreInterval) * time.Second)
	go func() {
		for {
			<-tickDump.C
			if err := controller.WriteDump(); err != nil {
				controller.Error().Err(err)
			}
		}
	}()
	select {}
}
