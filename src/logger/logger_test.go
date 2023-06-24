package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestInit(t *testing.T) {
	Init()
	assert.NotNil(t, zap.S())
}

func TestInitTwice(t *testing.T) {
	Init()
	initialLogger := zap.S()
	Init()
	assert.NotEqual(t, initialLogger, zap.S())
}
