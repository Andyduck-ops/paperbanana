package sqlite

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// LocalAssetStore implements persistence.AssetStore using the local filesystem.
// Files are stored under a root directory with opaque storage keys.
type LocalAssetStore struct {
	root string
}

// NewLocalAssetStore creates a new LocalAssetStore.
func NewLocalAssetStore(root string) *LocalAssetStore {
	return &LocalAssetStore{root: root}
}

// Write stores the provided bytes under the given storage key.
func (s *LocalAssetStore) Write(ctx context.Context, storageKey string, data []byte) error {
	fullPath := filepath.Join(s.root, storageKey)

	// Ensure parent directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create directory %s: %w", dir, err)
	}

	// Write file (create or truncate)
	if err := os.WriteFile(fullPath, data, 0o644); err != nil {
		return fmt.Errorf("write file %s: %w", storageKey, err)
	}

	return nil
}

// Read retrieves the bytes for the given storage key.
func (s *LocalAssetStore) Read(ctx context.Context, storageKey string) ([]byte, error) {
	fullPath := filepath.Join(s.root, storageKey)

	data, err := os.ReadFile(fullPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("asset not found: %s", storageKey)
		}
		return nil, fmt.Errorf("read file %s: %w", storageKey, err)
	}

	return data, nil
}

// Delete removes the file for the given storage key.
func (s *LocalAssetStore) Delete(ctx context.Context, storageKey string) error {
	fullPath := filepath.Join(s.root, storageKey)

	if err := os.Remove(fullPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// Already deleted, not an error
			return nil
		}
		return fmt.Errorf("delete file %s: %w", storageKey, err)
	}

	return nil
}

// Exists checks whether a file exists for the storage key.
func (s *LocalAssetStore) Exists(ctx context.Context, storageKey string) (bool, error) {
	fullPath := filepath.Join(s.root, storageKey)

	_, err := os.Stat(fullPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, fmt.Errorf("stat file %s: %w", storageKey, err)
	}

	return true, nil
}
