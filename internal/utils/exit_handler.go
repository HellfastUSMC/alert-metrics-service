package utils

import (
	"fmt"
	"os"
	"syscall"
)

func ExitHandler(signal os.Signal) {
	if signal == syscall.SIGTERM || signal == syscall.SIGINT || signal == syscall.SIGQUIT {
		fmt.Println("Got signal. ")
		fmt.Println("Program will terminate now.")
		os.Exit(0)
	} else {
		fmt.Println("Ignoring signal: ", signal)
	}
}
