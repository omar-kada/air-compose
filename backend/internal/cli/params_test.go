package cli

import (
	"testing"

	"omar-kada/air-compose/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestGetParamsWithDefaults_AllCliValuesProvided(t *testing.T) {
	// When all CLI values are provided, they should be returned as-is
	params := RunParams{
		DeploymentParams: models.DeploymentParams{
			ConfigFile:   "custom.yaml",
			WorkingDir:   "/custom/work",
			ServicesDir:  "/custom/services",
			AddWritePerm: "1",
		},
		ServerParams: models.ServerParams{
			Port: 9090,
		},
	}

	result := params.getParamsWithDefaults()

	assert.Equal(t, "/custom/work/custom.yaml", result.ConfigFile)
	assert.Equal(t, "/custom/work", result.WorkingDir)
	assert.Equal(t, "/custom/services", result.ServicesDir)
	assert.Equal(t, "1", result.AddWritePerm)
	assert.Equal(t, 9090, result.Port)
}

func TestGetParamsWithDefaults_UseEnvVariablesWhenCliEmpty(t *testing.T) {
	t.Setenv("AIR_COMPOSE_WORKING_DIR", "/env/work")
	t.Setenv("AIR_COMPOSE_SERVICES_DIR", "/env/services")
	t.Setenv("AIR_COMPOSE_CONFIG_FILE", "env1.yaml")
	t.Setenv("AIR_COMPOSE_PORT", "5005")

	params := RunParams{}

	result := params.getParamsWithDefaults()

	assert.Equal(t, "/env/work/env1.yaml", result.ConfigFile)
	assert.Equal(t, "/env/work", result.WorkingDir)
	assert.Equal(t, "/env/services", result.ServicesDir)
	assert.Equal(t, "false", result.AddWritePerm) // default Value
	assert.Equal(t, 5005, result.Port)            // Value from env
}

func TestGetParamsWithDefaults_UseDefaultsWhenCliAndEnvEmpty(t *testing.T) {
	// Clear environment variables
	t.Setenv("AIR_COMPOSE_WORKING_DIR", "")
	t.Setenv("AIR_COMPOSE_SERVICES_DIR", "")
	t.Setenv("AIR_COMPOSE_CONFIG_FILE", "")
	t.Setenv("AIR_COMPOSE_PORT", "")

	params := RunParams{}

	result := params.getParamsWithDefaults()

	// Check defaults are applied
	assert.Equal(t, "data/config.yaml", result.ConfigFile)
	assert.Equal(t, "./data", result.WorkingDir)
	assert.Equal(t, ".", result.ServicesDir)
	assert.Equal(t, "false", result.AddWritePerm) // Default value
	assert.Equal(t, 5005, result.Port)            // Default value
}

func TestGetParamsWithDefaults_CliPriority(t *testing.T) {
	// CLI values should take priority over env variables and defaults
	t.Setenv("AIR_COMPOSE_CONFIG_BRANCH", "env-branch")
	t.Setenv("AIR_COMPOSE_ADD_WRITE_PERM", "false")
	t.Setenv("AIR_COMPOSE_PORT", "5005")

	params := RunParams{
		DeploymentParams: models.DeploymentParams{
			ServicesDir:  "/s",
			AddWritePerm: "true",
		},
		ServerParams: models.ServerParams{
			Port: 9090,
		},
	}

	result := params.getParamsWithDefaults()

	// CLI value should win
	assert.Equal(t, "/s", result.ServicesDir)
	assert.Equal(t, "true", result.AddWritePerm)
	assert.Equal(t, 9090, result.Port)
}

func TestGetParamsWithDefaults_MixedSources(t *testing.T) {
	// Test a mix of CLI values, env variables, and defaults
	t.Setenv("AIR_COMPOSE_CONFIG_BRANCH", "env-branch")
	t.Setenv("AIR_COMPOSE_WORKING_DIR", "")
	t.Setenv("AIR_COMPOSE_ADD_WRITE_PERM", "true")
	t.Setenv("AIR_COMPOSE_PORT", "5005")

	params := RunParams{
		DeploymentParams: models.DeploymentParams{
			ConfigFile:  "cli.yaml", // From CLI
			ServicesDir: "/s",       // From CLI (overrides env)
			// WorkingDir && addWritePerm not provided, should use env or default
		},
		ServerParams: models.ServerParams{
			Port: 9090, // From CLI
		},
	}

	result := params.getParamsWithDefaults()

	assert.Equal(t, "data/cli.yaml", result.ConfigFile)
	assert.Equal(t, "/s", result.ServicesDir)
	assert.Equal(t, "./data", result.WorkingDir) // Should use default
	assert.Equal(t, "true", result.AddWritePerm)
	assert.Equal(t, 9090, result.Port)
}
func TestGetParamsWithDefaults_RedeployParams_AllCliValuesProvided(t *testing.T) {
	// When all CLI values are provided, they should be returned as-is
	params := RedeployParams{
		DeploymentParams: models.DeploymentParams{
			ConfigFile:   "custom.yaml",
			WorkingDir:   "/custom/work",
			ServicesDir:  "/custom/services",
			AddWritePerm: "1",
		},
		LoggerParams: models.LoggerParams{
			LogLevel: "DEBUG",
			LogFile:  "/custom/logs/app.log",
			HumanLog: "true",
		},
	}

	result := params.getParamsWithDefaults()

	assert.Equal(t, "/custom/work/custom.yaml", result.ConfigFile)
	assert.Equal(t, "/custom/work", result.WorkingDir)
	assert.Equal(t, "/custom/services", result.ServicesDir)
	assert.Equal(t, "1", result.AddWritePerm)
	assert.Equal(t, "DEBUG", result.LogLevel)
	assert.Equal(t, "/custom/logs/app.log", result.LogFile)
	assert.Equal(t, "true", result.HumanLog)
}

