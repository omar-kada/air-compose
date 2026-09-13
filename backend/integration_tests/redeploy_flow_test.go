package integrationtests

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"omar-kada/air-compose/internal/docker"
	"omar-kada/air-compose/internal/events"
	"omar-kada/air-compose/internal/shell"
	"omar-kada/air-compose/testutil"

	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"
)

func TestRedeploy(t *testing.T) {
	ctx := context.Background()
	t.Setenv("AIR_COMPOSE_DISPLAY_CMD_LOGS", "true")
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))

	// Given
	baseDir, err := os.MkdirTemp("", "redeploy_test_*")
	assert.NoError(t, err)

	t.Cleanup(func() {
		err = os.RemoveAll(baseDir)
		// print warning in case of errors that may occur in ci
		if err != nil {
			fmt.Printf("WARN : error while cleaning up test files : %v\n", err)
		}
	})

	servicesDir := filepath.Join(baseDir, "services")
	airComposeDir := filepath.Join(servicesDir, "air-compose")
	dataDir := filepath.Join(baseDir, "data")
	dbDir := filepath.Join(dataDir, "db")
	err = os.Mkdir(servicesDir, 0o750)
	assert.NoError(t, err)
	err = os.Mkdir(airComposeDir, 0o750)
	assert.NoError(t, err)
	err = os.Mkdir(dataDir, 0o750)
	assert.NoError(t, err)
	err = os.Mkdir(dbDir, 0o750)
	assert.NoError(t, err)

	err = os.WriteFile(filepath.Join(dataDir, "config.yaml"),
		[]byte(strings.Join([]string{
			"settings:",
			"  git:",
			"    repo: \"https://github.com/omar-kada/air-compose-config\"",
			"    branch: \"test\"",
			"  schedule:",
			"    cron: \"* * * * *\"",
			"environment:",
			"services:",
			"  air-compose:",
		}, "\n")), 0o750)
	assert.NoError(t, err, "error while creating config file")

	// When
	executor := shell.NewExecutor().WithLogs()
	_, err = executor.Exec("docker", "build", "-t", "air-compose:local", "../..")
	assert.NoError(t, err, "error while building docker image")
	_, err = executor.Exec("curl", "-L", "-o", filepath.Join(airComposeDir, "compose.yaml"), "https://github.com/omar-kada/air-compose-config/raw/test/services/air-compose/compose.yaml")
	assert.NoError(t, err, "error while downloading compose file")

	// Create .env file with additional environment variables
	envFilePath := filepath.Join(airComposeDir, ".env")

	err = os.WriteFile(envFilePath, []byte(strings.Join([]string{
		"AIR_COMPOSE_SERVICES_DIR=" + servicesDir,
		"AIR_COMPOSE_WORKING_DIR=" + dataDir,
		"AIR_COMPOSE_ADD_WRITE_PERM=true",
		"AIR_COMPOSE_DISPLAY_CMD_LOGS=true",
		"AIR_COMPOSE_LOG_LEVEL=DEBUG",
		"AIR_COMPOSE_DOCKER_IMAGE=air-compose:local",
		"UID=" + fmt.Sprint(os.Getuid()),
		"GID=" + fmt.Sprint(os.Getgid()),
	}, "\n")), 0o750)
	assert.NoError(t, err, "error while creating .env file")

	_, err = executor.Exec("docker", "compose", "--project-directory", airComposeDir, "--progress", "quiet", "up", "-d", "--quiet-pull")
	assert.NoError(t, err, "error while running docker compose up")

	oldContainer, ok := testutil.WaitForContainer(ctx, "/air-compose", 2*time.Minute)
	assert.True(t, ok, "old container not found")

	// Append new environment variable to the .env file to trigger a real update
	appendToFile(t, envFilePath, "\nAIR_COMPOSE_REDEPLOY_TEST=true")

	_, ok = testutil.WaitForContainer(ctx, "/air-compose-updater", 2*time.Minute)
	assert.True(t, ok, "updater container not found")

	ok = testutil.WaitForDeadContainer(ctx, "/air-compose-updater", 2*time.Minute)
	assert.True(t, ok, "updater container was not removed")

	newContainer, ok := testutil.WaitForContainer(ctx, "/air-compose", 2*time.Minute)
	assert.True(t, ok, "new container not found")

	assert.ElementsMatch(t, oldContainer.Names, newContainer.Names)
	assert.NotEqual(t, oldContainer.ID, newContainer.ID)

	t.Cleanup(func() {
		/// cleanup homepage container after test finishes
		dockerDeployer := docker.NewDeployer(events.NewBus(1), executor)
		dockerDeployer.RemoveServices([]string{"air-compose"}, servicesDir)
		executor.Exec("docker", "rm", "air-compose-updater")
		executor.Exec("docker", "rm", "air-compose")
	})
}

