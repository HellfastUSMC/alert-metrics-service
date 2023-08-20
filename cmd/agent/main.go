package main

import (
	"github.com/rs/zerolog"
	"os"
	"strconv"
	"time"

	"github.com/HellfastUSMC/alert-metrics-service/internal/agent-storage"
	"github.com/HellfastUSMC/alert-metrics-service/internal/config"
	"github.com/HellfastUSMC/alert-metrics-service/internal/controllers"
)

func main() {
	log := zerolog.New(os.Stdout).Level(zerolog.InfoLevel)
	conf, err := config.NewConfig()
	if err != nil {
		log.Error().Err(err)
	}
	if conf.ServerAddress == "" || conf.PollInterval == 0 || conf.ReportInterval == 0 {
		if err := conf.ParseAgentFlags(); err != nil {
			log.Error().Err(err)
		}
	}
	controller := controllers.NewAgentController(&log, conf, agentstorage.NewMetricsStorage())
	controller.Info().Msg(
		"Starting agent with remote server addr: %s, poll interval: %d, report interval: %d\n" +
			controller.Config.ServerAddress +
			strconv.FormatInt(controller.Config.PollInterval, 10) +
			strconv.FormatInt(controller.Config.ReportInterval, 10))
	tickPoll := time.NewTicker(time.Duration(controller.Config.PollInterval) * time.Second)
	tickReport := time.NewTicker(time.Duration(controller.Config.ReportInterval) * time.Second)
	go func() {
		for {
			<-tickPoll.C
			controller.RenewMetrics()
		}
	}()
	go func() {
		for {
			<-tickReport.C
			if err := controller.SendMetrics("http://" + controller.Config.ServerAddress); err != nil {
				controller.Error().Err(err)
			}
		}
	}()
	select {}
}
