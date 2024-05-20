package http

import (
	"net"
	"net/http"
	"time"
)

// defaultServerOptions represents the default serverOptions used when creating
// a new http Server.
var defaultServerOptions = serverOptions{
	IdleTimeout:       5 * time.Minute,
	MaxHeaderBytes:    http.DefaultMaxHeaderBytes,
	ReadHeaderTimeout: 10 * time.Second,
	ReadTimeout:       20 * time.Second,
	WriteTimeout:      30 * time.Second,
}

// serverOptions represents the possible configuration options for creating a
// new http Server.
type serverOptions struct {
	IdleTimeout       time.Duration
	MaxHeaderBytes    int
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
}

// ServerOption represents an option to modify a setting when creating a new
// http Server.
type ServerOption interface {
	modifyServer(*serverOptions)
}

type modifyServerFn func(*serverOptions)

func (m modifyServerFn) modifyServer(o *serverOptions) { m(o) }

// WithIdleTimeout returns a ServerOption that sets the idle timeout to 'd'.
// Default: 5 minutes.
func WithIdleTimeout(d time.Duration) ServerOption {
	return modifyServerFn(func(o *serverOptions) {
		o.IdleTimeout = d
	})
}

// WithMaxHeaderBytes returns a ServerOption that sets the maximum number of
// header bytes to 'n'.
// Default: http.DefaultMaxHeaderBytes.
func WithMaxHeaderBytes(n int) ServerOption {
	return modifyServerFn(func(o *serverOptions) {
		o.MaxHeaderBytes = n
	})
}

// WithReadHeaderTimeout returns a ServerOption that sets the read header
// timeout to 'd'.
// Default: 10 seconds.
func WithReadHeaderTimeout(d time.Duration) ServerOption {
	return modifyServerFn(func(o *serverOptions) {
		o.ReadHeaderTimeout = d
	})
}

// WithReadTimeout returns a ServerOption that sets the read timeout to 'd'.
// Default: 20 seconds.
func WithReadTimeout(d time.Duration) ServerOption {
	return modifyServerFn(func(o *serverOptions) {
		o.ReadTimeout = d
	})
}

// WithWriteTimeout returns a ServerOption that sets the write timeout to 'd'.
// Default: 30 seconds.
func WithWriteTimeout(d time.Duration) ServerOption {
	return modifyServerFn(func(o *serverOptions) {
		o.WriteTimeout = d
	})
}

// NewServer returns a new http Server given the provided address, handler,
// and optional ServerOptions.
func NewServer(addr string, h http.Handler, ops ...ServerOption) *http.Server {
	op := defaultServerOptions
	for _, o := range ops {
		o.modifyServer(&op)
	}
	return &http.Server{
		Addr:              addr,
		Handler:           h,
		IdleTimeout:       op.IdleTimeout,
		MaxHeaderBytes:    op.MaxHeaderBytes,
		ReadHeaderTimeout: op.ReadHeaderTimeout,
		ReadTimeout:       op.ReadTimeout,
		WriteTimeout:      op.WriteTimeout,
	}
}

// ListenAndServe starts an http Server using the provided TCP keep alive
// duration.
func ListenAndServe(s *http.Server, keepAlive time.Duration) error {
	addr := s.Addr
	if addr == "" {
		addr = ":http"
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	tcpLn := &tcpKeepAliveListener{
		TCPListener: ln.(*net.TCPListener),
		keepAlive:   keepAlive,
	}
	return s.Serve(tcpLn)
}

type tcpKeepAliveListener struct {
	*net.TCPListener
	keepAlive time.Duration
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	_ = tc.SetKeepAlive(true)
	_ = tc.SetKeepAlivePeriod(ln.keepAlive)
	return tc, nil
}
