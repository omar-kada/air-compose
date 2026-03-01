package mappers

import (
	"testing"
	"time"

	"omar-kada/air-compose/api"
	"omar-kada/air-compose/models"

	"github.com/stretchr/testify/assert"
)

func TestDeploymentMapper_Map(t *testing.T) {
	// Setup
	deploymentMapper := NewDeploymentMapper()

	// Test data
	deployment := models.Deployment{
		ID:      1,
		Author:  "testAuthor",
		Diff:    "testDiff",
		Status:  models.DeploymentStatusSuccess,
		Time:    time.Now(),
		EndTime: time.Now().Add(time.Hour),
		Title:   "testTitle",
		Events:  []models.Event{{Type: models.EventError, Msg: "testEvent", Time: time.Now()}},
		Files:   []models.FileDiff{{ID: 1, Diff: "testDiff", NewFile: "testNewFile", OldFile: "testOldFile"}},
	}

	// Expected result
	expected := api.Deployment{
		Author:  "testAuthor",
		Diff:    "testDiff",
		Id:      "1",
		Status:  api.DeploymentStatusSuccess,
		Time:    deployment.Time,
		EndTime: deployment.EndTime,
		Title:   "testTitle",
	}

	// Execute
	actual := deploymentMapper.Map(deployment)

	// Assert
	assert.Equal(t, expected, actual)
}

func TestDeploymentMapper_MapToPageInfo(t *testing.T) {
	// Setup
	deploymentMapper := NewDeploymentMapper()

	// Test data
	deployments := []models.Deployment{
		{
			ID:      1,
			Author:  "testAuthor1",
			Diff:    "testDiff1",
			Status:  models.DeploymentStatusSuccess,
			Time:    time.Now(),
			EndTime: time.Now().Add(time.Hour),
			Title:   "testTitle1",
			Events:  []models.Event{{Type: models.EventError, Msg: "testEvent1", Time: time.Now()}},
			Files:   []models.FileDiff{{ID: 1, Diff: "testDiff1", NewFile: "testNewFile1", OldFile: "testOldFile1"}},
		},
		{
			ID:      2,
			Author:  "testAuthor2",
			Diff:    "testDiff2",
			Status:  models.DeploymentStatusSuccess,
			Time:    time.Now(),
			EndTime: time.Now().Add(time.Hour),
			Title:   "testTitle2",
			Events:  []models.Event{{Type: models.EventError, Msg: "testEvent2", Time: time.Now()}},
			Files:   []models.FileDiff{{ID: 2, Diff: "testDiff2", NewFile: "testNewFile2", OldFile: "testOldFile2"}},
		},
	}

	// Test cases
	tests := []struct {
		name        string
		deployments []models.Deployment
		limit       int
		expected    api.PageInfo
	}{
		{
			name:        "No deployments",
			deployments: []models.Deployment{},
			limit:       2,
			expected: api.PageInfo{
				HasNextPage: false,
				EndCursor:   "",
			},
		},
		{
			name:        "Less deployments than limit",
			deployments: deployments,
			limit:       3,
			expected: api.PageInfo{
				HasNextPage: false,
				EndCursor:   "2",
			},
		},
		{
			name:        "Equal deployments to limit",
			deployments: deployments,
			limit:       2,
			expected: api.PageInfo{
				HasNextPage: true,
				EndCursor:   "2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			actual := deploymentMapper.MapToPageInfo(tt.deployments, tt.limit)

			// Assert
			assert.Equal(t, tt.expected, actual)
		})
	}
}
