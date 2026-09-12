// Package main is the entry point for AirCompose.
package main

import (
	"log/slog"
	"os"

	"omar-kada/air-compose/internal/cli"
	"omar-kada/air-compose/internal/shell"
)

func main() {
	retcode := 0
	defer func() { os.Exit(retcode) }()

	rootCmd := cli.NewRootCmd(shell.NewExecutor())
	if err := rootCmd.Execute(); err != nil {
		slog.Error("error executing root command", "error", err)
		retcode = 1 // it exits with code 1
	}
}
