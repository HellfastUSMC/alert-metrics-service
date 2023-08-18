package config

import (
	"flag"
	"os"

	"github.com/caarlos0/env/v6"
)

type SysConfig struct {
	PollInterval   int64  `env:"POLL_INTERVAL"`
	ReportInterval int64  `env:"REPORT_INTERVAL"`
	ServerAddress  string `env:"ADDRESS"`
}

func (c *SysConfig) ParseServerAddr() {
	flag.StringVar(&c.ServerAddress, "a", "localhost:8080", "Address and port of server")
	flag.Parse()
}

func (c *SysConfig) ParseAgentFlags() error {
	agentFlags := flag.NewFlagSet("agent config", flag.ExitOnError)
	agentFlags.StringVar(&c.ServerAddress, "a", "localhost:8080", "Address and port of server")
	agentFlags.Int64Var(&c.ReportInterval, "r", 2, "Report interval in seconds")
	agentFlags.Int64Var(&c.PollInterval, "p", 1, "Metric poll interval in seconds")
	if err := agentFlags.Parse(os.Args[1:]); err != nil {
		os.Exit(1)
		return err
	}
	return nil
}

func NewConfig() (*SysConfig, error) {
	config := SysConfig{}
	if err := env.Parse(&config); err != nil {
		return &config, err
	}
	return &config, nil
}
