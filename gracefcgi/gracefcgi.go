// FastCgi服务
// fcgi.go和child.go从golang源码拷贝过来
//   变更历史
//     2017-03-15  lixiaoya  新建
package gracefcgi

const (
	GRACEFUL_ENVIRON_KEY    = "FCGI_GRACE"
	GRACEFUL_ENVIRON_STRING = GRACEFUL_ENVIRON_KEY + "=1"

	DEFAULT_SHUT_TIMEOUT = 20
)
