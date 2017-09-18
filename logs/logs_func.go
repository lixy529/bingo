// logs的公用函数
//   变更历史
//     2017-03-01  lixiaoya  新建
package logs

import (
	"strings"
)

// formatMsg 生成格式化的串
func formatMsg(n int) string {
	return strings.TrimRight(strings.Repeat("%v"+MsgSep, n), MsgSep)
}

// getLevelName 获取级别对应的名称
//   参数
//     level: 日志级别编码
//   返回
//     日志级别名称
func getLevelName(level int) string {
	switch level {
	case LevelDebug:
		return DebugName
	case LevelInfo:
		return InfoName
	case LevelWarn:
		return WarnName
	case LevelError:
		return ErrorName
	case LevelFatal:
		return FatalName
	default:
		return InfoName
	}
}

// getLevelCode 获取级别对应的编号
//   参数
//     name: 日志级别名称
//   返回
//     日志级别编码
func getLevelCode(name string) int {
	switch strings.ToUpper(name) {
	case DebugName:
		return LevelDebug
	case InfoName:
		return LevelInfo
	case WarnName:
		return LevelWarn
	case ErrorName:
		return LevelError
	case FatalName:
		return LevelFatal
	default:
		return LevelInfo
	}
}
