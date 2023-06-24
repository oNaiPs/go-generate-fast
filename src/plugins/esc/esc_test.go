package plugin_esc

import (
	"os"
	"path"
	"testing"

	"github.com/oNaiPs/go-generate-fast/src/plugins"
	util_test "github.com/oNaiPs/go-generate-fast/src/test"
	"github.com/stretchr/testify/assert"
)

func TestEscPlugin_Name(t *testing.T) {
	plugin := &EscPlugin{}
	name := plugin.Name()
	assert.Equal(t, "esc", name, "Expected plugin name to be 'esc'")
}

func TestEscPlugin_Matches(t *testing.T) {
	tests := []struct {
		opts     plugins.GenerateOpts
		expected bool
	}{
		{
			opts: plugins.GenerateOpts{
				ExecutableName: "esc",
			},
			expected: true,
		},
		{
			opts: plugins.GenerateOpts{
				GoPackage: "github.com/mjibson/esc",
			},
			expected: true,
		},
		{
			opts: plugins.GenerateOpts{
				ExecutableName: "something",
				GoPackage:      "github.com/mjibson/esc",
			},
			expected: true,
		},
		{
			opts: plugins.GenerateOpts{
				ExecutableName: "esc",
				GoPackage:      "github.com/some/other/pkg",
			},
			expected: true,
		},
		{
			opts: plugins.GenerateOpts{
				ExecutableName: "not-esc",
				GoPackage:      "github.com/some/other/pkg",
			},
			expected: false,
		},
	}

	plugin := &EscPlugin{}
	for _, test := range tests {
		matches := plugin.Matches(test.opts)
		assert.Equal(t, test.expected, matches, "Unexpected match result")
	}
}

func TestEscPlugin_ComputeInputOutputFiles(t *testing.T) {
	tempFile := util_test.WriteTempFile(t, "test")

	err := os.Chdir(path.Dir(tempFile.Name()))
	assert.NoError(t, err)
	opts := plugins.GenerateOpts{
		SanitizedArgs: []string{
			"-o", "output.go",
			"-pkg", "my_pkg",
			"-prefix", "/my/prefix",
			"-ignore", "\\test",
			"-include", "go-generate-fast-test",
			tempFile.Name(),
		},
	}
	plugin := &EscPlugin{}
	ioFiles := plugin.ComputeInputOutputFiles(opts)
	assert.NotNil(t, ioFiles, "Expected non-nil input/output files")
	assert.Equal(t, []string{tempFile.Name()}, ioFiles.InputFiles, "Unexpected input files")
	assert.Equal(t, []string{"output.go"}, ioFiles.OutputFiles, "Unexpected output files")
}

func TestEscPlugin_ComputeInputOutputFiles_DefaultOut(t *testing.T) {
	tempFile := util_test.WriteTempFile(t, "test")

	err := os.Chdir(path.Dir(tempFile.Name()))
	assert.NoError(t, err)
	opts := plugins.GenerateOpts{
		SanitizedArgs: []string{
			tempFile.Name(),
		},
	}
	plugin := &EscPlugin{}
	ioFiles := plugin.ComputeInputOutputFiles(opts)
	assert.NotNil(t, ioFiles, "Expected non-nil input/output files")
	assert.Equal(t, []string{"static.go"}, ioFiles.OutputFiles, "Unexpected output files")
}
