package api

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

type shortWriter struct {
	max int
	buf bytes.Buffer
}

func (w *shortWriter) Write(p []byte) (int, error) {
	if len(p) > w.max {
		p = p[:w.max]
	}
	return w.buf.Write(p)
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("synthetic write failure")
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) {
	return 0, errors.New("synthetic read failure")
}

func TestWriteAllHandlesPartialAndFailedWrites(t *testing.T) {
	writer := &shortWriter{max: 2}
	payload := []byte("websocket-handshake")
	if err := writeAll(writer, payload); err != nil {
		t.Fatalf("writeAll returned error: %v", err)
	}
	if got := writer.buf.String(); got != string(payload) {
		t.Fatalf("writeAll payload = %q, want %q", got, payload)
	}

	if err := writeAll(failingWriter{}, payload); err == nil {
		t.Fatal("writeAll accepted a failed write")
	}
}

func TestCopyAllPropagatesReaderAndWriterResult(t *testing.T) {
	var dst bytes.Buffer
	if err := copyAll(&dst, bytes.NewBufferString("session-bytes")); err != nil {
		t.Fatalf("copyAll returned error: %v", err)
	}
	if got := dst.String(); got != "session-bytes" {
		t.Fatalf("copyAll payload = %q", got)
	}
	if err := copyAll(failingWriter{}, bytes.NewBufferString("session-bytes")); err == nil {
		t.Fatal("copyAll accepted a failed writer")
	}
	if err := copyAll(io.Discard, failingReader{}); err == nil {
		t.Fatal("copyAll accepted a failed reader")
	}
}
