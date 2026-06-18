package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"

	"omar-kada/air-compose/internal/config"
	"omar-kada/air-compose/internal/deployments"
	"omar-kada/air-compose/internal/docker"
	"omar-kada/air-compose/internal/events"
	"omar-kada/air-compose/internal/git"
	"omar-kada/air-compose/internal/models"
	"omar-kada/air-compose/internal/process"
	"omar-kada/air-compose/internal/server"
	"omar-kada/air-compose/internal/server/handlers"
	"omar-kada/air-compose/internal/server/socket"
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
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

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

	eventBus := events.NewBus(10)
	configStore, err := config.NewConfigStore(params.ConfigFile, eventBus)
	if err != nil {
		return fmt.Errorf("error creating config storage %w", err)
	}
	eventBus.SetTransform(events.NewEventTransformer(configStore).HandleEvent)

	inspector, err := docker.NewInspector(params.ServicesDir, configStore)
	if err != nil {
		return fmt.Errorf("couldn't init docker client %w", err)
	}

	fetcher := git.NewFetcher(params.GetAddWritePerm(), params.GetRepoDir(), configStore)
	deploymentService := process.NewDeploymentService(
		params.DeploymentParams,
		docker.NewDeployer(eventBus, run.executor),
		fetcher,
		deploymentStore,
		configStore,
		eventBus)

	repoWatcher := process.NewRepoWatcher(fetcher, configStore, deploymentService, eventBus, process.NewCronScheduler())
	healthChecker := docker.NewHealthChecker(configStore, inspector, eventBus)
	healthTransitionHandler := process.NewHealthTransitionHandler(configStore, deploymentService)

	oidcService := users.NewOidcService(configStore, authStore)
	userService := users.NewService(authStore, eventBus)

	// register events consumers
	eventBus.Register(
		process.NewConfigurationUpdatedHandler(deploymentService, eventBus),
		events.HandlerFunc(healthTransitionHandler.HandleHealthCheck),
		events.NewLoggingEventHandler(),
		events.NewNotificationEventHandler(configStore, eventStore),
		events.HandlerFunc(func(_ context.Context, event models.Event) {
			if event.Type == models.EventConfigurationUpdated {
				data, ok := event.Data.(models.EventDataChange[models.Config])
				if !ok {
					slog.Error("issue with event data type ", "event", event)
				} else if data.Old.Settings.Schedule.Cron != data.New.Settings.Schedule.Cron {
					slog.Debug("Rescheduling after cron changed",
						"oldCron", data.Old.Settings.Schedule.Cron,
						"newCron", data.New.Settings.Schedule.Cron)
					repoWatcher.Schedule()
				}
				healthTransitionHandler.ResetRetries()
			}
		}),
	)

	go eventBus.Run(ctx)

	// launch event publishers
	err = configStore.WatchFile()
	if err != nil {
		slog.Error("error watching config file", "err", err)
	}
	go repoWatcher.Schedule()
	go healthChecker.ScheduleStateRefresh(ctx)

	// launch server

	businessHandler := handlers.NewBusinessHandler(
		configStore, deploymentService, userService,
		fetcher, healthChecker, repoWatcher,
		eventStore, deploymentStore)

	socketHandler := socket.NewWebSocketHandler()

	eventBus.Register(events.HandlerFunc(socketHandler.BroadcastEvent))

	// create websocket handler and then register consumer of events to send to client
	server := server.NewServer()
	return server.Serve(params.ServerParams, businessHandler, socketHandler, userService, oidcService)
}
