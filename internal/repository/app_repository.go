package repository

import (
	"github.com/cachatto/master-slave-server/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AppRepository handles database operations for the app registry.
type AppRepository struct {
	db *gorm.DB
}

// NewAppRepository creates a new AppRepository.
func NewAppRepository(db *gorm.DB) *AppRepository {
	return &AppRepository{db: db}
}

// GetPermittedApps returns all apps a user is authorized to access.
func (r *AppRepository) GetPermittedApps(userID uuid.UUID) ([]models.App, error) {
	var apps []models.App
	result := r.db.
		Joins("JOIN user_app_permissions ON user_app_permissions.app_id = app_registry.id").
		Where("user_app_permissions.user_id = ?", userID).
		Find(&apps)
	if result.Error != nil {
		return nil, result.Error
	}
	return apps, nil
}

// FindByID retrieves an app by its UUID.
func (r *AppRepository) FindByID(id uuid.UUID) (*models.App, error) {
	var app models.App
	result := r.db.First(&app, "id = ?", id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &app, nil
}

// FindByPackageID retrieves an app by its package identifier.
func (r *AppRepository) FindByPackageID(packageID string) (*models.App, error) {
	var app models.App
	result := r.db.Where("package_id = ?", packageID).First(&app)
	if result.Error != nil {
		return nil, result.Error
	}
	return &app, nil
}

// HasPermission checks if a user has permission to use a specific app.
func (r *AppRepository) HasPermission(userID, appID uuid.UUID) (bool, error) {
	var count int64
	result := r.db.Model(&models.UserAppPermission{}).
		Where("user_id = ? AND app_id = ?", userID, appID).
		Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}
