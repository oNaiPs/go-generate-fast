package plugin_genny

import (
	"testing"

	"github.com/oNaiPs/go-generate-fast/src/plugins"
	"github.com/stretchr/testify/assert"
)

func TestName(t *testing.T) {
	var g GennyPlugin

	assert.Equal(t, "genny", g.Name())
}

func TestMatches(t *testing.T) {
	var g GennyPlugin
	var option plugins.GenerateOpts

	option.ExecutableName = "genny"
	assert.True(t, g.Matches(option))

	option.ExecutableName = "geny"
	assert.False(t, g.Matches(option))

	option.GoPackage = "github.com/cheekybits/genny"
	assert.True(t, g.Matches(option))

	option.GoPackage = "github.com/cheekybits/geny"
	assert.False(t, g.Matches(option))
}

func TestComputeInputOutputFiles(t *testing.T) {
	var g GennyPlugin
	var option plugins.GenerateOpts

	option.SanitizedArgs = []string{"wrongCommand", "int=integer"}

	assert.Nil(t, g.ComputeInputOutputFiles(option))

	option.SanitizedArgs = []string{"gen", "wrongTypeSet"}

	assert.Nil(t, g.ComputeInputOutputFiles(option))

	option.SanitizedArgs = []string{"-in=in.txt", "-out", "out.txt", "gen", "int=integer"}

	assert.NotNil(t, g.ComputeInputOutputFiles(option))

	result := g.ComputeInputOutputFiles(option)

	expectedIOFiles := &plugins.InputOutputFiles{
		InputFiles:  []string{"in.txt"},
		OutputFiles: []string{"out.txt"},
		Extra:       nil,
	}
	assert.Equal(t, expectedIOFiles, result)
}
