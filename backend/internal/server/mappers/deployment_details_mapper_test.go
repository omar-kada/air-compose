// Package mappers provides functionality for mapping between different data models.
package mappers

import (
	"testing"
	"time"

	"omar-kada/air-compose/api"
	"omar-kada/air-compose/models"

	"github.com/stretchr/testify/assert"
)

func TestDeploymentDetailsMapper_Map(t *testing.T) {
	// Setup
	diffMapper := DiffMapper{}
	eventMapper := EventMapper{}
	deploymentMapper := NewDeploymentDetailsMapper(diffMapper, eventMapper)

	// Test data
	deployment := models.Deployment{
		ID:      1,
		Author:  "testAuthor",
		Diff:    "testDiff",
		Status:  models.DeploymentStatusSuccess,
		Time:    time.Now(),
		EndTime: time.Now().Add(time.Hour),
		Title:   "testTitle",
		Events:  []models.Event{{Type: models.EventMisc, Msg: "testEvent", Time: time.Now()}},
		Files:   []models.FileDiff{{ID: 1, Diff: "testDiff", NewFile: "testNewFile", OldFile: "testOldFile"}},
	}

	// Expected result
	expected := api.DeploymentWithDetails{
		Author:  "testAuthor",
		Diff:    "testDiff",
		Id:      "1",
		Status:  api.DeploymentStatusSuccess,
		Time:    deployment.Time,
		EndTime: deployment.EndTime,
		Title:   "testTitle",
		Events:  []api.Event{{Type: api.EventTypeMISC, Msg: "testEvent", Time: deployment.Events[0].Time}},
		Files:   []api.FileDiff{{Diff: "testDiff", NewFile: "testNewFile", OldFile: "testOldFile"}},
	}

	// Execute
	actual := deploymentMapper.Map(deployment)

	// Assert
	assert.Equal(t, expected, actual)
}
