package main

import (
	"net/http"
	"time"
)

type httpServerTimeouts struct {
	readHeader time.Duration
	read       time.Duration
	write      time.Duration
	idle       time.Duration
}

var defaultHTTPServerTimeouts = httpServerTimeouts{
	readHeader: 5 * time.Second,
	read:       15 * time.Second,
	write:      30 * time.Second,
	idle:       60 * time.Second,
}

func newHTTPServer(addr string, handler http.Handler) *http.Server {
	return newHTTPServerWithTimeouts(addr, handler, defaultHTTPServerTimeouts)
}

func newHTTPServerWithTimeouts(addr string, handler http.Handler, timeouts httpServerTimeouts) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: timeouts.readHeader,
		ReadTimeout:       timeouts.read,
		WriteTimeout:      timeouts.write,
		IdleTimeout:       timeouts.idle,
	}
}
