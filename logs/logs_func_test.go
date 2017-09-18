// 日志共用函数测试
//   变更历史
//     2017-03-03  lixiaoya  新建
package logs

import (
	"testing"
)

// TestGetLevelName getLevelName测试
func TestGetLevelName(t *testing.T) {
	name := getLevelName(LevelDebug)
	if name != DebugName {
		t.Errorf("getLevelCode err, Got %s, export %s", name, DebugName)
	}

	name = getLevelName(LevelInfo)
	if name != InfoName {
		t.Errorf("getLevelCode err, Got %s, export %s", name, InfoName)
	}

	name = getLevelName(LevelWarn)
	if name != WarnName {
		t.Errorf("getLevelCode err, Got %s, export %s", name, WarnName)
	}

	name = getLevelName(LevelError)
	if name != ErrorName {
		t.Errorf("getLevelCode err, Got %s, export %s", name, ErrorName)
	}

	name = getLevelName(LevelFatal)
	if name != FatalName {
		t.Errorf("getLevelCode err, Got %s, export %s", name, FatalName)
	}
}

// TestGetLevelCode getLevelCode测试
func TestGetLevelCode(t *testing.T) {
	level := getLevelCode("debug")
	if level != LevelDebug {
		t.Errorf("getLevelCode err, Got %d, export %d", level, LevelDebug)
	}

	level = getLevelCode("info")
	if level != LevelInfo {
		t.Errorf("getLevelCode err, Got %d, export %d", level, LevelInfo)
	}

	level = getLevelCode("warn")
	if level != LevelWarn {
		t.Errorf("getLevelCode err, Got %d, export %d", level, LevelWarn)
	}

	level = getLevelCode("error")
	if level != LevelError {
		t.Errorf("getLevelCode err, Got %d, export %d", level, LevelError)
	}

	level = getLevelCode("fatal")
	if level != LevelFatal {
		t.Errorf("getLevelCode err, Got %d, export %d", level, LevelFatal)
	}
}
