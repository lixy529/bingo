// 日志输出到Syslog-ng
//   变更历史
//     2017-03-02  lixiaoya  新建
package logs

import (
	"legitlab.letv.cn/uc_tp/goweb/utils"
	"container/list"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

const (
	DefSockFile    = "/var/run/php-syslog-ng.sock"
	DefMaxConns    = 0
	DefMaxIdle     = 100
	DefIdleTimeout = 3
	DefConnTimeout = 3
)

// SyslogNgLogs Syslog-Ng日志
type SyslogNgLogs struct {
	SockFile    string        `json:"sockfile"`    // sock文件
	LocalFile   string        `json:"localfile"`   // syslog-ng异常时写本地文件
	MaxConns    int           `json:"maxconns"`    // 最大连接数，默认为0，0不限，如果小于0则不使用连接池
	MaxIdle     int           `json:"maxidle"`     // 最大空闲连接数，默认为100，0不限
	IdleTimeout time.Duration `json:"idletimeout"` // 空闲连接超时时间，默认为3秒，0不限，单位为秒
	Wait        bool          `json:"wait"`        // 没有空闲且达到最大连接数时，是否等待，默认为不等待，直接返回错误
	Level       int           `json:"level"`       // 日志级别
	ShowCall    bool          `json:"showcall"`    // 是否显示调用代码的文件名和行数
	Depth       int           `json:"depth"`       // 调用函数深度

	pool *SockPool  // 连接池
	mux  sync.Mutex // 写本地文件锁
}

// 初始化
//   参数
//     {
//     "sockfile":"/var/run/php-syslog-ng.sock",
//     "localfile":"/tmp/gomessages",
//     "maxconns":20
//     "maxidle":10
//     "idletimeout":3
//     "level":1,
//     "showcall":true, 是否显示调用堆栈信息（文件名和行数）
//     "depth":3
//     }
//   返回
//     成功返回nil，失败返回错误信息
func (l *SyslogNgLogs) Init(config string) error {
	l.SockFile = DefSockFile
	l.MaxConns = DefMaxConns
	l.MaxIdle = DefMaxIdle
	l.IdleTimeout = DefIdleTimeout
	l.Wait = false
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

	if l.SockFile == "" {
		l.SockFile = DefSockFile
	}

	isFile, err := utils.IsFile(l.SockFile)
	if err != nil {
		return err
	} else if !isFile {
		return fmt.Errorf("SyslogNgLogs: [%s] is not file.", l.SockFile)
	}

	// 初始化连接池， >=0时才使用连接池
	if l.MaxConns >= 0 {
		l.pool = NewSockPool(l.SockFile, l.MaxConns, l.MaxIdle, l.IdleTimeout, l.Wait)
		if l.pool == nil {
			return fmt.Errorf("SyslogNgLogs: NewSockPool failed [%s].", l.SockFile)
		}
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
func (l *SyslogNgLogs) WriteMsg(level int, fmtStr string, v ...interface{}) error {
	if level < l.Level {
		return nil
	}

	msg := fmt.Sprintf(fmtStr, v...)
	if l.ShowCall {
		file, line := utils.GetCall(l.Depth)
		msg += MsgSep + fmt.Sprintf("(%s:%d)", file, line)
	}
	msg += "\n"

	var conn net.Conn
	var err error
	forceClose := false

	if l.pool == nil {
		// 不使用连接池
		conn, err = net.Dial("unix", l.SockFile)
		if err != nil || conn == nil {
			l.writeLocalFile(msg)
			return err
		} else {
			conn.SetDeadline(time.Now().Add(DefConnTimeout * time.Second))
		}
		defer conn.Close()
	} else{
		// 使用连接池
		conn, err = l.pool.Get()
		defer l.pool.Put(conn, forceClose)
		if err != nil {
			l.writeLocalFile(msg)
			forceClose = true
			return err
		}
	}

	// 写日志
	_, err = conn.Write([]byte(msg))
	if err != nil {
		l.writeLocalFile(msg)
		forceClose = true
		return err
	}

	return nil
}

// writeLocalFile写本地文件
//   参数
//     msg:  日志内容
//   返回
//     成功返回nil，失败返回错误信息
func (l *SyslogNgLogs) writeLocalFile(msg string) error {
	l.mux.Lock()
	defer l.mux.Unlock()

	fd, err := os.OpenFile(l.LocalFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.FileMode(0660))
	if err != nil {
		return err
	}

	_, err = fd.Write([]byte(msg))
	fd.Close()
	return err
}

// Debug Debug日志
//   参数
//     v: 参数
//   返回
//
func (l *SyslogNgLogs) Debug(v ...interface{}) {
	fmtStr := formatMsg(len(v))
	l.WriteMsg(LevelDebug, fmtStr, v...)
}

// Info Info日志
//   参数
//     v: 参数
//   返回
//
func (l *SyslogNgLogs) Info(v ...interface{}) {
	fmtStr := formatMsg(len(v))
	l.WriteMsg(LevelInfo, fmtStr, v...)
}

// Warn Warn日志
//   参数
//     v: 参数
//   返回
//
func (l *SyslogNgLogs) Warn(v ...interface{}) {
	fmtStr := formatMsg(len(v))
	l.WriteMsg(LevelWarn, fmtStr, v...)
}

// Error Error日志
//   参数
//     v: 参数
//   返回
//
func (l *SyslogNgLogs) Error(v ...interface{}) {
	fmtStr := formatMsg(len(v))
	l.WriteMsg(LevelError, fmtStr, v...)
}

// Fatal Fatal日志
//   参数
//     v: 参数
//   返回
//
func (l *SyslogNgLogs) Fatal(v ...interface{}) {
	fmtStr := formatMsg(len(v))
	l.WriteMsg(LevelFatal, fmtStr, v...)
}

// Debugf Debug日志
//   参数
//     fmtStr: 格式串
//     v:      参数
//   返回
//
func (l *SyslogNgLogs) Debugf(fmtStr string, v ...interface{}) {
	l.WriteMsg(LevelDebug, fmtStr, v...)
}

// Infof Info日志
//   参数
//     fmtStr: 格式串
//     v:      参数
//   返回
//
func (l *SyslogNgLogs) Infof(fmtStr string, v ...interface{}) {
	l.WriteMsg(LevelInfo, fmtStr, v...)
}

// Warnf Warn日志
//   参数
//     fmtStr: 格式串
//     v:      参数
//   返回
//
func (l *SyslogNgLogs) Warnf(fmtStr string, v ...interface{}) {
	l.WriteMsg(LevelWarn, fmtStr, v...)
}

// Errorf Error日志
//   参数
//     fmtStr: 格式串
//     v:      参数
//   返回
//
func (l *SyslogNgLogs) Errorf(fmtStr string, v ...interface{}) {
	l.WriteMsg(LevelError, fmtStr, v...)
}

// Fatalf Fatal日志
//   参数
//     fmtStr: 格式串
//     v:      参数
//   返回
//
func (l *SyslogNgLogs) Fatalf(fmtStr string, v ...interface{}) {
	l.WriteMsg(LevelFatal, fmtStr, v...)
}

// Destroy Console不需要处理
//   参数
//
//   返回
//
func (l *SyslogNgLogs) Destroy() {
	if l.pool != nil {
		l.pool.Close()
	}
}

// Flush Console不需要处理
//   参数
//
//   返回
//
func (l *SyslogNgLogs) Flush() {
}

// init 注册adapter
//   参数
//
//   返回
//
func init() {
	Register(AdapterSyslogNg, &SyslogNgLogs{Level: LevelDebug})
}

var nowFunc = time.Now
var ErrPoolExhausted = errors.New("SockPool: Connection pool exhausted")

// IdleConn 空闲连接结构
type IdleConn struct {
	c net.Conn
	t time.Time
}

// SockPool Uninx sock 连接池
type SockPool struct {
	sockFile    string
	maxConns    int           // 最大连接数
	maxIdle     int           // 最大空闲连接数
	idleTimeout time.Duration // 最大空闲超时时间

	idleConns list.List

	mu       sync.Mutex
	cond     *sync.Cond
	wait     bool // 为true时，没有连接则等待，否则返回错误
	curConns int  // 打开的连接数
	closed   bool // 连接池是否已关闭
}

// NewSockPool 实例化一个连接池
//   参数
//     sockFile:    unix sock文件
//     maxConns:    最大连接数
//     maxIdle:     最大空闲连接数
//     idleTimeout: 最大空闲时间
//     wait:        没有空闲连接且已大最大连接数时，是否等待
//   返回
//
func NewSockPool(sockFile string, maxConns, maxIdle int, idleTimeout time.Duration, wait bool) *SockPool {
	return &SockPool{
		sockFile:    sockFile,
		maxConns:    maxConns,
		maxIdle:     maxIdle,
		idleTimeout: idleTimeout * time.Second,
		curConns:    0,
		closed:      false,
		wait:        wait,
	}
}

// Get 返回一个连接
//   参数
//
//   返回
//     成功时返回连接串，失败返回错误信息
func (p *SockPool) Get() (net.Conn, error) {
	p.mu.Lock()

	// 查询超时连接
	if timeout := p.idleTimeout; timeout > 0 {
		for i, n := 0, p.idleConns.Len(); i < n; i++ {
			e := p.idleConns.Back()
			if e == nil {
				break
			}
			ic := e.Value.(IdleConn)
			if ic.t.Add(timeout).After(nowFunc()) {
				break
			}
			p.idleConns.Remove(e)
			p.release()
			p.mu.Unlock()
			ic.c.Close()
			p.mu.Lock()
		}
	}

	for {
		// 如果有空闲，则取一个空闲连接
		for i, n := 0, p.idleConns.Len(); i < n; i++ {
			e := p.idleConns.Front()
			if e == nil {
				break
			}
			ic := e.Value.(IdleConn)
			p.idleConns.Remove(e)
			// todo: 可以做一个测试连接的可用性
			p.mu.Unlock()
			return ic.c, nil
		}

		// 判断连接池是否关闭
		if p.closed {
			p.mu.Unlock()
			return nil, errors.New("SockPool: Pool is closed")
		}

		// 没有空闲，但连接数小于最大连接数，则新建一个连接
		if p.maxConns == 0 || p.curConns < p.maxConns {
			p.curConns += 1
			p.mu.Unlock()
			c, err := net.Dial("unix", p.sockFile)
			if err != nil || c == nil {
				p.mu.Lock()
				p.release()
				p.mu.Unlock()
				c = nil
			} else {
				c.SetDeadline(time.Now().Add(DefConnTimeout * time.Second))
			}

			return c, err
		}

		// 如果不等待直接返回错误
		if !p.wait {
			p.mu.Unlock()
			return nil, ErrPoolExhausted
		}

		// 等待连接
		if p.cond == nil {
			p.cond = sync.NewCond(&p.mu)
		}
		p.cond.Wait()
	}

	return nil, errors.New("SockPool: error")
}

// release 释放一个连接，有过期或者新建连接失败时会调用
//   参数
//
//   返回
//
func (p *SockPool) release() {
	p.curConns -= 1
	if p.cond != nil {
		p.cond.Signal()
	}
}

// Put 将连接放回连接池
//   参数
//     c:          连接串
//     forceClose: 是否强制关闭
//   返回
//
func (p *SockPool) Put(c net.Conn, forceClose bool) {
	p.mu.Lock()
	// 如果连接为nil，则连接数-1，并唤醒信号
	if c == nil {
		p.release()
		p.mu.Unlock()
		return
	}

	isPut := false
	if !forceClose && !p.closed {
		p.idleConns.PushFront(IdleConn{t: nowFunc(), c: c})
		if p.maxIdle > 0 && p.idleConns.Len() > p.maxIdle {
			// 空闲数超出最大空闲数，则直接删除
			c = p.idleConns.Remove(p.idleConns.Back()).(IdleConn).c
		} else {
			c = nil
			isPut = true
		}
	}

	// 返回空闲池的直接唤醒
	if isPut {
		if p.cond != nil {
			p.cond.Signal()
		}
		p.mu.Unlock()
		return
	}

	p.release()
	p.mu.Unlock()

	if c != nil {
		c.Close()
	}

	return
}

// Close 关闭连接池
//   参数
//
//   返回
//
func (p *SockPool) Close() {
	p.mu.Lock()
	idle := p.idleConns
	p.idleConns.Init()
	p.closed = true
	p.curConns -= idle.Len()
	if p.cond != nil {
		p.cond.Broadcast()
	}
	p.mu.Unlock()
	for e := idle.Front(); e != nil; e = e.Next() {
		e.Value.(IdleConn).c.Close()
	}
	return
}
