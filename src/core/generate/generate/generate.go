// Package generate implements caching for go generate commands.
// Instead of reimplementing go generate logic, we use the standard
// go generate command for compatibility while adding intelligent caching.
package generate

import (
	"bufio"
	"bytes"
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/oNaiPs/go-generate-fast/src/core/cache"
	"github.com/oNaiPs/go-generate-fast/src/core/config"
	"github.com/oNaiPs/go-generate-fast/src/core/generate/base"
	"github.com/oNaiPs/go-generate-fast/src/core/generate/cfg"
	"github.com/oNaiPs/go-generate-fast/src/core/golist"
	"github.com/oNaiPs/go-generate-fast/src/plugins"
	"github.com/oNaiPs/go-generate-fast/src/utils/fs"
	"go.uber.org/zap"
)

var (
	generateRunFlag string         // generate -run flag
	generateRunRE   *regexp.Regexp // compiled expression for -run

	generateSkipFlag string         // generate -skip flag
	generateSkipRE   *regexp.Regexp // compiled expression for -skip
)

func RunGenerate(args []string) {
	if generateRunFlag != "" {
		var err error
		generateRunRE, err = regexp.Compile(generateRunFlag)
		if err != nil {
			log.Fatalf("generate: %s", err)
		}
	}
	if generateSkipFlag != "" {
		var err error
		generateSkipRE, err = regexp.Compile(generateSkipFlag)
		if err != nil {
			log.Fatalf("generate: %s", err)
		}
	}

	for _, pkg := range golist.ModulesAndErrors(args) {
		if pkg.Error != nil {
			fmt.Println(*pkg.Error)
			continue
		}

		if !generate(pkg.Package) {
			break
		}
	}
}

type directiveInfo struct {
	lineNum     int
	command     string
	opts        plugins.GenerateOpts
	cacheResult cache.VerifyResult
	needsRun    bool
	canCache    bool
}

func generate(absFile string) bool {
	src, err := os.ReadFile(absFile)
	if err != nil {
		log.Fatalf("generate: %s", err)
	}

	filePkg, err := parser.ParseFile(token.NewFileSet(), "", src, parser.PackageClauseOnly)
	if err != nil {
		return true
	}

	if cfg.BuildV {
		zap.S().Debug(absFile)
	}

	directives, err := scanDirectives(absFile, src, filePkg.Name.String())
	if err != nil {
		zap.S().Errorf("error scanning %s: %s", absFile, err)
		base.SetExitStatus(1)
		return false
	}

	if len(directives) == 0 {
		return true
	}

	cwd, err := os.Getwd()
	if err != nil {
		zap.S().Fatalf("cannot get working directory: %s", err)
	}

	anyNeedRun := false
	for i := range directives {
		dir := directives[i].opts.Dir()

		if err := os.Chdir(dir); err != nil {
			zap.S().Fatalf("cannot chdir to command directory: %s", err)
		}

		checkDirectiveCache(&directives[i])
		if directives[i].needsRun {
			anyNeedRun = true
		}

		if err := os.Chdir(cwd); err != nil {
			zap.S().Fatalf("cannot restore working directory: %s", err)
		}
	}

	if anyNeedRun {
		if config.Get().ForceUseCache {
			zap.S().Errorf("force_use_cache mode but cache miss for %s", absFile)
			base.SetExitStatus(1)
		} else {
			if !executeGoGenerate(absFile) {
				base.SetExitStatus(1)
				return false
			}
		}
	}

	for i := range directives {
		dir := directives[i].opts.Dir()
		if err := os.Chdir(dir); err != nil {
			zap.S().Fatalf("cannot chdir: %s", err)
		}

		saveAndReportDirective(&directives[i], absFile, cwd, anyNeedRun)

		if err := os.Chdir(cwd); err != nil {
			zap.S().Fatalf("cannot restore working directory: %s", err)
		}
	}

	return true
}

