package main

import (
	"fmt"
	"net/http"
	"os"
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
	dumper := connectors.NewFileDump(conf.DumpPath, conf.Recover, &log)
	memStore := serverstorage.NewMemStorage(dumper, &log)
	controller := controllers.NewServerController(&log, conf, memStore)
	if conf.DBPath != "" {
		db, err := connectors.NewConnectionPGSQL(*conf)
		if err != nil {
			log.Error().Err(err)
		}
		controller.DB = db
		defer db.Close()
	}
	tickDump := time.NewTicker(time.Duration(conf.StoreInterval) * time.Second)
	go func() {
		for {
			<-tickDump.C
			if err := memStore.WriteDump(); err != nil {
				log.Error().Err(err)
			}
		}
	}()
	if err := memStore.ReadDump(); err != nil {
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
