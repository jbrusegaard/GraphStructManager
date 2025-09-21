package log

import (
	"testing"
)

func TestLog(t *testing.T) {
	logger := InitializeLogger()
	logger.Info("Hello, world!")
}
