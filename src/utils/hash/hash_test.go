package hash

import (
	"testing"

	util_test "github.com/oNaiPs/go-generate-fast/src/test"
	"github.com/stretchr/testify/assert"
)

func TestHashString(t *testing.T) {
	hash, err := HashString("test string")
	assert.NoError(t, err)
	assert.Equal(t, hash, "7c106d42ca17fdbfb03f6b45b91effcef2cff61215a3552dbc1ab8fd46817719")
}

func TestHashFile(t *testing.T) {
	tmpFile := util_test.WriteTempFile(t, "test string")

	hash, err := HashFile(tmpFile.Name())
	assert.NoError(t, err)
	assert.Equal(t, hash, "7c106d42ca17fdbfb03f6b45b91effcef2cff61215a3552dbc1ab8fd46817719")

	_, err = HashFile("bad_file")
	assert.ErrorContains(t, err, "no such file or directory")
}
