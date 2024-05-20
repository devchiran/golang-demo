package postgres

import (
	"context"
	"time"

	"github.com/twitsprout/tools/sync/semaphore"
)

// Semaphore represents the interface for a generic semaphore implementation.
// Acquire is called to acquire the semaphore, blocking until it is successful
// or the provided context is cancelled. If a nil error is returned, the caller
// must call Release when it is finished with the protected operation.
type Semaphore interface {
	Acquire(context.Context) error
	Release()
}

// Clock represents the interface for returning the current time.
type Clock interface {
	Now() time.Time
}

// Config represents the basic options when initializing a DB instance.
type Config struct {
	DisableSSL bool
	Host       string
	Name       string
	Password   string
	Port       int
	Username   string
}

// Option represents various optional Options that can be used when initializing
// a DB instance. All Options provided by this package start with a "With"
// prefix.
type Option func(*options)

// WithClock sets Clock that the DB uses to 'c'.
func WithClock(c Clock) Option {
	return func(ops *options) {
		ops.clock = c
	}
}

// WithIdleConns sets the DB to have at most 'n' idle connections at a time.
func WithIdleConns(n int) Option {
	return func(ops *options) {
		ops.maxIdleConns = n
	}
}

// WithOnComplete will set the DB to invoke 'fn' after every call to the DB's
// Do method.
func WithOnComplete(fn func(context.Context, string, time.Time, error) error) Option {
	return func(ops *options) {
		ops.onComplete = fn
	}
}

// WithMaxConnLifetime sets the maximum connection lifetime to 'dur'.
func WithMaxConnLifetime(dur time.Duration) Option {
	return func(ops *options) {
		ops.maxConnLifetime = dur
	}
}

// WithSemaphore sets the DB to use the provided Semaphore instance. It is valid
// to provide a "nil" value.
func WithSemaphore(s Semaphore) Option {
	return func(ops *options) {
		ops.semaphore = s
	}
}

// WithTimeout sets a timeout of 'dur' in the context that the function provided
// to the Do method is invoked with. It is valid to provide a dur of <= 0 if no
// timeout is wanted.
func WithTimeout(dur time.Duration) Option {
	return func(ops *options) {
		ops.timeout = dur
	}
}

type options struct {
	clock           Clock
	maxConnLifetime time.Duration
	maxIdleConns    int
	onComplete      func(context.Context, string, time.Time, error) error
	semaphore       Semaphore
	timeout         time.Duration
}

func defaultOptions() options {
	return options{
		maxConnLifetime: 30 * time.Minute,
		maxIdleConns:    30,
		onComplete:      nil,
		semaphore:       semaphore.New(30, 420),
		timeout:         120 * time.Second,
	}
}
