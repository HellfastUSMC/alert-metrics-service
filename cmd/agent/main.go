package main

import (
	"github.com/HellfastUSMC/alert-metrics-service/internal/config"
	"os"
	"time"

	"github.com/HellfastUSMC/alert-metrics-service/internal/controllers"
	"github.com/HellfastUSMC/alert-metrics-service/internal/storage"

	"github.com/sirupsen/logrus"
)

func main() {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	log.SetOutput(os.Stdout)
	conf, err := config.NewConfig()
	controller := controllers.AgentController{
		Config:  conf,
		Logger:  log,
		Storage: storage.NewMetricsStorage(),
	}
	if err != nil {
		controller.Logger.Warning(err)
	}
	if controller.Config.ServerAddress == "" || controller.Config.PollInterval == 0 || controller.Config.ReportInterval == 0 {
		if err := controller.Config.ParseAgentFlags(); err != nil {
			controller.Logger.Warning(err)
		}
	}
	controller.Logger.Infof(
		"Starting agent with remote server addr: %s, poll interval: %d, report interval: %d\n",
		controller.Config.ServerAddress,
		controller.Config.PollInterval,
		controller.Config.ReportInterval,
	)
	tickPoll := time.NewTicker(time.Duration(controller.Config.PollInterval) * time.Second)
	tickReport := time.NewTicker(time.Duration(controller.Config.ReportInterval) * time.Second)
	go func() {
		for {
			<-tickPoll.C
			controller.Storage.RenewMetrics()
		}
	}()
	go func() {
		for {
			<-tickReport.C
			if err := controller.Storage.SendMetrics("http://" + controller.Config.ServerAddress); err != nil {
				controller.Logger.Error(err)
			}
		}
	}()
	select {}
}
