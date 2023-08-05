package flags

import (
	"flag"
	"fmt"
	"os"
)

var ServerAddr string
var AgentServerAddr string
var AgentReportInterval int64
var AgentPollInterval int64

func ParseServerAddr() {
	flag.StringVar(&ServerAddr, "a", "localhost:8080", "Address and port of server")
	flag.Parse()
}

func ParseAgentFlags() {
	agentFlags := flag.NewFlagSet("agent flags", flag.ExitOnError)
	agentFlags.StringVar(&AgentServerAddr, "a", "localhost:8080", "Address and port of server")
	agentFlags.Int64Var(&AgentReportInterval, "r", 10, "Report interval in seconds")
	agentFlags.Int64Var(&AgentPollInterval, "p", 2, "Metric poll interval in seconds")
	if err := agentFlags.Parse(os.Args[1:]); err != nil {
		fmt.Println(err, agentFlags.ErrorHandling())
		os.Exit(1)
	}

}
