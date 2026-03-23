package sqlite

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	domainconfig "github.com/paperbanana/paperbanana/internal/domain/config"
	"gorm.io/gorm"
)

// ModelInfo represents a model configuration.
type ModelInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	MaxTokens   int    `json:"max_tokens,omitempty"`
	SupportsVision bool  `json:"supports_vision,omitempty"`
	Enabled     bool   `json:"enabled"`
}

// ModelInfoList is a slice of ModelInfo that implements sql.Scanner and driver.Valuer.
type ModelInfoList []ModelInfo

// Scan implements sql.Scanner.
func (m *ModelInfoList) Scan(value interface{}) error {
	if value == nil {
		*m = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, m)
}

// Value implements driver.Valuer.
func (m ModelInfoList) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

// ProviderModel is the GORM model for providers.
type ProviderModel struct {
	ID             string         `gorm:"primaryKey;type:varchar(36)"`
	Type           string         `gorm:"type:varchar(32);not null;default:'openai-compatible'"`
	Name           string         `gorm:"type:varchar(64);uniqueIndex;not null"`
	DisplayName    string         `gorm:"type:varchar(128)"`
	APIHost        string         `gorm:"type:varchar(512)"`
	APIKey         string         `gorm:"type:varchar(512)"` // Encrypted or reference to API key
	Models         ModelInfoList  `gorm:"type:json"`
	Enabled        bool           `gorm:"not null;default:false"`
	IsSystem       bool           `gorm:"not null;default:false"`
	IsDefault      bool           `gorm:"not null;default:false"`
	TimeoutMs      int            `gorm:"not null;default:60000"`
	QueryModel     string         `gorm:"type:varchar(128)"`
	GenModel       string         `gorm:"type:varchar(128)"`
	CreatedAt      time.Time      `gorm:"not null"`
	UpdatedAt      time.Time      `gorm:"not null"`
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}

// TableName returns the table name for ProviderModel.
func (ProviderModel) TableName() string {
	return "providers"
}

// ToDomain converts the model to a domain entity.
func (m *ProviderModel) ToDomain() *domainconfig.Provider {
	models := make([]domainconfig.ModelInfo, len(m.Models))
	for i, model := range m.Models {
		models[i] = domainconfig.ModelInfo{
			ID:            model.ID,
			Name:          model.Name,
			MaxTokens:     model.MaxTokens,
			SupportsVision: model.SupportsVision,
			Enabled:       model.Enabled,
		}
	}
	return &domainconfig.Provider{
		ID:          m.ID,
		Type:        domainconfig.ProviderType(m.Type),
		Name:        m.Name,
		DisplayName: m.DisplayName,
		APIHost:     m.APIHost,
		APIKey:      m.APIKey,
		Models:      models,
		Enabled:     m.Enabled,
		IsSystem:    m.IsSystem,
		IsDefault:   m.IsDefault,
		TimeoutMs:   m.TimeoutMs,
		QueryModel:  m.QueryModel,
		GenModel:    m.GenModel,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// FromDomain converts a domain entity to a model.
func (m *ProviderModel) FromDomain(p *domainconfig.Provider) {
	m.ID = p.ID
	m.Type = string(p.Type)
	m.Name = p.Name
	m.DisplayName = p.DisplayName
	m.APIHost = p.APIHost
	m.APIKey = p.APIKey
	m.Enabled = p.Enabled
	m.IsSystem = p.IsSystem
	m.IsDefault = p.IsDefault
	m.TimeoutMs = p.TimeoutMs
	m.QueryModel = p.QueryModel
	m.GenModel = p.GenModel
	m.CreatedAt = p.CreatedAt
	m.UpdatedAt = p.UpdatedAt

	models := make(ModelInfoList, len(p.Models))
	for i, model := range p.Models {
		models[i] = ModelInfo{
			ID:            model.ID,
			Name:          model.Name,
			MaxTokens:     model.MaxTokens,
			SupportsVision: model.SupportsVision,
			Enabled:       model.Enabled,
		}
	}
	m.Models = models
}
