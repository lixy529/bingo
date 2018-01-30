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

// TestGetLevelNameById 测试GetLevelNameById
func TestGetLevelNameById(t *testing.T) {
	id := LevelDebug
	name := GetLevelNameById(id)
	if name != DebugName {
		t.Errorf("GetLevleNameById err, Got: %s, export: %s", name, DebugName)
		return
	}

	id = LevelInfo
	name = GetLevelNameById(id)
	if name != InfoName {
		t.Errorf("GetLevleNameById err, Got: %s, export: %s", name, InfoName)
		return
	}

	id = LevelWarn
	name = GetLevelNameById(id)
	if name != WarnName {
		t.Errorf("GetLevleNameById err, Got: %s, export: %s", name, WarnName)
		return
	}

	id = LevelError
	name = GetLevelNameById(id)
	if name != ErrorName {
		t.Errorf("GetLevleNameById err, Got: %s, export: %s", name, ErrorName)
		return
	}

	id = LevelFatal
	name = GetLevelNameById(id)
	if name != FatalName {
		t.Errorf("GetLevleNameById err, Got: %s, export: %s", name, FatalName)
		return
	}
}

// TestGetLevelIdByName 测试GetLevelIdByName
func TestGetLevelIdByName(t *testing.T) {
	name := DebugName
	id := GetLevelIdByName(name)
	if id != LevelDebug {
		t.Errorf("GetLevleNameById err, Got: %d, export: %d", id, LevelDebug)
		return
	}

	name = InfoName
	id = GetLevelIdByName(name)
	if id != LevelInfo {
		t.Errorf("GetLevleNameById err, Got: %d, export: %d", id, LevelInfo)
		return
	}

	name = WarnName
	id = GetLevelIdByName(name)
	if id != LevelWarn {
		t.Errorf("GetLevleNameById err, Got: %d, export: %d", id, LevelWarn)
		return
	}

	name = ErrorName
	id = GetLevelIdByName(name)
	if id != LevelError {
		t.Errorf("GetLevleNameById err, Got: %d, export: %d", id, LevelError)
		return
	}

	name = FatalName
	id = GetLevelIdByName(name)
	if id != LevelFatal {
		t.Errorf("GetLevleNameById err, Got: %d, export: %d", id, LevelFatal)
		return
	}
}
