package cli

import (
	"context"
	"fmt"
	"log/slog"

	"omar-kada/air-compose/internal/docker"
	"omar-kada/air-compose/internal/events"
	"omar-kada/air-compose/internal/git"
	"omar-kada/air-compose/internal/models"
	"omar-kada/air-compose/internal/process"
	"omar-kada/air-compose/internal/server"
	"omar-kada/air-compose/internal/server/handlers"
	"omar-kada/air-compose/internal/shell"
	"omar-kada/air-compose/internal/storage"
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

	eventStore, err := storage.NewEventStorage(db)
	if err != nil {
		return fmt.Errorf("couldn't init EventStorage %w", err)
	}
	deploymentStore, err := storage.NewDeploymentStorage(db)
	if err != nil {
		return fmt.Errorf("couldn't init DeploymentStorage %w", err)
	}
	userStore, err := storage.NewUsersStorage(db)
	if err != nil {
		return fmt.Errorf("couldn't init UserStorage %w", err)
	}
	sessionStore, err := storage.NewSessionStorage(db)
	if err != nil {
		return fmt.Errorf("couldn't init SessionStorage %w", err)
	}
	authStore, err := storage.NewAuthStorage(userStore, sessionStore, storage.NewTokenHolder())
	if err != nil {
		return fmt.Errorf("couldn't init AuthStorage %w", err)
	}

	configStore, err := storage.NewConfigStore(params.ConfigFile)
	if err != nil {
		return fmt.Errorf("error creating config storage %w", err)

	}
	err = configStore.WatchFile()
	if err != nil {
		slog.Error("error watching config file", "err", err)
	}
	config := configStore.Get()
	dispatcher := events.NewDefaultDispatcher([]events.EventHandler{
		events.NewLoggingEventHandler(),
		events.NewNotificationEventHandler(configStore, eventStore),
	})
	scheduler := process.NewConfigScheduler(configStore)

	inspector, err := docker.NewInspector(params.ServicesDir, configStore)
	if err != nil {
		return fmt.Errorf("couldn't init docker client %w", err)
	}
	healthChecker := docker.NewHealthChecker(configStore, inspector, dispatcher)
	go healthChecker.ScheduleStateRefresh(context.Background())

	fetcher := git.NewFetcher(params.GetAddWritePerm(), params.GetRepoDir(), configStore)
	deploymentService := process.NewDeploymentService(
		params.DeploymentParams,
		docker.NewDeployer(dispatcher, run.executor),
		inspector,
		fetcher,
		deploymentStore,
		configStore,
		dispatcher,
		scheduler)
	healthHandler := process.NewHealthTransitionHandler(configStore, deploymentService, healthChecker.GetChannel())
	oidcService := users.NewOidcService(config.Settings.Oidc, authStore)
	userService := users.NewService(authStore)
	go func() {
		_, err = scheduler.Schedule(func() {
			_, err := deploymentService.SyncDeployment()
			if err != nil {
				slog.Error(err.Error())
			}
		})
		if err != nil {
			slog.Warn(err.Error())
		}
	}()

	configStore.SetOnChange(func(oldCfg, cfg models.Config) {
		dispatcher.Dispatch(context.Background(), models.EventConfigurationUpdated, "")
		if oldCfg.Settings.Schedule.Cron != cfg.Settings.Schedule.Cron {
			slog.Debug("Rescheduling after cron changed", "oldCron", oldCfg.Settings.Schedule.Cron, "newCron", cfg.Settings.Schedule.Cron)
			scheduler.ReSchedule()
		}
		oidcService.OnConfigChanged(cfg.Settings.Oidc)
		healthHandler.ResetRetries()
	})

	businessHandler := handlers.NewBusinessHandler(configStore, deploymentService, userService, fetcher, inspector, eventStore, deploymentStore)
	server := server.NewServer()
	return server.Serve(params.ServerParams, businessHandler, userService, oidcService)
}
