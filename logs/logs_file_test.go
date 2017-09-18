// 日志输出到文件测试
//   变更历史
//     2017-03-01  lixiaoya  新建
package logs

import (
	"fmt"
	"testing"
)

func TestFile(t *testing.T) {
	l := &FileLogs{}
	err := l.Init(`{"filepath":"./log","filename":"demo.log","maxlines":100,"maxsize":4,"perm":"0660","level":1, "showcall":true, "depth":3}`)
	if err != nil {
		t.Errorf("logFile Init failed. err:%s.", err.Error())
		return
	}

	l.Debug(LevelDebug, "LevelDebug")
	l.Info(LevelInfo, "LevelInfo")
	l.Warn(LevelWarn, "LevelWarn")
	l.Error(LevelError, "LevelError")
	l.Fatal(LevelFatal, "LevelFatal")

	l.Debugf("Debugf %d-%s", LevelDebug, "LevelDebug")
	l.Infof("Infof %d-%s", LevelInfo, "LevelInfo")
	l.Warnf("Warnf %d-%s", LevelWarn, "LevelWarn")
	l.Errorf("Errorf %d-%s", LevelError, "LevelError")
	l.Fatalf("Fatalf %d-%s", LevelFatal, "LevelFatal")

	fmt.Println(l)

	l.Flush()
	l.Destroy()
}
