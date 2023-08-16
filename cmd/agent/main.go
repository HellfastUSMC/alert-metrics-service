package main

import (
	"os"
	"time"

	"github.com/HellfastUSMC/alert-metrics-service/internal/agent-storage"
	"github.com/HellfastUSMC/alert-metrics-service/internal/config"
	"github.com/HellfastUSMC/alert-metrics-service/internal/controllers"
	"github.com/sirupsen/logrus"
)

func main() {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	log.SetOutput(os.Stdout)
	conf, err := config.NewConfig()
	if err != nil {
		log.Warning(err)
	}
	if conf.ServerAddress == "" || conf.PollInterval == 0 || conf.ReportInterval == 0 {
		if err := conf.ParseAgentFlags(); err != nil {
			log.Warning(err)
		}
	}
	controller := controllers.NewAgentController(log, conf, agentstorage.NewMetricsStorage())
	controller.Infof(
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
			controller.RenewMetrics()
		}
	}()
	go func() {
		for {
			<-tickReport.C
			if err := controller.SendMetrics("http://" + controller.Config.ServerAddress); err != nil {
				controller.Error(err)
			}
		}
	}()
	select {}
}
