package plugin_gqlgen

import (
	"os"
	"path"
	"testing"

	"github.com/oNaiPs/go-generate-fast/src/plugins"
	"github.com/stretchr/testify/assert"
)

func TestName(t *testing.T) {
	var g GqlgenPlugin

	assert.Equal(t, "Gqlgen", g.Name())
}

func TestMatches(t *testing.T) {
	var g GqlgenPlugin
	var option plugins.GenerateOpts

	option.ExecutableName = "gqlgen"
	assert.True(t, g.Matches(option))

	option.ExecutableName = "gqlgen1"
	assert.False(t, g.Matches(option))

	option.GoPackage = "github.com/99designs/gqlgen"
	assert.True(t, g.Matches(option))

	option.GoPackage = "github.com/99designs/gqlgen1"
	assert.False(t, g.Matches(option))
}

func TestComputeInputOutputFiles(t *testing.T) {
	tempDir := t.TempDir()
	configFile := path.Join(tempDir, "gqlgen.yml")
	err := os.WriteFile(configFile, []byte("---"), 0666)
	assert.NoError(t, err)

	schemaFile := path.Join(tempDir, "schema.graphql")
	err = os.WriteFile(schemaFile, []byte("type Test {text: String!}"), 0666)
	assert.NoError(t, err)

	err = os.Chdir(tempDir)
	assert.NoError(t, err)

	var g GqlgenPlugin
	option := plugins.GenerateOpts{
		Path: path.Join(tempDir, "test.go"),
	}

	option.SanitizedArgs = []string{"wrongCommand"}

	assert.Nil(t, g.ComputeInputOutputFiles(option))

	option.SanitizedArgs = []string{"generate", "gen"}

	assert.NotNil(t, g.ComputeInputOutputFiles(option))
}
