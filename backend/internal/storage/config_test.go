package storage

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"omar-kada/air-compose/models"

	"github.com/stretchr/testify/assert"
)

func TestDecodeConfig(t *testing.T) {
	input := map[string]any{
		"environment": map[string]string{
			"AIR_COMPOSE_HOST": "localhost",
			"DATA_PATH":    "/data",
		},
		"services": map[string]map[string]string{
			"svc1": {
				"PORT":      "8080",
				"VERSION":   "v1",
				"NEW_FIELD": "new_value",
			},
			"svc2": {
				"Disabled": "true",
				"Port":     "9090",
				"Version":  "v2",
			},
		},
	}
	want := models.Config{
		Environment: models.Environment{
			"AIR_COMPOSE_HOST": "localhost",
			"DATA_PATH":    "/data",
		},
		Services: map[string]models.ServiceConfig{
			"svc1": {
				"NEW_FIELD": "new_value",
				"PORT":      "8080",
				"VERSION":   "v1",
			},
			"svc2": {
				"Disabled": "true",
				"Port":     "9090",
				"Version":  "v2",
			},
		},
	}

	cfg, err := decodeConfig(input)
	if err != nil {
		t.Fatalf("decodeConfig failed: %v", err)
	}

	if !reflect.DeepEqual(cfg, want) {
		t.Fatalf("decodeConfig mismatch\nwant=%#v\ngot =%#v", want, cfg)
	}
}

func TestUpdateConfig(t *testing.T) {
	t.Run("successful update", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "config.yaml")
		store := NewConfigStore(filePath)

		input := models.Config{
			Settings: models.Settings{
				Branch:            models.DefaultBranch,
				NotificationTypes: []models.EventType{},
			},
			Environment: models.Environment{
				"AIR_COMPOSE_HOST": "localhost",
				"DATA_PATH":    "/data",
			},
			Services: map[string]models.ServiceConfig{
				"svc1": {
					"PORT":    "8080",
					"VERSION": "v1",
				},
			},
		}

		err := store.Update(input)
		assert.NoError(t, err)

		// Verify the file was written correctly
		_, err = os.ReadFile(filePath)
		assert.NoError(t, err)

		cfg, err := store.Get()
		assert.NoError(t, err)

		// Use deepEqual to compare the result with the expected value
		if !reflect.DeepEqual(cfg, input) {
			t.Errorf("Expected %v, got %v", input, cfg)
		}
	})

	t.Run("on config update callback", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "config.yaml")
		store := NewConfigStore(filePath)

		// Initial config
		initialCfg := models.Config{
			Environment: models.Environment{
				"AIR_COMPOSE_HOST": "localhost",
			},
			Services: map[string]models.ServiceConfig{},
			Settings: models.Settings{
				NotificationTypes: []models.EventType{},
			},
		}
		err := store.Update(initialCfg)
		assert.NoError(t, err)

		// Set up the callback
		var called bool
		var oldCfg, newCfg models.Config
		store.SetOnChange(func(oldC, newC models.Config) {
			called = true
			oldCfg = oldC
			newCfg = newC
		})

		// Update the config
		updatedCfg := models.Config{
			Environment: models.Environment{
				"AIR_COMPOSE_HOST": "new-host",
			},
			Settings: models.Settings{
				NotificationTypes: []models.EventType{},
			},
		}
		err = store.Update(updatedCfg)
		assert.NoError(t, err)

		// Verify the callback was called
		assert.True(t, called)
		assert.Equal(t, initialCfg, oldCfg)
		assert.Equal(t, updatedCfg, newCfg)
	})

	t.Run("file write error", func(t *testing.T) {
		// Create a directory that we can't write to
		tmpDir := t.TempDir()
		readOnlyDir := filepath.Join(tmpDir, "readonly")
		err := os.Mkdir(readOnlyDir, 0o540)
		assert.NoError(t, err)

		filePath := filepath.Join(readOnlyDir, "config.yaml")
		store := NewConfigStore(filePath)
		called := false
		store.SetOnChange(func(_, _ models.Config) {
			called = true
		})

		input := models.Config{
			Environment: models.Environment{
				"AIR_COMPOSE_HOST": "localhost",
			},
			Settings: models.Settings{
				NotificationTypes: []models.EventType{},
			},
		}

		err = store.Update(input)
		assert.Error(t, err)
		assert.False(t, called)
	})

	t.Run("obfuscate token on update", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "config.yaml")
		store := NewConfigStore(filePath)

		input := models.Config{
			Settings: models.Settings{
				Token:           "my-secret-token-12345",
				NotificationURL: "https://example.com/webhook?token=12345",
			},
			Environment: models.Environment{},
			Services:    map[string]models.ServiceConfig{},
		}

		err := store.Update(input)
		assert.NoError(t, err)

		// Verify the token was obfuscated in the stored config
		storedCfg, err := store.Get()
		assert.NoError(t, err)

		storedToken := storedCfg.Settings.Token
		assert.Equal(t, "my-secret-token-12345", storedToken)

		storedURL := storedCfg.Settings.NotificationURL
		assert.Equal(t, "https://example.com/webhook?token=12345", storedURL)

		input2 := models.Config{
			Settings: models.Settings{
				Token:           "my-secret-********************",
				NotificationURL: "https://ex********************",
			},
			Environment: models.Environment{},
			Services:    map[string]models.ServiceConfig{},
		}

		err = store.Update(input2)
		assert.NoError(t, err)

		// Verify the notification URL was obfuscated in the stored config
		storedCfg, err = store.Get()
		assert.NoError(t, err)

		storedToken = storedCfg.Settings.Token
		assert.Equal(t, "my-secret-token-12345", storedToken)

		storedURL = storedCfg.Settings.NotificationURL
		assert.Equal(t, "https://example.com/webhook?token=12345", storedURL)
	})
}

func TestLoadConfig_FileError(t *testing.T) {
	t.Run("missing file", func(t *testing.T) {
		configStore := NewConfigStore("/does/not/exist.yaml")
		cfg, err := configStore.Get()
		assert.NoError(t, err)
		assert.Equal(t, models.Config{}, cfg)
	})

	t.Run("invalid yaml", func(t *testing.T) {
		// create a temporary file with invalid YAML content so the test
		tmp := t.TempDir()
		f := filepath.Join(tmp, "invalid.yaml")
		invalid := []byte("this: is: not: valid: yaml")
		err := os.WriteFile(f, invalid, 0o644)
		assert.NoError(t, err)

		configStore := NewConfigStore(f)

		_, err = configStore.Get()
		assert.Error(t, err)
	})
}
