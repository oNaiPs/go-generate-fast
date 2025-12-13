package copy

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCopyFile(t *testing.T) {
	content := []byte("some content\n")
	tmpfile, err := os.CreateTemp("", "example")
	assert.NoError(t, err)

	_, err = tmpfile.Write(content)
	assert.NoError(t, err)

	err = tmpfile.Close()
	assert.NoError(t, err)

	err = CopyFile(tmpfile.Name(), "testfile.tmp")
	assert.NoError(t, err)

	_, err = os.Stat("testfile.tmp")
	assert.Falsef(t, os.IsNotExist(err), "copyFile() failed: %s should have been created", "testfile.tmp")

	dat, err := os.ReadFile("testfile.tmp")
	assert.NoError(t, err)

	assert.Falsef(t, string(dat) != string(content), "copyFile() failed: content mismatch: got %v, wanted %v", string(dat), string(content))

	_ = os.Remove("testfile.tmp")
	_ = os.Remove(tmpfile.Name())
}

func TestCopyHashFile(t *testing.T) {
	content := []byte("some content\n")
	tmpfile, err := os.CreateTemp("", "example")
	assert.NoError(t, err)

	_, err = tmpfile.Write(content)
	assert.NoError(t, err)

	err = tmpfile.Close()
	assert.NoError(t, err)

	hash, err := CopyHashFile(tmpfile.Name(), "testfile.tmp")
	assert.NoError(t, err)

	assert.Equal(t, hash, "d3471f00c65d2ae6c70d94d9fdee32dc9fab025707ad5dd3c26f7d31f22a85da")

	_, err = os.Stat("testfile.tmp")
	assert.Falsef(t, os.IsNotExist(err), "copyFile() failed: %s should have been created", "testfile.tmp")

	dat, err := os.ReadFile("testfile.tmp")
	assert.NoError(t, err)

	assert.Falsef(t, string(dat) != string(content), "copyFile() failed: content mismatch: got %v, wanted %v", string(dat), string(content))

	_ = os.Remove("testfile.tmp")
	_ = os.Remove(tmpfile.Name())
}
