package main

import (
	"github.com/HellfastUSMC/alert-metrics-service/internal/config"
	"github.com/HellfastUSMC/alert-metrics-service/internal/controllers"
	"github.com/HellfastUSMC/alert-metrics-service/internal/server-storage"
	"github.com/rs/zerolog"
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
	//controller.StartDumping()
	tickDump := time.NewTicker(time.Duration(controller.Config.StoreInterval) * time.Second)
	go func() {
		for {
			<-tickDump.C
			if err := controller.WriteDump(); err != nil {
				controller.Error().Err(err)
			}
		}
	}()
	controller.StartServer()
	select {}
}
