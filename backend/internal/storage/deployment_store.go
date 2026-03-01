package storage

import (
	"errors"
	"time"

	"omar-kada/air-compose/models"

	"gorm.io/gorm"
)

// gormDeploymentStorage implements the Storage interface using GORM
type gormDeploymentStorage struct {
	db *gorm.DB
}

// DeploymentStorage is an abstraction of all deployment database operations
type DeploymentStorage interface {
	GetDeployments(c Cursor[uint64]) ([]models.Deployment, error)
	GetDeployment(id uint64) (models.Deployment, error)
	InitDeployment(title string, author string, diff string, files []models.FileDiff) (models.Deployment, error)
	EndDeployment(deploymentID uint64, status models.DeploymentStatus) error
	GetLastDeployment() (models.Deployment, error)
}

// NewDeploymentStorage creates a storage for deployments using gorm
func NewDeploymentStorage(db *gorm.DB) (DeploymentStorage, error) {
	// Auto-migrate models types
	if err := db.AutoMigrate(&models.Deployment{}, &models.FileDiff{}); err != nil {
		return nil, err
	}
	return &gormDeploymentStorage{db: db}, nil
}

// GetDeployments retrieves all deployments with their associated files and events
func (s *gormDeploymentStorage) GetDeployments(c Cursor[uint64]) ([]models.Deployment, error) {
	var deps []models.Deployment
	if err := s.db.
		Scopes(Paginate(c)).Order("Time desc").Find(&deps).Error; err != nil {
		return nil, err
	}
	return deps, nil
}

// GetDeployment retrieves a specific deployment by ID with its associated files and events
func (s *gormDeploymentStorage) GetDeployment(id uint64) (models.Deployment, error) {
	var dep models.Deployment
	if err := s.db.Preload("Files").Preload("Events").First(&dep, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Deployment{}, nil
		}
		return models.Deployment{}, err
	}
	return dep, nil
}

// InitDeployment creates a new deployment with the given parameters
func (s *gormDeploymentStorage) InitDeployment(title string, author string, diff string, files []models.FileDiff) (models.Deployment, error) {
	dep := models.Deployment{
		Title:  title,
		Author: author,
		Diff:   diff,
		Status: models.DeploymentStatusRunning,
		Time:   time.Now(),
		Files:  files,
		Events: []models.Event{},
	}
	if err := s.db.Create(&dep).Error; err != nil {
		return models.Deployment{}, err
	}
	return dep, nil
}

// EndDeployment updates a deployment's status and sets its end time
func (s *gormDeploymentStorage) EndDeployment(deploymentID uint64, status models.DeploymentStatus) error {
	var dep models.Deployment
	if err := s.db.First(&dep, deploymentID).Error; err != nil {
		return err
	}
	dep.Status = status
	dep.EndTime = time.Now()
	return s.db.Save(&dep).Error
}

// GetLastDeployment returns the most recent deployment based on Time (or ID) descending
func (s *gormDeploymentStorage) GetLastDeployment() (models.Deployment, error) {
	var dep models.Deployment
	req := s.db.Preload("Files").Preload("Events").Order("time DESC")
	if err := req.First(&dep).Error; err != nil {
		return models.Deployment{}, err
	}
	return dep, nil
}
