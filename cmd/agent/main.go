package main

import (
	"fmt"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/HellfastUSMC/alert-metrics-service/internal/agent-storage"
	"github.com/HellfastUSMC/alert-metrics-service/internal/config"
	"github.com/HellfastUSMC/alert-metrics-service/internal/controllers"
	"github.com/rs/zerolog"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

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

	jobsChan := make(chan int, conf.RateLimit)
	jobNum := 0
	var wg sync.WaitGroup
	sender := func(id int, jobs chan int) {
		defer wg.Done()
		for jNum := range jobs {
			controller.Logger.Info().Msg(fmt.Sprintf("Starting worker №%d with job number %d", id, jNum))
			if err1 := controller.SendMetrics(conf.Key, "http://"+controller.Config.ServerAddress); err != nil {
				log.Error().Err(err1).Msg("Error when sending metrics to server")
				f := func() error {
					err2 := controller.SendMetrics(conf.Key, "http://"+controller.Config.ServerAddress)
					if err2 != nil {
						return err
					}
					return nil
				}
				err = agentstorage.RetryFunc(&log, intervals, errorsList, f)
				log.Error().Err(err).Msg(fmt.Sprintf("Error after %d retries", len(intervals)+1))
			}
			jobNum += 1
		}
	}
	for i := 0; i < int(conf.RateLimit); i++ {
		wg.Add(1)
		go sender(i, jobsChan)
	}

	wg.Add(1)
	go func() {
		for {
			<-tickPoll.C
			controller.RenewMetrics()
		}
	}()
	wg.Add(1)
	go func() {
		for {
			<-tickPoll.C
			controller.RenewMemCPUMetrics()
		}
	}()
	wg.Add(1)
	go func() {
		for {
			<-tickReport.C
			jobsChan <- jobNum
		}
	}()
	wg.Wait()
}
