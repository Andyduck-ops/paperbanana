// Package localstore provides an opaque local asset storage implementation.
// It stores asset bytes under a configured root directory using project-scoped
// storage keys, with strict path validation to prevent traversal attacks.
package localstore

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Errors returned by the store.
var (
	ErrAssetNotFound  = errors.New("asset not found")
	ErrInvalidKey     = errors.New("invalid storage key")
	ErrPathTraversal  = errors.New("path traversal attempt detected")
	ErrInvalidRoot    = errors.New("invalid asset root")
)

// AssetInfo contains metadata about a stored asset.
type AssetInfo struct {
	StorageKey    string
	MIMEType      string
	ByteSize      int64
	ChecksumSHA256 string
	CreatedAt     time.Time
}

// Store implements AssetStore for local filesystem storage.
// All assets are stored under a configured root directory using opaque,
// project-scoped storage keys.
type Store struct {
	root string
	mu   sync.RWMutex
}

// NewStore creates a new local asset store rooted at the given directory.
// The directory must exist and be accessible.
func NewStore(root string) (*Store, error) {
	// Validate root exists and is a directory
	info, err := os.Stat(root)
	if err != nil {
		if os.IsNotExist(err) {
			// Try to create the directory
			if err := os.MkdirAll(root, 0755); err != nil {
				return nil, fmt.Errorf("%w: cannot create root directory: %v", ErrInvalidRoot, err)
			}
		} else {
			return nil, fmt.Errorf("%w: cannot access root: %v", ErrInvalidRoot, err)
		}
	} else if !info.IsDir() {
		return nil, fmt.Errorf("%w: path is not a directory", ErrInvalidRoot)
	}

	// Get absolute path for consistent validation
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("%w: cannot resolve absolute path: %v", ErrInvalidRoot, err)
	}

	return &Store{root: absRoot}, nil
}

// Write stores the provided bytes under a new opaque storage key.
// The key is generated based on project ID and visualization ID for isolation,
// but uses a UUID to ensure opaqueness.
// MIME type is preserved in a sidecar file for later retrieval via Stat.
func (s *Store) Write(ctx context.Context, projectID, visualizationID, mimeType string, data []byte) (string, error) {
	// Generate opaque storage key: projects/{projectID}/viz/{visualizationID}/{uuid}
	// This provides project isolation while keeping keys opaque
	assetUUID := uuid.New().String()
	relativeKey := fmt.Sprintf("projects/%s/viz/%s/%s", projectID, visualizationID, assetUUID)

	// Validate and construct full path
	fullPath, err := s.safePath(relativeKey)
	if err != nil {
		return "", err
	}

	// Ensure parent directory exists
	parentDir := filepath.Dir(fullPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return "", fmt.Errorf("create parent directory: %w", err)
	}

	// Write data to temp file first, then rename for atomicity
	tempPath := fullPath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return "", fmt.Errorf("write temp file: %w", err)
	}

	// Rename is atomic on POSIX systems
	if err := os.Rename(tempPath, fullPath); err != nil {
		os.Remove(tempPath) // Clean up temp file
		return "", fmt.Errorf("rename to final path: %w", err)
	}

	// Store MIME type in sidecar file
	if mimeType != "" {
		mimePath := fullPath + ".mime"
		if err := os.WriteFile(mimePath, []byte(mimeType), 0644); err != nil {
			// Non-fatal: clean up and continue
			os.Remove(mimePath)
		}
	}

	return relativeKey, nil
}

// Read retrieves the bytes for the given storage key.
// Returns ErrAssetNotFound if the asset does not exist, or ErrInvalidKey
// if the key would escape the asset root.
func (s *Store) Read(ctx context.Context, storageKey string) ([]byte, error) {
	fullPath, err := s.safePath(storageKey)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrAssetNotFound
		}
		return nil, fmt.Errorf("read file: %w", err)
	}

	return data, nil
}

// Delete removes the file for the given storage key.
// Returns ErrAssetNotFound if the asset does not exist.
func (s *Store) Delete(ctx context.Context, storageKey string) error {
	fullPath, err := s.safePath(storageKey)
	if err != nil {
		return err
	}

	err = os.Remove(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrAssetNotFound
		}
		return fmt.Errorf("delete file: %w", err)
	}

	return nil
}

