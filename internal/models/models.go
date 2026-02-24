package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a registered user in the system.
type User struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Email        string         `gorm:"uniqueIndex;not null;size:255" json:"email"`
	PasswordHash string         `gorm:"not null" json:"-"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// App represents a registered application (Slave app) in the system.
type App struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AppName        string    `gorm:"not null;size:255" json:"app_name"`
	PackageID      string    `gorm:"uniqueIndex;not null;size:255" json:"package_id"`
	DeepLinkScheme string    `gorm:"not null;size:255" json:"deep_link_scheme"`
	CreatedAt      time.Time `json:"created_at"`
}

// TableName overrides the default table name for App.
func (App) TableName() string {
	return "app_registry"
}

// UserAppPermission links a user to a slave app they are authorized to use.
type UserAppPermission struct {
	ID     uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	AppID  uuid.UUID `gorm:"type:uuid;not null;index" json:"app_id"`
	User   User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	App    App       `gorm:"foreignKey:AppID;constraint:OnDelete:CASCADE" json:"-"`
}

// TableName overrides the default table name.
func (UserAppPermission) TableName() string {
	return "user_app_permissions"
}

// OneTimeCode represents a short-lived code for the OTC handshake.
type OneTimeCode struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	AppID     uuid.UUID `gorm:"type:uuid;not null" json:"app_id"`
	Code      string    `gorm:"uniqueIndex;not null;size:32" json:"code"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	Claimed   bool      `gorm:"default:false" json:"claimed"`
	User      User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	App       App       `gorm:"foreignKey:AppID;constraint:OnDelete:CASCADE" json:"-"`
}

// TableName overrides the default table name.
func (OneTimeCode) TableName() string {
	return "one_time_codes"
}
