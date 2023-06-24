package cache_test

import (
	"os"
	"testing"
	"time"

	"github.com/oNaiPs/go-generate-fast/src/core/cache"
	"github.com/stretchr/testify/assert"
)

func TestGetConfigFilePath(t *testing.T) {
	assert.Equal(t,
		"/path/to/cache/hit/dir/cache.json",
		cache.GetConfigFilePath("/path/to/cache/hit/dir"),
		"Incorrect config file path")
}

func TestSaveAndLoadConfig(t *testing.T) {
	cacheHitDir := t.TempDir()

	// Create a sample cache config
	config := cache.CacheConfig{
		OutputFiles: []cache.CacheConfigOutputFileInfo{
			{
				Hash:    "abc123",
				Path:    "/path/to/output/file1",
				ModTime: time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC),
			},
			{
				Hash:    "def456",
				Path:    "/path/to/output/file2",
				ModTime: time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	err := cache.SaveConfig(config, cacheHitDir)
	assert.NoError(t, err, "Failed to save cache config")

	cacheFilePath := cache.GetConfigFilePath(cacheHitDir)
	_, err = os.Stat(cacheFilePath)
	assert.False(t, os.IsNotExist(err), "Cache file doesn't exist")

	loadedConfig, err := cache.LoadConfig(cacheHitDir)
	assert.NoError(t, err, "Failed to load cache config")

	assert.Equal(t, config, loadedConfig, "Loaded cache config does not match the original")
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	_, err := cache.LoadConfig("/non-existent/cache/dir")

	assert.Error(t, err, "Expected an error for non-existent cache config file")
	assert.Contains(t, err.Error(), "cannot read cache config file")
}

func TestLoadConfig_InvalidData(t *testing.T) {
	cacheHitDir := t.TempDir()

	err := os.WriteFile(cache.GetConfigFilePath(cacheHitDir), []byte{}, 0644)
	assert.NoError(t, err, "Failed to create empty cache config file")

	_, err = cache.LoadConfig(cacheHitDir)

	assert.Error(t, err, "Expected an error for invalid cache config data")
	assert.Contains(t, err.Error(), "cannot unmarshal cache config file")
}
