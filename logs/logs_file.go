// 日志输出到文件
//   变更历史
//     2017-03-01  lixiaoya  新建
package logs

import (
	"legitlab.letv.cn/uc_tp/goweb/utils"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	SizeUnit    = 1024 * 1024
	DefMaxLines = 1000000
	DefMaxSize  = 256 //256 MB 后面会统一*SizeUnit
	DefPerm     = "0660"
)

// logsFile 文件日志
type FileLogs struct {
	sync.RWMutex

	FilePath   string `json:"filepath"`
	FileName   string `json:"filename"`
	fullFile   string // 文件全路径，会在在FilePath下建一个日期[YYYYMMDD]目录
	fileWriter *os.File

	MaxLines int       `json:"maxlines"` // 文件最大行数，0不限
	curLines int       // 文件当前行数
	MaxSize  int       `json:"maxsize"` // 文件最大大小，单位M，0不限
	curSize  int       // 文件当前大小
	Perm     string    `json:"perm"` // 文件权限
	openTime time.Time // 当前文件的打开的时间

	Level    int  `json:"level"`
	ShowCall bool `json:"showcall"`
	Depth    int  `json:"depth"`
}

// newFileWriter create a FileLogWriter returning as LoggerInterface.
func newLogsFile() Logger {
	l := &FileLogs{
		MaxLines: DefMaxLines,
		MaxSize:  DefMaxSize * SizeUnit,
		Perm:     DefPerm,
		Level:    LevelDebug,
		ShowCall: false,
		Depth:    DefDepth,
	}
	return l
}

// 初始化
//   参数
//     config: 配置json串，如：
//	     {
//       "filepath":"/tmp/log"
//	     "filename":"web.log",
//	     "maxLines":10000, 为0时不判断
//	     "maxsize":500,    单位为M，为0时不判断
//       "perm":"0600"
//	     }
//   返回
//     成功返回nil，失败返回错误信息
func (l *FileLogs) Init(config string) error {
	// 初始化配置文件
	l.MaxLines = DefMaxLines
	l.MaxSize = DefMaxSize
	l.Perm = DefPerm
	l.Level = LevelDebug
	l.ShowCall = false
	l.Depth = DefDepth

	if len(config) == 0 {
		l.MaxSize = l.MaxSize * SizeUnit
		return nil
	}

	err := json.Unmarshal([]byte(config), l)
	if err != nil {
		return err
	}

	if l.FilePath == "" && l.FileName == "" {
		return errors.New("FileLogs: filepath or filename is empty")
	}

	l.MaxSize = l.MaxSize * SizeUnit

	// 初始化日志文件
	err = l.startLogger()

	return nil
}

// initLogger 开始一个新的日志文件
//   参数
//
//   返回
//     成功返回nil，失败返回错误信息
func (l *FileLogs) startLogger() error {
	f, err := l.createLogFile()
	if err != nil {
		return err
	}

	if l.fileWriter != nil {
		l.fileWriter.Close()
	}

	l.fileWriter = f

	return l.initFd()
}

// initFd 初始化日志文件信息
//   参数
//
//   返回
//     成功返回nil，失败返回错误信息
func (l *FileLogs) initFd() error {
	fd := l.fileWriter
	fInfo, err := fd.Stat()
	if err != nil {
		return fmt.Errorf("FileLogs: Get stat err: %s\n", err)
	}
	l.curSize = int(fInfo.Size())
	l.openTime = time.Now()
	if fInfo.Size() > 0 {
		count, err := l.lines()
		if err != nil {
			return err
		}
		l.curLines = count
	} else {
		l.curLines = 0
	}

	// 记一个协程，定时更新超过一天的日志，暂时未启动，让写日志时自动处理
	//go l.dailyChange(l.openTime)

	return nil
}

