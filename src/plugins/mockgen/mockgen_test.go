package plugin_mockgen

import (
	"os"
	"path"
	"testing"

	"github.com/oNaiPs/go-generate-fast/src/plugins"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockgenPlugin_Name(t *testing.T) {
	p := &MockgenPlugin{}
	assert.Equal(t, "mockgen", p.Name())
}

func TestMockgenPlugin_Matches(t *testing.T) {
	p := &MockgenPlugin{}
	assert.True(t, p.Matches(plugins.GenerateOpts{
		ExecutableName: "mockgen",
	}))
}

func TestMockgenPlugin_ComputeInputOutputFiles_SourceMode(t *testing.T) {
	p := &MockgenPlugin{}
	tempDir := t.TempDir()

	err := os.WriteFile(path.Join(tempDir, "go.mod"), []byte("module example.com/mod"), 0644)
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
		ExecutableName: "mockgen",
		SanitizedArgs: []string{
			"-source", "input_file1.go",
			"-destination", "output_file.go",
			"-aux_files", "aa=aa.go,b=ab.go,ac=a/b/ac.go",
			"-imports", "ia=ia.go,b=ib.go,ic=a/b/ic.go",
			"-copyright_file", "cop.txt",
		},
	}

	ioFiles := p.ComputeInputOutputFiles(opts)
	require.NotNil(t, ioFiles)
	assert.Equal(t, []string{
		"aa.go",
		"ab.go",
		"a/b/ac.go",
		"cop.txt",
		"ia.go",
		"ib.go",
		"a/b/ic.go",
		"input_file1.go",
	}, ioFiles.InputFiles)
	assert.Equal(t, []string{"output_file.go"}, ioFiles.OutputFiles)
}

func TestMockgenPlugin_ComputeInputOutputFiles_ReflectMode(t *testing.T) {
	p := &MockgenPlugin{}
	tempDir := t.TempDir()

	err := os.WriteFile(path.Join(tempDir, "go.mod"), []byte("module example_pkg"), 0644)
	assert.NoError(t, err)
	err = os.WriteFile(path.Join(tempDir, "input_file1.go"), []byte(`
package example_pkg
type Driver interface {
	Bar(x int) int
}
`), 0644)
	assert.NoError(t, err)

	opts := plugins.GenerateOpts{
		Path:           path.Join(tempDir, "test.go"),
		ExecutableName: "mockgen",
		SanitizedArgs:  []string{"-destination", "output_file.go", "example_pkg", "Driver"},
	}

	err = os.Chdir(tempDir)
	assert.NoError(t, err)
	ioFiles := p.ComputeInputOutputFiles(opts)
	require.NotNil(t, ioFiles)
	assert.Equal(t, []string{path.Join(tempDir, "input_file1.go")}, ioFiles.InputFiles)
	assert.Equal(t, []string{"output_file.go"}, ioFiles.OutputFiles)
}
