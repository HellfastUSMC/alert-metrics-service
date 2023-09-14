package main

import (
	"errors"
	"fmt"
	"github.com/HellfastUSMC/alert-metrics-service/internal/logger"
	"net"
	"os"
	"time"

	"github.com/HellfastUSMC/alert-metrics-service/internal/agent-storage"
	"github.com/HellfastUSMC/alert-metrics-service/internal/config"
	"github.com/HellfastUSMC/alert-metrics-service/internal/controllers"
	"github.com/rs/zerolog"
)

func checkErr(errorsToRetry []error, err error) bool {
	for _, errr := range errorsToRetry {
		if errors.As(err, &errr) {
			return true
		}
	}
	return false
}

func retryFunc(logger logger.CLogger, intervals []int, errorsToRetry []error, function func() error) error {
	err := function(url)
	if err != nil && checkErr(errorsToRetry, err) {
		for i, interval := range intervals {
			logger.Info().Msg(fmt.Sprintf("Error %v. Attempt #%d with interval %d", err, i, intervals))
			time.Sleep(time.Second * time.Duration(interval))
			errOK := checkErr(errorsToRetry, err)
			if errOK {
				err = function(url)
				if err == nil {
					return nil
				}
			}
		}
	}
	return err
}

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
				f := func() error {
					err := controller.SendMetrics("http://" + controller.Config.ServerAddress)
					if err != nil {
						return err
					}
					return nil
				}
				err = retryFunc(&log, []int{1, 3, 5}, []error{&netErr}, f)
			}
		}
	}()
	select {}
}
