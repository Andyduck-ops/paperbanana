package sqlite

import (
	"time"

	"gorm.io/gorm"
)

// APIKeyModel is the GORM model for API keys.
type APIKeyModel struct {
	ID           string         `gorm:"primaryKey;type:varchar(36)"`
	ProviderID   string         `gorm:"type:varchar(36);not null;index;constraint:OnDelete:CASCADE"`
	EncryptedKey string         `gorm:"type:text;not null"`
	KeyPrefix    string         `gorm:"type:varchar(10)"`
	KeySuffix    string         `gorm:"type:varchar(10)"`
	IsActive     bool           `gorm:"not null;default:true"`
	LastUsedAt   *time.Time
	CreatedAt    time.Time      `gorm:"not null"`
	UpdatedAt    time.Time      `gorm:"not null"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

// TableName returns the table name for APIKeyModel.
func (APIKeyModel) TableName() string {
	return "api_keys"
}
