// http 服务相关
//   变更历史
//     2017-02-06  lixiaoya  新建
package gracehttp

import (
	"net/http"
)

const (
	GRACEFUL_ENVIRON_KEY    = "IS_GRACEFUL"
	GRACEFUL_ENVIRON_STRING = GRACEFUL_ENVIRON_KEY + "=1"

	DEFAULT_READ_TIMEOUT  = 60
	DEFAULT_WRITE_TIMEOUT = DEFAULT_READ_TIMEOUT
	DEFAULT_SHUT_TIMEOUT  = 20
)

// ListenAndServe 启动http服务
//   参数
//     addr:    监听地址
//     handler: 路由接口实现类
//   返回
//     成功-启动监听，失败-返回错误信息
func ListenAndServe(addr string, handler http.Handler) error {
	server := NewServer(addr, handler, DEFAULT_READ_TIMEOUT, DEFAULT_WRITE_TIMEOUT, DEFAULT_SHUT_TIMEOUT)
	return server.ListenAndServe()
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
	server := NewServer(addr, handler, DEFAULT_READ_TIMEOUT, DEFAULT_WRITE_TIMEOUT, DEFAULT_SHUT_TIMEOUT)
	return server.ListenAndServeTLS(certFile, keyFile)
}
