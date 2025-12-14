package cache

import (
	"os"
	"path"
	"plugin"
	"strings"
	"testing"
	"time"

	"github.com/oNaiPs/go-generate-fast/src/core/config"
	"github.com/oNaiPs/go-generate-fast/src/plugins"
	util_test "github.com/oNaiPs/go-generate-fast/src/test"
	"github.com/oNaiPs/go-generate-fast/src/utils/hash"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	config.Init()
	code := m.Run()
	os.Exit(code)
}

type TestPlugin struct {
	plugin.Plugin
	t *testing.T
}

func (p *TestPlugin) Name() string {
	return "test"
}

func (p *TestPlugin) Matches(opts plugins.GenerateOpts) bool {
	return opts.ExecutableName == "test"
}

func (p *TestPlugin) ComputeInputOutputFiles(opts plugins.GenerateOpts) *plugins.InputOutputFiles {

	input1 := util_test.WriteTempFile(p.t, "input1")
	output1 := util_test.WriteTempFile(p.t, "output1")

	ioFiles := plugins.InputOutputFiles{
		InputFiles:  []string{input1.Name()},
		OutputFiles: []string{output1.Name()},
	}
	return &ioFiles
}

func TestNoCacheVerify(t *testing.T) {
	plugins.ClearPlugins()
	testPlugin := TestPlugin{t: t}
	plugins.RegisterPlugin(&testPlugin)

	t.Chdir(t.TempDir())
	result, err := Verify(plugins.GenerateOpts{
		ExecutableName: "test",
		Path:           path.Join(t.TempDir(), "test.go"),
	})

	assert.NoError(t, err)
	assert.NotEmpty(t, result.CacheHitDir)
	assert.NoDirExists(t, result.CacheHitDir)
}
func TestCacheVerify(t *testing.T) {
	plugins.ClearPlugins()
	testPlugin := TestPlugin{t: t}
	plugins.RegisterPlugin(&testPlugin)

	// TODO do real cache verifications here

	t.Chdir(t.TempDir())
	result, err := Verify(plugins.GenerateOpts{
		ExecutableName: "test",
		Path:           path.Join(t.TempDir(), "test.go"),
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, result.CacheHitDir)
	assert.NoDirExists(t, result.CacheHitDir)
}

func newVerifyResult(t *testing.T) VerifyResult {
	file1 := util_test.WriteTempFile(t, "some-content")
	file2 := util_test.WriteTempFile(t, "some-other-content")

	tempDir := t.TempDir()
	err := os.WriteFile(path.Join(tempDir, "test1.txt"), []byte("test"), 0755)
	assert.NoError(t, err)

	return VerifyResult{
		CacheHit:    false,
		CacheHitDir: t.TempDir(),
		IoFiles: plugins.InputOutputFiles{
			InputFiles:     []string{},
			OutputFiles:    []string{file1.Name(), file2.Name()},
			OutputPatterns: []string{tempDir + "/**.txt"},
		},
	}
}

func TestSave(t *testing.T) {
	verifyRes := newVerifyResult(t)

	err := Save(verifyRes)
	assert.NoError(t, err)

	cacheConfig, err := LoadConfig(verifyRes.CacheHitDir)
	assert.NoError(t, err)

	assert.Equal(t, 2, len(verifyRes.IoFiles.OutputFiles))
	//cache shall also have the found file with the glob
	assert.Equal(t, 3, len(cacheConfig.OutputFiles))

	for index, outputFile := range verifyRes.IoFiles.OutputFiles {
		hash, err := hash.HashFile(outputFile)
		assert.NoError(t, err)

		assert.Equal(t, cacheConfig.OutputFiles[index], CacheConfigOutputFileInfo{
			Hash:    hash,
			Path:    outputFile,
			ModTime: time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC),
		})
	}
}

