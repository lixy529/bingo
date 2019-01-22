// 日志输出到Syslog-ng测试
//   变更历史
//     2017-03-02  lixiaoya  新建
package logs

import (
	"fmt"
	"github.com/lixy529/gotools/utils"
	"sync"
	"testing"
	"time"
)

func TestSyslogNg(t *testing.T) {
	l := &SyslogNgLogs{}
	err := l.Init(`{"sockfile":"/var/run/php-syslog-ng.sock", "localfile":"/tmp/gomessages", "maxconns":20, "maxidle":10, "idletimeout":3, "level":1, "showcall":true, "depth":3}`)
	if err != nil {
		t.Errorf("Init error: %s", err.Error())
		return
	}

	l.Debug("<182>" + "sso" + "[10001]:" + "usrlog|" + "LevelDebug")
	l.Info("<182>" + "sso" + "[10002]:" + "usrlog|" + "LevelInfo")
	l.Warn("<182>" + "sso" + "[10003]:" + "usrlog|" + "LevelWarn")
	l.Error("<182>" + "sso" + "[10004]:" + "usrlog|" + "LevelError")
	l.Fatal("<182>" + "sso" + "[10005]:" + "usrlog|" + "LevelFatal")

	l.Debugf("<182>sso[%d]:%s|%s\t[%s]\t%s", 20001, "usrlog", utils.CurTime(), "DEBUG", "LevelDebug")
	l.Infof("<182>sso[%d]:%s|%s\t[%s]\t%s", 20002, "usrlog", utils.CurTime(), "INFO", "LevelInfo")
	l.Warnf("<182>sso[%d]:%s|%s\t[%s]\t%s", 20003, "usrlog", utils.CurTime(), "WARN", "LevelWarn")
	l.Errorf("<182>sso[%d]:%s|%s\t[%s]\t%s", 20004, "usrlog", utils.CurTime(), "ERROR", "LevelError")
	l.Fatalf("<182>sso[%d]:%s|%s\t[%s]\t%s", 20005, "usrlog", utils.CurTime(), "FATAL", "LevelFatal")
}

var wg sync.WaitGroup

// TestSockPool 连接池测试
func TestSockPool(t *testing.T) {
	return
	sockFile := "/tmp/echo.sock"
	maxConns := 10
	maxIdle := 5
	idleTimeout := 3
	wait := true
	p := NewSockPool(sockFile, maxConns, maxIdle, time.Duration(idleTimeout), wait)
	if p == nil {
		t.Error("NewSockPool error")
		return
	}

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go sendMsg(p, i)
	}
	time.Sleep(1 * time.Second)

	wg.Wait()

	//time.Sleep(3*time.Second)
	//p.Close()
}

func sendMsg(p *SockPool, i int) {
	defer wg.Done()
	forceClose := false
	c, err := p.Get()
	defer p.Put(c, forceClose)
	if err != nil {
		fmt.Printf("SockPool >>> get err %s\n", err.Error())
		forceClose = true
		return
	}

	//msg := fmt.Sprintf("Hi server [%d]", i)
	msg := fmt.Sprintf("Hi[%d]", i)
	//fmt.Printf(msg + "\n")

	_, err = c.Write([]byte(msg))
	if err != nil {
		forceClose = true
		fmt.Printf("SockPool >>> write error:%s", err.Error())
	}
	//fmt.Println(p)
}
