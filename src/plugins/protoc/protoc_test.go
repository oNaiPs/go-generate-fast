package plugin_protoc

import (
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/oNaiPs/go-generate-fast/src/plugins"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProtocMatches(t *testing.T) {
	p := ProtocPlugin{}

	assert.True(t, p.Matches(plugins.GenerateOpts{
		ExecutableName: "protoc",
	}))

	// Test with non-matching args
	assert.False(t, p.Matches(plugins.GenerateOpts{
		ExecutableName: "foo",
	}))
}

func TestProtocSearchFile(t *testing.T) {
	includeDirs := []string{os.TempDir() + "path/to/include", "relative/path"}
	baseDir := t.TempDir()

	// Test with non-existent file
	filePath := "nonexistent.proto"
	assert.Empty(t, searchFile(filePath, includeDirs, baseDir))

	// Create a temporary directory for the test
	tempDir := filepath.Join(baseDir, "relative/path")
	err := os.MkdirAll(tempDir, os.ModePerm)
	assert.NoError(t, err)

	// Test with a temporary file in the second include directory
	tmpfile, err := os.Create(filepath.Join(tempDir, "temp.proto"))
	assert.NoError(t, err)
	defer func() { _ = os.Remove(tmpfile.Name()) }()

	filePath = "temp.proto"
	assert.Equal(t, tmpfile.Name(), searchFile(filePath, includeDirs, baseDir))
}

func TestProtocParseProtoFile(t *testing.T) {
	// Create a temporary file with example content
	tmpfile, err := os.CreateTemp("", "example.proto")
	require.Nil(t, err)
	defer func() { _ = os.Remove(tmpfile.Name()) }()

	exampleContent := `syntax = "proto3";
package example;
import "test/import1.proto";
import "import2.proto";
option go_package = "package1";`
	_, err = tmpfile.Write([]byte(exampleContent))
	require.Nil(t, err)

	// Test the ParseProtoFile function
	goFile, err := parseProtoFile(tmpfile.Name())
	require.Nil(t, err)

	expectedImports := []string{"test/import1.proto", "import2.proto"}
	require.Equal(t, len(expectedImports), len(goFile.Imports))

	assert.Equal(t, expectedImports, goFile.Imports)

	assert.Equal(t, "package1", goFile.GoPackage)
}

