package main

import (
	"fmt"
	"github.com/HellfastUSMC/alert-metrics-service/internal/config"
	"github.com/HellfastUSMC/alert-metrics-service/internal/controllers"
	"github.com/HellfastUSMC/alert-metrics-service/internal/server-storage"
	"github.com/rs/zerolog"
	"os"
)

func main() {
	log := zerolog.New(os.Stdout).Level(zerolog.TraceLevel).With().Timestamp().Logger()
	conf, err := config.GetServerConfigData()
	if err != nil {
		log.Error().Err(err)
	}
	controller := controllers.NewServerController(&log, conf, serverstorage.NewMemStorage())
	if err := controller.ReadDump(); err != nil {
		fmt.Println(controller.MemStore)
		controller.Error().Err(err)
	}
	//time.Sleep(time.Duration(controller.Config.StoreInterval) * time.Second)

	controller.StartDumping()
	controller.StartServer()
	select {}
}
