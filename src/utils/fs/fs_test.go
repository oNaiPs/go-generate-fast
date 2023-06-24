package fs

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_IsDir(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "Test Directory Exists",
			path: t.TempDir(),
			want: true,
		},
		{
			name: "Test Directory Does Not Exist",
			path: "/path/that/does/not/exist",
			want: false,
		},
		{
			name: "Test Not Directory",
			path: "fs_test.go", // this file
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsDir(tt.path))
		})
	}
}

func TestFindExecutablePath(t *testing.T) {
	t.Run("test with empty executable", func(t *testing.T) {
		_, err := FindExecutablePath("")
		require.Error(t, err, "Expected error, got nil.")
	})

	t.Run("test with non-path executable", func(t *testing.T) {
		_, err := FindExecutablePath("non-exist")
		require.Error(t, err, "Expected error for non-existent executable, got nil.")
	})

	t.Run("test with go executable in GOROOT/bin", func(t *testing.T) {
		_, err := FindExecutablePath("go")
		require.NoError(t, err, "Expected nil, got error: %v.", err)
	})

	t.Run("test with path separator", func(t *testing.T) {
		execPath, _ := FindExecutablePath(string(os.PathSeparator) + "usr" + string(os.PathSeparator) + "bin" + string(os.PathSeparator) + "env")
		require.NotEmpty(t, execPath, "Expected path, got empty string.")
	})
}
