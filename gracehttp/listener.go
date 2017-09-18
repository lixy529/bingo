// http 监听
//   变更历史
//     2017-02-06  lixiaoya  新建
package gracehttp

import (
	"net"
	"time"
)

// tcpKeepAliveListener 代码从http复制
type webTcpListener struct {
	*net.TCPListener
}

// NewWebTcpListener 返回webTcpListener实例
//   参数
//     tl: TCP监听
//   返回
//     实例化的webTcpListener对象
func NewWebTcpListener(tl *net.TCPListener) net.Listener {
	return &webTcpListener{
		TCPListener: tl,
	}
}

// Accept
//   参数
//
//   返回
//     成功时返回链接，失败时返回错误信息
func (ln *webTcpListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}

// Fd 获取listener的fd
//   参数
//
//   返回
//     成功时返回监听的fd，失败时返回错误信息
func (ln *webTcpListener) Fd() (uintptr, error) {
	f, err := ln.File()
	if err != nil {
		return 0, err
	}
	return f.Fd(), nil
}