func scanDirectives(absFile string, src []byte, pkg string) ([]directiveInfo, error) {
	var directives []directiveInfo
	input := bufio.NewReader(bytes.NewReader(src))

	var extraInputPatterns []string
	var extraOutputPatterns []string
	lineNum := 0

	for {
		lineNum++
		buf, err := input.ReadSlice('\n')
		if err == bufio.ErrBufferFull {
			for err == bufio.ErrBufferFull {
				_, err = input.ReadSlice('\n')
			}
			if err != nil && err != io.EOF {
				return nil, err
			}
			continue
		}
		if err != nil && err != io.EOF {
			return nil, err
		}

		if !bytes.HasPrefix(buf, []byte("//go:generate")) {
			if err == io.EOF {
				break
			}
			continue
		}

		if bytes.HasPrefix(buf, []byte("//go:generate_input ")) {
			patterns := parseSimpleLine(buf, len("//go:generate_input "))
			extraInputPatterns = append(extraInputPatterns, patterns...)
			if err == io.EOF {
				break
			}
			continue
		}
		if bytes.HasPrefix(buf, []byte("//go:generate_output ")) {
			patterns := parseSimpleLine(buf, len("//go:generate_output "))
			extraOutputPatterns = append(extraOutputPatterns, patterns...)
			if err == io.EOF {
				break
			}
			continue
		}

		if !bytes.HasPrefix(buf, []byte("//go:generate ")) && !bytes.HasPrefix(buf, []byte("//go:generate\t")) {
			if err == io.EOF {
				break
			}
			continue
		}

		if generateRunFlag != "" && !generateRunRE.Match(bytes.TrimSpace(buf)) {
			if err == io.EOF {
				break
			}
			continue
		}
		if generateSkipFlag != "" && generateSkipRE.Match(bytes.TrimSpace(buf)) {
			if err == io.EOF {
				break
			}
			continue
		}

		command := strings.TrimSpace(string(buf[len("//go:generate"):]))
		if command == "" {
			if err == io.EOF {
				break
			}
			continue
		}

		words := strings.Fields(command)
		if len(words) == 0 {
			if err == io.EOF {
				break
			}
			continue
		}

		if words[0] == "-command" {
			if err == io.EOF {
				break
			}
			continue
		}

		if cfg.BuildN || cfg.BuildX {
			zap.S().Info(command)
		}
		if cfg.BuildN {
			if err == io.EOF {
				break
			}
			continue
		}

		opts := plugins.GenerateOpts{
			Path:                absFile,
			Words:               words,
			ExtraInputPatterns:  append([]string{}, extraInputPatterns...),
			ExtraOutputPatterns: append([]string{}, extraOutputPatterns...),
		}

		// Handle env variable prefix (e.g., "env VAR=value command args...")
		actualWords := skipEnvPrefix(words)

		executablePath, _ := fs.FindExecutablePath(actualWords[0])
		if executablePath != "" {
			opts.ExecutablePath = executablePath
		} else {
			opts.ExecutablePath = actualWords[0]
		}

		// Create a temporary opts with actualWords for parsing
		tempOpts := opts
		tempOpts.Words = actualWords

		if !parseGoToolCommand(&tempOpts) && !parseGoRunCommand(&tempOpts) {
			opts.ExecutableName = filepath.Base(actualWords[0])
			opts.SanitizedArgs = actualWords[1:]
		} else {
			// Copy parsed values back to opts
			opts.ExecutableName = tempOpts.ExecutableName
			opts.SanitizedArgs = tempOpts.SanitizedArgs
			opts.GoPackage = tempOpts.GoPackage
			opts.GoPackageVersion = tempOpts.GoPackageVersion
		}

		directives = append(directives, directiveInfo{
			lineNum: lineNum,
			command: command,
			opts:    opts,
		})

		extraInputPatterns = nil
		extraOutputPatterns = nil

		if err == io.EOF {
			break
		}
	}

	return directives, nil
}

func parseSimpleLine(buf []byte, stripLen int) []string {
	line := string(buf[stripLen:])
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}
	return strings.Fields(line)
}

