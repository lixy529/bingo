// 日志输出到终端
//   变更历史
//     2017-03-01  lixiaoya  新建
package logs

import (
	"github.com/lixy529/bingo/utils"
	"encoding/json"
	"fmt"
)

type ConsoleLogs struct {
	Level    int  `json:"level"`    // 日志级别
	ShowCall bool `json:"showcall"` // 是否显示调用代码的文件名和行数
	Depth    int  `json:"depth"`    // 调用函数深度
}

// 初始化
//   参数
//     config: 配置json串，如：
//      {
//      "level":1,       级别
//      "showcall":true, 是否显示调用堆栈信息（文件名和行数）
//      "depth":3        调用深度
//      }
//   返回
//     成功返回nil，失败返回错误信息
func (l *ConsoleLogs) Init(config string) error {
	l.Level = LevelDebug
	l.ShowCall = false
	l.Depth = DefDepth

	if len(config) == 0 {
		return nil
	}

	err := json.Unmarshal([]byte(config), l)
	if err != nil {
		return err
	}

	return nil
}

// WriteMsg 写日志
//   参数
//     level:  日志级别
//     fmtStr: 格式串
//     v:      参数
//   返回
//     成功返回nil，失败返回错误信息
func (l *ConsoleLogs) WriteMsg(level int, fmtStr string, v ...interface{}) error {
	if level < l.Level {
		return nil
	}

	msg := fmt.Sprintf(fmtStr, v...)
	curTime := utils.CurTime()
	levelName := getLevelName(level)

	strTmpFmt := "%s" + MsgSep + "[%s]" + MsgSep

	if l.ShowCall {
		file, line := utils.GetCall(l.Depth)
		if level == LevelWarn {
			fmt.Printf(strTmpFmt+"%c[1;00;33m%s%c[0m"+MsgSep+"(%s:%d)\n", curTime, levelName, 0x1B, msg, 0x1B, file, line)
		} else if level > LevelWarn {
			fmt.Printf(strTmpFmt+"%c[1;00;31m%s%c[0m"+MsgSep+"(%s:%d)\n", curTime, levelName, 0x1B, msg, 0x1B, file, line)
		} else {
			fmt.Printf(strTmpFmt+"%s"+MsgSep+"(%s:%d)\n", curTime, levelName, msg, file, line)
		}
	} else {
		if level == LevelWarn {
			fmt.Printf(strTmpFmt+"%c[1;00;33m%s%c[0m\n", curTime, levelName, 0x1B, msg, 0x1B)
		} else if level > LevelWarn {
			fmt.Printf(strTmpFmt+"%c[1;00;31m%s%c[0m\n", curTime, levelName, 0x1B, msg, 0x1B)
		} else {
			fmt.Printf(strTmpFmt+"%s\n", curTime, levelName, msg)
		}
	}

	return nil
}

// Debug Debug日志
//   参数
//     v: 参数
//   返回
//
func (l *ConsoleLogs) Debug(v ...interface{}) {
	fmtStr := formatMsg(len(v))
	l.WriteMsg(LevelDebug, fmtStr, v...)
}

// Info Info日志
//   参数
//     v: 参数
//   返回
//
func (l *ConsoleLogs) Info(v ...interface{}) {
	fmtStr := formatMsg(len(v))
	l.WriteMsg(LevelInfo, fmtStr, v...)
}

// Warn Warn日志
//   参数
//     v: 参数
//   返回
//
func (l *ConsoleLogs) Warn(v ...interface{}) {
	fmtStr := formatMsg(len(v))
	l.WriteMsg(LevelWarn, fmtStr, v...)
}

// Error Error日志
func (l *ConsoleLogs) Error(v ...interface{}) {
	fmtStr := formatMsg(len(v))
	l.WriteMsg(LevelError, fmtStr, v...)
}

// Fatal Fatal日志
//   参数
//     v: 参数
//   返回
//
func (l *ConsoleLogs) Fatal(v ...interface{}) {
	fmtStr := formatMsg(len(v))
	l.WriteMsg(LevelFatal, fmtStr, v...)
}

// Debugf Debug日志
//   参数
//     fmtStr: 格式串
//     v:      参数
//   返回
//
func (l *ConsoleLogs) Debugf(fmtStr string, v ...interface{}) {
	l.WriteMsg(LevelDebug, fmtStr, v...)
}

// Infof Info日志
//   参数
//     fmtStr: 格式串
//     v:      参数
//   返回
//
func (l *ConsoleLogs) Infof(fmtStr string, v ...interface{}) {
	l.WriteMsg(LevelInfo, fmtStr, v...)
}

// Warnf Warn日志
//   参数
//     fmtStr: 格式串
//     v:      参数
//   返回
//
func (l *ConsoleLogs) Warnf(fmtStr string, v ...interface{}) {
	l.WriteMsg(LevelWarn, fmtStr, v...)
}

// Errorf Error日志
//   参数
//     fmtStr: 格式串
//     v:      参数
//   返回
//
func (l *ConsoleLogs) Errorf(fmtStr string, v ...interface{}) {
	l.WriteMsg(LevelError, fmtStr, v...)
}

// Fatalf Fatal日志
//   参数
//     fmtStr: 格式串
//     v:      参数
//   返回
//
func (l *ConsoleLogs) Fatalf(fmtStr string, v ...interface{}) {
	l.WriteMsg(LevelFatal, fmtStr, v...)
}

// Destroy Console不需要处理
//   参数
//
//   返回
//
func (l *ConsoleLogs) Destroy() {
}

// Flush Console不需要处理
//   参数
//
//   返回
//
func (l *ConsoleLogs) Flush() {
}

// init 注册adapter
//   参数
//
//   返回
//
func init() {
	Register(AdapterConsole, &ConsoleLogs{Level: LevelDebug})
}
