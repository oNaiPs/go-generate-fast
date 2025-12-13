package config

import (
	"os"
	"path"
	"strconv"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestConfigGet(t *testing.T) {
	expectedConfigDir := t.TempDir()
	expectedCacheDir := t.TempDir()
	expectedDisable := true
	expectedReadOnly := false
	expectedReCache := true

	t.Setenv("GO_GENERATE_FAST_DIR", expectedConfigDir)
	t.Setenv("GO_GENERATE_FAST_CACHE_DIR", expectedCacheDir)
	t.Setenv("GO_GENERATE_FAST_DISABLE", strconv.FormatBool(expectedDisable))
	t.Setenv("GO_GENERATE_FAST_READ_ONLY", strconv.FormatBool(expectedReadOnly))
	t.Setenv("GO_GENERATE_FAST_RECACHE", strconv.FormatBool(expectedReCache))

	Init()

	// Ensure viper's current configuration is cleared after the test
	defer viper.Reset()

	config := Get()

	assert.Equal(t, expectedConfigDir, config.ConfigDir)
	assert.Equal(t, expectedCacheDir, config.CacheDir)
	assert.Equal(t, expectedDisable, config.Disable)
	assert.Equal(t, expectedReadOnly, config.ReadOnly)
	assert.Equal(t, expectedReCache, config.ReCache)
}

func TestConfigCreateDirIfNotExists(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := path.Join(t.TempDir(), "create-dir-test")

	// Call the function with the test directory
	CreateDirIfNotExists(tmpDir)

	// Assert that the directory exists
	_, err := os.Stat(tmpDir)
	assert.False(t, os.IsNotExist(err), "directory doesn't exist")

	// Assert that the directory has the correct permissions
	info, _ := os.Stat(tmpDir)
	assert.True(t, info.IsDir(), "path is not a directory")
	assert.Equal(t, os.FileMode(0700), info.Mode().Perm(), "incorrect directory permissions")
}
