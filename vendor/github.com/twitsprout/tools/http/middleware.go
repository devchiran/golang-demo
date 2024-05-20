package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/twitsprout/tools"
	"github.com/twitsprout/tools/requestid"
	"github.com/twitsprout/tools/runtime"
)

// TimeoutMiddleware is an HTTP middleware function that sets a timeout in the
// request's context.
func TimeoutMiddleware(dur time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), dur)
			defer cancel()
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

// RequestIDMiddleware is an HTTP middleware function that sets the request ID
// in the request's context.
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := requestid.WithRequestID(r.Context())
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

// RealIPMiddleware is an HTTP middleware function that sets the request's real
// IP address as the "RemoteAddr" in the request.
func RealIPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.RemoteAddr = realIP(r)
		next.ServeHTTP(w, r)
	})
}

// LimitReaderMiddleware is an HTTP middleware function that limits the number
// of bytes that can be read from the request body.
func LimitReaderMiddleware(limit int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = newLimitedReadCloser(limit, r.Body)
			next.ServeHTTP(w, r)
		})
	}
}

// LoggingMiddleware is an HTTP middleware function that logs the start and end
// of a request.
func LoggingMiddleware(logger tools.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			start := time.Now()

			// Get request source & URI.
			uri := r.URL.RequestURI()

			// Log initial request.
			logger.Debug("request received",
				"request_id", requestid.Get(ctx),
				"method", r.Method,
				"uri", uri,
				"source", r.RemoteAddr,
			)

			// Use a custom response writer.
			wr := withResponseWriter(w)

			// Log request information when the ServeHTTP returns.
			defer func() {
				logger.Info("request complete",
					"request_id", requestid.Get(ctx),
					"method", r.Method,
					"uri", uri,
					"source", r.RemoteAddr,
					"code", wr.Code,
					"written", wr.Written,
					"duration", time.Since(start),
				)
			}()

			next.ServeHTTP(wr, r)
		})
	}
}

// StatsMiddleware is an HTTP middleware function that records the HTTP duration
// in the provided StatsClient.
func StatsMiddleware(sc tools.StatsClient, name string) func(http.Handler) http.Handler {
	return StatsRouteMiddleware(sc, name, nil)
}

// StatsRouteMiddleware is an HTTP middleware function that records the HTTP
// duration and the route label in the provided StatsClient.
func StatsRouteMiddleware(sc tools.StatsClient, name string, labelFn func(r *http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Use a custom response writer.
			wr := withResponseWriter(w)

			// Record duration after request has completed.
			defer func() {
				durSeconds := float64(time.Since(start)) / float64(time.Second)
				labels := make([]string, 0, 2)
				labels = append(labels, strconv.Itoa(wr.Code))
				if labelFn != nil {
					labels = append(labels, labelFn(r))
				}
				sc.Histogram(name, durSeconds, labels)
			}()

			next.ServeHTTP(wr, r)
		})
	}
}

// RecoverMiddleware is an HTTP middleware function that gracefully recovers
// from panics and writes a 500 response if nothing has been written yet.
func RecoverMiddleware(logger tools.ErrorLogger, fn http.HandlerFunc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Use a custom response writer.
			wr := withResponseWriter(w)

			// Recover from any panics and respond with
			defer func() {
				rec := recover()
				if rec == nil {
					return
				}
				ctx := r.Context()
				trace := runtime.Stacktrace(0)
				logger.Error("recovered from panic",
					"request_id", requestid.Get(ctx),
					"details", fmt.Sprintf("%v", rec),
					"stack_trace", trace,
				)
				if wr.Code == 0 {
					// No response written yet, write one.
					fn(wr, r)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// MaxConnectionsMiddleware limits the number of open connections by responding
// with the provided HandlerFunc when the limit is reached.
func MaxConnectionsMiddleware(limit int, fn http.HandlerFunc) func(http.Handler) http.Handler {
	var open int32
	limit32 := int32(limit)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			val := atomic.AddInt32(&open, 1)
			defer atomic.AddInt32(&open, -1)
			if val > limit32 {
				fn(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ConcurrentLimitMiddleware limits the number of concurrent connections to the
// provided limit by blocking the current request (if necessary). If the context
// of the http request is cancelled, the provided HandlerFunc is invoked.
func ConcurrentLimitMiddleware(limit int, fn http.HandlerFunc) func(http.Handler) http.Handler {
	chSem := make(chan struct{}, limit)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			select {
			case chSem <- struct{}{}:
			case <-r.Context().Done():
				fn(w, r)
				return
			}
			defer func() { <-chSem }()
			next.ServeHTTP(w, r)
		})
	}
}

var (
	xForwardedFor = http.CanonicalHeaderKey("X-Forwarded-For")
	xRealIP       = http.CanonicalHeaderKey("X-Real-IP")
)

// realIP returns the HTTP request's source IP address as a string.
func realIP(r *http.Request) string {
	if xff := r.Header.Get(xForwardedFor); xff != "" {
		i := strings.Index(xff, ", ")
		if i == -1 {
			i = len(xff)
		}
		return xff[:i]
	}
	if xrip := r.Header.Get(xRealIP); xrip != "" {
		return xrip
	}
	return r.RemoteAddr
}

// responseWriter implements the HTTP ResponseWriter interface, while keeping
// track of the status code and number of bytes written.
type responseWriter struct {
	http.ResponseWriter
	Code    int
	Written int
}

func withResponseWriter(w http.ResponseWriter) *responseWriter {
	if wr, ok := w.(*responseWriter); ok {
		return wr
	}
	return &responseWriter{ResponseWriter: w}
}

// Header calls the underlying ResponseWriter's Header method.
func (rw *responseWriter) Header() http.Header {
	return rw.ResponseWriter.Header()
}

// Write ensures that WriteHeader has been called and then uses the underlying
// ResponseWriter's Write method, keeping track of bytes written.
func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.Code == 0 {
		rw.WriteHeader(200)
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.Written += n
	return n, err
}

// WriteHeader calls the underlying ResponseWriter's WriteHeader method, keeping
// track of the status code.
func (rw *responseWriter) WriteHeader(code int) {
	rw.Code = code
	rw.ResponseWriter.WriteHeader(code)
}

// limitedReadCloser wraps an io.ReadCloser and limits the number of bytes that
// can be read.
type limitedReadCloser struct {
	io.ReadCloser
	n int
}

func newLimitedReadCloser(limit int, rc io.ReadCloser) *limitedReadCloser {
	return &limitedReadCloser{
		ReadCloser: rc,
		n:          limit,
	}
}

func (l *limitedReadCloser) Read(p []byte) (int, error) {
	if l.n <= 0 {
		return 0, io.EOF
	}
	if len(p) > l.n {
		p = p[0:l.n]
	}
	n, err := l.ReadCloser.Read(p)
	l.n -= n
	return n, err
}
