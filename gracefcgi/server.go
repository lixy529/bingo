// FastCgi server
// fcgi.go and child.go copy from source of golang.
package gracefcgi

import (
	"fmt"
	"github.com/lixy529/gotools/utils"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"syscall"
	"time"
)

var (
	wg sync.WaitGroup
)

type Server struct {
	addr        string
	port        int
	listener    net.Listener
	handler     http.Handler
	shutTimeout time.Duration // Close timeout will be forced to close

	isGraceful bool
	endRunning chan bool
	isStop     bool
	isRestart  bool
}

func NewServer(addr string, port int, handler http.Handler, shutTimeout time.Duration) *Server {
	isGraceful := false
	if os.Getenv(GRACEFUL_ENVIRON_KEY) != "" {
		isGraceful = true
	}

	if handler == nil {
		handler = http.DefaultServeMux
	}

	if shutTimeout <= 0 {
		shutTimeout = DEFAULT_SHUT_TIMEOUT
	}

	return &Server{
		addr:        addr,
		port:        port,
		handler:     handler,
		shutTimeout: shutTimeout * time.Second,

		isGraceful: isGraceful,
		endRunning: make(chan bool, 1),
		isStop:     false,
		isRestart:  false,
	}
}

// ListenAndServe start listen and services
// Standard I/O when srv.addr is empty,
// Listen ip when srv.port > 0, Otherwise, is sock file.
func (srv *Server) ListenAndServe() error {
	var err error

	if srv.addr == "" {
		srv.listener = nil
	} else if srv.port > 0 {
		laddr := fmt.Sprintf("%s:%d", srv.addr, srv.port)
		srv.listener, err = srv.getListener(laddr)
		if err != nil {
			return err
		}
	} else {
		tAddr := srv.addr + "_tmp"
		_, err = os.Stat(tAddr)
		if err == nil {
			os.Remove(tAddr)
		}
		srv.listener, err = net.Listen("unix", tAddr)
		if err != nil {
			return err
		}
		err = os.Rename(tAddr, srv.addr)
		if err != nil {
			return err
		}

		// Modify file permissions
		err = os.Chmod(srv.addr, 0777)
		if err != nil {
			return err
		}
	}

	return srv.Start()
}

// getListener return listener.
// Listen from fd if restart.
func (srv *Server) getListener(addr string) (*net.TCPListener, error) {
	var ln net.Listener
	var err error

	if srv.isGraceful {
		file := os.NewFile(3, "")
		ln, err = net.FileListener(file)
		if err != nil {
			err = fmt.Errorf("GraceFcgi: net.FileListener error: %v", err)
			return nil, err
		}
	} else {
		ln, err = net.Listen("tcp", addr)
		if err != nil {
			err = fmt.Errorf("GraceFcgi: net.Listen error: %v", err)
			return nil, err
		}
	}
	return ln.(*net.TCPListener), nil
}

// Serve start service.
func (srv *Server) Start() error {
	go srv.handleSignals()
	go srv.Serve(srv.listener, srv.handler)

	<-srv.endRunning

	// Start a sub process.
	if srv.isRestart {
		srv.startNewProcess()
	}

	// Waiting...
	chanStop := make(chan bool)
	go func() {
		wg.Wait()
		close(chanStop)
	}()

	select {
	case <-time.After(srv.shutTimeout):
	case <-chanStop:
	}

	return nil
}

// Serve start service.
func (srv *Server) Serve(l net.Listener, handler http.Handler) error {
	if l == nil {
		var err error
		l, err = net.FileListener(os.Stdin)
		if err != nil {
			return err
		}
		defer l.Close()
	}
	if handler == nil {
		handler = http.DefaultServeMux
	}
	for {
		rw, err := l.Accept()
		if err != nil {
			return err
		}
		c := newChild(rw, handler)
		wg.Add(1)
		go c.serve()
	}
}

// startNewProcess start a new process.
func (srv *Server) startNewProcess() error {
	log.Println("GraceFcgi: Start new process begin...")

	argv0 := os.Args[0]
	var listenerFd uintptr

	if srv.port > 0 {
		//listenerFd, err := srv.listener.(*webTcpListener).Fd()
		f, err := srv.listener.(*net.TCPListener).File()
		if err != nil {
			return err
		}
		listenerFd = f.Fd()
	}

	// Setting environment variables, mark restart.
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
		return fmt.Errorf("GraceFcgi: Failed to forkexec: %v", err)
	}

	log.Printf("GraceFcgi: Start new process success, pid %d.", fork)

	return nil
}

// handleSignals capture signal.
func (srv *Server) handleSignals() {
	sigCode, sigName := utils.HandleSignals()
	log.Printf("GraceFcgi: Pid %d received %s.\n", os.Getpid(), sigName)
	if sigCode == syscall.SIGHUP || sigCode == syscall.SIGUSR2 {
		srv.isStop = true
		srv.isRestart = true
		srv.endRunning <- true
	} else {
		srv.isStop = true
		srv.isRestart = false
		srv.endRunning <- true
	}
}
