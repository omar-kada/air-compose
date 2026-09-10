package cli

import (
	"omar-kada/air-compose/internal/cli/defaults"
	"omar-kada/air-compose/internal/models"
)

const (
	_configFile   defaults.VarKey = "config-file"
	_workingDir   defaults.VarKey = "working-dir"
	_servicesDir  defaults.VarKey = "services-dir"
	_image        defaults.VarKey = "image"
	_addWritePerm defaults.VarKey = "add-write-perm"
	_port         defaults.VarKey = "port"
	_logLevel     defaults.VarKey = "log-level"
	_logFile      defaults.VarKey = "log-file"
	_humanLog     defaults.VarKey = "human-log"
)

var varInfoMap = defaults.VariableInfoMap{
	_configFile:   {EnvKey: "AIR_COMPOSE_CONFIG_FILE", DefaultValue: "./config.yaml"},
	_workingDir:   {EnvKey: "AIR_COMPOSE_WORKING_DIR", DefaultValue: "./data"},
	_servicesDir:  {EnvKey: "AIR_COMPOSE_SERVICES_DIR", DefaultValue: "."},
	_image:        {EnvKey: "AIR_COMPOSE_DOCKER_IMAGE", DefaultValue: "ghcr.io/omar-kada/air-compose"},
	_addWritePerm: {EnvKey: "AIR_COMPOSE_ADD_WRITE_PERM", DefaultValue: "false"},
	_port:         {EnvKey: "AIR_COMPOSE_PORT", DefaultValue: 5005},
	_logLevel:     {EnvKey: "AIR_COMPOSE_LOG_LEVEL", DefaultValue: "INFO"},
	_logFile:      {EnvKey: "AIR_COMPOSE_LOG_FILE", DefaultValue: "./air-compose.log"},
	_humanLog:     {EnvKey: "AIR_COMPOSE_HUMAN_LOG", DefaultValue: "false"},
}

// RunParams contain parameters of the run command
type RunParams struct {
	models.DeploymentParams
	models.ServerParams
	models.LoggerParams
}

func (p RunParams) getParamsWithDefaults() RunParams {
	return RunParams{
		DeploymentParams: getDeploymentParamsWithDefaults(p.DeploymentParams),
		ServerParams: models.ServerParams{
			Port:     varInfoMap.EnvOrDefaultInt(p.Port, _port),
			FrontDir: "/app/frontend/dist",
		},
		LoggerParams: getLoggerParamsWithDefaults(p.LoggerParams),
	}
}

func getDeploymentParamsWithDefaults(p models.DeploymentParams) models.DeploymentParams {
	return models.DeploymentParams{
		ConfigFile:      varInfoMap.EnvOrDefault(p.ConfigFile, _configFile),
		WorkingDir:      varInfoMap.EnvOrDefault(p.WorkingDir, _workingDir),
		ServicesDir:     varInfoMap.EnvOrDefault(p.ServicesDir, _servicesDir),
		AddWritePerm:    varInfoMap.EnvOrDefault(p.AddWritePerm, _addWritePerm),
		AirComposeImage: varInfoMap.EnvOrDefault(p.AirComposeImage, _image),
	}
}

func getLoggerParamsWithDefaults(p models.LoggerParams) models.LoggerParams {
	return models.LoggerParams{
		LogLevel:          varInfoMap.EnvOrDefault(p.LogLevel, _logLevel),
		LogFile:           varInfoMap.EnvOrDefault(p.LogFile, _logFile),
		HumanLog:          varInfoMap.EnvOrDefault(p.HumanLog, _humanLog),
		MaxLogFileSize:    1,
		MaxLogHistorySize: 200,
	}
}

// RedeployParams contain parameters of the redeploy command
type RedeployParams struct {
	models.DeploymentParams
	models.LoggerParams
}

func (p RedeployParams) getParamsWithDefaults() RedeployParams {
	return RedeployParams{
		DeploymentParams: getDeploymentParamsWithDefaults(p.DeploymentParams),
		LoggerParams:     getLoggerParamsWithDefaults(p.LoggerParams),
	}
}
