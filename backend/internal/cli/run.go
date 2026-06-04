package cli

import (
	"context"
	"fmt"
	"log/slog"

	"omar-kada/air-compose/internal/config"
	"omar-kada/air-compose/internal/deployments"
	"omar-kada/air-compose/internal/docker"
	"omar-kada/air-compose/internal/events"
	"omar-kada/air-compose/internal/git"
	"omar-kada/air-compose/internal/models"
	"omar-kada/air-compose/internal/process"
	"omar-kada/air-compose/internal/server"
	"omar-kada/air-compose/internal/server/handlers"
	"omar-kada/air-compose/internal/shell"
	"omar-kada/air-compose/internal/users"

	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

type runCommand struct {
	executor  shell.Executor
	dbCreator func(params RunParams) (*gorm.DB, error)

	cmd    *cobra.Command
	params RunParams
}

// NewRunCommand creates a new run
func NewRunCommand(executor shell.Executor, dbCreator func(params RunParams) (*gorm.DB, error)) *cobra.Command {
	run := runCommand{
		params:    RunParams{},
		executor:  executor,
		dbCreator: dbCreator,
	}

	run.cmd = &cobra.Command{
		Use:   "run",
		Short: "Run with optional config file",
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := run.doRun(); err != nil {
				slog.Error(err.Error())
				return err
			}
			return nil
		},
	}
	run.cmd.Flags().StringVarP(&run.params.ConfigFile, string(_file), "f", "",
		varInfoMap.GetDefaultString("YAML config file", _file))
	run.cmd.Flags().StringVarP(&run.params.WorkingDir, string(_workingDir), "d", "",
		varInfoMap.GetDefaultString("directory where air-compose data will be stored", _workingDir))
	run.cmd.Flags().StringVarP(&run.params.ServicesDir, string(_servicesDir), "s", "",
		varInfoMap.GetDefaultString("directory where services compose stacks will be stored", _servicesDir))
	run.cmd.Flags().StringVarP(&run.params.AddWritePerm, string(_addWritePerm), "w", "",
		varInfoMap.GetDefaultString("when true, the tool adds write permission to files it creates", _addWritePerm))
	run.cmd.Flags().IntVarP(&run.params.Port, string(_port), "p", 0,
		varInfoMap.GetDefaultString("port that will be used for exposing the API/UI", _port))

	return run.cmd
}

func (run *runCommand) doRun() error {
	params := getParamsWithDefaults(run.params)
	slog.Debug("running params : ", "params", params)

	db, err := run.dbCreator(params)
	if err != nil {
		return fmt.Errorf("couldn't init storage %w", err)
	}

	eventStore, err := events.NewEventStorage(db)
	if err != nil {
		return fmt.Errorf("couldn't init EventStorage %w", err)
	}
	deploymentStore, err := deployments.NewDeploymentStorage(db)
	if err != nil {
		return fmt.Errorf("couldn't init DeploymentStorage %w", err)
	}
	userStore, err := users.NewUsersStorage(db)
	if err != nil {
		return fmt.Errorf("couldn't init UserStorage %w", err)
	}
	sessionStore, err := users.NewSessionStorage(db)
	if err != nil {
		return fmt.Errorf("couldn't init SessionStorage %w", err)
	}
	authStore, err := users.NewAuthStorage(userStore, sessionStore, users.NewTokenHolder())
	if err != nil {
		return fmt.Errorf("couldn't init AuthStorage %w", err)
	}

	configStore, err := config.NewConfigStore(params.ConfigFile)
	if err != nil {
		return fmt.Errorf("error creating config storage %w", err)
	}
	err = configStore.WatchFile()
	if err != nil {
		slog.Error("error watching config file", "err", err)
	}
	dispatcher := events.NewDefaultDispatcher([]events.Handler{
		events.NewLoggingEventHandler(),
		events.NewNotificationEventHandler(configStore, eventStore),
	})

	inspector, err := docker.NewInspector(params.ServicesDir, configStore)
	if err != nil {
		return fmt.Errorf("couldn't init docker client %w", err)
	}

	fetcher := git.NewFetcher(params.GetAddWritePerm(), params.GetRepoDir(), configStore)
	deploymentService := process.NewDeploymentService(
		params.DeploymentParams,
		docker.NewDeployer(dispatcher, run.executor),
		fetcher,
		deploymentStore,
		configStore,
		dispatcher)
	dispatcher.AddHandler(process.NewConfigurationUpdatedHandler(deploymentService))
	watcher := process.NewRepoWatcher(fetcher, configStore, deploymentService, dispatcher, process.NewCronScheduler())
	go func() {
		_, err := watcher.Schedule()
		if err != nil {
			slog.Error("error scheduling polling job", "err", err)
			dispatcher.Dispatch(context.Background(), models.EventError,
				fmt.Sprintf("failed to schedule repo polling %v", err))
		}
	}()

	healthChecker := docker.NewHealthChecker(configStore, inspector, dispatcher)
	go healthChecker.ScheduleStateRefresh(context.Background())
	healthHandler := process.NewHealthTransitionHandler(configStore, deploymentService, healthChecker.GetChannel())

	oidcService := users.NewOidcService(configStore, authStore)
	userService := users.NewService(authStore)

	configStore.SetOnChange(func(oldCfg, cfg models.Config) {
		dispatcher.Dispatch(context.Background(), models.EventConfigurationUpdated, "")
		if oldCfg.Settings.Schedule.Cron != cfg.Settings.Schedule.Cron {
			slog.Debug("Rescheduling after cron changed", "oldCron", oldCfg.Settings.Schedule.Cron, "newCron", cfg.Settings.Schedule.Cron)
			watcher.Schedule()
		}
		healthHandler.ResetRetries()
	})

	businessHandler := handlers.NewBusinessHandler(
		configStore, deploymentService, userService,
		fetcher, inspector, watcher,
		eventStore, deploymentStore)
	server := server.NewServer()
	return server.Serve(params.ServerParams, businessHandler, userService, oidcService)
}
