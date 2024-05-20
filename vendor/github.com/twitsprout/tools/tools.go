package tools

import (
	"context"
	"net/http"
)

// DebugLogger is the interface for logging a debug message.
type DebugLogger interface {
	Debug(string, ...interface{})
}

// ErrorLogger is the interface for logging an error message.
type ErrorLogger interface {
	Error(string, ...interface{})
}

// InfoLogger is the interface for logging an info message.
type InfoLogger interface {
	Info(string, ...interface{})
}

// WarnLogger is the interface for logging a warning message.
type WarnLogger interface {
	Warn(string, ...interface{})
}

// Logger represents the generic logging interface.
type Logger interface {
	DebugLogger
	ErrorLogger
	Handler() http.Handler
	InfoLogger
	WarnLogger
}

// DebugLoggerCtx is the interface for logging a contextual debug message.
type DebugLoggerCtx interface {
	DebugCtx(context.Context, string, ...interface{})
}

// ErrorLoggerCtx is the interface for logging a contextual error message.
type ErrorLoggerCtx interface {
	ErrorCtx(context.Context, string, ...interface{})
}

// InfoLoggerCtx is the interface for logging a contextual info message.
type InfoLoggerCtx interface {
	InfoCtx(context.Context, string, ...interface{})
}

// WarnLoggerCtx is the interface for logging a contextual warning message.
type WarnLoggerCtx interface {
	WarnCtx(context.Context, string, ...interface{})
}

// LoggerCtx represents the generic contextual logging interface.
type LoggerCtx interface {
	Logger
	DebugLoggerCtx
	ErrorLoggerCtx
	InfoLoggerCtx
	WarnLoggerCtx
}

// StatsClient is the interface for the metrics collecting client.
type StatsClient interface {
	Count(name string, incBy float64, labels []string)
	Gauge(name string, value float64, labels []string)
	Handler() http.Handler
	Histogram(name string, value float64, labels []string)
}
