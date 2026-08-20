package main

import (
	"net"
	"testing"
)

func TestCR02LoopbackBindHostRejectsNonLoopbackHosts(t *testing.T) {
	for _, host := range []string{"", "localhost", "0.0.0.0", "192.0.2.10", "::"} {
		if _, err := loopbackBindHost(host); err == nil {
			t.Fatalf("loopbackBindHost(%q) accepted a non-literal-loopback host", host)
		}
	}
}

func TestCR02LoopbackBindHostBindsLocalSocket(t *testing.T) {
	host, err := loopbackBindHost("127.0.0.1")
	if err != nil {
		t.Fatalf("loopbackBindHost: %v", err)
	}
	listener, err := net.Listen("tcp", net.JoinHostPort(host, "0"))
	if err != nil {
		t.Fatalf("listen loopback: %v", err)
	}
	defer listener.Close()
	addr, ok := listener.Addr().(*net.TCPAddr)
	if !ok || !addr.IP.IsLoopback() {
		t.Fatalf("listener address = %v, want loopback", listener.Addr())
	}
}
