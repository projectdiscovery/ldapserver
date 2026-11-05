package ldapserver

import (
	"errors"
	"net"
	"sync/atomic"
	"testing"
	"time"
)

// flakyListener fails Accept a few times with a non-timeout error, as a
// listener does when the process runs out of file descriptors.
type flakyListener struct {
	net.Listener
	fails atomic.Int32
}

func (l *flakyListener) Accept() (net.Conn, error) {
	if l.fails.Add(-1) >= 0 {
		return nil, &net.OpError{Op: "accept", Net: "tcp", Err: errors.New("too many open files")}
	}
	return l.Listener.Accept()
}

func TestServeSurvivesAcceptErrors(t *testing.T) {
	inner, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ln := &flakyListener{Listener: inner}
	ln.fails.Store(3)

	s := NewServer()
	s.Handle(NewRouteMux())
	done := make(chan error, 1)
	go func() { done <- s.Serve(ln) }()

	select {
	case err := <-done:
		t.Fatalf("server stopped after a transient accept error: %v", err)
	case <-time.After(300 * time.Millisecond):
	}
	conn, err := net.Dial("tcp", inner.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()
	s.Stop()
	if err := <-done; err != nil {
		t.Fatalf("Stop: %v", err)
	}
}
