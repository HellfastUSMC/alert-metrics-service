package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/HellfastUSMC/alert-metrics-service/internal/agent-storage"
	"github.com/HellfastUSMC/alert-metrics-service/internal/config"
	"github.com/HellfastUSMC/alert-metrics-service/internal/controllers"
	"github.com/HellfastUSMC/alert-metrics-service/internal/logger"
	"github.com/rs/zerolog"
)

func checkErr(errorsToRetry []any, err error) bool {
	for _, cErr := range errorsToRetry {
		if errors.As(err, &cErr) {
			return true
		}
	}
	return false
}

func retryFunc(logger logger.CLogger, intervals []int, errorsToRetry []any, function func() error) error {
	err := function()
	if err != nil && checkErr(errorsToRetry, err) {
		for i, interval := range intervals {
			logger.Info().Msg(fmt.Sprintf("Error %v. Attempt #%d with interval %ds", err, i, interval))
			time.Sleep(time.Second * time.Duration(interval))
			errOK := checkErr(errorsToRetry, err)
			if errOK {
				err = function()
				if err == nil {
					return nil
				}
			}
		}
	}
	return err
}

func main() {
	var (
		urlErr     url.Error
		intervals  = []int{1, 3, 5}
		errorsList = []any{&urlErr}
	)
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
				f := func() error {
					err := controller.SendMetrics("http://" + controller.Config.ServerAddress)
					if err != nil {
						return err
					}
					return nil
				}
				err = retryFunc(&log, intervals, errorsList, f)
				log.Error().Err(err).Msg(fmt.Sprintf("Error after %d retries", len(intervals)+1))
			}
		}
	}()
	select {}
}
