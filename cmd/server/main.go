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
		log.Error().Err(err)
	}
	var dumper serverstorage.Dumper
	if conf.DBPath != "" {
		dumper, err = connectors.NewConnectionPGSQL(conf.DBPath, &log)
		if err != nil {
			log.Error().Err(err).Msg("")
		}
		log.Info().Msg("Using DB dumper")
	} else if conf.DumpPath != "" {
		dumper = connectors.NewFileDump(conf.DumpPath, conf.Recover, &log)
		log.Info().Msg("Using file dumper")
	} else {
		log.Info().Msg("Using memory dumper")
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
					log.Error().Err(err).Msg("")
				}
			}
		}()
		if conf.Recover {
			if err := memStore.ReadDump(); err != nil {
				log.Error().Err(err).Msg("")
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
