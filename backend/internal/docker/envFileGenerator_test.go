package docker

import (
	"omar-kada/air-compose/models"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateEnvFile(t *testing.T) {

	generator := NewEnvGenerator()
	tempDir := t.TempDir()
	serviceDir := filepath.Join(tempDir, "testService")
	err := os.Mkdir(serviceDir, 0750)
	assert.NoError(t, err)

	tests := []struct {
		name         string
		envFileLines []string // if not nil, write these lines to the .env file before test
		cfg          models.Config
		wantContent  []string // lines that should exist in the file, in order, and only these
	}{
		{
			name:         "No existing .env file",
			envFileLines: nil,
			cfg:          models.Config{},
			wantContent:  []string{OverrideHeader, ""},
		},
		{
			name:         "Existing .env file with some variables",
			envFileLines: []string{"VAR1=value1", "VAR2=value2"},
			cfg: models.Config{
				Services: map[string]models.ServiceConfig{
					"testService": {
						"VAR1": "newValue1",
						"VAR3": "value3",
					},
				},
			},
			wantContent: []string{
				"# VAR1=value1 # Overridden by AirCompose ",
				"VAR2=value2",
				OverrideHeader,
				"VAR3=value3",
				"VAR1=newValue1",
				"",
			},
		},
		{
			name:         "Env file already contains OVERRIDE_HEADER and more values",
			envFileLines: []string{"VAR1=value1", OverrideHeader, "VAR2=value2", "VAR3=value3"},
			cfg: models.Config{
				Services: map[string]models.ServiceConfig{
					"testService": {
						"VAR2": "newValue2",
						"VAR4": "value4",
					},
				},
			},
			wantContent: []string{
				"VAR1=value1",
				OverrideHeader,
				"VAR4=value4",
				"VAR2=newValue2",
				"",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envFilePath := filepath.Join(serviceDir, ".env")
			// Clean up before each test
			_ = os.Remove(envFilePath)
			if tt.envFileLines != nil {
				err := os.WriteFile(envFilePath, []byte(strings.Join(tt.envFileLines, "\n")), 0644)
				assert.NoError(t, err)
			}

			err := generator.generateEnvFile(tt.cfg, tempDir, "testService")
			assert.NoError(t, err)

			content, err := os.ReadFile(envFilePath)
			assert.NoError(t, err)

			assert.ElementsMatch(t, tt.wantContent, strings.Split(string(content), "\n"))

		})
	}
}
