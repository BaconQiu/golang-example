// with reference to https://tomaz.lovrec.eu/posts/graceful-server-restart/
package main

import (
	"net"
	"net/http"
	"time"
)

//var cfg *srvCfg

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

func main() {
	initCfg("/tmp/api.sock", ":8000")

	serve(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`Hello, world!`))
	}))
}