// skipEnvPrefix skips over "env" and environment variable assignments
// to find the actual command. For example:
// ["env", "VAR=value", "mockgen", "args"] -> ["mockgen", "args"]
// ["mockgen", "args"] -> ["mockgen", "args"] (unchanged)
func skipEnvPrefix(words []string) []string {
	if len(words) == 0 {
		return words
	}

	// Check if command starts with "env"
	if words[0] != "env" {
		return words
	}

	// Skip "env" and any VAR=value assignments
	i := 1
	for i < len(words) && strings.Contains(words[i], "=") {
		i++
	}

	// Return remaining words (the actual command)
	if i < len(words) {
		return words[i:]
	}

	// If we only have env variables with no command, return as-is
	return words
}

func checkDirectiveCache(d *directiveInfo) {
	if config.Get().Disable {
		d.needsRun = true
		return
	}

	cacheResult, err := cache.Verify(d.opts)
	d.cacheResult = cacheResult
	d.canCache = err == nil

	if err != nil {
		zap.S().Debugf("cannot verify cache: %s", err)
		d.needsRun = true
		return
	}

	if config.Get().ReCache {
		d.needsRun = true
		return
	}

	if cacheResult.CacheHit {
		if err := cache.Restore(cacheResult); err != nil {
			zap.S().Errorf("cannot restore cache: %s", err)
			d.needsRun = true
		} else {
			d.needsRun = false
		}
	} else {
		d.needsRun = true
	}
}

func executeGoGenerate(absFile string) bool {
	cmd := exec.Command("go", "generate", absFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = filepath.Dir(absFile)
	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		zap.S().Errorf("go generate failed for %s: %s", absFile, err)
		return false
	}
	return true
}

func saveAndReportDirective(d *directiveInfo, absFile string, cwd string, fileWasGenerated bool) {
	var cachedInfo []string
	start := time.Now()

	if fileWasGenerated && d.canCache && d.cacheResult.CanSave &&
		!config.Get().ReadOnly && !config.Get().ForceUseCache {
		if err := cache.Save(d.cacheResult); err != nil {
			zap.S().Errorf("cannot save cache: %s", err)
		} else {
			cachedInfo = append(cachedInfo, "saved")
		}
	}

	if !d.needsRun {
		cachedInfo = append(cachedInfo, "cached")
	} else if fileWasGenerated {
		cachedInfo = append(cachedInfo, "generated")
	}

	if config.Get().Disable {
		cachedInfo = append(cachedInfo, "disabled")
	} else if d.cacheResult.PluginMatch == nil {
		cachedInfo = append(cachedInfo, "noplugin")
	}

	cachedInfo = append(cachedInfo, fmt.Sprintf("%dms", time.Since(start).Milliseconds()))

	relPath, _ := filepath.Rel(cwd, absFile)
	if relPath == "" {
		relPath = absFile
	}
	zap.S().Infof("%s: %s (%s)", relPath, d.command, strings.Join(cachedInfo, ", "))
}

func parseGoToolCommand(opts *plugins.GenerateOpts) bool {
	if len(opts.Words) < 3 {
		return false
	}
	if opts.Words[0] != "go" || opts.Words[1] != "tool" {
		return false
	}
	opts.ExecutableName = filepath.Base(opts.Words[2])
	opts.SanitizedArgs = opts.Words[3:]
	return true
}

func parseGoRunCommand(opts *plugins.GenerateOpts) bool {
	if len(opts.Words) < 3 {
		return false
	}
	if opts.Words[0] != "go" || opts.Words[1] != "run" {
		return false
	}

	pkgIdx := 2
	for pkgIdx < len(opts.Words) && strings.HasPrefix(opts.Words[pkgIdx], "-") {
		pkgIdx++
		// Skip flag value if this flag takes an argument
		if pkgIdx < len(opts.Words) && !strings.HasPrefix(opts.Words[pkgIdx], "-") {
			pkgIdx++
		}
	}

	if pkgIdx >= len(opts.Words) {
		return false
	}

	packageAndVersion := strings.Split(opts.Words[pkgIdx], "@")
	opts.GoPackage = packageAndVersion[0]
	if len(packageAndVersion) > 1 {
		opts.GoPackageVersion = packageAndVersion[1]
	}
	opts.SanitizedArgs = opts.Words[pkgIdx+1:]
	return true
}
