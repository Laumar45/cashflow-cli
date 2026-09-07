package main

import (
	"os"

	"cashflow/internal/adapters/cli"
)

func main() {
	exitCode := cli.Execute()
	if exitCode != 0 {
		os.Exit(exitCode)
	}
}
