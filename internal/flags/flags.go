package flags

import (
	"flag"
	"fmt"
	"os"
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

func (c *SysConfig) ParseAgentFlags() {
	agentFlags := flag.NewFlagSet("agent flags", flag.ExitOnError)
	agentFlags.StringVar(&c.ServerAddress, "a", "localhost:8080", "Address and port of server")
	agentFlags.Int64Var(&c.ReportInterval, "r", 10, "Report interval in seconds")
	agentFlags.Int64Var(&c.PollInterval, "p", 2, "Metric poll interval in seconds")
	if err := agentFlags.Parse(os.Args[1:]); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}

func NewConfig() *SysConfig {
	return &SysConfig{
		PollInterval:   0,
		ReportInterval: 0,
		ServerAddress:  "",
	}
}
