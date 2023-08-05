package main

import (
	"fmt"
	"github.com/HellfastUSMC/alert-metrics-service/internal/flags"
	"github.com/HellfastUSMC/alert-metrics-service/internal/storage"
	"time"
)

func main() {
	flags.ParseAgentFlags()
	var stats storage.Metrics
	conf := storage.SysConfig{
		PollInterval:   flags.AgentPollInterval,
		ReportInterval: flags.AgentReportInterval,
	}
	for {
		for i := int64(1); i <= conf.ReportInterval; i++ {
			if i%conf.PollInterval == 0 {
				stats.RenewMetrics()
			}
			if i%conf.ReportInterval == 0 {
				err := stats.SendMetrics("http://" + flags.AgentServerAddr)
				if err != nil {
					fmt.Printf("there's an error in sending metrics - %e", err)
				}
			}
			time.Sleep(time.Duration(1) * time.Second)
		}
	}
}
