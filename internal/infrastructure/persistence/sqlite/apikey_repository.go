package sqlite

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	domainconfig "github.com/paperbanana/paperbanana/internal/domain/config"
	domaincrypto "github.com/paperbanana/paperbanana/internal/domain/crypto"
	"gorm.io/gorm"
)

// APIKeyRepository implements APIKeyRepository using SQLite/GORM.
type APIKeyRepository struct {
	db      *gorm.DB
	encrypt domaincrypto.EncryptionService
}

// NewAPIKeyRepository creates a new API key repository.
func NewAPIKeyRepository(db *gorm.DB, encrypt domaincrypto.EncryptionService) *APIKeyRepository {
	return &APIKeyRepository{db: db, encrypt: encrypt}
}

// Create creates a new API key with encryption.
func (r *APIKeyRepository) Create(ctx interface{}, key *domainconfig.APIKey, plaintext string) error {
	if key.ID == "" {
		key.ID = uuid.New().String()
	}

	// Get context for encryption
	var encCtx context.Context
	switch v := ctx.(type) {
	case context.Context:
		encCtx = v
	default:
		encCtx = context.Background()
	}

	// Encrypt the plaintext key
	encrypted, err := r.encrypt.Encrypt(encCtx, plaintext)
	if err != nil {
		return fmt.Errorf("failed to encrypt API key: %w", err)
	}
	key.EncryptedKey = encrypted

	// Store prefix and suffix for display
	if len(plaintext) >= 10 {
		key.KeyPrefix = plaintext[:6]
		key.KeySuffix = plaintext[len(plaintext)-4:]
	}

	now := time.Now()
	key.CreatedAt = now
	key.UpdatedAt = now

	model := r.toModel(key)
	if err := r.db.Create(model).Error; err != nil {
		return fmt.Errorf("failed to create API key: %w", err)
	}
	return nil
}

// GetByID gets an API key by ID.
func (r *APIKeyRepository) GetByID(id string) (*domainconfig.APIKey, error) {
	var model APIKeyModel
	if err := r.db.First(&model, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}
	return r.toDomain(&model), nil
}

// GetDecrypted gets and decrypts an API key.
func (r *APIKeyRepository) GetDecrypted(ctx interface{}, id string) (string, error) {
	var model APIKeyModel
	if err := r.db.First(&model, "id = ?", id).Error; err != nil {
		return "", fmt.Errorf("failed to get API key: %w", err)
	}

	var encCtx context.Context
	switch v := ctx.(type) {
	case context.Context:
		encCtx = v
	default:
		encCtx = context.Background()
	}

	return r.encrypt.Decrypt(encCtx, model.EncryptedKey)
}

// ListByProvider lists all API keys for a provider.
func (r *APIKeyRepository) ListByProvider(providerID string) ([]*domainconfig.APIKey, error) {
	var models []APIKeyModel
	if err := r.db.Where("provider_id = ?", providerID).Order("created_at ASC").Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to list API keys: %w", err)
	}

	keys := make([]*domainconfig.APIKey, len(models))
	for i, m := range models {
		keys[i] = r.toDomain(&m)
	}
	return keys, nil
}

// GetActiveKeys gets all active API keys for a provider.
func (r *APIKeyRepository) GetActiveKeys(providerID string) ([]*domainconfig.APIKey, error) {
	var models []APIKeyModel
	if err := r.db.Where("provider_id = ? AND is_active = ?", providerID, true).
		Order("last_used_at ASC NULLS FIRST, created_at ASC").
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get active API keys: %w", err)
	}

	keys := make([]*domainconfig.APIKey, len(models))
	for i, m := range models {
		keys[i] = r.toDomain(&m)
	}
	return keys, nil
}

// GetNextKey gets the next API key for rotation.
func (r *APIKeyRepository) GetNextKey(ctx interface{}, providerID string) (*domainconfig.APIKey, string, error) {
	keys, err := r.GetActiveKeys(providerID)
	if err != nil {
		return nil, "", err
	}
	if len(keys) == 0 {
		return nil, "", fmt.Errorf("no active API keys for provider %s", providerID)
	}

	// Get the first key (least recently used)
	key := keys[0]

	var encCtx context.Context
	switch v := ctx.(type) {
	case context.Context:
		encCtx = v
	default:
		encCtx = context.Background()
	}

	// Decrypt the key
	plaintext, err := r.encrypt.Decrypt(encCtx, key.EncryptedKey)
	if err != nil {
		return nil, "", fmt.Errorf("failed to decrypt API key: %w", err)
	}

	// Mark as used
	_ = r.MarkUsed(key.ID) // Ignore error, not critical

	return key, plaintext, nil
}

// Update updates an API key.
func (r *APIKeyRepository) Update(key *domainconfig.APIKey) error {
	key.UpdatedAt = time.Now()
	model := r.toModel(key)
	if err := r.db.Save(model).Error; err != nil {
		return fmt.Errorf("failed to update API key: %w", err)
	}
	return nil
}

// Delete deletes an API key.
func (r *APIKeyRepository) Delete(id string) error {
	if err := r.db.Delete(&APIKeyModel{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete API key: %w", err)
	}
	return nil
}

// MarkUsed marks an API key as used.
func (r *APIKeyRepository) MarkUsed(id string) error {
	now := time.Now()
	if err := r.db.Model(&APIKeyModel{}).Where("id = ?", id).Update("last_used_at", now).Error; err != nil {
		return fmt.Errorf("failed to mark API key as used: %w", err)
	}
	return nil
}

func (r *APIKeyRepository) toModel(k *domainconfig.APIKey) *APIKeyModel {
	return &APIKeyModel{
		ID:           k.ID,
		ProviderID:   k.ProviderID,
		EncryptedKey: k.EncryptedKey,
		KeyPrefix:    k.KeyPrefix,
		KeySuffix:    k.KeySuffix,
		IsActive:     k.IsActive,
		LastUsedAt:   k.LastUsedAt,
		CreatedAt:    k.CreatedAt,
		UpdatedAt:    k.UpdatedAt,
	}
}

func (r *APIKeyRepository) toDomain(m *APIKeyModel) *domainconfig.APIKey {
	return &domainconfig.APIKey{
		ID:           m.ID,
		ProviderID:   m.ProviderID,
		EncryptedKey: m.EncryptedKey,
		KeyPrefix:    m.KeyPrefix,
		KeySuffix:    m.KeySuffix,
		IsActive:     m.IsActive,
		LastUsedAt:   m.LastUsedAt,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}
