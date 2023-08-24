package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/HellfastUSMC/alert-metrics-service/internal/config"
	"github.com/HellfastUSMC/alert-metrics-service/internal/controllers"
	"github.com/HellfastUSMC/alert-metrics-service/internal/server-storage"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

func main() {
	log := zerolog.New(os.Stdout).Level(zerolog.TraceLevel).With().Timestamp().Logger()
	conf, err := config.GetServerConfigData()
	if err != nil {
		log.Error().Err(err)
	}
	memStore := serverstorage.NewMemStorage(serverstorage.NewDump())
	controller := controllers.NewServerController(&log, conf, memStore)
	tickDump := time.NewTicker(time.Duration(controller.Config.StoreInterval) * time.Second)
	go func() {
		for {
			<-tickDump.C
			if err := controller.MemStore(
				controller.Config.DumpPath,
				controller.Logger,
				memStore,
			); err != nil {
				log.Error().Err(err)
			}
		}
	}()
	if err := controller.MemStore.ReadDump(
		controller.Config.DumpPath,
		controller.Config.Recover,
		controller.Logger,
		memStore,
	); err != nil {
		log.Error().Err(err)
	}
	router := chi.NewRouter()
	router.Mount("/", controller.Route())
	log.Info().Msg(fmt.Sprintf(
		"Starting server at %s with store interval %ds, dump path %s and recover state is %v",
		controller.Config.ServerAddress,
		controller.Config.StoreInterval,
		controller.Config.DumpPath,
		controller.Config.Recover,
	))
	err = http.ListenAndServe(controller.Config.ServerAddress, controller.Route())
	if err != nil {
		log.Error().Err(err)
	}
	select {}
}
