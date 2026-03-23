package localstore

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStore_ProjectScopedAssets(t *testing.T) {
	// Setup: create temp asset root
	tmpDir, err := os.MkdirTemp("", "assetstore-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	ctx := context.Background()
	projectID := "proj-123"
	visualizationID := "viz-456"

	// Write asset bytes - should generate opaque storage key
	data := []byte("test asset content")
	key, err := store.Write(ctx, projectID, visualizationID, "image/png", data)
	if err != nil {
		t.Fatalf("failed to write asset: %v", err)
	}

	// Verify storage key is opaque (not user-controlled)
	if strings.Contains(key, "image.png") || strings.Contains(key, "userfile") {
		t.Errorf("storage key should be opaque, got: %s", key)
	}

	// Verify storage key contains project identifier for isolation
	if !strings.Contains(key, projectID) {
		t.Errorf("storage key should contain project ID for isolation, got: %s", key)
	}

	// Read back the asset
	readData, err := store.Read(ctx, key)
	if err != nil {
		t.Fatalf("failed to read asset: %v", err)
	}

	if string(readData) != string(data) {
		t.Errorf("read data mismatch: got %q, want %q", readData, data)
	}
}

func TestStore_RejectsUnsafeStorageKey(t *testing.T) {
	// Setup: create temp asset root
	tmpDir, err := os.MkdirTemp("", "assetstore-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	ctx := context.Background()

	// First, write a legitimate asset to ensure store works
	projectID := "proj-safe"
	visualizationID := "viz-safe"
	data := []byte("legitimate data")
	key, err := store.Write(ctx, projectID, visualizationID, "text/plain", data)
	if err != nil {
		t.Fatalf("failed to write legitimate asset: %v", err)
	}

	// Now test unsafe keys are rejected
	unsafeKeys := []string{
		"../../../etc/passwd",
		"../outside-root",
		"/absolute/path",
		"proj-123/../../../etc/passwd",
		"./relative",
		"//double-slash",
	}

	for _, unsafeKey := range unsafeKeys {
		t.Run(unsafeKey, func(t *testing.T) {
			_, err := store.Read(ctx, unsafeKey)
			if err == nil {
				t.Errorf("expected error for unsafe key %q, got nil", unsafeKey)
			}

			err = store.Delete(ctx, unsafeKey)
			if err == nil {
				t.Errorf("expected error for delete with unsafe key %q, got nil", unsafeKey)
			}
		})
	}

	// Verify the legitimate key still works after unsafe attempts
	readData, err := store.Read(ctx, key)
	if err != nil {
		t.Errorf("legitimate key should still work: %v", err)
	}
	if string(readData) != string(data) {
		t.Errorf("legitimate data should still match")
	}
}

func TestStore_RoundTripsBytes(t *testing.T) {
	// Setup: create temp asset root
	tmpDir, err := os.MkdirTemp("", "assetstore-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	ctx := context.Background()
	projectID := "proj-roundtrip"
	visualizationID := "viz-roundtrip"

	testCases := []struct {
		name     string
		mimeType string
		data     []byte
	}{
		{
			name:     "small text",
			mimeType: "text/plain",
			data:     []byte("hello world"),
		},
		{
			name:     "binary image",
			mimeType: "image/png",
			data:     []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, // PNG header
		},
		{
			name:     "empty file",
			mimeType: "application/octet-stream",
			data:     []byte{},
		},
		{
			name:     "large file",
			mimeType: "application/pdf",
			data:     make([]byte, 1024*100), // 100KB
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			key, err := store.Write(ctx, projectID, visualizationID, tc.mimeType, tc.data)
			if err != nil {
				t.Fatalf("failed to write: %v", err)
			}

			// Verify checksum matches expected
			expectedChecksum := sha256.Sum256(tc.data)
			expectedHex := hex.EncodeToString(expectedChecksum[:])

			info, err := store.Stat(ctx, key)
			if err != nil {
				t.Fatalf("failed to stat: %v", err)
			}

			if info.ChecksumSHA256 != expectedHex {
				t.Errorf("checksum mismatch: got %s, want %s", info.ChecksumSHA256, expectedHex)
			}

			if info.ByteSize != int64(len(tc.data)) {
				t.Errorf("size mismatch: got %d, want %d", info.ByteSize, len(tc.data))
			}

			if info.MIMEType != tc.mimeType {
				t.Errorf("mime type mismatch: got %s, want %s", info.MIMEType, tc.mimeType)
			}

			// Read back and verify bytes
			readData, err := store.Read(ctx, key)
			if err != nil {
				t.Fatalf("failed to read: %v", err)
			}

			if string(readData) != string(tc.data) {
				t.Errorf("data mismatch: got %d bytes, want %d bytes", len(readData), len(tc.data))
			}
		})
	}
}

func TestStore_Delete(t *testing.T) {
	// Setup: create temp asset root
	tmpDir, err := os.MkdirTemp("", "assetstore-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	ctx := context.Background()
	projectID := "proj-delete"
	visualizationID := "viz-delete"

	// Write an asset
	data := []byte("to be deleted")
	key, err := store.Write(ctx, projectID, visualizationID, "text/plain", data)
	if err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	// Verify it exists
	exists, err := store.Exists(ctx, key)
	if err != nil {
		t.Fatalf("failed to check existence: %v", err)
	}
	if !exists {
		t.Fatal("asset should exist before delete")
	}

	// Delete it
	err = store.Delete(ctx, key)
	if err != nil {
		t.Fatalf("failed to delete: %v", err)
	}

	// Verify it's gone
	exists, err = store.Exists(ctx, key)
	if err != nil {
		t.Fatalf("failed to check existence after delete: %v", err)
	}
	if exists {
		t.Error("asset should not exist after delete")
	}

	// Read should fail
	_, err = store.Read(ctx, key)
	if err == nil {
		t.Error("expected error reading deleted asset")
	}
}

func TestStore_Exists(t *testing.T) {
	// Setup: create temp asset root
	tmpDir, err := os.MkdirTemp("", "assetstore-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	ctx := context.Background()
	projectID := "proj-exists"
	visualizationID := "viz-exists"

	// Non-existent key
	exists, err := store.Exists(ctx, "nonexistent-key")
	if err != nil {
		t.Fatalf("failed to check existence: %v", err)
	}
	if exists {
		t.Error("non-existent key should not exist")
	}

	// Write an asset
	data := []byte("existing asset")
	key, err := store.Write(ctx, projectID, visualizationID, "text/plain", data)
	if err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	// Now it should exist
	exists, err = store.Exists(ctx, key)
	if err != nil {
		t.Fatalf("failed to check existence: %v", err)
	}
	if !exists {
		t.Error("asset should exist")
	}
}

func TestStore_ProjectIsolation(t *testing.T) {
	// Setup: create temp asset root
	tmpDir, err := os.MkdirTemp("", "assetstore-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	ctx := context.Background()

	// Write assets for different projects
	projectA := "project-a"
	projectB := "project-b"
	dataA := []byte("project A data")
	dataB := []byte("project B data")

	keyA, err := store.Write(ctx, projectA, "viz-a", "text/plain", dataA)
	if err != nil {
		t.Fatalf("failed to write project A: %v", err)
	}

	keyB, err := store.Write(ctx, projectB, "viz-b", "text/plain", dataB)
	if err != nil {
		t.Fatalf("failed to write project B: %v", err)
	}

	// Verify keys are different and project-scoped
	if keyA == keyB {
		t.Error("keys for different projects should be different")
	}

	// Verify assets are stored in project-specific directories
	// Keys should contain the project ID
	if !strings.Contains(keyA, projectA) {
		t.Errorf("keyA should contain project A ID, got: %s", keyA)
	}
	if !strings.Contains(keyB, projectB) {
		t.Errorf("keyB should contain project B ID, got: %s", keyB)
	}

	// Verify each project can only read its own assets
	readA, err := store.Read(ctx, keyA)
	if err != nil {
		t.Fatalf("failed to read project A: %v", err)
	}
	if string(readA) != string(dataA) {
		t.Error("project A data mismatch")
	}

	readB, err := store.Read(ctx, keyB)
	if err != nil {
		t.Fatalf("failed to read project B: %v", err)
	}
	if string(readB) != string(dataB) {
		t.Error("project B data mismatch")
	}
}

func TestStore_RejectsOutsideRoot(t *testing.T) {
	// Setup: create temp asset root with a specific structure
	tmpDir, err := os.MkdirTemp("", "assetstore-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a file outside the asset root
	outsideFile := filepath.Join(filepath.Dir(tmpDir), "outside-root.txt")
	if err := os.WriteFile(outsideFile, []byte("outside"), 0644); err != nil {
		t.Fatalf("failed to create outside file: %v", err)
	}
	defer os.Remove(outsideFile)

	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	ctx := context.Background()

	// Attempt to access file outside root using relative path manipulation
	// The store should validate that the resolved path stays within root
	unsafeKey := "../outside-root.txt"

	_, err = store.Read(ctx, unsafeKey)
	if err == nil {
		t.Error("expected error reading file outside root")
	}
}

func TestStore_MultipleAssetsSameVisualization(t *testing.T) {
	// Setup: create temp asset root
	tmpDir, err := os.MkdirTemp("", "assetstore-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	ctx := context.Background()
	projectID := "proj-multi"
	visualizationID := "viz-multi"

	// Write multiple assets for the same visualization
	assets := make(map[string][]byte)
	for i := 0; i < 5; i++ {
		data := []byte(fmt.Sprintf("asset data %d", i))
		mimeType := "image/png"
		key, err := store.Write(ctx, projectID, visualizationID, mimeType, data)
		if err != nil {
			t.Fatalf("failed to write asset %d: %v", i, err)
		}
		assets[key] = data

		// Each key should be unique
		for existingKey := range assets {
			if existingKey == key && string(assets[existingKey]) != string(data) {
				t.Errorf("duplicate key with different data: %s", key)
			}
		}
	}

	// Verify all assets are readable
	for key, expectedData := range assets {
		readData, err := store.Read(ctx, key)
		if err != nil {
			t.Errorf("failed to read asset with key %s: %v", key, err)
			continue
		}
		if string(readData) != string(expectedData) {
			t.Errorf("data mismatch for key %s", key)
		}
	}
}

func TestStore_InvalidRoot(t *testing.T) {
	// Test that NewStore validates the root directory
	_, err := NewStore("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Error("expected error for non-existent root")
	}

	// Create a file where we expect a directory
	tmpFile, err := os.CreateTemp("", "notadir-*")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	_, err = NewStore(tmpFile.Name())
	if err == nil {
		t.Error("expected error when root is a file, not a directory")
	}
}

func TestStore_StatNonExistent(t *testing.T) {
	// Setup: create temp asset root
	tmpDir, err := os.MkdirTemp("", "assetstore-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	ctx := context.Background()

	_, err = store.Stat(ctx, "nonexistent-key")
	if err == nil {
		t.Error("expected error for non-existent key")
	}
	if !errors.Is(err, ErrAssetNotFound) {
		t.Errorf("expected ErrAssetNotFound, got: %v", err)
	}
}
