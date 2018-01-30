// FastCgi服务
// fcgi.go和child.go从golang源码拷贝过来
//   变更历史
//     2017-03-14  lixiaoya  新建
package gracefcgi

import (
	"fmt"
	"github.com/bingo/utils"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"syscall"
	"time"
)

var (
	wg sync.WaitGroup
)

type Server struct {
	addr        string
	port        int
	listener    net.Listener
	handler     http.Handler
	shutTimeout time.Duration // 关闭超过时间将强制关闭

	isGraceful bool
	endRunning chan bool
	isStop     bool
	isRestart  bool
}

func NewServer(addr string, port int, handler http.Handler, shutTimeout time.Duration) *Server {
	isGraceful := false
	if os.Getenv(GRACEFUL_ENVIRON_KEY) != "" {
		isGraceful = true
	}

	if handler == nil {
		handler = http.DefaultServeMux
	}

	if shutTimeout <= 0 {
		shutTimeout = DEFAULT_SHUT_TIMEOUT
	}

	return &Server{
		addr:        addr,
		port:        port,
		handler:     handler,
		shutTimeout: shutTimeout * time.Second,

		isGraceful: isGraceful,
		endRunning: make(chan bool, 1),
		isStop:     false,
		isRestart:  false,
	}
}

// ListenAndServe 启动监听和服务
//   参数
//     addr: 为空时为标准I/O，当port>0j时为监听的ip地址，否则为sock file
//   返回
//     成功返回nil，失败返回错误信息
func (srv *Server) ListenAndServe() error {
	var err error

	if srv.addr == "" {
		srv.listener = nil
	} else if srv.port > 0 {
		laddr := fmt.Sprintf("%s:%d", srv.addr, srv.port)
		srv.listener, err = srv.getListener(laddr)
		if err != nil {
			return err
		}
	} else {
		_, err = os.Stat(srv.addr)
		if err == nil {
			os.Remove(srv.addr)
		}

		srv.listener, err = net.Listen("unix", srv.addr)
		if err != nil {
			return err
		}

		// 修改文件权限
		err = os.Chmod(srv.addr, 0777)
		if err != nil {
			return err
		}
	}

	return srv.Start()
}

// getListener 获取监听
// 如果是重启则直接从fd接收监听
//   参数
//     addr:    监听地址
//   返回
//     成功-监听信息，失败-返回错误信息
func (srv *Server) getListener(addr string) (*net.TCPListener, error) {
	var ln net.Listener
	var err error

	if srv.isGraceful {
		file := os.NewFile(3, "")
		ln, err = net.FileListener(file)
		if err != nil {
			err = fmt.Errorf("GraceFcgi: net.FileListener error: %v", err)
			return nil, err
		}
	} else {
		ln, err = net.Listen("tcp", addr)
		if err != nil {
			err = fmt.Errorf("GraceFcgi: net.Listen error: %v", err)
			return nil, err
		}
	}
	return ln.(*net.TCPListener), nil
}

// Serve 启动服务
//   参数
//
//   返回
//
func (srv *Server) Start() error {
	go srv.handleSignals()
	go srv.Serve(srv.listener, srv.handler)

	<-srv.endRunning

	// 启动一个子进程
	if srv.isRestart {
		srv.startNewProcess()
	}

	// 等待业务处理结束
	chanStop := make(chan bool)
	go func() {
		wg.Wait()
		close(chanStop)
	}()

	select {
	case <-time.After(srv.shutTimeout):
	case <-chanStop:
	}

	return nil
}

// Serve 启动服务
//   参数
//
//   返回
//
func (srv *Server) Serve(l net.Listener, handler http.Handler) error {
	if l == nil {
		var err error
		l, err = net.FileListener(os.Stdin)
		if err != nil {
			return err
		}
		defer l.Close()
	}
	if handler == nil {
		handler = http.DefaultServeMux
	}
	for {
		rw, err := l.Accept()
		if err != nil {
			return err
		}
		c := newChild(rw, handler)
		wg.Add(1)
		go c.serve()
	}
}

// startNewProcess 启动一个这样的进程
//   参数
//
//   返回
//     成功-启动一个新进程，失败-返回错误信息，继续使用老进程
func (srv *Server) startNewProcess() error {
	log.Println("GraceFcgi: Start new process begin...")

	argv0 := os.Args[0] // 执行命令
	var listenerFd uintptr

	if srv.port > 0 {
		//listenerFd, err := srv.listener.(*webTcpListener).Fd()
		f, err := srv.listener.(*net.TCPListener).File()
		if err != nil {
			return err
		}
		listenerFd = f.Fd()
	}

	// 设置标识优雅重启的环境变量
	environList := []string{}
	for _, value := range os.Environ() {
		if value != GRACEFUL_ENVIRON_STRING {
			environList = append(environList, value)
		}
	}
	environList = append(environList, GRACEFUL_ENVIRON_STRING)

	execSpec := &syscall.ProcAttr{
		Env:   environList,
		Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd(), listenerFd},
	}

	fork, err := syscall.ForkExec(argv0, os.Args, execSpec)
	if err != nil {
		return fmt.Errorf("GraceFcgi: Failed to forkexec: %v", err)
	}

	log.Printf("GraceFcgi: Start new process success, pid %d.", fork)

	return nil
}

// handleSignals 捕获信号
//   参数
//
//   返回
//
func (srv *Server) handleSignals() {
	sigCode, sigName := utils.HandleSignals()
	log.Printf("GraceFcgi: Pid %d received %s.\n", os.Getpid(), sigName)
	if sigCode == syscall.SIGHUP || sigCode == syscall.SIGUSR2 {
		srv.isStop = true
		srv.isRestart = true
		srv.endRunning <- true
	} else {
		srv.isStop = true
		srv.isRestart = false
		srv.endRunning <- true
	}
}
