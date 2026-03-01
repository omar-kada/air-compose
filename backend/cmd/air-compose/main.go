// Package main is the entry point for AirCompose.
package main

import (
	"log/slog"
	"os"
	"strings"

	"omar-kada/air-compose/internal/cli"
	"omar-kada/air-compose/internal/shell"
)

func main() {
	retcode := 0
	defer func() { os.Exit(retcode) }()

	isDev := strings.ToUpper(os.Getenv("ENV")) == "DEV"
	if isDev {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))
	} else {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))
	}

	// Add subcommands
	rootCmd := cli.NewRootCmd(shell.NewExecutor())
	if err := rootCmd.Execute(); err != nil {
		slog.Error("error executing root command", "error", err)
		retcode = 1 // it exits with code 1
	}
}
