// Package main is the entry point for AirCompose.
package main

import (
	"log/slog"
	"os"
	"strings"
	"time"

	"omar-kada/air-compose/internal/cli"
	"omar-kada/air-compose/internal/shell"

	"github.com/lmittmann/tint"
)

func main() {
	retcode := 0
	defer func() { os.Exit(retcode) }()

	isDev := strings.ToUpper(os.Getenv("ENV")) == "DEV"
	if isDev {
		slog.SetDefault(slog.New(
			tint.NewHandler(os.Stdout, &tint.Options{
				Level:      slog.LevelDebug,
				TimeFormat: time.Kitchen,
				AddSource:  true,
			}),
		))
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
