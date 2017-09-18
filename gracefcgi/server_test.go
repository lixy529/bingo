// FastCgi服务测试
//   变更历史
//     2017-03-14  lixiaoya  新建
package gracefcgi

import (
	"net/http"
	"testing"
)

type FastCGIServer struct{}

func (s FastCGIServer) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	resp.Write([]byte("<h1>Hello, 世界</h1>\n<p>Behold my Go web app.</p>"))
}

func TestServer(t *testing.T) {
	srv := &FastCGIServer{}
	err := NewServer("127.0.0.1", 9090, srv, 50).ListenAndServe()
	if err != nil {
		t.Errorf("Start failed, err: %s", err.Error())
	}
}