func TestRestore(t *testing.T) {
	//test with bundled save function
	verifyRes := newVerifyResult(t)

	err := Save(verifyRes)
	assert.NoError(t, err)

	err = Restore(verifyRes)
	assert.NoError(t, err)

	//test for file corruption
	file1 := util_test.WriteTempFile(t, "some-content")

	cacheDir := t.TempDir()

	verifyRes = VerifyResult{
		CacheHit:    false,
		CacheHitDir: cacheDir,
		IoFiles: plugins.InputOutputFiles{
			InputFiles:  []string{},
			OutputFiles: []string{file1.Name()},
		},
	}

	err = os.Rename(file1.Name(), path.Join(cacheDir, "bad_hash"))
	assert.NoError(t, err)

	cacheConfig := CacheConfig{
		OutputFiles: []CacheConfigOutputFileInfo{
			{
				Hash:    "bad_hash",
				Path:    file1.Name(),
				ModTime: time.Now(),
			},
		},
	}
	err = SaveConfig(cacheConfig, cacheDir)

	assert.NoError(t, err)

	err = Restore(verifyRes)
	assert.ErrorContains(t, err, "file hash is different, corruption")
}

func TestCalculateCacheDirectoryFromInputData(t *testing.T) {
	// Create stable test files with fixed content
	tmpDir := t.TempDir()
	inputFile := path.Join(tmpDir, "input.txt")
	err := os.WriteFile(inputFile, []byte("test content"), 0644)
	assert.NoError(t, err)

	opts := plugins.GenerateOpts{
		Words:          []string{"mockgen", "-source=input.go"},
		ExecutableName: "mockgen",
		Path:           path.Join(tmpDir, "test.go"),
	}

	ioFiles := plugins.InputOutputFiles{
		InputFiles:  []string{inputFile},
		OutputFiles: []string{"output.go"},
	}

	// Test 1: Same inputs should produce same cache directory
	dir1, err := calculateCacheDirectoryFromInputData(opts, ioFiles)
	assert.NoError(t, err)
	assert.NotEmpty(t, dir1)

	dir2, err := calculateCacheDirectoryFromInputData(opts, ioFiles)
	assert.NoError(t, err)
	assert.Equal(t, dir1, dir2, "Same inputs should produce same cache directory")

	// Test 2: Cache directory should have expected structure (cachedir/a/bc/defg...)
	assert.Contains(t, dir1, config.Get().CacheDir)
	// Should have format: cachedir/{1char}/{2char}/{rest}
	relPath := dir1[len(config.Get().CacheDir)+1:]
	parts := strings.Split(relPath, string(os.PathSeparator))
	assert.Len(t, parts, 3, "Cache directory should have 3-level structure")
	assert.Len(t, parts[0], 1, "First level should be 1 character")
	assert.Len(t, parts[1], 2, "Second level should be 2 characters")
	assert.Greater(t, len(parts[2]), 0, "Third level should have remaining hash")

	// Test 3: Different input content should produce different cache directory
	err = os.WriteFile(inputFile, []byte("different content"), 0644)
	assert.NoError(t, err)

	dir3, err := calculateCacheDirectoryFromInputData(opts, ioFiles)
	assert.NoError(t, err)
	assert.NotEqual(t, dir1, dir3, "Different input content should produce different cache directory")

	// Test 4: Different command should produce different cache directory
	opts.Words = []string{"mockgen", "-source=other.go"}
	dir4, err := calculateCacheDirectoryFromInputData(opts, ioFiles)
	assert.NoError(t, err)
	assert.NotEqual(t, dir3, dir4, "Different command should produce different cache directory")
}

func TestExecutableFileInfo(t *testing.T) {
	tmpFile := util_test.WriteTempFile(t, "test string")

	//os.LookPath needs file to be executable
	err := os.Chmod(tmpFile.Name(), 0700)
	assert.Nil(t, err, "Failed to chmod file")

	info, err := getExecutableDetails(tmpFile.Name())

	assert.NoError(t, err)
	assert.Equal(t, info, tmpFile.Name()+"00000000000000000111990-01-01T00:00:00Z")

	_, err = getExecutableDetails("bad_file")
	assert.ErrorContains(t, err, "executable file not found in $PATH")
}


func TestExecutableFileInfoGoTool(t *testing.T) {
	info, err := getExecutableDetails("go tool compile")
	assert.NoError(t, err)
	assert.NotEmpty(t, info)
}
