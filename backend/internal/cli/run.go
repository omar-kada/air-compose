package cli

import (
	"context"
	"fmt"
	"log/slog"

	"omar-kada/air-compose/internal/docker"
	"omar-kada/air-compose/internal/events"
	"omar-kada/air-compose/internal/files"
	"omar-kada/air-compose/internal/git"
	"omar-kada/air-compose/internal/process"
	"omar-kada/air-compose/internal/server"
	"omar-kada/air-compose/internal/shell"
	"omar-kada/air-compose/internal/storage"
	"omar-kada/air-compose/internal/users"
	"omar-kada/air-compose/models"

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

	configStore := storage.NewConfigStore(params.ConfigFile)
	dispatcher := events.NewDefaultDispatcher([]events.EventHandler{
		events.NewLoggingEventHandler(),
		events.NewNotificationEventHandler(configStore, eventStore),
	})
	scheduler := process.NewConfigScheduler(configStore)
	configStore.SetOnChange(func(oldCfg, cfg models.Config) {
		oldYamlCfg, _ := configStore.ToYaml(oldCfg)
		newYamlCfg, _ := configStore.ToYaml(cfg)
		dispatcher.Dispatch(context.Background(), models.EventConfigurationUpdated,
			files.DiffText(string(oldYamlCfg), string(newYamlCfg)))
		if oldCfg.Settings.Cron != cfg.Settings.Cron {
			slog.Debug("Rescheduling after cron changed", "oldCron", oldCfg.Settings.Cron, "newCron", cfg.Settings.Cron)
			scheduler.ReSchedule()
		}
	})
	inspector, err := docker.NewInspector()
	if err != nil {
		return fmt.Errorf("couldn't init docker client %w", err)
	}
	service := process.NewService(
		params.DeploymentParams,
		docker.NewDeployer(dispatcher, run.executor),
		inspector,
		git.NewFetcher(params.GetAddWritePerm(), params.GetRepoDir(), configStore),
		deploymentStore,
		eventStore,
		configStore,
		dispatcher,
		scheduler)
	userService := users.NewService(userStore)
	go func() {
		_, err = scheduler.Schedule(func() {
			_, err := service.SyncDeployment()
			if err != nil {
				slog.Error(err.Error())
			}
		})
		if err != nil {
			slog.Warn(err.Error())
		}
	}()
	server := server.NewServer(configStore, service, userService)
	return server.Serve(params.ServerParams)
}