func TestGetParamsWithDefaults_RedeployParams_UseEnvVariablesWhenCliEmpty(t *testing.T) {
	t.Setenv("AIR_COMPOSE_WORKING_DIR", "/env/work")
	t.Setenv("AIR_COMPOSE_SERVICES_DIR", "/env/services")
	t.Setenv("AIR_COMPOSE_CONFIG_FILE", "env1.yaml")
	t.Setenv("AIR_COMPOSE_LOG_LEVEL", "WARN")
	t.Setenv("AIR_COMPOSE_LOG_FILE", "/env/logs/app.log")

	params := RedeployParams{}

	result := params.getParamsWithDefaults()

	assert.Equal(t, "/env/work/env1.yaml", result.ConfigFile)
	assert.Equal(t, "/env/work", result.WorkingDir)
	assert.Equal(t, "/env/services", result.ServicesDir)
	assert.Equal(t, "false", result.AddWritePerm) // default Value
	assert.Equal(t, "WARN", result.LogLevel)      // Value from env
	assert.Equal(t, "/env/logs/app.log", result.LogFile)
	assert.Equal(t, "false", result.HumanLog) // default Value
}

func TestGetParamsWithDefaults_RedeployParams_UseDefaultsWhenCliAndEnvEmpty(t *testing.T) {
	// Clear environment variables
	t.Setenv("AIR_COMPOSE_WORKING_DIR", "")
	t.Setenv("AIR_COMPOSE_SERVICES_DIR", "")
	t.Setenv("AIR_COMPOSE_CONFIG_FILE", "")
	t.Setenv("AIR_COMPOSE_LOG_LEVEL", "")
	t.Setenv("AIR_COMPOSE_LOG_FILE", "")

	params := RedeployParams{}

	result := params.getParamsWithDefaults()

	// Check defaults are applied
	assert.Equal(t, "data/config.yaml", result.ConfigFile)
	assert.Equal(t, "./data", result.WorkingDir)
	assert.Equal(t, ".", result.ServicesDir)
	assert.Equal(t, "false", result.AddWritePerm) // Default value
	assert.Equal(t, "INFO", result.LogLevel)      // Default value
	assert.Equal(t, "data/air-compose.log", result.LogFile)
	assert.Equal(t, "false", result.HumanLog) // Default value
}

func TestGetParamsWithDefaults_RedeployParams_CliPriority(t *testing.T) {
	// CLI values should take priority over env variables and defaults
	t.Setenv("AIR_COMPOSE_CONFIG_FILE", "env.yaml")
	t.Setenv("AIR_COMPOSE_ADD_WRITE_PERM", "false")
	t.Setenv("AIR_COMPOSE_LOG_LEVEL", "ERROR")

	params := RedeployParams{
		DeploymentParams: models.DeploymentParams{
			ServicesDir:  "/s",
			AddWritePerm: "true",
		},
		LoggerParams: models.LoggerParams{
			LogLevel: "DEBUG",
		},
	}

	result := params.getParamsWithDefaults()

	// CLI value should win
	assert.Equal(t, "/s", result.ServicesDir)
	assert.Equal(t, "true", result.AddWritePerm)
	assert.Equal(t, "DEBUG", result.LogLevel)
}

func TestGetParamsWithDefaults_RedeployParams_MixedSources(t *testing.T) {
	// Test a mix of CLI values, env variables, and defaults
	t.Setenv("AIR_COMPOSE_WORKING_DIR", "/env/work")
	t.Setenv("AIR_COMPOSE_ADD_WRITE_PERM", "true")
	t.Setenv("AIR_COMPOSE_LOG_FILE", "/env/logs/app.log")

	params := RedeployParams{
		DeploymentParams: models.DeploymentParams{
			ConfigFile:  "cli.yaml", // From CLI
			ServicesDir: "/s",       // From CLI (overrides env)
			// WorkingDir && addWritePerm not provided, should use env or default
		},
		LoggerParams: models.LoggerParams{
			LogLevel: "WARN", // From CLI
			// LogFile not provided, should use env or default
		},
	}

	result := params.getParamsWithDefaults()

	assert.Equal(t, "/env/work/cli.yaml", result.ConfigFile)
	assert.Equal(t, "/s", result.ServicesDir)
	assert.Equal(t, "/env/work", result.WorkingDir) // Should use env
	assert.Equal(t, "true", result.AddWritePerm)
	assert.Equal(t, "WARN", result.LogLevel)
	assert.Equal(t, "/env/logs/app.log", result.LogFile) // Should use env
	assert.Equal(t, "false", result.HumanLog)            // Should use default
}
