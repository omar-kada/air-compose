package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadFeatures_Defaults(t *testing.T) {
	t.Setenv("AIR_COMPOSE_DISPLAY_CONFIG", "")
	t.Setenv("AIR_COMPOSE_EDIT_CONFIG", "")
	t.Setenv("AIR_COMPOSE_EDIT_SETTINGS", "")
	t.Setenv("AIR_COMPOSE_DISPLAY_CMD_LOGS", "")

	features := LoadFeatures()
	assert.True(t, features.DisplayConfig)
	assert.True(t, features.EditConfig)
	assert.True(t, features.EditSettings)
	assert.False(t, features.DisplayCmdLogs)
}

func TestLoadFeatures_Disabled(t *testing.T) {
	t.Setenv("AIR_COMPOSE_DISPLAY_CONFIG", "false")
	t.Setenv("AIR_COMPOSE_EDIT_CONFIG", "false")
	t.Setenv("AIR_COMPOSE_EDIT_SETTINGS", "false")
	t.Setenv("AIR_COMPOSE_DISPLAY_CMD_LOGS", "true")

	features := LoadFeatures()
	assert.False(t, features.DisplayConfig)
	assert.False(t, features.EditConfig)
	assert.False(t, features.EditSettings)
	assert.True(t, features.DisplayCmdLogs)
}

func TestLoadFeatures_InvalidBool(t *testing.T) {
	t.Setenv("AIR_COMPOSE_DISPLAY_CONFIG", "notabool")

	features := LoadFeatures()
	assert.True(t, features.DisplayConfig) // falls back to default
}

func TestGetBool(t *testing.T) {
	t.Setenv("TEST_BOOL_KEY", "true")
	assert.True(t, getBool("TEST_BOOL_KEY", false))

	t.Setenv("TEST_BOOL_KEY", "false")
	assert.False(t, getBool("TEST_BOOL_KEY", true))

	t.Setenv("TEST_BOOL_KEY", "")
	assert.True(t, getBool("TEST_BOOL_KEY", true)) // empty → default

	t.Setenv("TEST_BOOL_KEY", "invalid")
	assert.False(t, getBool("TEST_BOOL_KEY", false)) // invalid → default
}
