package plugin_stringer

import (
	"os"
	"path"
	"testing"

	"github.com/oNaiPs/go-generate-fast/src/plugins"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringerPlugin_Name(t *testing.T) {
	p := &StringerPlugin{}
	assert.Equal(t, "stringer", p.Name())
}

func TestStringerPlugin_Matches(t *testing.T) {
	p := &StringerPlugin{}
	assert.True(t, p.Matches(plugins.GenerateOpts{
		ExecutableName: "stringer",
	}))
}

func TestStringerPlugin_ComputeInputOutputFiles(t *testing.T) {
	p := &StringerPlugin{}
	tempDir := t.TempDir()

	err := os.WriteFile(path.Join(tempDir, "go.mod"), []byte("module example.com/mod"), 0644)
	assert.NoError(t, err)
	err = os.WriteFile(path.Join(tempDir, "input_file1.go"), []byte("package ex"), 0644)
	assert.NoError(t, err)
	err = os.WriteFile(path.Join(tempDir, "input_file2.go"), []byte("package ex"), 0644)
	assert.NoError(t, err)

	opts := plugins.GenerateOpts{
		Path:           path.Join(tempDir, "test.go"),
		ExecutableName: "stringer",
		SanitizedArgs:  []string{"-type", "MyType", "-output", "output_file.go"},
	}

	err = os.Chdir(tempDir)
	assert.NoError(t, err)
	cwd, err := os.Getwd()
	assert.NoError(t, err)
	ioFiles := p.ComputeInputOutputFiles(opts)
	require.NotNil(t, ioFiles)
	assert.Equal(t, []string{
		path.Join(cwd, "input_file1.go"),
		path.Join(cwd, "input_file2.go"),
	}, ioFiles.InputFiles)
	assert.Equal(t, []string{"output_file.go"}, ioFiles.OutputFiles)
}
