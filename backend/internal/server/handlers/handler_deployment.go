package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"strconv"

	"omar-kada/air-compose/api"
	"omar-kada/air-compose/internal/deployments"
	"omar-kada/air-compose/internal/docker"
	"omar-kada/air-compose/internal/events"
	"omar-kada/air-compose/internal/git"
	"omar-kada/air-compose/internal/models"
	"omar-kada/air-compose/internal/process"
	"omar-kada/air-compose/internal/server/mappers"
	"omar-kada/air-compose/internal/storage"
)

// DeploymentHandler handles deployment-related operations.
type DeploymentHandler struct {
	processService  process.DeploymentService
	deploymentStore deployments.DeploymentStorage
	eventStore      events.EventStorage
	configStore     models.ConfigGetter
	fetcher         git.Fetcher
	inspector       docker.Inspector
	watcher         process.RepoWatcher

	depMapper        mappers.PageMapper[models.Deployment, api.Deployment]
	depDetailsMapper mappers.Mapper[models.Deployment, api.DeploymentWithDetails]
	statusMapper     mappers.Mapper[models.ContainerSummary, api.ContainerStatus]
	stateMapper      mappers.Mapper[models.State, api.State]
	eventMapper      mappers.PageMapper[models.Event, api.Event]
	diffMapper       mappers.Mapper[models.FileDiff, api.FileDiff]
}

// DeployementAPIList lists deployments with pagination support
func (h *DeploymentHandler) DeployementAPIList(_ context.Context, request api.DeployementAPIListRequestObject) (api.DeployementAPIListResponseObject, error) {
	offset, err := validateCursorOffset(request.Params.Offset)
	if err != nil {
		return nil, fmt.Errorf("invalid after value")
	}

	if request.Params.Limit <= 0 {
		return nil, fmt.Errorf("invalid first value")
	}

	deps, err := h.deploymentStore.GetDeployments(storage.NewIDCursor(int(request.Params.Limit), offset))

	return api.DeployementAPIList200JSONResponse{
		Items:    models.ListMapper(h.depMapper.Map)(deps),
		PageInfo: h.depMapper.MapToPageInfo(deps, int(request.Params.Limit)),
	}, err
}

// DeployementAPIRead retrieves details of a specific deployment
func (h *DeploymentHandler) DeployementAPIRead(_ context.Context, request api.DeployementAPIReadRequestObject) (api.DeployementAPIReadResponseObject, error) {
	id, err := strconv.ParseUint(request.Id, 10, 64)
	if err != nil {
		return nil, err
	}
	dep, err := h.deploymentStore.GetDeployment(id)
	if err != nil {
		return nil, err
	} else if dep.ID == 0 {
		return api.DeployementAPIReaddefaultJSONResponse{
			Body: api.Error{
				Code: api.ErrorCodeNOTFOUND,
			},
			StatusCode: http.StatusNotFound,
		}, nil
	}

	events, err := h.eventStore.GetEvents(dep.ID)
	if err != nil {
		return nil, err
	}

	depDTO := h.depDetailsMapper.Map(dep)
	depDTO.Events = models.ListMapper(h.eventMapper.Map)(events)
	return api.DeployementAPIRead200JSONResponse(depDTO), err
}

// DeployementAPISync syncs the deployment
func (h *DeploymentHandler) DeployementAPISync(_ context.Context, _ api.DeployementAPISyncRequestObject) (api.DeployementAPISyncResponseObject, error) {
	dep, err := h.processService.DoDeploy(process.DeploymentTriggerManual, models.Patch{})
	if err != nil {
		slog.Error(err.Error())
	} else if reflect.DeepEqual(models.Deployment{}, dep) {
		return api.DeployementAPISync204Response{}, nil
	}
	return api.DeployementAPISync200JSONResponse(h.depDetailsMapper.Map(dep)), err
}

// StatusAPIGet retrieves the status of managed stacks
func (h *DeploymentHandler) StatusAPIGet(_ context.Context, _ api.StatusAPIGetRequestObject) (api.StatusAPIGetResponseObject, error) {
	stacks, err := h.inspector.GetManagedStacks()
	if err != nil {
		return nil, err
	}

	result := models.MapMapper[string](
		models.MapMapper[string](h.statusMapper.Map),
	)(map[string]map[string]models.ContainerSummary(stacks))
	return api.StatusAPIGet200JSONResponse(result), nil
}

// StateAPIGet retrieves the state of AirCompose
func (h *DeploymentHandler) StateAPIGet(_ context.Context, _ api.StateAPIGetRequestObject) (api.StateAPIGetResponseObject, error) {
	dep, _ := h.deploymentStore.GetLastDeployment()
	cfg := h.configStore.Get()
	stackstate, _ := h.inspector.GetManagedStacks()

	return api.StateAPIGet200JSONResponse(h.stateMapper.Map(models.State{
		LastStatus:  dep.Status,
		NextDeploy:  h.watcher.GetNext(),
		Health:      stackstate.GetGlobalHealth(),
		Initialized: cfg.Settings.Git.Repo != "",
	})), nil
}

// DiffAPIGet retrieves the differences in files
func (h *DeploymentHandler) DiffAPIGet(_ context.Context, _ api.DiffAPIGetRequestObject) (api.DiffAPIGetResponseObject, error) {
	diff, err := h.fetcher.DiffWithRemote()
	if err != nil {
		return nil, err
	}
	return api.DiffAPIGet200JSONResponse(models.ListMapper(h.diffMapper.Map)(diff.Files)), nil
}

// NotificationsAPIList lists notifications with pagination support
func (h *DeploymentHandler) NotificationsAPIList(_ context.Context, request api.NotificationsAPIListRequestObject) (api.NotificationsAPIListResponseObject, error) {
	offset, err := validateCursorOffset(request.Params.Offset)
	if err != nil {
		return nil, fmt.Errorf("invalid after value")
	}

	if request.Params.Limit <= 0 {
		return nil, fmt.Errorf("invalid first value")
	}

	events, err := h.eventStore.GetNotifications(storage.NewIDCursor(int(request.Params.Limit), offset))

	return api.NotificationsAPIList200JSONResponse{
		Items:    models.ListMapper(h.eventMapper.Map)(events),
		PageInfo: h.eventMapper.MapToPageInfo(events, int(request.Params.Limit)),
	}, err
}
