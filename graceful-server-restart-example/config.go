package main

import (
	"net"
	"time"
)

type SrvCfg struct {
	// Socket file location
	SockFile string

	// Listen address
	Addr string

	// Listener
	Ln net.Listener

	// Amount of time allowed for requests to finish before server shutdown
	ShutdownTimeout time.Duration

	// Amount of time allowed for a child to properly spin up and request the listener
	ChildTimeout time.Duration
}

var cfg *SrvCfg

func initCfg(sockFile, addr string) {
	cfg = &SrvCfg{
		SockFile: sockFile,
		Addr: addr,
		ShutdownTimeout: 5 * time.Second,
		ChildTimeout: 5 * time.Second,
	}

	var err error
	cfg.Ln, err = getListener()
	if err != nil {
		panic(err)
	}
}