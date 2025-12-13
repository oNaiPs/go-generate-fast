package plugin_moq

import (
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/oNaiPs/go-generate-fast/src/plugins"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMoqPlugin_Name(t *testing.T) {
	p := &MoqPlugin{}
	assert.Equal(t, "moq", p.Name())
}

func TestMoqPlugin_Matches(t *testing.T) {
	p := &MoqPlugin{}
	assert.True(t, p.Matches(plugins.GenerateOpts{
		ExecutableName: "moq",
	}))
}

func TestMoqPlugin_ComputeInputOutputFiles(t *testing.T) {
	p := &MoqPlugin{}
	tempDir := t.TempDir()

	// Resolve symlinks (on macOS /var -> /private/var)
	tempDir, err := filepath.EvalSymlinks(tempDir)
	require.NoError(t, err)

	err = os.WriteFile(path.Join(tempDir, "go.mod"), []byte("module example.com/mod"), 0644)
	assert.NoError(t, err)
	err = os.WriteFile(path.Join(tempDir, "input_file1.go"), []byte(`
package example
type Foo interface {
	Bar(x int) int
}
`), 0644)
	assert.NoError(t, err)

	opts := plugins.GenerateOpts{
		Path:           path.Join(tempDir, "test.go"),
		ExecutableName: "moq",
		SanitizedArgs: []string{
			"-out", "output_file.go",
			"-pkg", "example",
			".",
			"Foo",
		},
	}

	_ = os.Chdir(tempDir)
	ioFiles := p.ComputeInputOutputFiles(opts)
	require.NotNil(t, ioFiles)
	assert.Equal(t, []string{
		path.Join(tempDir, "input_file1.go"),
	}, ioFiles.InputFiles)
	assert.Equal(t, []string{"output_file.go"}, ioFiles.OutputFiles)
}
