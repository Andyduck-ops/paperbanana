package sqlite

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	domainconfig "github.com/paperbanana/paperbanana/internal/domain/config"
	"gorm.io/gorm"
)

// ProviderRepository implements ProviderRepository using SQLite/GORM.
type ProviderRepository struct {
	db *gorm.DB
}

// NewProviderRepository creates a new provider repository.
func NewProviderRepository(db *gorm.DB) *ProviderRepository {
	return &ProviderRepository{db: db}
}

// Create creates a new provider.
func (r *ProviderRepository) Create(provider *domainconfig.Provider) error {
	if provider.ID == "" {
		provider.ID = uuid.New().String()
	}
	now := time.Now()
	provider.CreatedAt = now
	provider.UpdatedAt = now

	model := r.toModel(provider)
	if err := r.db.Create(model).Error; err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}
	return nil
}

// GetByID gets a provider by ID.
func (r *ProviderRepository) GetByID(id string) (*domainconfig.Provider, error) {
	var model ProviderModel
	if err := r.db.First(&model, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}
	return r.toDomain(&model), nil
}

// GetByName gets a provider by name.
func (r *ProviderRepository) GetByName(name string) (*domainconfig.Provider, error) {
	var model ProviderModel
	if err := r.db.Where("name = ?", name).First(&model).Error; err != nil {
		return nil, fmt.Errorf("failed to get provider by name: %w", err)
	}
	return r.toDomain(&model), nil
}

// List lists all providers.
func (r *ProviderRepository) List() ([]*domainconfig.Provider, error) {
	var models []ProviderModel
	if err := r.db.Order("is_default DESC, name ASC").Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to list providers: %w", err)
	}

	providers := make([]*domainconfig.Provider, len(models))
	for i, m := range models {
		providers[i] = r.toDomain(&m)
	}
	return providers, nil
}

// Update updates a provider.
func (r *ProviderRepository) Update(provider *domainconfig.Provider) error {
	provider.UpdatedAt = time.Now()
	model := r.toModel(provider)
	if err := r.db.Save(model).Error; err != nil {
		return fmt.Errorf("failed to update provider: %w", err)
	}
	return nil
}

// Delete deletes a provider.
func (r *ProviderRepository) Delete(id string) error {
	if err := r.db.Delete(&ProviderModel{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete provider: %w", err)
	}
	return nil
}

// SetDefault sets the default provider.
func (r *ProviderRepository) SetDefault(id string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Unset current default
		if err := tx.Model(&ProviderModel{}).Where("is_default = ?", true).Update("is_default", false).Error; err != nil {
			return err
		}
		// Set new default
		if err := tx.Model(&ProviderModel{}).Where("id = ?", id).Update("is_default", true).Error; err != nil {
			return err
		}
		return nil
	})
}

// GetDefault gets the default provider.
func (r *ProviderRepository) GetDefault() (*domainconfig.Provider, error) {
	var model ProviderModel
	if err := r.db.Where("is_default = ?", true).First(&model).Error; err != nil {
		return nil, fmt.Errorf("failed to get default provider: %w", err)
	}
	return r.toDomain(&model), nil
}

// ListEnabled lists all enabled providers.
func (r *ProviderRepository) ListEnabled() ([]*domainconfig.Provider, error) {
	var models []ProviderModel
	if err := r.db.Where("enabled = ?", true).Order("is_default DESC, name ASC").Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to list enabled providers: %w", err)
	}

	providers := make([]*domainconfig.Provider, len(models))
	for i, m := range models {
		providers[i] = r.toDomain(&m)
	}
	return providers, nil
}

// InitializeSystemProviders initializes system provider presets.
func (r *ProviderRepository) InitializeSystemProviders() error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, preset := range domainconfig.SystemProviderPresets() {
			// Check if already exists (including soft-deleted records)
			var count int64
			if err := tx.Unscoped().Model(&ProviderModel{}).Where("name = ?", preset.Name).Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				continue // Already exists
			}

			// Create system provider
			models := make(ModelInfoList, len(preset.DefaultModels))
			for i, m := range preset.DefaultModels {
				models[i] = ModelInfo{
					ID:             m.ID,
					Name:           m.Name,
					MaxTokens:      m.MaxTokens,
					SupportsVision: m.SupportsVision,
					Enabled:        m.Enabled,
				}
			}

			now := time.Now()
			provider := &ProviderModel{
				ID:          uuid.New().String(),
				Type:        string(preset.Type),
				Name:        preset.Name,
				DisplayName: preset.DisplayName,
				APIHost:     preset.APIHost,
				Models:      models,
				Enabled:     false, // Disabled by default, user needs to enable
				IsSystem:    true,
				IsDefault:   false,
				TimeoutMs:   60000,
				CreatedAt:   now,
				UpdatedAt:   now,
			}

			if err := tx.Create(provider).Error; err != nil {
				return fmt.Errorf("failed to create system provider %s: %w", preset.Name, err)
			}
		}
		return nil
	})
}

func (r *ProviderRepository) toModel(p *domainconfig.Provider) *ProviderModel {
	m := &ProviderModel{}
	m.FromDomain(p)
	return m
}

func (r *ProviderRepository) toDomain(m *ProviderModel) *domainconfig.Provider {
	return m.ToDomain()
}
