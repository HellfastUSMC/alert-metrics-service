package main

import (
	"fmt"
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
	log := zerolog.New(os.Stdout).Level(zerolog.TraceLevel).With().Timestamp().Logger()
	conf, err := config.GetServerConfigData()
	if err != nil {
		log.Error().Err(err)
	}
	controller := controllers.NewServerController(&log, conf, serverstorage.NewMemStorage())
	if err := controller.ReadDump(); err != nil {
		controller.Error().Err(err)
	}
	router := chi.NewRouter()
	router.Mount("/", controller.Route())
	controller.Info().Msg(fmt.Sprintf(
		"Starting server at %s with store interval %ds, dump path %s and recover state is %v",
		controller.Config.ServerAddress,
		controller.Config.StoreInterval,
		controller.Config.DumpPath,
		controller.Config.Recover,
	))
	tickDump := time.NewTicker(time.Duration(controller.Config.StoreInterval) * time.Second)
	go func() {
		for {
			<-tickDump.C
			if err := controller.WriteDump(); err != nil {
				controller.Error().Err(err)
			}
		}
	}()
	//go func() {
	err = http.ListenAndServe(controller.Config.ServerAddress, router)
	if err != nil {
		controller.Error().Err(err)
	}
	//}()
}
