// FastCgi server
// fcgi.go and child.go copy from source of golang.
package gracefcgi

const (
	GRACEFUL_ENVIRON_KEY    = "FCGI_GRACE"
	GRACEFUL_ENVIRON_STRING = GRACEFUL_ENVIRON_KEY + "=1"

	DEFAULT_SHUT_TIMEOUT = 20
)
