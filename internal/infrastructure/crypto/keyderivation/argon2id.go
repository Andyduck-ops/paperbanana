// Package keyderivation provides Argon2id-based key derivation for encryption.
package keyderivation

import (
	"crypto/rand"
	"crypto/sha256"

	"golang.org/x/crypto/argon2"
)

const (
	saltLength  = 16
	memory      = 64 * 1024 // 64 MB
	iterations  = 3
	parallelism = 4
	keyLength   = 32 // 256 bits
)

// Argon2idKDF implements KeyDerivationService using Argon2id.
type Argon2idKDF struct{}

// NewArgon2idKDF creates a new Argon2id key derivation function.
func NewArgon2idKDF() *Argon2idKDF {
	return &Argon2idKDF{}
}

// DeriveKey derives a 256-bit key from a password using Argon2id.
func (k *Argon2idKDF) DeriveKey(password string, salt []byte) []byte {
	return argon2.IDKey(
		[]byte(password),
		salt,
		iterations,
		memory,
		parallelism,
		keyLength,
	)
}

// GenerateSalt generates a cryptographically secure random salt.
func (k *Argon2idKDF) GenerateSalt() ([]byte, error) {
	salt := make([]byte, saltLength)
	_, err := rand.Read(salt)
	return salt, err
}

// DeriveSaltFromKey creates a deterministic salt from a key for consistent key derivation.
// This ensures the same encryption key always produces the same derived key.
func DeriveSaltFromKey(key string) []byte {
	h := sha256.Sum256([]byte("paperbanana-encryption-salt-" + key))
	return h[:16]
}
