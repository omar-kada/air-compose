package cli

import (
	"omar-kada/air-compose/internal/cli/defaults"
	"omar-kada/air-compose/models"
)

const (
	_file         defaults.VarKey = "file"
	_workingDir   defaults.VarKey = "working-dir"
	_servicesDir  defaults.VarKey = "services-dir"
	_addWritePerm defaults.VarKey = "add-write-perm"
	_port         defaults.VarKey = "port"
	_httpMode     defaults.VarKey = "http-mode"
)

var varInfoMap = defaults.VariableInfoMap{
	_file:         {EnvKey: "AIR_COMPOSE_CONFIG_FILE", DefaultValue: "/data/config.yaml"},
	_workingDir:   {EnvKey: "AIR_COMPOSE_WORKING_DIR", DefaultValue: "./config"},
	_servicesDir:  {EnvKey: "AIR_COMPOSE_SERVICES_DIR", DefaultValue: "."},
	_addWritePerm: {EnvKey: "AIR_COMPOSE_ADD_WRITE_PERM", DefaultValue: "false"},
	_port:         {EnvKey: "AIR_COMPOSE_PORT", DefaultValue: 5005},
	_httpMode:     {EnvKey: "AIR_COMPOSE_HTTP_MODE", DefaultValue: "false"},
}

// RunParams contain parameters of the run command
type RunParams struct {
	models.DeploymentParams
	models.ServerParams
	ConfigFile string
}

func getParamsWithDefaults(p RunParams) RunParams {
	return RunParams{
		ConfigFile: varInfoMap.EnvOrDefault(p.ConfigFile, _file),
		DeploymentParams: models.DeploymentParams{
			WorkingDir:   varInfoMap.EnvOrDefault(p.WorkingDir, _workingDir),
			ServicesDir:  varInfoMap.EnvOrDefault(p.ServicesDir, _servicesDir),
			AddWritePerm: varInfoMap.EnvOrDefault(p.AddWritePerm, _addWritePerm),
		},
		ServerParams: models.ServerParams{
			Port:     varInfoMap.EnvOrDefaultInt(p.Port, _port),
			HTTPMode: varInfoMap.EnvOrDefault(p.HTTPMode, _httpMode),
		},
	}
}
