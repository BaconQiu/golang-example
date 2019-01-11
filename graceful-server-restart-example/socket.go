package main

import (
	"fmt"
	"net"
	"time"
)

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

	case <-time.After(cfg.ChildTimeout):
		fmt.Println("timeout occurred waiting for connection from child")
	}

	return
}

// create a new UNIX domain socket and handle communication
func socketListener(chn chan<- string, errChn chan<- error) {
	ln, err := net.Listen("unix", cfg.SockFile)
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

		err := sendListener(c)
		if err != nil {
			fmt.Println("Unable to send http listener socket over the unix domain socket")
			errChn <- err
			return
		}

		chn <- "listener_sent"
	}
}