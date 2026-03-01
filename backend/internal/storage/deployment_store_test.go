package storage

import (
	"fmt"
	"testing"
	"time"

	"omar-kada/air-compose/models"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupDeploymentStorage(t *testing.T) (DeploymentStorage, *gorm.DB) {
	db, err := NewGormDb(":memory:", 0o000)
	if err != nil {
		t.Fatalf("new db: %v", err)
	}
	deploymentStore, err := NewDeploymentStorage(db)
	if err != nil {
		t.Fatalf("new storage: %v", err)
	}
	return deploymentStore, db
}

func TestDeploymentStorage_Migrates(t *testing.T) {
	_, db := setupDeploymentStorage(t)
	// ensure migrations created the deployments table
	has := db.Migrator().HasTable(&models.Deployment{})
	assert.True(t, has)
}

func TestInitAndGetDeployment(t *testing.T) {
	s, _ := setupDeploymentStorage(t)
	files := []models.FileDiff{{Diff: "d1", NewFile: "n1", OldFile: "o1"}}
	dep, err := s.InitDeployment("title1", "author1", "diff1", files)
	assert.NoError(t, err)
	assert.NotZero(t, dep.ID)
	assert.Equal(t, models.DeploymentStatusRunning, dep.Status)

	got, err := s.GetDeployment(dep.ID)
	assert.NoError(t, err)
	assert.Equal(t, dep.ID, got.ID)
	assert.Equal(t, files[0].Diff, got.Files[0].Diff)
	assert.Equal(t, files[0].NewFile, got.Files[0].NewFile)
	assert.Equal(t, files[0].OldFile, got.Files[0].OldFile)
	assert.Empty(t, got.Events)
}

func TestGetDeployment_NoNExisting(t *testing.T) {
	s, _ := setupDeploymentStorage(t)

	dep, err := s.GetDeployment(999999)

	// Verify that no error is returned and the deployment is empty
	assert.NoError(t, err)
	assert.Equal(t, models.Deployment{}, dep)
}

func TestGetLastAndGetDeploymentsOrdering(t *testing.T) {
	s, _ := setupDeploymentStorage(t)
	_, err := s.InitDeployment("title1", "author1", "diff1", nil)
	assert.NoError(t, err)

	// small sleep to ensure time difference
	time.Sleep(2 * time.Millisecond)
	dep2, _ := s.InitDeployment("title2", "author2", "diff2", nil)
	time.Sleep(2 * time.Millisecond)
	dep3, _ := s.InitDeployment("title3", "author3", "diff3", nil)

	last, err := s.GetLastDeployment()
	assert.NoError(t, err)
	assert.Equal(t, dep3.ID, last.ID)

	deps, err := s.GetDeployments(Cursor[uint64]{Limit: 10, Offset: 999999})
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(deps), 3)
	assert.Equal(t, dep3.ID, deps[0].ID)
	assert.Equal(t, dep2.ID, deps[1].ID)
}

func TestEndDeploymentAndErrorCases(t *testing.T) {
	s, _ := setupDeploymentStorage(t)
	dep, err := s.InitDeployment("title1", "author1", "diff1", nil)
	assert.NoError(t, err)

	assert.NoError(t, s.EndDeployment(dep.ID, models.DeploymentStatusSuccess))
	d, err := s.GetDeployment(dep.ID)
	assert.NoError(t, err)
	assert.Equal(t, models.DeploymentStatusSuccess, d.Status)
	assert.False(t, d.EndTime.IsZero())
}

func TestGetDeployments_Pagination(t *testing.T) {
	cases := []struct {
		name     string
		seed     int
		cursor   Cursor[uint64]
		expected int
	}{
		{"DefaultOffset", 15, NewIDCursor(5, 0), 5},
		{"LimitCapped", 50, NewIDCursor(200, 0), 50},
		{"OffsetWithinRange", 15, NewIDCursor(10, 10), 9},

		{"ExactLimit", 10, NewIDCursor(3, 0), 3},
		{"NoResults", 10, Cursor[uint64]{Limit: 10, Offset: 1}, 0},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s, _ := setupDeploymentStorage(t)
			for i := 1; i <= c.seed; i++ {
				_, err := s.InitDeployment(fmt.Sprintf("t%d", i), "author", "diff", nil)
				assert.NoError(t, err)
			}

			deps, err := s.GetDeployments(c.cursor)
			assert.NoError(t, err)
			assert.Equal(t, c.expected, len(deps))

			for _, d := range deps {
				assert.Less(t, d.ID, c.cursor.Offset)
			}
		})
	}
}
