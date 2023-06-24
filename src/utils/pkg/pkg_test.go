package pkg

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadPackages_Success(t *testing.T) {
	tempDir := t.TempDir()

	err := os.WriteFile(path.Join(tempDir, "go.mod"), []byte("module example.com/mod"), 0644)
	assert.NoError(t, err)
	err = os.WriteFile(path.Join(tempDir, "file1.go"), []byte("package ex"), 0644)
	assert.NoError(t, err)
	err = os.WriteFile(path.Join(tempDir, "file2.go"), []byte("package ex"), 0644)
	assert.NoError(t, err)

	p := LoadPackages(tempDir, []string{"file1.go", "file2.go"}, []string{})

	assert.NotNil(t, p)
	assert.Equal(t, []string{path.Join(tempDir, "file1.go"), path.Join(tempDir, "file2.go")}, p.CompiledGoFiles)
}

func TestLoadPackages_Error(t *testing.T) {
	tempDir := t.TempDir()

	p := LoadPackages(tempDir, []string{"blah"}, []string{"tag1"})

	assert.NotNil(t, p)
	assert.Nil(t, p.CompiledGoFiles)
	assert.Len(t, p.Errors, 1)
	assert.Contains(t, p.Errors[0].Msg, "package blah is not in std")
}
