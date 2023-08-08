package flags

import (
	"flag"
	"fmt"
	"github.com/HellfastUSMC/alert-metrics-service/internal/storage"
	"os"
)

func ParseServerAddr(c *storage.SysConfig) {
	flag.StringVar(&c.ServerAddress, "a", "localhost:8080", "Address and port of server")
	flag.Parse()
}

func ParseAgentFlags(c *storage.SysConfig) {
	agentFlags := flag.NewFlagSet("agent flags", flag.ExitOnError)
	agentFlags.StringVar(&c.ServerAddress, "a", "localhost:8080", "Address and port of server")
	agentFlags.Int64Var(&c.ReportInterval, "r", 10, "Report interval in seconds")
	agentFlags.Int64Var(&c.PollInterval, "p", 2, "Metric poll interval in seconds")
	if err := agentFlags.Parse(os.Args[1:]); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
