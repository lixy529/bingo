// 原生服务器类
//   变更历史
//     2017-02-07  lixiaoya  新建
package bingo

import (
	"context"
	"github.com/lixy529/gotools/utils"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	DEFAULT_READ_TIMEOUT  = 60
	DEFAULT_WRITE_TIMEOUT = DEFAULT_READ_TIMEOUT
	DEFAULT_SHUT_TIMEOUT  = 20
)

// WebHttp
type WebHttp struct {
	httpServer  *http.Server
	shutTimeout time.Duration // 关闭超过时间将强制关闭
	endRunning  chan bool
	err         error
}

// ListenAndServe 启动http服务
//   参数
//     addr:    监听地址
//     handler: 路由接口实现类
// 返回
//     成功-启动监听，失败-返回错误信息
func ListenAndServe(addr string, handler http.Handler) error {
	return NewWebServer(addr, handler, DEFAULT_READ_TIMEOUT, DEFAULT_WRITE_TIMEOUT, DEFAULT_SHUT_TIMEOUT).ListenAndServe()

}

// ListenAndServeTLS 启动https服务
//   参数
//     addr:    监听地址
//     certFile: cert证书文件
//     keyFile:  key证书文件
//     handler: 路由接口实现类
//   返回
//     成功-启动监听，失败-返回错误信息
func ListenAndServeTLS(addr, certFile, keyFile string, handler http.Handler) error {
	return NewWebServer(addr, handler, DEFAULT_READ_TIMEOUT, DEFAULT_WRITE_TIMEOUT, DEFAULT_SHUT_TIMEOUT).ListenAndServeTLS(certFile, keyFile)
}

// NewServer 实例化Server
//   参数
//     addr:         监听地址
//     readTimeout:  读超时时间
//     writeTimeout: 写超时时间
//     shutTimeout:  关闭监听超时时间
//   返回
//     实例化的WebHttp对象
func NewWebServer(addr string, handler http.Handler, readTimeout, writeTimeout, shutTimeout time.Duration) *WebHttp {
	if readTimeout <= 0 {
		readTimeout = DEFAULT_READ_TIMEOUT
	}

	if writeTimeout <= 0 {
		writeTimeout = DEFAULT_WRITE_TIMEOUT
	}

	return &WebHttp{
		httpServer: &http.Server{
			Addr:    addr,
			Handler: handler,

			ReadTimeout:  readTimeout * time.Second,
			WriteTimeout: writeTimeout * time.Second,
		},
		shutTimeout: shutTimeout * time.Second,
		endRunning:  make(chan bool, 1),
	}
}

// ListenAndServe 启动http服务
//   参数
//
//   返回
//     成功-启动监听，失败-返回错误信息
func (srv *WebHttp) ListenAndServe() error {
	go srv.handleSignals() // 捕获信号
	go func() {
		srv.err = srv.httpServer.ListenAndServe()
		if srv.err != nil {
			log.Println(srv.err)
		}
		srv.endRunning <- true
	}() // 启动服务
	<-srv.endRunning

	// 关闭老程序
	pid := os.Getpid()
	ctx, _ := context.WithTimeout(context.Background(), srv.shutTimeout)
	srv.httpServer.Shutdown(ctx)
	Flogger.Infof("Server: Listener of pid %d closed.", pid)

	return nil
}

// ListenAndServeTLS 启动https服务
//   参数
//     certFile: cert证书文件
//     keyFile:  key证书文件
//   返回
//     成功-启动监听，失败-返回错误信息
func (srv *WebHttp) ListenAndServeTLS(certFile, keyFile string) error {
	go srv.handleSignals() // 捕获信号
	go func() {
		srv.err = srv.httpServer.ListenAndServeTLS(certFile, keyFile)
		if srv.err != nil {
			log.Println(srv.err)
		}
		srv.endRunning <- true
	}() // 启动服务
	<-srv.endRunning

	// 关闭老程序
	pid := os.Getpid()
	ctx, _ := context.WithTimeout(context.Background(), srv.shutTimeout)
	srv.httpServer.Shutdown(ctx)
	Flogger.Infof("Server: Listener of pid %d closed.", pid)

	return srv.err
}

// handleSignals 捕获信号
//   参数
//
//   返回
//
func (srv *WebHttp) handleSignals() {
	_, sigName := utils.HandleSignals()
	Flogger.Infof("Server: Pid %d received %s.", os.Getpid(), sigName)
	srv.endRunning <- true
}
