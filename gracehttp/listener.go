// Copy from http
package gracehttp

import (
	"net"
	"time"
)

// tcpKeepAliveListener
type webTcpListener struct {
	*net.TCPListener
}

// NewWebTcpListener return webTcpListener object.
func NewWebTcpListener(tl *net.TCPListener) net.Listener {
	return &webTcpListener{
		TCPListener: tl,
	}
}

// Accept
func (ln *webTcpListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}

// Fd return listener fd.
func (ln *webTcpListener) Fd() (uintptr, error) {
	f, err := ln.File()
	if err != nil {
		return 0, err
	}
	return f.Fd(), nil
}
