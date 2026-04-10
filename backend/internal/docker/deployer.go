// Package docker implements operations for docker containers
package docker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"omar-kada/air-compose/internal/events"
	"omar-kada/air-compose/internal/files"
	"omar-kada/air-compose/internal/shell"
	"omar-kada/air-compose/models"
)

// Deployer defines methods for managing containerized services.
type Deployer interface {
	WithCtx(ctx context.Context) Deployer
	RemoveServices(services []string, servicesDir string) map[string]error
	DeployServices(cfg models.Config, params models.DeploymentParams) map[string]error
	RemoveAndDeployStacks(oldCfg, cfg models.Config, params models.DeploymentParams) error
}

// NewDeployer creates an instance of Manager for docker containers
func NewDeployer(dispatcher events.Dispatcher, executor shell.Executor) Deployer {
	return &deployer{
		cmdExecuter:  executor,
		envGenerator: NewEnvGenerator(),
		copier:       files.NewCopier(),
		dispatcher:   dispatcher,
		ctx:          context.Background(),
	}
}

// deployer manages Docker Compose services.
type deployer struct {
	cmdExecuter  shell.Executor
	envGenerator *EnvGenerator
	copier       files.Copier
	dispatcher   events.Dispatcher
	ctx          context.Context
}

// WithCtx sets the logger for the Deployer
func (d deployer) WithCtx(ctx context.Context) Deployer {
	newDeployer := d
	newDeployer.ctx = ctx
	return newDeployer
}

// RemoveServices stops and removes Docker Compose services.
func (d deployer) RemoveServices(services []string, servicesDir string) map[string]error {
	d.dispatcher.Dispatch(d.ctx, models.EventMisc, fmt.Sprintf("these services will be removed if running %v", services))
	errors := make(map[string]error)
	for _, service := range services {

		composeDir := filepath.Join(servicesDir, service)

		if info, err := os.Stat(composeDir); os.IsNotExist(err) || !info.IsDir() {
			d.dispatcher.Dispatch(d.ctx, models.EventMisc, fmt.Sprintf("Skipping docker compose down for %s: directory does not exist", service))
			continue
		}

		err := d.composeDown(composeDir)
		if err != nil {
			d.dispatcher.Dispatch(d.ctx, models.EventError, fmt.Sprintf("Error running docker compose down for %s: %v", service, err))
			errors[service] = err
		}
	}

	return errors
}

// DeployServices generates .env files and runs Docker Compose for enabled services.
func (d deployer) DeployServices(cfg models.Config, params models.DeploymentParams) map[string]error {
	enabledServices := cfg.GetEnabledServices()
	if len(enabledServices) == 0 {
		d.dispatcher.Dispatch(d.ctx, models.EventMisc, "No enabled services specified in config. Skipping .env generation and compose up.")
		return nil
	}

	errors := make(map[string]error)
	for _, service := range enabledServices {

		if err := d.copyServiceFiles(service, params); err != nil {
			d.dispatcher.Dispatch(d.ctx, models.EventError, fmt.Sprintf("Error copying service files for %s : %v", service, err))
			errors[service] = err
			continue
		}
		if err := d.envGenerator.generateEnvFile(cfg, params.ServicesDir, service); err != nil {
			d.dispatcher.Dispatch(d.ctx, models.EventError, fmt.Sprintf("Error creating env file for %s : %v", service, err))
			errors[service] = err
			continue
		}
		if err := d.composeUp(filepath.Join(params.ServicesDir, service)); err != nil {
			d.dispatcher.Dispatch(d.ctx, models.EventError, fmt.Sprintf("Error running docker compose for %s : %v", service, err))
			errors[service] = err
		}
	}
	return errors
}

func (d deployer) composeUp(composePath string) error {
	args := []string{"-i", "docker", "compose", "--project-directory", composePath, "up", "-d"}
	if _, err := d.cmdExecuter.Exec("env", args...); err != nil {
		return fmt.Errorf("failed to run docker compose up : %w", err)
	}
	return nil
}

func (d deployer) composeDown(composePath string) error {
	args := []string{"-i", "docker", "compose", "--project-directory", composePath, "down"}
	if _, err := d.cmdExecuter.Exec("env", args...); err != nil {
		return fmt.Errorf("failed to run docker compose down : %w", err)
	}
	return nil
}

// RemoveAndDeployStacks handles the deployment/removal of services based on the current and new configuration.
func (d deployer) RemoveAndDeployStacks(oldCfg, cfg models.Config, params models.DeploymentParams) error {
	toBeRemoved := getUnusedServices(oldCfg, cfg)
	if len(toBeRemoved) > 0 {
		if errs := d.RemoveServices(toBeRemoved, params.ServicesDir); len(errs) > 0 {
			return fmt.Errorf("error while removing services : %v", errs)
		}
	}

	enabledServiecs := cfg.GetEnabledServices()

	d.dispatcher.Dispatch(d.ctx, models.EventMisc, fmt.Sprintf("deploying enabled services %v", enabledServiecs))
	if errs := d.DeployServices(cfg, params); len(errs) > 0 {
		return fmt.Errorf("error(s) while deploying services : %v", errs)
	}
	return nil
}

func (d deployer) copyServiceFiles(serviceName string, params models.DeploymentParams) error {
	src := filepath.Join(params.GetRepoDir(), "services", serviceName)
	dst := filepath.Join(params.ServicesDir, serviceName)
	if err := d.copier.Copy(src, dst); err != nil {
		return err
	}
	return nil
}

func getUnusedServices(oldCfg, cfg models.Config) []string {
	var unusedServices []string
	currentlyEnabled := oldCfg.GetEnabledServices()
	shouldBeEnabled := cfg.GetEnabledServices()
	for _, serviceName := range currentlyEnabled {
		if !slices.Contains(shouldBeEnabled, serviceName) {
			unusedServices = append(unusedServices, serviceName)
		}
	}
	return unusedServices
}