// dailyChange 每天0点进行日志更新
//   参数
//     openTime: 日志打开时间
//   返回
//
func (l *FileLogs) dailyChange(openTime time.Time) {
	y, m, d := openTime.Add(24 * time.Hour).Date()
	nextDay := time.Date(y, m, d, 0, 0, 0, 0, openTime.Location())
	tm := time.NewTimer(time.Duration(nextDay.UnixNano() - openTime.UnixNano() + 100))
	select {
	case <-tm.C:
		l.Lock()
		if l.needChange(time.Now().Day()) {
			if err := l.doChange(); err != nil {
				fmt.Fprintf(os.Stderr, "FileLogs: %q[%s]\n", l.fullFile, err)
			}
		}
		l.Unlock()
	}
}

// lines 返回文件行数
//   参数
//
//   返回
//     成功返回文件行数，失败返回错误信息
func (l *FileLogs) lines() (int, error) {
	fd, err := os.Open(l.fullFile)
	if err != nil {
		return 0, err
	}
	defer fd.Close()

	buf := make([]byte, 32768) // 32k
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := fd.Read(buf)
		if err != nil && err != io.EOF {
			return count, err
		}

		count += bytes.Count(buf[:c], lineSep)

		if err == io.EOF {
			break
		}
	}

	return count, nil
}

// needChange 是否需要更新文件
//   参数
//     day: 天数
//   返回
//     true-需要更新文件，false-不需要更新文件
func (l *FileLogs) needChange(day int) bool {
	return (l.MaxLines > 0 && l.curLines >= l.MaxLines) ||
		(l.MaxSize > 0 && l.curSize >= l.MaxSize) ||
		(day != l.openTime.Day())
}

// backFile 备份日志文件
//   参数
//
//   返回
//     成功返回nil，失败返回错误信息
func (l *FileLogs) backFile() error {
	var err error

	_, err = os.Lstat(l.fullFile)
	if err != nil {
		return nil
	}

	num := 1
	fName := ""
	for ; err == nil && num <= 9999; num++ {
		fName = fmt.Sprintf("%s_%04d", l.fullFile, num)
		_, err = os.Lstat(fName)
		if err != nil {
			goto RENAME
		}
	}

	return fmt.Errorf("FileLogs: Cannot find free log number to rename %s\n", l.fullFile)

RENAME:
	l.fileWriter.Close()
	err = os.Rename(l.fullFile, fName)

	return err
}

// doChange 更新新文件写日志
//   参数
//
//   返回
//     成功返回nil，失败返回错误信息
func (l *FileLogs) doChange() error {
	err := l.backFile()
	if err != nil {
		return err
	}

	err = l.startLogger()
	if err != nil {
		return fmt.Errorf("FileLogs: startLogger error %s\n", err)
	}

	return err
}

// createLogFile 创建新文件
//   参数
//
//   返回
//     成功返回文件句柄，失败返回错误信息
func (l *FileLogs) createLogFile() (*os.File, error) {
	perm, err := strconv.ParseInt(l.Perm, 8, 64)
	if err != nil {
		return nil, err
	}

	// 拼文件路径
	l.fullFile = path.Join(l.FilePath, utils.CurTime("060102"), l.FileName)
	err = utils.MkDir(l.fullFile, 0760, true)
	if err != nil {
		return nil, err
	}

	fd, err := os.OpenFile(l.fullFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, os.FileMode(perm))
	if err == nil {
		os.Chmod(l.fullFile, os.FileMode(perm))
	}

	return fd, err
}

