package gracehttp

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/lixy529/gotools/utils"
	"log"
	"net"
	"net/http"
	"os"
	"syscall"
	"time"
)

type Server struct {
	httpServer *http.Server
	listener   net.Listener
	tlsConfig  *tls.Config

	isGraceful  bool
	shutTimeout time.Duration // Close timeout will be forced to close
	endRunning  chan bool
	isRestart   bool
	isHttps     bool
	err         error
}

// NewServer return Server object.
func NewServer(addr string, handler http.Handler, readTimeout, writeTimeout, shutTimeout time.Duration) *Server {
	isGraceful := false
	if os.Getenv(GRACEFUL_ENVIRON_KEY) != "" {
		isGraceful = true
	}

	if readTimeout <= 0 {
		readTimeout = DEFAULT_READ_TIMEOUT
	}

	if writeTimeout <= 0 {
		writeTimeout = DEFAULT_WRITE_TIMEOUT
	}

	if shutTimeout <= 0 {
		shutTimeout = DEFAULT_SHUT_TIMEOUT
	}

	return &Server{
		httpServer: &http.Server{
			Addr:    addr,
			Handler: handler,

			ReadTimeout:  readTimeout * time.Second,
			WriteTimeout: writeTimeout * time.Second,
		},
		isGraceful:  isGraceful,
		shutTimeout: shutTimeout * time.Second,
		endRunning:  make(chan bool, 1),
		isRestart:   false,
		isHttps:     false,
	}
}

// ListenAndServe start listen and http services
func (srv *Server) ListenAndServe() error {
	addr := srv.httpServer.Addr
	if addr == "" {
		addr = ":http"
	}

	ln, err := srv.getListener(addr)
	if err != nil {
		return err
	}

	srv.listener = NewWebTcpListener(ln)

	return srv.Serve()
}

// ListenAndServeTLS start listen and https services
func (srv *Server) ListenAndServeTLS(certFile, keyFile string) error {
	srv.isHttps = true
	addr := srv.httpServer.Addr
	if addr == "" {
		addr = ":https"
	}

	srv.tlsConfig = &tls.Config{}
	if srv.httpServer.TLSConfig != nil {
		*srv.tlsConfig = *srv.httpServer.TLSConfig
	}
	if srv.tlsConfig.NextProtos == nil {
		srv.tlsConfig.NextProtos = []string{"http/1.1"}
	}

	var err error
	srv.tlsConfig.Certificates = make([]tls.Certificate, 1)
	srv.tlsConfig.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}

	ln, err := srv.getListener(addr)
	if err != nil {
		return err
	}

	srv.listener = NewWebTcpListener(ln)

	return srv.Serve()
}

// Server start service.
func (srv *Server) Serve() error {
	go srv.handleSignals()
	go func() {
		if srv.isHttps {
			ln := tls.NewListener(srv.listener, srv.tlsConfig)
			srv.err = srv.httpServer.Serve(ln)
		} else {
			srv.err = srv.httpServer.Serve(srv.listener)
		}
		if srv.err != nil {
			log.Println(srv.err)
		}
		srv.endRunning <- true
	}()
	<-srv.endRunning

	// If it is restarted, start a new process first, then close the old process, otherwise close the old process directly.
	pid := os.Getpid()
	if srv.isRestart {
		err := srv.startNewProcess()
		if err != nil {
			log.Printf("GraceHttp: Start new process failed[%v], pid[%d] continue serve.\n", err, pid)
			return err
		}
	}
	ctx, _ := context.WithTimeout(context.Background(), srv.shutTimeout)
	srv.httpServer.Shutdown(ctx)
	log.Printf("GraceHttp: Listener of pid %d closed.\n", pid)

	return srv.err
}

// getListener return listener.
// If restart, listen from FD.
func (srv *Server) getListener(addr string) (*net.TCPListener, error) {
	var ln net.Listener
	var err error

	if srv.isGraceful {
		file := os.NewFile(3, "")
		ln, err = net.FileListener(file)
		if err != nil {
			err = fmt.Errorf("GraceHttp: net.FileListener error: %v", err)
			return nil, err
		}
	} else {
		ln, err = net.Listen("tcp", addr)
		if err != nil {
			err = fmt.Errorf("GraceHttp: net.Listen error: %v", err)
			return nil, err
		}
	}
	return ln.(*net.TCPListener), nil
}

// startNewProcess start a new process.
// If startup fails, continue using the old process.
func (srv *Server) startNewProcess() error {
	log.Println("Start new process begin...")

	listenerFd, err := srv.listener.(*webTcpListener).Fd()
	if err != nil {
		return fmt.Errorf("failed to get socket file descriptor: %v", err)
	}

	argv0 := os.Args[0]

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
		return fmt.Errorf("GraceHttp: Failed to forkexec: %v", err)
	}

	log.Printf("GraceHttp: Start new process success, pid %d.", fork)

	return nil
}

// handleSignals capture signal.
func (srv *Server) handleSignals() {
	sigCode, sigName := utils.HandleSignals()
	log.Printf("GraceHttp: Pid %d received %s.\n", os.Getpid(), sigName)
	if sigCode == syscall.SIGHUP || sigCode == syscall.SIGUSR2 {
		srv.isRestart = true
		srv.endRunning <- true
	} else {
		srv.isRestart = false
		srv.endRunning <- true
	}
}
