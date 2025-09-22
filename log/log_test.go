package log

import (
	"testing"
)

func TestLog(t *testing.T) {
	t.Parallel()
	logger := InitializeLogger()
	logger.Info("Hello, world!")
}
