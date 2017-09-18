// 日志处理测试
//   变更历史
//     2017-03-01  lixiaoya  新建
package logs

import (
	"testing"
)

// TestLogger
func TestLogger(t *testing.T) {
	logger := Log("console")
	if logger == nil {
		t.Errorf("Logger get err")
	}

	err := logger.Init(`{"level":1, "showcall":true, "depth":3}`)
	if err != nil {
		t.Errorf("Logger Init err")
	}

	logger.Debug(LevelDebug, "LevelDebug")
	logger.Info(LevelInfo, "LevelInfo")
	logger.Warn(LevelWarn, "LevelWarn")
	logger.Error(LevelError, "LevelError")
	logger.Fatal(LevelFatal, "LevelFatal")
}
