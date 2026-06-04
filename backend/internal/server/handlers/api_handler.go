// Package handlers provides HTTP request handlers for the application.
package handlers

import (
	"omar-kada/air-compose/internal/config"
	"omar-kada/air-compose/internal/deployments"
	"omar-kada/air-compose/internal/docker"
	"omar-kada/air-compose/internal/events"
	"omar-kada/air-compose/internal/git"
	"omar-kada/air-compose/internal/process"
	"omar-kada/air-compose/internal/server/mappers"
	"omar-kada/air-compose/internal/users"
	"strconv"
)

// BusinessHandler implements the generated strict server interface
type BusinessHandler struct {
	*AuthUserHandler
	*DeploymentHandler
	*SettingsConfigHandler
}

// NewBusinessHandler creates a new Handler
func NewBusinessHandler(
	configStore config.Store,
	processService process.DeploymentService,
	userService users.AccountService,
	fetcher git.Fetcher,
	inspector docker.Inspector,
	watcher process.RepoWatcher,
	eventStore events.EventStorage,
	deploymentStore deployments.DeploymentStorage,
) *BusinessHandler {
	diffMapper := mappers.DiffMapper{}
	eventMapper := mappers.EventMapper{}

	return &BusinessHandler{
		&AuthUserHandler{
			accountService: userService,
			configStore:    configStore,
			userMapper:     mappers.UserMapper{},
		},
		&DeploymentHandler{
			processService:   processService,
			deploymentStore:  deploymentStore,
			eventStore:       eventStore,
			configStore:      configStore,
			fetcher:          fetcher,
			inspector:        inspector,
			watcher:          watcher,
			depMapper:        mappers.DeploymentMapper{},
			depDetailsMapper: mappers.NewDeploymentDetailsMapper(diffMapper, eventMapper),
			statusMapper:     mappers.StatusMapper{},
			stateMapper:      mappers.StateMapper{},
			eventMapper:      eventMapper,
			diffMapper:       diffMapper,
		},
		&SettingsConfigHandler{
			configStore:    configStore,
			fetcher:        fetcher,
			configMapper:   mappers.ConfigMapper{},
			settingsMapper: mappers.SettingsMapper{},
			featuresMapper: mappers.FeaturesMapper{},
		},
	}
}

func validateCursorOffset(offsetStr *string) (uint64, error) {
	offset := uint64(0)
	var err error
	if offsetStr != nil && *offsetStr != "" {
		offset, err = strconv.ParseUint(*offsetStr, 10, 64)
	}
	return offset, err
}
