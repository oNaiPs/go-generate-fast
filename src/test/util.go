//go:build test

package util_test

import (
	"os"
	"testing"
	"time"
)

func WriteTempFile(t *testing.T, content string) *os.File {
	t.Helper()

	tmpFile, err := os.CreateTemp(t.TempDir(), "go-generate-fast-test-")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %s", err)
	}

	text := []byte(content)
	if _, err = tmpFile.Write(text); err != nil {
		t.Fatalf("Failed to write to temporary file: %s", err)
	}

	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temporary file: %s", err)
	}

	desiredTime := time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC)
	err = os.Chtimes(tmpFile.Name(), desiredTime, desiredTime)
	if err != nil {
		t.Fatalf("Failed to set static time: %s", err)
	}

	return tmpFile
}
