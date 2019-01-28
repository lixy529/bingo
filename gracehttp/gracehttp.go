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

// ListenAndServe start http service.
func ListenAndServe(addr string, handler http.Handler) error {
	server := NewServer(addr, handler, DEFAULT_READ_TIMEOUT, DEFAULT_WRITE_TIMEOUT, DEFAULT_SHUT_TIMEOUT)
	return server.ListenAndServe()
}

// ListenAndServeTLS start https service.
func ListenAndServeTLS(addr, certFile, keyFile string, handler http.Handler) error {
	server := NewServer(addr, handler, DEFAULT_READ_TIMEOUT, DEFAULT_WRITE_TIMEOUT, DEFAULT_SHUT_TIMEOUT)
	return server.ListenAndServeTLS(certFile, keyFile)
}
