// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package httptest

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	internalapi "github.com/tetratelabs/func-e/internal/api"
)

// Server is re-exported so callers importing this package don't also need to
// import the stdlib net/http/httptest just to name the return type.
type Server = httptest.Server

// NewServer starts an httptest server backed by net.Pipe.
func NewServer(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()
	// Swap the default TCP listener for a net.Pipe-backed to stay in-process.
	listener := newMemoryListener()
	ts := httptest.NewUnstartedServer(handler)
	ts.Listener = listener
	ts.Start()

	// Keep-alives are off so each request gets a pipe pair instead of pinning.
	ts.Client().Transport = &http.Transport{
		DialContext:       listener.DialContext,
		DisableKeepAlives: true,
	}

	t.Cleanup(ts.Close)
	return ts
}

// HandlerFactory returns a factory for a client that serves requests through
// handler in the caller's goroutine.
func HandlerFactory(handler http.Handler) internalapi.HTTPClientFunc {
	return func() *http.Client {
		return &http.Client{Transport: handlerTransport{handler: handler}}
	}
}

type memoryListener struct {
	addr memoryAddr

	conns  chan net.Conn
	closed chan struct{}
	once   sync.Once
}

var _ net.Listener = (*memoryListener)(nil)

type handlerTransport struct {
	handler http.Handler
}

func (t handlerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	recorder := httptest.NewRecorder()
	t.handler.ServeHTTP(recorder, req)
	if err := req.Context().Err(); err != nil {
		return nil, err
	}
	return recorder.Result(), nil
}

func newMemoryListener() *memoryListener {
	return &memoryListener{
		addr:   memoryAddr{},
		conns:  make(chan net.Conn),
		closed: make(chan struct{}),
	}
}

func (l *memoryListener) Accept() (net.Conn, error) {
	select {
	case <-l.closed:
		return nil, &net.OpError{
			Op:   "accept",
			Net:  l.addr.Network(),
			Addr: l.addr,
			Err:  net.ErrClosed,
		}
	case server := <-l.conns:
		return server, nil
	}
}

func (l *memoryListener) Close() error {
	l.once.Do(func() {
		close(l.closed)
	})
	return nil
}

func (l *memoryListener) Addr() net.Addr {
	return l.addr
}

func (l *memoryListener) DialContext(ctx context.Context, _, _ string) (net.Conn, error) {
	server, client := net.Pipe()
	select {
	case <-ctx.Done():
		closeConn(server)
		closeConn(client)
		return nil, &net.OpError{Op: "dial", Net: l.addr.Network(), Err: ctx.Err()}
	case l.conns <- server:
		return client, nil
	case <-l.closed:
		closeConn(server)
		closeConn(client)
		return nil, &net.OpError{Op: "dial", Net: l.addr.Network(), Err: net.ErrClosed}
	}
}

func closeConn(conn net.Conn) {
	err := conn.Close()
	if err == nil || errors.Is(err, net.ErrClosed) {
		return
	}
}

var _ net.Addr = memoryAddr{}

type memoryAddr struct{}

func (memoryAddr) Network() string {
	return "memory"
}

func (memoryAddr) String() string {
	return "127.0.0.1:0"
}
