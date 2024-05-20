package mock

import (
	"context"
	"net/http"

	"github.com/twitsprout/tools"
)

// Logger implements the classic Logger interface for mocking purposes.
type Logger struct {
	DebugFn   func(string, ...interface{})
	ErrorFn   func(string, ...interface{})
	HandlerFn func() http.Handler
	InfoFn    func(string, ...interface{})
	WarnFn    func(string, ...interface{})
}

// Debug calls the Logger's DebugFn.
func (l *Logger) Debug(msg string, keyVals ...interface{}) {
	l.DebugFn(msg, keyVals...)
}

// Error calls the Logger's ErrorFn.
func (l *Logger) Error(msg string, keyVals ...interface{}) {
	l.ErrorFn(msg, keyVals...)
}

// Handler calls the Logger's HandlerFn.
func (l *Logger) Handler() http.Handler {
	return l.HandlerFn()
}

// Info calls the Logger's InfoFn.
func (l *Logger) Info(msg string, keyVals ...interface{}) {
	l.InfoFn(msg, keyVals...)
}

// Warn calls the Logger's WarnFn.
func (l *Logger) Warn(msg string, keyVals ...interface{}) {
	l.WarnFn(msg, keyVals...)
}

// LoggerCtx implements the LoggerCtx interface for mocking purposes.
type LoggerCtx struct {
	Logger
	DebugCtxFn func(context.Context, string, ...interface{})
	ErrorCtxFn func(context.Context, string, ...interface{})
	InfoCtxFn  func(context.Context, string, ...interface{})
	WarnCtxFn  func(context.Context, string, ...interface{})
}

// Debug calls the LoggerCtx's DebugFn.
func (l *LoggerCtx) Debug(msg string, keyVals ...interface{}) {
	l.DebugFn(msg, keyVals...)
}

// DebugCtx calls the LoggerCtx's DebugCtxFn.
func (l *LoggerCtx) DebugCtx(ctx context.Context, msg string, keyVals ...interface{}) {
	l.DebugCtxFn(ctx, msg, keyVals...)
}

// Error calls the LoggerCtx's ErrorFn.
func (l *LoggerCtx) Error(msg string, keyVals ...interface{}) {
	l.ErrorFn(msg, keyVals...)
}

// ErrorCtx calls the LoggerCtx's ErrorCtxFn.
func (l *LoggerCtx) ErrorCtx(ctx context.Context, msg string, keyVals ...interface{}) {
	l.ErrorCtxFn(ctx, msg, keyVals...)
}

// Handler calls the LoggerCtx's HandlerFn.
func (l *LoggerCtx) Handler() http.Handler {
	return l.HandlerFn()
}

// Info calls the LoggerCtx's InfoFn.
func (l *LoggerCtx) Info(msg string, keyVals ...interface{}) {
	l.InfoFn(msg, keyVals...)
}

// InfoCtx calls the LoggerCtx's InfoCtxFn.
func (l *LoggerCtx) InfoCtx(ctx context.Context, msg string, keyVals ...interface{}) {
	l.InfoCtxFn(ctx, msg, keyVals...)
}

// Warn calls the LoggerCtx's WarnFn.
func (l *LoggerCtx) Warn(msg string, keyVals ...interface{}) {
	l.WarnFn(msg, keyVals...)
}

// WarnCtx calls the LoggerCtx's WarnCtxFn.
func (l *LoggerCtx) WarnCtx(ctx context.Context, msg string, keyVals ...interface{}) {
	l.WarnCtxFn(ctx, msg, keyVals...)
}

// Ensure mocks implement the desired interfaces.
var _ tools.Logger = (*Logger)(nil)
var _ tools.LoggerCtx = (*LoggerCtx)(nil)

// NopLogger represents an empty logger that does nothing!
var NopLogger = &Logger{
	DebugFn:   func(string, ...interface{}) {},
	ErrorFn:   func(string, ...interface{}) {},
	HandlerFn: func() http.Handler { return nil },
	InfoFn:    func(string, ...interface{}) {},
	WarnFn:    func(string, ...interface{}) {},
}

// NopLoggerCtx represents an empty contextual logger that does nothing!
var NopLoggerCtx = &LoggerCtx{
	Logger:     *NopLogger,
	DebugCtxFn: func(context.Context, string, ...interface{}) {},
	ErrorCtxFn: func(context.Context, string, ...interface{}) {},
	InfoCtxFn:  func(context.Context, string, ...interface{}) {},
	WarnCtxFn:  func(context.Context, string, ...interface{}) {},
}
