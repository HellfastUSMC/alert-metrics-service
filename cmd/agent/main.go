package main

import (
	"github.com/HellfastUSMC/alert-metrics-service/internal/storage"
	"time"
)

func main() {
	var stats storage.Metrics
	conf := storage.SysConfig{
		PollInterval:   2,
		ReportInterval: 10,
	}
	for {
		for i := int64(1); i <= conf.ReportInterval; i++ {
			if i%conf.PollInterval == 0 {
				stats.RenewMetrics()
			}
			if i%conf.ReportInterval == 0 {
				stats.SendMetrics("http://localhost:8088")
			}
			time.Sleep(time.Duration(1) * time.Second)
		}
	}
}
