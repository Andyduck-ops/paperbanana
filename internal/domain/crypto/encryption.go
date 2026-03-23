// Package crypto provides encryption and key derivation interfaces for secure data storage.
package crypto

import "context"

// EncryptionService provides secure encryption/decryption for sensitive data.
type EncryptionService interface {
	// Encrypt encrypts plaintext and returns base64-encoded ciphertext.
	Encrypt(ctx context.Context, plaintext string) (string, error)

	// Decrypt decrypts base64-encoded ciphertext and returns plaintext.
	Decrypt(ctx context.Context, ciphertext string) (string, error)

	// Mask returns a masked version of the plaintext for display.
	// Example: "sk-abc123xyz" -> "sk-abc****xyz"
	Mask(plaintext string) string
}

// KeyDerivationService derives encryption keys from passwords/secrets.
type KeyDerivationService interface {
	// DeriveKey derives a 256-bit key from a password using Argon2id.
	DeriveKey(password string, salt []byte) []byte

	// GenerateSalt generates a cryptographically secure random salt.
	GenerateSalt() ([]byte, error)
}
