package aesgcm

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptDecrypt(t *testing.T) {
	os.Setenv("PAPERBANANA_ENCRYPTION_KEY", "dGVzdC1lbmNyeXB0aW9uLWtleS0zMi1ieXRlcy1sb25n")
	defer os.Unsetenv("PAPERBANANA_ENCRYPTION_KEY")

	svc, err := NewService()
	require.NoError(t, err)

	tests := []struct {
		name      string
		plaintext string
	}{
		{"simple", "sk-abc123"},
		{"long key", "sk-proj-abcdefghijklmnopqrstuvwxyz1234567890"},
		{"empty", ""},
		{"special chars", "sk-!@#$%^&*()_+-=[]{}|;':\",./<>?"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := svc.Encrypt(context.Background(), tt.plaintext)
			require.NoError(t, err)
			assert.NotEqual(t, tt.plaintext, encrypted)

			decrypted, err := svc.Decrypt(context.Background(), encrypted)
			require.NoError(t, err)
			assert.Equal(t, tt.plaintext, decrypted)
		})
	}
}

func TestMask(t *testing.T) {
	svc := &Service{}

	tests := []struct {
		input    string
		expected string
	}{
		{"sk-abc123xyz", "sk-abc****3xyz"},
		{"short", "****"},
		{"sk-proj-verylongapikey123", "sk-pro****y123"},
		{"AIzaSyABCDEFGHIJKLMNOPQRSTUVWXYZ", "AIzaSy****WXYZ"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, svc.Mask(tt.input))
		})
	}
}

func TestMaskAPIKey(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"sk-abc123xyz", "sk-abc****3xyz"},
		{"short", "****"},
		{"1234567890", "****"},
		{"12345678901", "123456****8901"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, MaskAPIKey(tt.input))
		})
	}
}

func TestNonDeterministicEncryption(t *testing.T) {
	os.Setenv("PAPERBANANA_ENCRYPTION_KEY", "dGVzdC1lbmNyeXB0aW9uLWtleS0zMi1ieXRlcy1sb25n")
	defer os.Unsetenv("PAPERBANANA_ENCRYPTION_KEY")

	svc, err := NewService()
	require.NoError(t, err)

	// Same plaintext should produce different ciphertext (due to random nonce)
	encrypted1, _ := svc.Encrypt(context.Background(), "test-key")
	encrypted2, _ := svc.Encrypt(context.Background(), "test-key")

	assert.NotEqual(t, encrypted1, encrypted2, "encryption should be non-deterministic")

	// But both should decrypt to the same value
	decrypted1, _ := svc.Decrypt(context.Background(), encrypted1)
	decrypted2, _ := svc.Decrypt(context.Background(), encrypted2)
	assert.Equal(t, decrypted1, decrypted2)
}

func TestDecryptInvalidBase64(t *testing.T) {
	os.Setenv("PAPERBANANA_ENCRYPTION_KEY", "dGVzdC1lbmNyeXB0aW9uLWtleS0zMi1ieXRlcy1sb25n")
	defer os.Unsetenv("PAPERBANANA_ENCRYPTION_KEY")

	svc, err := NewService()
	require.NoError(t, err)

	_, err = svc.Decrypt(context.Background(), "not-valid-base64!!!")
	assert.Error(t, err)
}

func TestDecryptTooShort(t *testing.T) {
	os.Setenv("PAPERBANANA_ENCRYPTION_KEY", "dGVzdC1lbmNyeXB0aW9uLWtleS0zMi1ieXRlcy1sb25n")
	defer os.Unsetenv("PAPERBANANA_ENCRYPTION_KEY")

	svc, err := NewService()
	require.NoError(t, err)

	// Valid base64 but too short to be valid ciphertext
	_, err = svc.Decrypt(context.Background(), "YWJj") // "abc"
	assert.Error(t, err)
}
