package main

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/HellfastUSMC/alert-metrics-service/internal/config"
	"github.com/HellfastUSMC/alert-metrics-service/internal/connectors"
	"github.com/HellfastUSMC/alert-metrics-service/internal/controllers"
	"github.com/HellfastUSMC/alert-metrics-service/internal/server-storage"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

func main() {
	log := zerolog.New(os.Stdout).Level(zerolog.TraceLevel).With().Timestamp().Logger()
	conf, err := config.GetServerConfigData()
	if err != nil {
		log.Error().Err(err).Msg("config create error")
	}
	dumper, err := connectors.GetDumper(&log, conf)
	if err != nil {
		log.Error().Err(err).Msg("dumper create error")
	}
	memStore := serverstorage.NewMemStorage(dumper, &log)
	controller := controllers.NewServerController(&log, conf, memStore)
	tickDump := time.NewTicker(time.Duration(conf.StoreInterval) * time.Second)
	if dumper != nil {
		go func() {
			defer runtime.Goexit()
			for {
				<-tickDump.C
				if err := memStore.WriteDump(); err != nil {
					log.Error().Err(err).Msg("dump write error")
				}
			}
		}()
		if conf.Recover {
			if err := memStore.ReadDump(); err != nil {
				log.Error().Err(err).Msg("dump read error")
			}
		}
	}
	router := chi.NewRouter()
	router.Mount("/", controller.Route())
	log.Info().Msg(fmt.Sprintf(
		"Starting server at %s with store interval %ds, dump path %s, DB path %s and recover state is %v",
		controller.Config.ServerAddress,
		controller.Config.StoreInterval,
		controller.Config.DumpPath,
		controller.Config.DBPath,
		controller.Config.Recover,
	))
	err = http.ListenAndServe(controller.Config.ServerAddress, controller.Route())
	if err != nil {
		log.Error().Err(err)
	}
	select {}
}
