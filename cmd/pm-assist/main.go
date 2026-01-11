package main

import (
	"os"

	"github.com/pm-assist/pm-assist/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(cli.ExitUnknownError)
	}
}
