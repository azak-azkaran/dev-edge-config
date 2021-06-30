package main

import (
	"testing"

	"go.uber.org/zap"
)

func CleanUpOpenshiftTest() {
}

func TestOpenshiftConfig(t *testing.T) {
	t.Cleanup(CleanUpOpenshiftTest)
	LogConfig.Level.SetLevel(zap.DebugLevel)
	Sugar.Info("Test: UpdateGlobalConfig")

}
