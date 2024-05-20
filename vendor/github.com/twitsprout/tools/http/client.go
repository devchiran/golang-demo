package http

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"time"
)

// DefaultClient represents the default global HTTP client.
var DefaultClient = NewClient()

// defaultClientOptions represents the default clientOptions used when creating
// a new http Client.
var defaultClientOptions = clientOptions{
	Dialer: &net.Dialer{
		DualStack: true,
		KeepAlive: 30 * time.Second,
		Timeout:   20 * time.Second,
	},
	ExpectContinueTimeout: 1 * time.Second,
	IdleConnTimeout:       120 * time.Second,
	MaxIdleConns:          10,
	MaxOpenConns:          15,
	Proxy:                 nil,
	Timeout:               30 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
}

// clientOptions represents the possible configuration options for creating a
// new http Client.
type clientOptions struct {
	Dialer                *net.Dialer
	DisableHTTP2          bool
	ExpectContinueTimeout time.Duration
	IdleConnTimeout       time.Duration
	MaxIdleConns          int
	MaxIdleConnsPerHost   int
	MaxOpenConns          int
	Proxy                 func(*http.Request) (*url.URL, error)
	Timeout               time.Duration
	TLSHandshakeTimeout   time.Duration
}

// ClientOption represents an option to modify a setting when creating a new
// http Client.
type ClientOption interface {
	modifyClient(*clientOptions)
}

type modifyClientFn func(*clientOptions)

func (m modifyClientFn) modifyClient(o *clientOptions) { m(o) }

// WithDialer returns a ClientOption that sets the http Client's Dialer.
// Default: dualstack, 30 second keepalive, 20 second timeout.
func WithDialer(d *net.Dialer) ClientOption {
	return modifyClientFn(func(o *clientOptions) {
		o.Dialer = d
	})
}

// WithHTTP2Disabled returns a ClientOption that disables the HTTP2 protocol.
// Default: HTTP2 enabled.
func WithHTTP2Disabled() ClientOption {
	return modifyClientFn(func(o *clientOptions) {
		o.DisableHTTP2 = true
	})
}

// WithExpectContinueTimeout returns a ClientOption that sets the http Client's
// "expect continue timeout".
// Default: 1 second.
func WithExpectContinueTimeout(d time.Duration) ClientOption {
	return modifyClientFn(func(o *clientOptions) {
		o.ExpectContinueTimeout = d
	})
}

// WithIdleConnTimeout returns a ClientOption that sets the http Client's
// idle connection timeout to 'd'.
// Default: 2 minutes.
func WithIdleConnTimeout(d time.Duration) ClientOption {
	return modifyClientFn(func(o *clientOptions) {
		o.IdleConnTimeout = d
	})
}

// WithMaxIdleConns returns a ClientOption that sets the http Client's maximum
// number of idle connections to 'n'.
// Default: 10.
func WithMaxIdleConns(n int) ClientOption {
	return modifyClientFn(func(o *clientOptions) {
		o.MaxIdleConns = n
	})
}

// WithMaxIdleConnsPerHost returns a ClientOption that sets the http Client's
// maximum number of idle connections per host to 'n'.
// Default: unlimited.
func WithMaxIdleConnsPerHost(n int) ClientOption {
	return modifyClientFn(func(o *clientOptions) {
		o.MaxIdleConnsPerHost = n
	})
}

// WithMaxOpenConns returns a ClientOption that sets the http Client's maximum
// number of concurrently open connections to 'n'.
// Default: 15.
func WithMaxOpenConns(n int) ClientOption {
	return modifyClientFn(func(o *clientOptions) {
		o.MaxOpenConns = n
	})
}

// WithProxy reutrns a ClientOption that sets the http Client's transport proxy
// to the provided proxy function.
// Default: nil.
func WithProxy(proxy func(*http.Request) (*url.URL, error)) ClientOption {
	return modifyClientFn(func(o *clientOptions) {
		o.Proxy = proxy
	})
}

// WithTimeout returns a ClientOption that sets the http Client's total request
// timeout to 'd'.
// Default: 30 seconds.
func WithTimeout(d time.Duration) ClientOption {
	return modifyClientFn(func(o *clientOptions) {
		o.Timeout = d
	})
}

// WithTLSHandshakeTimeout returns a ClientOption that sets the http Client's
// TLS handshake timeout to 'd'.
// Default: 10 seconds.
func WithTLSHandshakeTimeout(d time.Duration) ClientOption {
	return modifyClientFn(func(o *clientOptions) {
		o.TLSHandshakeTimeout = d
	})
}

// NewClient creates a new HTTP client using any provided ClientOptions.
func NewClient(ops ...ClientOption) *http.Client {
	// Use default options and apply any provided custom options.
	op := defaultClientOptions
	for _, o := range ops {
		o.modifyClient(&op)
	}

	// Create http transport.
	t := &http.Transport{
		ExpectContinueTimeout: op.ExpectContinueTimeout,
		IdleConnTimeout:       op.IdleConnTimeout,
		MaxIdleConns:          op.MaxIdleConns,
		MaxIdleConnsPerHost:   op.MaxIdleConnsPerHost,
		Proxy:                 op.Proxy,
		TLSHandshakeTimeout:   op.TLSHandshakeTimeout,
	}
	if op.Dialer != nil {
		t.DialContext = op.Dialer.DialContext
	}
	if op.DisableHTTP2 {
		t.TLSNextProto = map[string]func(authority string, c *tls.Conn) http.RoundTripper{}
	}

	// Wrap roundtripper with custom type if MaxOpenConns set.
	var rt http.RoundTripper = t
	if op.MaxOpenConns > 0 {
		rt = &roundTripper{rt: t, ch: make(chan struct{}, op.MaxOpenConns)}
	}
	return &http.Client{
		Timeout:   op.Timeout,
		Transport: rt,
	}
}

// roundTripper wraps an underlying http.RoundTripper, limiting the total
// number of concurrent requests at a time.
type roundTripper struct {
	rt http.RoundTripper
	ch chan struct{}
}

// RoundTrip implements the http RoundTripper interface.
func (rt *roundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	ctx := r.Context()
	select {
	case rt.ch <- struct{}{}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	defer func() { <-rt.ch }()
	return rt.rt.RoundTrip(r)
}
