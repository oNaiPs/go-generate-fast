package plugin_gqlgen

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/99designs/gqlgen/codegen/config"
	"github.com/oNaiPs/go-generate-fast/src/plugins"
	"go.uber.org/zap"
)

type GqlgenPlugin struct {
	plugins.Plugin
}

func (p *GqlgenPlugin) Name() string {
	return "Gqlgen"
}

func (p *GqlgenPlugin) Matches(opts plugins.GenerateOpts) bool {
	return opts.ExecutableName == "gqlgen" ||
		opts.GoPackage == "github.com/99designs/gqlgen"
}

type GqlgenFlags struct {
	Config string
}

func (p *GqlgenPlugin) ComputeInputOutputFiles(opts plugins.GenerateOpts) *plugins.InputOutputFiles {
	flagSet := flag.NewFlagSet("Gqlgen", flag.ContinueOnError)

	flags := GqlgenFlags{}
	flagSet.StringVar(&flags.Config, "config", "", "config file")
	flagSet.StringVar(&flags.Config, "c", "", "config file")

	err := flagSet.Parse(opts.SanitizedArgs)
	if err != nil {
		zap.S().Warn("Cannot parse Gqlgen arguments: ", err.Error())
		return nil
	}

	if flagSet.NArg() > 0 && flagSet.Arg(0) != "generate" {
		zap.S().Info("gqlgen only supports generate command")
		return nil
	}

	cfg, cfgFile, err := getConfig(flags.Config)
	if err != nil {
		zap.S().Errorf("cannot get gqlgen config: %s", err)
		return nil
	}

	// LoadSchema calls private check() method that defaults some additional values in the config
	err = cfg.LoadSchema()
	if err != nil {
		zap.S().Errorf("failed loading schema: %s", err)
		return nil
	}

	ioFiles := plugins.InputOutputFiles{}

	ioFiles.InputFiles = append(ioFiles.InputFiles, cfgFile)
	ioFiles.InputFiles = append(ioFiles.InputFiles, cfg.SchemaFilename...)

	if cfg.Model.IsDefined() {
		ioFiles.OutputFiles = append(ioFiles.OutputFiles, cfg.Model.Filename)
	}

	// TODO cache will remove manually added resolver code from these files
	if cfg.Resolver.IsDefined() {
		if cfg.Resolver.Layout == config.LayoutSingleFile {
			ioFiles.OutputFiles = append(ioFiles.OutputFiles, cfg.Resolver.Filename)
		} else if cfg.Resolver.Layout == config.LayoutFollowSchema {
			for _, schemaFile := range cfg.SchemaFilename {
				ioFiles.OutputFiles = append(ioFiles.OutputFiles, path.Join(cfg.Resolver.DirName, filename(schemaFile, cfg.Resolver.FilenameTemplate)))
			}
		} else {
			zap.S().Error("unknown config resolver layout: ", cfg.Resolver.Layout)
		}
	}

	if cfg.Exec.IsDefined() {
		if cfg.Exec.Layout == config.ExecLayoutSingleFile {
			ioFiles.OutputFiles = append(ioFiles.OutputFiles, cfg.Exec.Filename)
		} else if cfg.Exec.Layout == config.ExecLayoutFollowSchema {
			ioFiles.OutputFiles = append(ioFiles.OutputFiles, path.Join(cfg.Exec.DirName, "root_.generated.go"))
			for _, schemaFile := range cfg.SchemaFilename {
				ioFiles.OutputFiles = append(ioFiles.OutputFiles, path.Join(cfg.Exec.DirName, filename(schemaFile, cfg.Exec.FilenameTemplate)))
			}
		} else {
			zap.S().Error("unknown config exec layout", cfg.Exec.Layout)
		}
	}

	if cfg.Federation.IsDefined() {
		ioFiles.OutputFiles = append(ioFiles.OutputFiles, cfg.Federation.Filename)
	}

	return &ioFiles
}

func filename(schemaFile string, filenameTemplate string) string {
	gqlname := filepath.Base(schemaFile)
	ext := filepath.Ext(schemaFile)
	name := strings.TrimSuffix(gqlname, ext)

	if filenameTemplate == "" {
		filenameTemplate = "{name}.generated.go"
	}

	return strings.ReplaceAll(filenameTemplate, "{name}", name)
}

var cfgFilenames = []string{".gqlgen.yml", "gqlgen.yml", "gqlgen.yaml"}

func findCfg() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("unable to get working dir to findCfg: %w", err)
	}

	cfg := findCfgInDir(dir)

	for cfg == "" && dir != filepath.Dir(dir) {
		dir = filepath.Dir(dir)
		cfg = findCfgInDir(dir)
	}

	if cfg == "" {
		return "", os.ErrNotExist
	}

	return cfg, nil
}

func findCfgInDir(dir string) string {
	for _, cfgName := range cfgFilenames {
		path := filepath.Join(dir, cfgName)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

func getConfig(configFile string) (*config.Config, string, error) {
	if configFile != "" {
		cfgFile, err := config.LoadConfig(configFile)
		return cfgFile, configFile, err
	} else {
		cfgFile, err := findCfg()
		if err != nil {
			return nil, cfgFile, err
		}
		err = os.Chdir(filepath.Dir(cfgFile))
		if err != nil {
			return nil, cfgFile, err
		}
		cfg, err := config.LoadConfig(cfgFile)
		if errors.Is(err, fs.ErrNotExist) {
			cfg, err = config.LoadDefaultConfig()
		}
		return cfg, cfgFile, err
	}
}

func init() {
	plugins.RegisterPlugin(&GqlgenPlugin{})
}
