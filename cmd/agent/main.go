package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/HellfastUSMC/alert-metrics-service/internal/agent-storage"
	"github.com/HellfastUSMC/alert-metrics-service/internal/config"
	"github.com/HellfastUSMC/alert-metrics-service/internal/controllers"
	"github.com/rs/zerolog"
)

func main() {
	log := zerolog.New(os.Stdout).Level(zerolog.InfoLevel)
	conf, err := config.GetAgentConfigData()
	if err != nil {
		log.Error().Err(err)
	}
	memStore := agentstorage.NewMetricsStorage()
	controller := controllers.NewAgentController(&log, conf, memStore)
	log.Info().Msg(
		fmt.Sprintf("Starting agent with remote server addr: %s, poll interval: %d, report interval: %d",
			controller.Config.ServerAddress,
			controller.Config.PollInterval,
			controller.Config.ReportInterval))
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
				log.Error().Err(err).Msg("Error when sending metrics to server")
				var netErr net.Error
				//fmt.Println(errors.As(err, &context.DeadlineExceeded))
				if errors.As(err, &netErr) {
					//if errors.As(err, &context.DeadlineExceeded) {
					time.Sleep(time.Second * 1)
					for n := 0; n < 3; n++ {
						if err := controller.SendMetrics("http://" + controller.Config.ServerAddress); err != nil {
							log.Error().Err(err).Msg(fmt.Sprintf("Error when sending metrics to server, retry %d", n+1))
						} else {
							log.Info().Msg("Metrics batch sent to server")
						}
						if n != 2 {
							time.Sleep(time.Second * 2)
						}
					}
					log.Error().Msg("Tried 4 times to send request to server")
				}
			}
		}
	}()
	select {}
}
