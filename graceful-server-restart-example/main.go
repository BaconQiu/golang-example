// with reference to https://tomaz.lovrec.eu/posts/graceful-server-restart/
package main

import (
	"net"
	"net/http"
	"time"
)

func main() {
	initCfg("/tmp/api.sock", ":8000")

	serve(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`Hello, world!`))
	}))
}
