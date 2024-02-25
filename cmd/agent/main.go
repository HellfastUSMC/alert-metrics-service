package main

import (
	"context"
	"fmt"
	"github.com/HellfastUSMC/alert-metrics-service/internal/utils"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
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
	if buildVersion != "" {
		fmt.Printf("Build version: %s\n", buildVersion)
	} else {
		fmt.Println("Build version: N/A")
	}
	if buildDate != "" {
		fmt.Printf("Build date: %s\n", buildDate)
	} else {
		fmt.Println("Build date: N/A")
	}
	if buildCommit != "" {
		fmt.Printf("Build commit: %s\n", buildCommit)
	} else {
		fmt.Println("Build commit: N/A")
	}
	tickPoll := time.NewTicker(time.Duration(controller.Config.PollInterval) * time.Second)
	tickReport := time.NewTicker(time.Duration(controller.Config.ReportInterval) * time.Second)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	jobsChan := make(chan int, conf.RateLimit)
	jobNum := 0
	var wg sync.WaitGroup
	sender := func(id int, jobs chan int) {
		defer wg.Done()
		for jNum := range jobs {
			controller.Logger.Info().Msg(fmt.Sprintf("Starting worker №%d with job number %d", id, jNum))
			if err1 := controller.SendMetrics(conf.KeyPath, "http://"+controller.Config.ServerAddress); err != nil {
				log.Error().Err(err1).Msg("Error when sending metrics to server")
				f := func() error {
					err2 := controller.SendMetrics(conf.KeyPath, "http://"+controller.Config.ServerAddress)
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
	sigChnl := make(chan os.Signal, 1)
	signal.Notify(sigChnl)
	go func() {
		for {
			s := <-sigChnl
			utils.ExitHandler(s)
		}
	}()
	<-ctx.Done()
	stop()
	log.Info().Msg("Agent about to stop working in 10 seconds...")
	ctxTimeOut, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	wg.Done()
	wg.Done()
	wg.Done()
	<-ctxTimeOut.Done()
	os.Exit(0)
	//wg.Wait()
}
