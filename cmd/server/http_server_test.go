package main

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
)

func startHTTPServerForTest(t *testing.T, srv *http.Server) (string, <-chan error) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	serveErr := make(chan error, 1)
	go func() { serveErr <- srv.Serve(listener) }()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	})
	return "http://" + listener.Addr().String(), serveErr
}

func TestHTTPServerReadHeaderTimeoutRejectsSlowRequest(t *testing.T) {
	const headerTimeout = 50 * time.Millisecond
	srv := newHTTPServerWithTimeouts("", http.NotFoundHandler(), httpServerTimeouts{
		readHeader: headerTimeout,
		read:       250 * time.Millisecond,
		write:      time.Second,
		idle:       time.Second,
	})
	baseURL, _ := startHTTPServerForTest(t, srv)
	conn, err := net.Dial("tcp", strings.TrimPrefix(baseURL, "http://"))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if _, err := io.WriteString(conn, "GET / HTTP/1.1\r\nHost: localhost\r\n"); err != nil {
		t.Fatal(err)
	}
	if err := conn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	response, err := io.ReadAll(conn)
	if err != nil {
		t.Fatalf("slow request did not terminate: %v", err)
	}
	// Go may close the connection without writing a 408 response. Either an
	// HTTP 408 or EOF is evidence that the server no longer waits for headers.
	_ = response
}

func TestHTTPServerObservesAbandonedClient(t *testing.T) {
	started := make(chan struct{})
	cancelled := make(chan struct{})
	handler := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		close(started)
		<-r.Context().Done()
		close(cancelled)
	})
	srv := newHTTPServerWithTimeouts("", handler, httpServerTimeouts{
		readHeader: time.Second,
		read:       2 * time.Second,
		write:      2 * time.Second,
		idle:       time.Second,
	})
	baseURL, _ := startHTTPServerForTest(t, srv)
	conn, err := net.Dial("tcp", strings.TrimPrefix(baseURL, "http://"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.WriteString(conn, "GET /abandoned HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n"); err != nil {
		conn.Close()
		t.Fatal(err)
	}
	select {
	case <-started:
	case <-time.After(time.Second):
		conn.Close()
		t.Fatal("handler did not start")
	}
	if err := conn.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-cancelled:
	case <-time.After(time.Second):
		t.Fatal("abandoned client did not cancel request context")
	}
}

func TestHTTPServerShutdownReturnsCleanly(t *testing.T) {
	srv := newHTTPServer("", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	baseURL, serveErr := startHTTPServerForTest(t, srv)
	response, err := http.Get(baseURL)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		t.Fatalf("graceful shutdown failed: %v", err)
	}
	select {
	case err := <-serveErr:
		if !errors.Is(err, http.ErrServerClosed) {
			t.Fatalf("Serve returned %v, want http.ErrServerClosed", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Serve did not return after graceful shutdown")
	}
}

func TestHTTPServerReadHeaderTimeoutIsConfigured(t *testing.T) {
	srv := newHTTPServer("127.0.0.1:19281", http.NotFoundHandler())
	if srv.ReadHeaderTimeout != 5*time.Second || srv.ReadTimeout != 15*time.Second || srv.WriteTimeout != 30*time.Second || srv.IdleTimeout != 60*time.Second {
		t.Fatalf("unexpected HTTP timeouts: header=%s read=%s write=%s idle=%s", srv.ReadHeaderTimeout, srv.ReadTimeout, srv.WriteTimeout, srv.IdleTimeout)
	}
}
