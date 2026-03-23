package config

import "time"

// APIKey represents an encrypted API key for a provider.
type APIKey struct {
	ID           string     `json:"id"`
	ProviderID   string     `json:"provider_id"`
	EncryptedKey string     `json:"-"` // Never expose in JSON
	KeyPrefix    string     `json:"key_prefix"`
	KeySuffix    string     `json:"key_suffix"`
	IsActive     bool       `json:"is_active"`
	LastUsedAt   *time.Time `json:"last_used_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// MaskedKey returns the masked representation for display.
func (k *APIKey) MaskedKey() string {
	if k.KeyPrefix == "" || k.KeySuffix == "" {
		return "****"
	}
	return k.KeyPrefix + "****" + k.KeySuffix
}

// APIKeyRepository defines the interface for API key persistence.
type APIKeyRepository interface {
	Create(ctx interface{}, key *APIKey, plaintext string) error
	GetByID(id string) (*APIKey, error)
	GetDecrypted(ctx interface{}, id string) (string, error)
	ListByProvider(providerID string) ([]*APIKey, error)
	GetActiveKeys(providerID string) ([]*APIKey, error)
	GetNextKey(ctx interface{}, providerID string) (*APIKey, string, error)
	Update(key *APIKey) error
	Delete(id string) error
	MarkUsed(id string) error
}
