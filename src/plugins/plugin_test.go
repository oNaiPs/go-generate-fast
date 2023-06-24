package plugins

import (
	"plugin"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestPlugin struct {
	plugin.Plugin
}

func (p *TestPlugin) Name() string {
	return "test"
}

func (p *TestPlugin) Matches(opts GenerateOpts) bool {
	return opts.ExecutableName == "test"
}

func (p *TestPlugin) ComputeInputOutputFiles(opts GenerateOpts) *InputOutputFiles {
	ioFiles := InputOutputFiles{}
	return &ioFiles
}

func TestGenerateOptsFile(t *testing.T) {
	opts := GenerateOpts{Path: "/a/b/c"}
	assert.Equal(t, opts.File(), "c")
}

func TestGenerateOptsDir(t *testing.T) {
	opts := GenerateOpts{Path: "/a/b/c"}
	assert.Equal(t, opts.Dir(), "/a/b")
}

func TestGenerateOptsCommand(t *testing.T) {
	opts := GenerateOpts{Words: []string{"command", "arg1"}}
	assert.Equal(t, opts.Command(), "command arg1")
}

func TestRegister(t *testing.T) {
	ClearPlugins()

	testPlugin := TestPlugin{}
	RegisterPlugin(&testPlugin)

	notMatched := MatchPlugin(GenerateOpts{ExecutableName: "unmatched"})
	assert.Nil(t, notMatched)

	matched := MatchPlugin(GenerateOpts{ExecutableName: "test"})
	assert.NotNil(t, matched)
	assert.Equal(t, matched.Name(), "test")

	assert.Panics(t, func() {
		RegisterPlugin(&testPlugin)
	})
}
