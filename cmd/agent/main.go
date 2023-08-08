package main

import (
	"fmt"
	"time"

	"github.com/HellfastUSMC/alert-metrics-service/internal/flags"
	"github.com/HellfastUSMC/alert-metrics-service/internal/storage"

	"github.com/caarlos0/env/v6"
)

func main() {
	conf := flags.SysConfig{}
	err := env.Parse(&conf)
	if err != nil {
		fmt.Println(err)
	}
	if conf.ServerAddress == "" || conf.PollInterval == 0 || conf.ReportInterval == 0 {
		conf.ParseAgentFlags()
	}
	fmt.Printf(
		"Starting agent with remote server addr: %s, poll interval: %d, report interval: %d\n",
		conf.ServerAddress,
		conf.PollInterval,
		conf.ReportInterval,
	)
	tickPoll := time.NewTicker(time.Duration(conf.PollInterval) * time.Second)
	tickReport := time.NewTicker(time.Duration(conf.ReportInterval) * time.Second)
	var stats storage.Metrics
	//for {
	//	for i := int64(1); i <= conf.ReportInterval; i++ {
	//		if i%conf.PollInterval == 0 {
	//			stats.RenewMetrics()
	//		}
	//		if i%conf.ReportInterval == 0 {
	//			err := stats.SendMetrics("http://" + conf.ServerAddress)
	//			if err != nil {
	//				fmt.Printf("there's an error in sending metrics - %e", err)
	//			}
	//		}
	//		time.Sleep(time.Duration(1) * time.Second)
	//	}
	//}
	go func() {
		for {
			<-tickPoll.C
			stats.RenewMetrics()
		}
	}()
	go func() {
		for {
			<-tickReport.C
			if err := stats.SendMetrics("http://" + conf.ServerAddress); err != nil {
				fmt.Println(err)
			}
		}
	}()
	select {}
}
