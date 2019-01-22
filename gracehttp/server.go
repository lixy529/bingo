// http 服务相关
//   变更历史
//     2017-02-06  lixiaoya  新建
package gracehttp

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/lixy529/gotools/utils"
	"log"
	"net"
	"net/http"
	"os"
	"syscall"
	"time"
)

type Server struct {
	httpServer *http.Server
	listener   net.Listener
	tlsConfig  *tls.Config

	isGraceful  bool
	shutTimeout time.Duration // 关闭超过时间将强制关闭
	endRunning  chan bool
	isRestart   bool // 是否是重启
	isHttps     bool // 是否https
	err         error
}

// NewServer 实例化Server
//   参数
//     addr:         监听地址
//     readTimeout:  读超时时间
//     writeTimeout: 写超时时间
//     shutTimeout:  关闭监听超时时间
//   返回
//     实例化的Server对象
func NewServer(addr string, handler http.Handler, readTimeout, writeTimeout, shutTimeout time.Duration) *Server {
	isGraceful := false
	if os.Getenv(GRACEFUL_ENVIRON_KEY) != "" {
		isGraceful = true
	}

	if readTimeout <= 0 {
		readTimeout = DEFAULT_READ_TIMEOUT
	}

	if writeTimeout <= 0 {
		writeTimeout = DEFAULT_WRITE_TIMEOUT
	}

	if shutTimeout <= 0 {
		shutTimeout = DEFAULT_SHUT_TIMEOUT
	}

	return &Server{
		httpServer: &http.Server{
			Addr:    addr,
			Handler: handler,

			ReadTimeout:  readTimeout * time.Second,
			WriteTimeout: writeTimeout * time.Second,
		},
		isGraceful:  isGraceful,
		shutTimeout: shutTimeout * time.Second,
		endRunning:  make(chan bool, 1),
		isRestart:   false,
		isHttps:     false,
	}
}

// ListenAndServe 启动http监听服务
//   参数
//
//   返回
//     成功-启动监听，失败-返回错误信息
func (srv *Server) ListenAndServe() error {
	addr := srv.httpServer.Addr
	if addr == "" {
		addr = ":http"
	}

	ln, err := srv.getListener(addr)
	if err != nil {
		return err
	}

	srv.listener = NewWebTcpListener(ln)

	return srv.Serve()
}

// ListenAndServeTLS 启动https监听服务
//   参数
//     certFile: cert证书文件
//     keyFile:  key证书文件
//   返回
//     成功-启动监听，失败-返回错误信息
func (srv *Server) ListenAndServeTLS(certFile, keyFile string) error {
	srv.isHttps = true
	addr := srv.httpServer.Addr
	if addr == "" {
		addr = ":https"
	}

	srv.tlsConfig = &tls.Config{}
	if srv.httpServer.TLSConfig != nil {
		*srv.tlsConfig = *srv.httpServer.TLSConfig
	}
	if srv.tlsConfig.NextProtos == nil {
		srv.tlsConfig.NextProtos = []string{"http/1.1"}
	}

	var err error
	srv.tlsConfig.Certificates = make([]tls.Certificate, 1)
	srv.tlsConfig.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}

	ln, err := srv.getListener(addr)
	if err != nil {
		return err
	}

	srv.listener = NewWebTcpListener(ln)

	return srv.Serve()
}

// Server 启动服务
//   参数
//
//   返回
//     成功-启动监听，失败-返回错误信息
func (srv *Server) Serve() error {
	go srv.handleSignals() // 捕获信号
	go func() {
		if srv.isHttps {
			ln := tls.NewListener(srv.listener, srv.tlsConfig)
			srv.err = srv.httpServer.Serve(ln)
		} else {
			srv.err = srv.httpServer.Serve(srv.listener)
		}
		if srv.err != nil {
			log.Println(srv.err)
		}
		srv.endRunning <- true
	}() // 启动服务
	<-srv.endRunning

	// 如果是重启，先启动一个新进程，再关闭老程序，否则直接关闭老进程
	pid := os.Getpid()
	if srv.isRestart {
		err := srv.startNewProcess()
		if err != nil {
			log.Printf("GraceHttp: Start new process failed[%v], pid[%d] continue serve.\n", err, pid)
			return err
		}
	}
	ctx, _ := context.WithTimeout(context.Background(), srv.shutTimeout)
	srv.httpServer.Shutdown(ctx)
	log.Printf("GraceHttp: Listener of pid %d closed.\n", pid)

	return srv.err
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
			err = fmt.Errorf("GraceHttp: net.FileListener error: %v", err)
			return nil, err
		}
	} else {
		ln, err = net.Listen("tcp", addr)
		if err != nil {
			err = fmt.Errorf("GraceHttp: net.Listen error: %v", err)
			return nil, err
		}
	}
	return ln.(*net.TCPListener), nil
}

// startNewProcess 启动一个这样的进程
//   参数
//
//   返回
//     成功-启动一个新进程，失败-返回错误信息，继续使用老进程
func (srv *Server) startNewProcess() error {
	log.Println("Start new process begin...")

	listenerFd, err := srv.listener.(*webTcpListener).Fd()
	if err != nil {
		return fmt.Errorf("failed to get socket file descriptor: %v", err)
	}

	argv0 := os.Args[0] // 执行命令

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
		return fmt.Errorf("GraceHttp: Failed to forkexec: %v", err)
	}

	log.Printf("GraceHttp: Start new process success, pid %d.", fork)

	return nil
}

// handleSignals 捕获信号
//   参数
//
//   返回
//
func (srv *Server) handleSignals() {
	sigCode, sigName := utils.HandleSignals()
	log.Printf("GraceHttp: Pid %d received %s.\n", os.Getpid(), sigName)
	if sigCode == syscall.SIGHUP || sigCode == syscall.SIGUSR2 {
		srv.isRestart = true
		srv.endRunning <- true
	} else {
		srv.isRestart = false
		srv.endRunning <- true
	}
}
