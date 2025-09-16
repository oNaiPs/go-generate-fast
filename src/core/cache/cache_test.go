package cache

import (
	"os"
	"path"
	"plugin"
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

	err := os.Chdir(t.TempDir())
	assert.NoError(t, err)
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

	err := os.Chdir(t.TempDir())
	assert.NoError(t, err)
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

func TestGetCacheHitDir(t *testing.T) {

	// TODO not working because temp folder is used in hash and keeps changing

	// execFile := util_test.WriteTempFile(t,"bin")
	// //os.LookPath needs file to be executable
	// err = os.Chmod(execFile.Name(), 0700)
	// if err != nil {
	// 	t.Fatalf("Failed to chmod file: %s", err)
	// }

	// tmpFile := util_test.WriteTempFile(t,"test string")
	// dir, err := calculateCacheDirectoryFromInputData(plugins.GenerateOpts{
	// 	Words: []string{execFile.Name()},
	// 	Dir:   "",
	// }, plugins.InputOutputFiles{
	// 	InputFiles:  []string{tmpFile.Name()},
	// 	OutputFiles: []string{"file2"},
	// })

	// assert.NoError(t, err)
	// assert.Equal(t, dir, config.Get().CacheDir+"/9/5a/94e7bf23896c36ca7237fb6d064fb11172e121acb9eeb96e506c82bf9a06f")
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
