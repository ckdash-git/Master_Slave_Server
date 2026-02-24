package repository

import (
	"time"

	"github.com/cachatto/master-slave-server/internal/models"
	"gorm.io/gorm"
)

// OTCRepository handles database operations for one-time codes.
type OTCRepository struct {
	db *gorm.DB
}

// NewOTCRepository creates a new OTCRepository.
func NewOTCRepository(db *gorm.DB) *OTCRepository {
	return &OTCRepository{db: db}
}

// Create stores a new one-time code in the database.
func (r *OTCRepository) Create(otc *models.OneTimeCode) error {
	return r.db.Create(otc).Error
}

// FindByCode retrieves a one-time code by its code string.
func (r *OTCRepository) FindByCode(code string) (*models.OneTimeCode, error) {
	var otc models.OneTimeCode
	result := r.db.Where("code = ?", code).First(&otc)
	if result.Error != nil {
		return nil, result.Error
	}
	return &otc, nil
}

// MarkClaimed marks a one-time code as claimed.
func (r *OTCRepository) MarkClaimed(otc *models.OneTimeCode) error {
	return r.db.Model(otc).Update("claimed", true).Error
}

// CleanExpired removes all expired one-time codes from the database.
func (r *OTCRepository) CleanExpired() error {
	return r.db.Where("expires_at < ? OR claimed = ?", time.Now(), true).
		Delete(&models.OneTimeCode{}).Error
}
