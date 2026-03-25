// Package aesgcm provides AES-256-GCM encryption service implementation.
package aesgcm

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/paperbanana/paperbanana/internal/infrastructure/crypto/keyderivation"
)

const nonceLength = 12
const defaultDevKeyPath = ".paperbanana/dev-encryption.key"

// Service implements EncryptionService using AES-256-GCM.
type Service struct {
	gcm cipher.AEAD
	kdf *keyderivation.Argon2idKDF
}

// NewService creates a new AES-256-GCM encryption service.
// It reads the encryption key from PAPERBANANA_ENCRYPTION_KEY environment variable.
// If not set, it loads or creates a persisted development key under .paperbanana/.
func NewService() (*Service, error) {
	encKey := os.Getenv("PAPERBANANA_ENCRYPTION_KEY")
	if encKey == "" {
		var err error
		encKey, err = loadOrCreateDevKey()
		if err != nil {
			return nil, err
		}
	}

	kdf := keyderivation.NewArgon2idKDF()
	salt := keyderivation.DeriveSaltFromKey(encKey)
	key := kdf.DeriveKey(encKey, salt)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	return &Service{gcm: gcm, kdf: kdf}, nil
}

// Encrypt encrypts plaintext and returns base64-encoded ciphertext.
func (s *Service) Encrypt(ctx context.Context, plaintext string) (string, error) {
	nonce := make([]byte, nonceLength)
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := s.gcm.Seal(nil, nonce, []byte(plaintext), nil)

	// Prepend nonce to ciphertext
	result := append(nonce, ciphertext...)
	return base64.StdEncoding.EncodeToString(result), nil
}

// Decrypt decrypts base64-encoded ciphertext and returns plaintext.
func (s *Service) Decrypt(ctx context.Context, ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	if len(data) < nonceLength+s.gcm.Overhead() {
		return "", errors.New("ciphertext too short")
	}

	nonce := data[:nonceLength]
	ct := data[nonceLength:]

	plaintext, err := s.gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// Mask returns a masked version of the plaintext for display.
// Shows first 6 and last 4 characters with **** in between.
func (s *Service) Mask(plaintext string) string {
	if len(plaintext) <= 10 {
		return "****"
	}
	return plaintext[:6] + "****" + plaintext[len(plaintext)-4:]
}

// MaskAPIKey is a standalone function for masking API keys.
func MaskAPIKey(key string) string {
	if len(key) <= 10 {
		return "****"
	}
	return key[:6] + "****" + key[len(key)-4:]
}

func generateDevKey() string {
	key := make([]byte, 32)
	_, _ = rand.Read(key)
	return base64.StdEncoding.EncodeToString(key)
}

func loadOrCreateDevKey() (string, error) {
	path := os.Getenv("PAPERBANANA_ENCRYPTION_KEY_FILE")
	if strings.TrimSpace(path) == "" {
		path = defaultDevKeyPath
	}

	if data, err := os.ReadFile(path); err == nil {
		key := strings.TrimSpace(string(data))
		if key != "" {
			fmt.Printf("WARNING: PAPERBANANA_ENCRYPTION_KEY not set, using persisted development key at %s\n", path)
			return key, nil
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("failed to read development encryption key %s: %w", path, err)
	}

	key := generateDevKey()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return "", fmt.Errorf("failed to create development key directory for %s: %w", path, err)
	}
	if err := os.WriteFile(path, []byte(key+"\n"), 0o600); err != nil {
		return "", fmt.Errorf("failed to persist development encryption key %s: %w", path, err)
	}

	fmt.Printf("WARNING: PAPERBANANA_ENCRYPTION_KEY not set, created persisted development key at %s\n", path)
	return key, nil
}