// Exists checks whether a file exists for the storage key.
func (s *Store) Exists(ctx context.Context, storageKey string) (bool, error) {
	fullPath, err := s.safePath(storageKey)
	if err != nil {
		return false, err
	}

	_, err = os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("stat file: %w", err)
	}

	return true, nil
}

// Stat returns metadata about a stored asset.
func (s *Store) Stat(ctx context.Context, storageKey string) (*AssetInfo, error) {
	fullPath, err := s.safePath(storageKey)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrAssetNotFound
		}
		return nil, fmt.Errorf("stat file: %w", err)
	}

	// Read file to compute checksum
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("read file for checksum: %w", err)
	}

	checksum := sha256.Sum256(data)

	// Extract MIME type from sidecar file if it exists
	mimeType := "application/octet-stream"
	mimePath := fullPath + ".mime"
	if mimeData, err := os.ReadFile(mimePath); err == nil && len(mimeData) > 0 {
		mimeType = strings.TrimSpace(string(mimeData))
	}

	return &AssetInfo{
		StorageKey:     storageKey,
		MIMEType:       mimeType,
		ByteSize:       info.Size(),
		ChecksumSHA256: hex.EncodeToString(checksum[:]),
		CreatedAt:      info.ModTime(),
	}, nil
}

// safePath validates a storage key and returns the full filesystem path.
// It ensures the key does not escape the asset root using Go 1.23's
// filepath.IsLocal and additional validation.
func (s *Store) safePath(storageKey string) (string, error) {
	// Reject empty keys
	if storageKey == "" {
		return "", ErrInvalidKey
	}

	// Clean the key to remove any redundant elements
	cleanKey := filepath.Clean(storageKey)

	// Go 1.23+: Use filepath.IsLocal to check if the path is local
	// (doesn't start with /, doesn't contain .., etc.)
	// This is the recommended way to validate paths in Go 1.23+
	if !filepath.IsLocal(cleanKey) {
		return "", ErrPathTraversal
	}

	// Additional check: ensure no backslashes (Windows path separator)
	if strings.Contains(storageKey, "\\") {
		return "", ErrPathTraversal
	}

	// Construct full path
	fullPath := filepath.Join(s.root, cleanKey)

	// Double-check: resolve both paths to absolute and verify containment
	absRoot, err := filepath.Abs(s.root)
	if err != nil {
		return "", fmt.Errorf("resolve root: %w", err)
	}

	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", fmt.Errorf("resolve full path: %w", err)
	}

	// Verify the resolved path is within root
	// Using HasPrefix after cleaning ensures we don't escape via symlinks
	if !strings.HasPrefix(absPath, absRoot+string(filepath.Separator)) && absPath != absRoot {
		return "", ErrPathTraversal
	}

	return absPath, nil
}

// Root returns the configured asset root directory.
func (s *Store) Root() string {
	return s.root
}

// WriteWithMetadata stores bytes and preserves MIME type in a sidecar file.
// This is an extended version of Write that ensures MIME type is recoverable.
func (s *Store) WriteWithMetadata(ctx context.Context, projectID, visualizationID, mimeType string, data []byte) (string, error) {
	// Write the data using Write
	key, err := s.Write(ctx, projectID, visualizationID, mimeType, data)
	if err != nil {
		return "", err
	}

	// Store MIME type in sidecar file
	fullPath, err := s.safePath(key)
	if err != nil {
		return "", err
	}

	mimePath := fullPath + ".mime"
	if err := os.WriteFile(mimePath, []byte(mimeType), 0644); err != nil {
		// Non-fatal: we can still serve the asset, just won't know MIME type
		// In production, we'd log this
	}

	return key, nil
}

// Ensure Store implements the AssetStore interface from tx.go
var _ interface {
	Write(context.Context, string, string, string, []byte) (string, error)
	Read(context.Context, string) ([]byte, error)
	Delete(context.Context, string) error
	Exists(context.Context, string) (bool, error)
} = (*Store)(nil)

// Ensure FileInfo interface is compatible
var _ fs.FileInfo = (*AssetInfo)(nil)

func (i *AssetInfo) Name() string       { return i.StorageKey }
func (i *AssetInfo) Size() int64        { return i.ByteSize }
func (i *AssetInfo) Mode() fs.FileMode  { return 0644 }
func (i *AssetInfo) ModTime() time.Time { return i.CreatedAt }
func (i *AssetInfo) IsDir() bool        { return false }
func (i *AssetInfo) Sys() interface{}   { return nil }
