package main

import (
	"fmt"
	"github.com/HellfastUSMC/alert-metrics-service/internal/utils"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/HellfastUSMC/alert-metrics-service/internal/config"
	"github.com/HellfastUSMC/alert-metrics-service/internal/connectors"
	"github.com/HellfastUSMC/alert-metrics-service/internal/controllers"
	"github.com/HellfastUSMC/alert-metrics-service/internal/server-storage"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
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
				if err1 := memStore.WriteDump(); err1 != nil {
					log.Error().Err(err1).Msg("dump write error")
				}
			}
		}()
		if conf.Recover {
			if err2 := memStore.ReadDump(); err2 != nil {
				log.Error().Err(err2).Msg("dump read error")
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
	switch buildVersion {
	case "":
		fmt.Println("Build version: N/A")
	default:
		fmt.Printf("Build version: %s\n", buildVersion)
	}
	switch buildDate {
	case "":
		fmt.Println("Build date: N/A")
	default:
		fmt.Printf("Build date: %s\n", buildDate)
	}
	switch buildCommit {
	case "":
		fmt.Println("Build commit: N/A")
	default:
		fmt.Printf("Build commit: %s\n", buildCommit)
	}
	sigChnl := make(chan os.Signal, 1)
	signal.Notify(sigChnl)
	go func() {
		for {
			s := <-sigChnl
			utils.ExitHandler(s)
		}
	}()
	err = http.ListenAndServe(controller.Config.ServerAddress, controller.Route())
	if err != nil {
		log.Error().Err(err)
	}
	select {}
}