func TestNoRedeploy(t *testing.T) {
	ctx := context.Background()

	t.Setenv("AIR_COMPOSE_DISPLAY_CMD_LOGS", "true")
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))

	// Given
	baseDir, err := os.MkdirTemp("", "redeploy_test_*")
	assert.NoError(t, err)

	t.Cleanup(func() {
		err = os.RemoveAll(baseDir)
		// print warning in case of errors that may occur in ci
		if err != nil {
			fmt.Printf("WARN : error while cleaning up test files : %v\n", err)
		}
	})

	servicesDir := filepath.Join(baseDir, "services")
	airComposeDir := filepath.Join(servicesDir, "air-compose")
	dataDir := filepath.Join(baseDir, "data")
	dbDir := filepath.Join(dataDir, "db")
	err = os.Mkdir(servicesDir, 0o750)
	assert.NoError(t, err)
	err = os.Mkdir(airComposeDir, 0o750)
	assert.NoError(t, err)
	err = os.Mkdir(dataDir, 0o750)
	assert.NoError(t, err)
	err = os.Mkdir(dbDir, 0o750)
	assert.NoError(t, err)

	err = os.WriteFile(filepath.Join(dataDir, "config.yaml"),
		[]byte(strings.Join([]string{
			"settings:",
			"  git:",
			"    repo: \"https://github.com/omar-kada/air-compose-config\"",
			"    branch: \"test\"",
			"  schedule:",
			"    cron: \"* * * * *\"",
			"environment:",
			"services:",
			"  air-compose:",
		}, "\n")), 0o750)
	assert.NoError(t, err, "error while creating config file")

	// When
	executor := shell.NewExecutor().WithLogs()
	_, err = executor.Exec("docker", "build", "-t", "air-compose:local", "../..")
	assert.NoError(t, err, "error while building docker image")
	_, err = executor.Exec("curl", "-L", "-o", filepath.Join(airComposeDir, "compose.yaml"), "https://github.com/omar-kada/air-compose-config/raw/test/services/air-compose/compose.yaml")
	assert.NoError(t, err, "error while downloading compose file")

	// Create .env file with additional environment variables
	envFilePath := filepath.Join(airComposeDir, ".env")

	err = os.WriteFile(envFilePath, []byte(strings.Join([]string{
		"AIR_COMPOSE_SERVICES_DIR=" + servicesDir,
		"AIR_COMPOSE_WORKING_DIR=" + dataDir,
		"AIR_COMPOSE_ADD_WRITE_PERM=true",
		"AIR_COMPOSE_DISPLAY_CMD_LOGS=true",
		"AIR_COMPOSE_LOG_LEVEL=DEBUG",
		"AIR_COMPOSE_DOCKER_IMAGE=air-compose:local",
		"UID=" + fmt.Sprint(os.Getuid()),
		"GID=" + fmt.Sprint(os.Getgid()),
	}, "\n")), 0o750)
	assert.NoError(t, err, "error while creating .env file")

	_, err = executor.Exec("docker", "compose", "--project-directory", airComposeDir, "--progress", "quiet", "up", "-d", "--quiet-pull")
	assert.NoError(t, err, "error while running docker compose up")

	oldContainer, ok := testutil.WaitForContainer(ctx, "/air-compose", 2*time.Minute)
	assert.True(t, ok, "old container not found")

	updater, ok := testutil.WaitForContainer(ctx, "/air-compose-updater", 2*time.Minute)
	assert.True(t, ok, "updater container not found")

	cli, err := client.New(client.FromEnv, client.WithUserAgent("integration-tests/1.0"))
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()

	testutil.FollowContainerLogs(ctx, t, cli, updater.ID)

	ok = testutil.WaitForDeadContainer(ctx, "/air-compose-updater", 2*time.Minute)
	assert.True(t, ok, "updater container was not removed")

	newContainer, ok := testutil.WaitForContainer(ctx, "/air-compose", 2*time.Minute)
	assert.True(t, ok, "new container not found")

	assert.ElementsMatch(t, oldContainer.Names, newContainer.Names)
	assert.Equal(t, oldContainer.ID, newContainer.ID) // shouldn't redeploy beacause nothing changed

	t.Cleanup(func() {
		/// cleanup homepage container after test finishes
		dockerDeployer := docker.NewDeployer(events.NewBus(1), executor)
		dockerDeployer.RemoveServices([]string{"air-compose"}, servicesDir)
		executor.Exec("docker", "rm", "air-compose-updater")
		executor.Exec("docker", "rm", "air-compose")
	})
}

func appendToFile(t *testing.T, filePath string, value string) {
	fileContent, err := os.ReadFile(filePath)
	assert.NoError(t, err, "error while reading %v file", filePath)

	fileContent = append(fileContent, []byte(value)...)
	err = os.WriteFile(filePath, fileContent, 0o750)
	assert.NoError(t, err, "error while updating %v file", filePath)
}
