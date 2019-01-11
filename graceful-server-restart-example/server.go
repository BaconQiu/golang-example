package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

// start the server
func serve(handler http.Handler) {
	srv := start(handler)

	var err error
	err = waitForSignals(srv)
	if err != nil {
		panic(err)
	}
}

// start the server and register the handler
func start(handler http.Handler) *http.Server {
	srv := &http.Server{
		Addr: cfg.Addr,
	}

	srv.Handler = handler

	go srv.Serve(cfg.Ln)

	return srv
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

// fork the process
func fork() (*os.Process, error) {
	// get the file descriptor and pack it up in the metadata
	lnFile, err := getListenerFile(cfg.Ln)
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

// gracefully shutdown the server
func shutdown(srv *http.Server) error {
	fmt.Println("Server shutting down")

	ctx, cancel := context.WithTimeout(context.Background(),
		cfg.ShutdownTimeout)
	defer cancel()

	return srv.Shutdown(ctx)
}

