package plugin_go_bindata

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/oNaiPs/go-generate-fast/src/plugins"
	"github.com/stretchr/testify/assert"
)

func TestGobindataPluginName(t *testing.T) {
	p := &GobindataPlugin{}

	assert.Equal(t, "go-bindata", p.Name(), "They should be equal")
}

func TestGobindataPluginMatches(t *testing.T) {
	p := &GobindataPlugin{}

	t.Run("Match by executable", func(t *testing.T) {
		opts := plugins.GenerateOpts{ExecutableName: "go-bindata"}

		assert.True(t, p.Matches(opts))
	})

	t.Run("Match by package", func(t *testing.T) {
		opts := plugins.GenerateOpts{GoPackage: "github.com/go-bindata/go-bindata/..."}

		assert.True(t, p.Matches(opts))
	})

	t.Run("No match", func(t *testing.T) {
		opts := plugins.GenerateOpts{ExecutableName: "not-bindata", GoPackage: "not-bindata"}

		assert.False(t, p.Matches(opts))
	})
}

func TestGobindataPluginComputeInputOutputFiles(t *testing.T) {
	p := &GobindataPlugin{}
	tmpDir := t.TempDir()

	inputFile1 := filepath.Join(tmpDir, "input1")
	err := os.WriteFile(inputFile1, []byte("dummy content"), 0600)
	assert.NoError(t, err)

	inputFile2 := filepath.Join(tmpDir, "input2")
	err = os.WriteFile(inputFile2, []byte("dummy content"), 0600)
	assert.NoError(t, err)

	outputFile := filepath.Join(tmpDir, "output")

	result := p.ComputeInputOutputFiles(plugins.GenerateOpts{
		SanitizedArgs: []string{"-o", outputFile, inputFile1},
	})

	assert.Equal(t, []string{inputFile1}, result.InputFiles)
	assert.Equal(t, []string{outputFile}, result.OutputFiles)

	result = p.ComputeInputOutputFiles(plugins.GenerateOpts{
		SanitizedArgs: []string{inputFile1},
	})

	assert.Equal(t, []string{inputFile1}, result.InputFiles)
	assert.Equal(t, []string{"./bindata.go"}, result.OutputFiles)

	result = p.ComputeInputOutputFiles(plugins.GenerateOpts{
		SanitizedArgs: []string{tmpDir + "/..."},
	})

	assert.Equal(t, []string{inputFile1, inputFile2}, result.InputFiles)
	assert.Equal(t, []string{"./bindata.go"}, result.OutputFiles)
}