// WriteMsg 写日志
//   参数
//     level:  日志级别
//     fmtStr: 格式串
//     v:      参数
//   返回
//     成功返回nil，失败返回错误信息
func (l *FileLogs) WriteMsg(level int, fmtStr string, v ...interface{}) error {
	if level < l.Level {
		return nil
	}

	now := time.Now()
	curTime := now.Format("2006-01-02 15:04:05")
	strMsg := fmt.Sprintf(fmtStr, v...)
	levelName := getLevelName(level)
	var arrMsg []string
	arrMsg = append(arrMsg, curTime)
	arrMsg = append(arrMsg, "["+levelName+"]")
	arrMsg = append(arrMsg, strMsg)
	if l.ShowCall {
		file, line := utils.GetCall(l.Depth)
		arrMsg = append(arrMsg, fmt.Sprintf("(%s:%d)", file, line))
	}
	msg := strings.Join(arrMsg, MsgSep)
	msg = msg + "\n"

	l.RLock()
	if l.needChange(now.Day()) {
		l.RUnlock()
		l.Lock()
		if l.needChange(now.Day()) {
			if err := l.doChange(); err != nil {
				fmt.Fprintf(os.Stderr, "FileLogs: %q[%s]\n", l.fullFile, err)
			}
		}
		l.Unlock()
	} else {
		l.RUnlock()
	}

	l.Lock()
	_, err := l.fileWriter.Write([]byte(msg))
	if err == nil {
		l.curLines++
		l.curSize += len(msg)
	}
	l.Unlock()

	return nil
}

// Debug Debug日志
//   参数
//     v: 参数
//   返回
//
func (l *FileLogs) Debug(v ...interface{}) {
	fmtStr := formatMsg(len(v))
	l.WriteMsg(LevelDebug, fmtStr, v...)
}

// Info Info日志
//   参数
//     v: 参数
//   返回
//
func (l *FileLogs) Info(v ...interface{}) {
	fmtStr := formatMsg(len(v))
	l.WriteMsg(LevelInfo, fmtStr, v...)
}

// Warn Warn日志
//   参数
//     v: 参数
//   返回
//
func (l *FileLogs) Warn(v ...interface{}) {
	fmtStr := formatMsg(len(v))
	l.WriteMsg(LevelWarn, fmtStr, v...)
}

// Error Error日志
//   参数
//     v: 参数
//   返回
//
func (l *FileLogs) Error(v ...interface{}) {
	fmtStr := formatMsg(len(v))
	l.WriteMsg(LevelError, fmtStr, v...)
}

// Fatal Fatal日志
//   参数
//     v: 参数
//   返回
//
func (l *FileLogs) Fatal(v ...interface{}) {
	fmtStr := formatMsg(len(v))
	l.WriteMsg(LevelFatal, fmtStr, v...)
}

// Debugf Debug日志
//   参数
//     fmtStr: 格式串
//     v:      参数
//   返回
//
func (l *FileLogs) Debugf(fmtStr string, v ...interface{}) {
	l.WriteMsg(LevelDebug, fmtStr, v...)
}

// Infof Info日志
//   参数
//     fmtStr: 格式串
//     v:      参数
//   返回
//
func (l *FileLogs) Infof(fmtStr string, v ...interface{}) {
	l.WriteMsg(LevelInfo, fmtStr, v...)
}

// Warnf Warn日志
//   参数
//     fmtStr: 格式串
//     v:      参数
//   返回
//
func (l *FileLogs) Warnf(fmtStr string, v ...interface{}) {
	l.WriteMsg(LevelWarn, fmtStr, v...)
}

// Errorf Error日志
//   参数
//     fmtStr: 格式串
//     v:      参数
//   返回
//
func (l *FileLogs) Errorf(fmtStr string, v ...interface{}) {
	l.WriteMsg(LevelError, fmtStr, v...)
}

// Fatalf Fatal日志
//   参数
//     fmtStr: 格式串
//     v:      参数
//   返回
//
func (l *FileLogs) Fatalf(fmtStr string, v ...interface{}) {
	l.WriteMsg(LevelFatal, fmtStr, v...)
}

// Destroy 关闭文件
//   参数
//
//   返回
//
func (l *FileLogs) Destroy() {
	l.fileWriter.Close()
	l.fileWriter = nil
}

// Flush flush文件
//   参数
//
//   返回
//
func (l *FileLogs) Flush() {
	l.fileWriter.Sync()
}

// init 注册adapter
//   参数
//
//   返回
//
func init() {
	Register(AdapterFile, newLogsFile())
	Register(AdapterFrame, newLogsFile())
}
