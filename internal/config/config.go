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
	StoreInterval  int64  `env:"STORE_INTERVAL"`
	DumpPath       string `env:"FILE_STORAGE_PATH"`
	Recover        bool   `env:"RESTORE"`
}

func (c *SysConfig) ParseServerFlags() error {
	serverFlags := flag.NewFlagSet("server config", flag.ExitOnError)
	serverFlags.StringVar(&c.ServerAddress, "a", "localhost:8080", "Address and port of server string")
	serverFlags.StringVar(&c.DumpPath, "f", "/tmp/metrics-db.json", "Path to dump file string")
	serverFlags.Int64Var(&c.StoreInterval, "i", 300, "Storing interval in seconds int")
	serverFlags.BoolVar(&c.Recover, "r", true, "Recover from file sign bool")
	if err := serverFlags.Parse(os.Args[1:]); err != nil {
		os.Exit(1)
		return err
	}
	return nil
}

func (c *SysConfig) ParseAgentFlags() error {
	agentFlags := flag.NewFlagSet("agent config", flag.ExitOnError)
	agentFlags.StringVar(&c.ServerAddress, "a", "localhost:8080", "Address and port of server")
	agentFlags.Int64Var(&c.ReportInterval, "r", 2, "Report interval in seconds")
	agentFlags.Int64Var(&c.PollInterval, "p", 10, "Metric poll interval in seconds")
	if err := agentFlags.Parse(os.Args[1:]); err != nil {
		os.Exit(1)
		return err
	}
	return nil
}

//func parseFlags(config *SysConfig) (*SysConfig, error) {
//	flags := flag.NewFlagSet("agent config", flag.ExitOnError)
//	flags.StringVar(&config.ServerAddress, "a", "localhost:8080", "Address and port of server")
//	flags.Int64Var(&config.ReportInterval, "r", 2, "Report interval in seconds")
//	flags.Int64Var(&config.PollInterval, "p", 10, "Metric poll interval in seconds")
//	flags.StringVar(&config.DumpPath, "f", "/tmp/metrics-db.json", "Path to dump file string")
//	flags.Int64Var(&config.StoreInterval, "i", 300, "Storing interval in seconds int")
//	flags.BoolVar(&config.Recover, "r", true, "Recover from file sign bool")
//	if err := flags.Parse(os.Args[1:]); err != nil {
//		os.Exit(1)
//		return nil, err
//	}
//	return config, nil
//}

func NewConfig() (*SysConfig, error) {
	config := SysConfig{}
	return &config, nil
}

func GetAgentConfigData() (*SysConfig, error) {
	conf, err := NewConfig()
	if err != nil {
		return nil, err
	}
	err = conf.ParseAgentFlags()
	if err != nil {
		return nil, err
	}
	if err := env.Parse(conf); err != nil {
		return conf, err
	}
	return conf, nil
}

func GetServerConfigData() (*SysConfig, error) {
	conf, err := NewConfig()
	if err != nil {
		return nil, err
	}
	err = conf.ParseServerFlags()
	if err != nil {
		return nil, err
	}
	if err := env.Parse(conf); err != nil {
		return conf, err
	}
	return conf, nil
}
