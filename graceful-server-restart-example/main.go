// with reference to https://tomaz.lovrec.eu/posts/graceful-server-restart/
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

var cfg *srvCfg

type srvCfg struct {
	// Socket file location
	sockFile string

	// Listen address
	addr string

	// Listener
	ln net.Listener

	// Amount of time allowed for requests to finish before server shutdown
	shutdownTimeout time.Duration

	// Amount of time allowed for a child to properly spin up and request the listener
	childTimeout time.Duration
}

type listener struct {
	// Listener address
	Addr string `json:"addr"`

	// Listener file descriptor
	FD int `json:"fd"`

	// Listener file name
	Filename string `json:"filename"`
}

// accept connections on the socket
func acceptConn(l net.Listener) (c net.Conn, err error) {
	chn := make(chan error)
	go func() {
		defer close(chn)
		c, err = l.Accept()
		if err != nil {
			chn <- err
		}
	}()

	select {
	case err = <-chn:
		if err != nil {
			fmt.Printf("error occurred when accepting socket connection: %v\n",
				err)
		}

	case <-time.After(cfg.childTimeout):
		fmt.Println("timeout occurred waiting for connection from child")
	}

	return
}

// create a new UNIX domain socket and handle communication
func socketListener(chn chan<- string, errChn chan<- error) {
	ln, err := net.Listen("unix", cfg.sockFile)
	if err != nil {
		errChn <- err
		return
	}
	defer ln.Close()

	// signal that we created a socket
	chn <- "socket_opened"

	// accept
	c, err := acceptConn(ln)
	if err != nil {
		errChn <- err
		return
	}

	// read from the socket
	buf := make([]byte, 512)
	nr, err := c.Read(buf)
	if err != nil {
		errChn <- err
		return
	}

	data := buf[0:nr]
	switch string(data) {
	case "get_listener":
		fmt.Println("get_listener received - sending listener information")
		// TODO: send listener
		chn <- "listener_sent"
	}
}

// handle SIGHUP - create socket, fork the child, and wait for child execution
func handleHangup() error {
	c := make(chan string)
	defer close(c)
	errChn := make(chan error)
	defer close(errChn)

	go socketListener(c, errChn)

	for {
		select {
		case cmd := <-c:
			switch cmd {
			case "socket_opened":
				// TODO: fork
				p, err := fork()
				if err != nil {
					fmt.Printf("unable to fork: %v\n", err)
					continue
				}
				fmt.Printf("forked (PID: %d), waiting for spinup", p.Pid)

			case "listener_sent":
				fmt.Println("listener sent - shutting down")

				return nil
			}

		case err := <-errChn:
			return err
		}
	}

	return nil
}

// listen for signals
func waitForSignals(srv *http.Server) error {
	sig := make(chan os.Signal, 1024)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)
	for {
		select {
		case s := <-sig:
			switch s {
			case syscall.SIGHUP:
				err := handleHangup()
				if err == nil {
					// no error occured - child spawned and started
					return shutdown(srv)
				}
			case syscall.SIGTERM, syscall.SIGINT:
				return shutdown(srv)
			}
		}
	}
}

// create a new listener and return it
func createListener() (net.Listener, error) {
	ln, err := net.Listen("tcp", cfg.addr)
	if err != nil {
		return nil, err
	}

	return ln, nil
}

// import the listener from the UNIX domain socket and attempt to
// recreate/rebuild the underlying *os.File
func importListener() (net.Listener, error) {
	c, err := net.Dial("unix", cfg.sockFile)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	var lnEnv string
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func(r io.Reader) {
		defer wg.Done()

		buf := make([]byte, 1024)
		n, err := r.Read(buf[:])
		if err != nil {
			return
		}

		lnEnv = string(buf[0:n])
	}(c)

	_, err = c.Write([]byte("get_listener"))
	if err != nil {
		return nil, err
	}

	wg.Wait()

	if lnEnv == "" {
		return nil, fmt.Errorf("Listener info not received from socket")
	}

	var l listener
	err = json.Unmarshal([]byte(lnEnv), &l)
	if err != nil {
		return nil, err
	}
	if l.Addr != cfg.addr {
		return nil, fmt.Errorf("unable to find listener for %v", cfg.addr)
	}

	// the file has already been passed to this process, extract the file
	// descriptor and name from the metadata to rebuild/find the *os.File for
	// the listener.
	lnFile := os.NewFile(uintptr(l.FD), l.Filename)
	if lnFile == nil {
		return nil, fmt.Errorf("unable to create listener file: %v", l.Filename)
	}
	defer lnFile.Close()

	// create a listener with the *os.File
	ln, err := net.FileListener(lnFile)
	if err != nil {
		return nil, err
	}

	return ln, nil
}

// obtain a network listener
func getListener() (net.Listener, error) {
	// try to import a listener if we are a fork
	ln, err := importListener()
	if err == nil {
		fmt.Printf("imported listener file descriptor for addr: %s\n", cfg.addr)
		return ln, nil
	}

	// couldn't import a listener, let's create one
	ln, err = createListener()
	if err != nil {
		return nil, err
	}

	return ln, err
}

// obtain the *os.File from the listener
func getListenerFile(ln net.Listener) (*os.File, error) {
	switch t := ln.(type) {
	case *net.TCPListener:
		return t.File()

	case *net.UnixListener:
		return t.File()
	}

	return nil, fmt.Errorf("unsupported listener: %T", ln)
}

// fork the process
func fork() (*os.Process, error) {
	// get the file descriptor and pack it up in the metadata
	lnFile, err := getListenerFile(cfg.ln)
	if err != nil {
		return nil, err
	}
	defer lnFile.Close()

	// pass the stdin, stdout, stderr, and the listener files to the child
	files := []*os.File{
		os.Stdin,
		os.Stdout,
		os.Stderr,
		lnFile,
	}

	// get process name and dir
	execName, err := os.Executable()
	if err != nil {
		return nil, err
	}
	execDir := filepath.Dir(execName)

	// spawn a child
	p, err := os.StartProcess(execName, []string{execName}, &os.ProcAttr{
		Dir:   execDir,
		Files: files,
		Sys:   &syscall.SysProcAttr{},
	})
	if err != nil {
		return nil, err
	}

	return p, nil
}

// start the server and register the handler
func start(handler http.Handler) *http.Server {
	srv := &http.Server{
		Addr: cfg.addr,
	}

	srv.Handler = handler

	go srv.Serve(cfg.ln)

	return srv
}

// gracefully shutdown the server
func shutdown(srv *http.Server) error {
	fmt.Println("Server shutting down")

	ctx, cancel := context.WithTimeout(context.Background(),
		cfg.shutdownTimeout)
	defer cancel()

	return srv.Shutdown(ctx)
}

// obtain a listener, start the server
func serve(config srvCfg, handler http.Handler) {
	cfg = &config

	var err error
	cfg.ln, err = getListener()
	if err != nil {
		panic(err)
	}

	srv := start(handler)

	err = waitForSignals(srv)
	if err != nil {
		panic(err)
	}
}

func main() {
	serve(srvCfg{
		sockFile:        "/tmp/api.sock",
		addr:            ":8000",
		shutdownTimeout: 5 * time.Second,
		childTimeout:    5 * time.Second,
	}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`Hello, world!`))
	}))
}
