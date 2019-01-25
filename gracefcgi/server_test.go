package gracefcgi

import (
	"net/http"
	"testing"
)

type FastCGIServer struct{}

func (s FastCGIServer) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	resp.Write([]byte("<h1>Hello, world!</h1>\n<p>Behold my Go web app.</p>"))
}

func TestServer(t *testing.T) {
	srv := &FastCGIServer{}
	err := NewServer("127.0.0.1", 9090, srv, 50).ListenAndServe()
	if err != nil {
		t.Errorf("Start failed, err: %s", err.Error())
	}
}
