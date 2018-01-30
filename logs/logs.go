// 日志处理
//   变更历史
//     2017-03-01  lixiaoya  新建
package logs

// 日志级别
const (
	LevelDebug = 1
	LevelInfo  = 2
	LevelWarn  = 3
	LevelError = 4
	LevelFatal = 5
)

const (
	DebugName = "DEBUG"
	InfoName  = "INFO"
	WarnName  = "WARN"
	ErrorName = "ERROR"
	FatalName = "FATAL"
)

// 日志输出类型
const (
	AdapterFrame    = "frame" // 这个也是写文件，给框架使用
	AdapterConsole  = "console"
	AdapterFile     = "file"
	AdapterSyslogNg = "syslog"
)

const (
	MsgSep   = "\t"
	DefDepth = 3
)

// LoggerInter 日志接口
type Logger interface {
	Init(config string) error
	Destroy()
	Flush()
	WriteMsg(level int, fmtStr string, v ...interface{}) error

	Debug(v ...interface{})
	Info(v ...interface{})
	Warn(v ...interface{})
	Error(v ...interface{})
	Fatal(v ...interface{})

	Debugf(fmtStr string, v ...interface{})
	Infof(fmtStr string, v ...interface{})
	Warnf(fmtStr string, v ...interface{})
	Errorf(fmtStr string, v ...interface{})
	Fatalf(fmtStr string, v ...interface{})
}

var adapters = make(map[string]Logger)

// Register 注册一个适配器
//   参数
//     name: 适配器名称
//     log:  适配器对象
//   返回
//
func Register(name string, log Logger) {
	if log == nil {
		panic("logs: Register provide is nil")
	}
	if _, dup := adapters[name]; dup {
		panic("logs: Register called twice for provider " + name)
	}
	adapters[name] = log
}

// Log 获取一个Logger
//   参数
//     name: 适配器名称
//   返回
//     适配器对象
func Log(name string) Logger {
	logger, ok := adapters[name]
	if ok {
		return logger
	}

	return nil
}

// GetLevelNameById 根据日志级别ID获取级别名称
//   参数
//     levelId: 日志级别ID
//   返回
//     日志级别名称
func GetLevelNameById(levelId int) string {
	if levelId == LevelDebug {
		return DebugName
	} else if levelId == LevelInfo {
		return InfoName
	} else if levelId == LevelWarn {
		return WarnName
	}  else if levelId == LevelError {
		return ErrorName
	}  else if levelId == LevelFatal {
		return FatalName
	}

	return ""
}

// GetLevelIdByName 根据日志级别名称获取级别ID
//   参数
//     levelName: 日志级别名称
//   返回
//     日志级别ID
func GetLevelIdByName(levelName string) int {
	if levelName == DebugName {
		return LevelDebug
	} else if levelName == InfoName {
		return LevelInfo
	} else if levelName == WarnName {
		return LevelWarn
	}  else if levelName == ErrorName {
		return LevelError
	}  else if levelName == FatalName {
		return LevelFatal
	}

	return -1
}
