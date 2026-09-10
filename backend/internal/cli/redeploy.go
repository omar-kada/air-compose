package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os/signal"
	"path/filepath"
	"syscall"

	"omar-kada/air-compose/internal/logs"
	"omar-kada/air-compose/internal/shell"

	"github.com/spf13/cobra"
)

type redeployCommand struct {
	executor shell.Executor

	cmd    *cobra.Command
	params RedeployParams
}

// NewRedeployCommand updates the container with the latetst configuration
func NewRedeployCommand(executor shell.Executor) *cobra.Command {
	redeploy := redeployCommand{
		params:   RedeployParams{},
		executor: executor,
	}

	redeploy.cmd = &cobra.Command{
		Use:   "redeploy",
		Short: "redeploy the current air-compose container with latest configuration",
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := redeploy.doRedeploy(); err != nil {
				slog.Error(err.Error())
				return err
			}
			return nil
		},
	}
	redeploy.cmd.Flags().StringVarP(&redeploy.params.WorkingDir, string(_workingDir), "d", "",
		varInfoMap.GetDefaultString("directory where air-compose data will be stored", _workingDir))
	redeploy.cmd.Flags().StringVarP(&redeploy.params.ServicesDir, string(_servicesDir), "s", "",
		varInfoMap.GetDefaultString("directory where services compose stacks will be stored", _servicesDir))

	redeploy.cmd.Flags().StringVarP(&redeploy.params.LogLevel, string(_logLevel), "l", "",
		varInfoMap.GetDefaultString("log level", _logLevel))
	redeploy.cmd.Flags().StringVarP(&redeploy.params.LogFile, string(_logFile), "f", "",
		varInfoMap.GetDefaultString("log file", _logFile))

	return redeploy.cmd
}

func (redeploy *redeployCommand) doRedeploy() error {
	_, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	params := redeploy.params.getParamsWithDefaults()
	logs.InitLogHub(params.LoggerParams)
	slog.Info("redeploy params", "params", params)

	args := []string{"compose", "--project-directory", filepath.Join(params.ServicesDir, "air-compose"), "up", "-d"}
	if _, err := redeploy.executor.Exec("docker", args...); err != nil {
		slog.Error("failed to redeploy air-compose", "err", err)
		return fmt.Errorf("failed to run docker compose up : %w", err)
	}
	return nil
}
