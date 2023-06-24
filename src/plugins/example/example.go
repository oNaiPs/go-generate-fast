package plugin_example

import "github.com/oNaiPs/go-generate-fast/src/plugins"

type ExamplePlugin struct {
	plugins.Plugin
}

func (p *ExamplePlugin) Name() string {
	return "example"
}

func (p *ExamplePlugin) Matches(opts plugins.GenerateOpts) bool {
	return opts.ExecutableName == "example"
}

func (p *ExamplePlugin) ComputeInputOutputFiles(opts plugins.GenerateOpts) *plugins.InputOutputFiles {
	return &plugins.InputOutputFiles{}
}

func init() {
	plugins.RegisterPlugin(&ExamplePlugin{})
}