func TestProtocComputeInputOutputFiles(t *testing.T) {
	type ProtoFile struct {
		FileName string
		Content  string
	}

	testCases := []struct {
		desc                string
		opts                plugins.GenerateOpts
		protoFiles          []ProtoFile
		expectedInputFiles  []string
		expectedOutputFiles []string
	}{
		{
			desc: "Test with valid arguments and single input file (default paths)",
			opts: plugins.GenerateOpts{
				ExecutableName: "protoc",
				SanitizedArgs:  []string{"file1.proto"},
			},
			protoFiles: []ProtoFile{
				{
					FileName: "file1.proto",
					Content: `
		syntax = "proto3";
		package example;
		option go_package = "github.com/yourusername/yourpackage/example";
		message MyMessage1 {}`,
				},
			},
			expectedInputFiles:  []string{"file1.proto"},
			expectedOutputFiles: []string{"github.com/yourusername/yourpackage/example/file1.pb.go"},
		},
		{
			desc: "Test with valid arguments and single input file (paths=import)",
			opts: plugins.GenerateOpts{
				ExecutableName: "protoc",
				SanitizedArgs:  []string{"--go_opt=paths=import", "file1.proto"},
			},
			protoFiles: []ProtoFile{
				{
					FileName: "file1.proto",
					Content: `
		syntax = "proto3";
		package example;
		option go_package = "github.com/yourusername/yourpackage/example";
		message MyMessage1 {}`,
				},
			},
			expectedInputFiles:  []string{"file1.proto"},
			expectedOutputFiles: []string{"github.com/yourusername/yourpackage/example/file1.pb.go"},
		},
		{
			desc: "Test with valid arguments and single input file (specify package on go_opt)",
			opts: plugins.GenerateOpts{
				ExecutableName: "protoc",
				SanitizedArgs:  []string{"--go_opt=Mfile.proto=pkg", "--go_opt=Mfile1.proto=github.com/yourusername/yourpackage/example", "file1.proto"},
			},
			protoFiles: []ProtoFile{
				{
					FileName: "file1.proto",
					Content: `
				syntax = "proto3";
				package example;
				message MyMessage1 {}`,
				},
			},
			expectedInputFiles:  []string{"file1.proto"},
			expectedOutputFiles: []string{"github.com/yourusername/yourpackage/example/file1.pb.go"},
		},
		{
			desc: "Test with valid arguments and single input file (paths=source_relative)",
			opts: plugins.GenerateOpts{
				ExecutableName: "protoc",
				SanitizedArgs:  []string{"--go_opt=paths=source_relative", "file1.proto"},
			},
			protoFiles: []ProtoFile{
				{
					FileName: "file1.proto",
					Content: `syntax = "proto3";
				message MyMessage1 {}`,
				},
			},
			expectedInputFiles:  []string{"file1.proto"},
			expectedOutputFiles: []string{"file1.pb.go"},
		},
		{
			desc: "Case 2: Test with multiple input files in different directories",
			opts: plugins.GenerateOpts{
				ExecutableName: "protoc",
				SanitizedArgs:  []string{"file1.proto", "folder/file2.proto"},
			},
			protoFiles: []ProtoFile{
				{
					FileName: "file1.proto",
					Content: `syntax = "proto3";
				option go_package = "github.com/yourusername/yourpackage/example";
				message MyMessage1 {}`,
				},
				{
					FileName: "folder/file2.proto",
					Content: `syntax = "proto3";
				option go_package = "github.com/yourusername/yourpackage/example/folder";
				message MyMessage2 {}`,
				},
			},
			expectedInputFiles: []string{
				"file1.proto",
				"folder/file2.proto",
			},
			expectedOutputFiles: []string{
				"github.com/yourusername/yourpackage/example/file1.pb.go",
				"github.com/yourusername/yourpackage/example/folder/file2.pb.go",
			},
		},
		{
			desc: "Case 2: Test with single input file that imports another one",
			opts: plugins.GenerateOpts{
				ExecutableName: "protoc",
				SanitizedArgs:  []string{"file1.proto"},
			},
			protoFiles: []ProtoFile{
				{
					FileName: "file1.proto",
					Content: `syntax = "proto3";
		option go_package = "github.com/yourusername/yourpackage/example";
		import "folder/file2.proto";
		message MyMessage1 {}`,
				},
				{
					FileName: "folder/file2.proto",
					Content: `syntax = "proto3";
		option go_package = "github.com/yourusername/yourpackage/example/folder";
		message MyMessage2 {}`,
				},
			},
			expectedInputFiles: []string{
				"file1.proto",
				"folder/file2.proto",
			},
			expectedOutputFiles: []string{
				"github.com/yourusername/yourpackage/example/file1.pb.go",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			p := ProtocPlugin{}

			baseDir := t.TempDir()

			tc.opts.Path = path.Join(baseDir, "test.go")

			// Generate proto files with provided contents and filenames
			for _, protoFile := range tc.protoFiles {
				if filepath.IsAbs(protoFile.FileName) {
					t.Fatal("Proto file ", protoFile.FileName, " shall not be absolute")
				}
				protoPath := filepath.Join(tc.opts.Dir(), protoFile.FileName)

				err := os.MkdirAll(filepath.Dir(protoPath), os.ModePerm)
				require.Nil(t, err)
				err = os.WriteFile(protoPath, []byte(protoFile.Content), 0644)
				require.Nil(t, err)
				defer func() { _ = os.Remove(protoPath) }()
			}

			err := os.Chdir(tc.opts.Dir())
			assert.NoError(t, err)
			ioFiles := p.ComputeInputOutputFiles(tc.opts)
			assert.NotNil(t, ioFiles)

			//add absolute path to expected io files
			for i, element := range tc.expectedInputFiles {
				if !filepath.IsAbs(element) {
					tc.expectedInputFiles[i] = filepath.Join(baseDir, tc.expectedInputFiles[i])
				}
			}
			for i, element := range tc.expectedOutputFiles {
				if !filepath.IsAbs(element) {
					tc.expectedOutputFiles[i] = filepath.Join(baseDir, tc.expectedOutputFiles[i])
				}
			}

			assert.ElementsMatch(t, tc.expectedInputFiles, ioFiles.InputFiles, tc.desc)
			assert.ElementsMatch(t, tc.expectedOutputFiles, ioFiles.OutputFiles, tc.desc)
		})
	}
}
