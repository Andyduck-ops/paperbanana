package sqlite

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	domaincrypto "github.com/paperbanana/paperbanana/internal/domain/crypto"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type apiKeyTestEncryption struct {
	decrypt map[string]string
	fail    map[string]error
}

func (e apiKeyTestEncryption) Encrypt(ctx context.Context, plaintext string) (string, error) {
	return plaintext, nil
}

func (e apiKeyTestEncryption) Decrypt(ctx context.Context, ciphertext string) (string, error) {
	if err, ok := e.fail[ciphertext]; ok {
		return "", err
	}
	if value, ok := e.decrypt[ciphertext]; ok {
		return value, nil
	}
	return "", errors.New("missing ciphertext")
}

func (e apiKeyTestEncryption) Mask(plaintext string) string {
	return plaintext
}

var _ domaincrypto.EncryptionService = (*apiKeyTestEncryption)(nil)

func TestAPIKeyRepositoryGetNextKey_SkipsUndecryptableKeys(t *testing.T) {
	t.Parallel()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&APIKeyModel{}))

	now := time.Now().UTC().Add(-time.Hour)
	require.NoError(t, db.Create(&APIKeyModel{
		ID:           "key-bad",
		ProviderID:   "provider-1",
		EncryptedKey: "bad-cipher",
		KeyPrefix:    "sk-bad",
		KeySuffix:    "fail",
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}).Error)
	require.NoError(t, db.Create(&APIKeyModel{
		ID:           "key-good",
		ProviderID:   "provider-1",
		EncryptedKey: "good-cipher",
		KeyPrefix:    "sk-goo",
		KeySuffix:    "pass",
		IsActive:     true,
		CreatedAt:    now.Add(time.Minute),
		UpdatedAt:    now.Add(time.Minute),
	}).Error)

	repo := NewAPIKeyRepository(db, apiKeyTestEncryption{
		decrypt: map[string]string{
			"good-cipher": "sk-test-good",
		},
		fail: map[string]error{
			"bad-cipher": errors.New("cipher mismatch"),
		},
	})

	key, plaintext, err := repo.GetNextKey(context.Background(), "provider-1")
	require.NoError(t, err)
	require.NotNil(t, key)
	require.Equal(t, "key-good", key.ID)
	require.Equal(t, "sk-test-good", plaintext)
}
