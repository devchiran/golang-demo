package semaphore

import (
	"context"
	"errors"
	"sync/atomic"

	"golang.org/x/sync/semaphore"
)

// ErrTooManyQueued represents the error that is returned from Acquire when the
// requester is unable to obtain the semaphore, and the maximum number of queued
// requests is exceeded.
var ErrTooManyQueued = errors.New("semaphore: exceeded the maximum number of queued requests")

// Semaphore represents a simple semaphore that allows for the concurrent
// limiting of an entity, with a maximum number of requests that can be
// "queued".
type Semaphore struct {
	max   int64
	total int64
	sem   *semaphore.Weighted
}

// New returns an initialized Semaphore using the provided number of active
// requests (acquires before being released) and queued requests (maximum number
// of requests waiting to acquire the semaphore). If active is less than one, or
// queued is less than zero, New will panic.
func New(active, queued int) *Semaphore {
	if active < 1 {
		panic("semaphore: active count must be positive")
	}
	if queued < 0 {
		panic("semaphore: queued count cannot be negative")
	}
	return &Semaphore{
		max: int64(active + queued),
		sem: semaphore.NewWeighted(int64(active)),
	}
}

// Acquire is an alias for AcquireN(ctx, 1).
func (s *Semaphore) Acquire(ctx context.Context) error {
	return s.AcquireN(ctx, 1)
}

// Release is an alias for ReleaseN(1).
func (s *Semaphore) Release() {
	s.ReleaseN(1)
}

// AcquireN attempts to acquire the semaphore with a weight of 'n'. If the
// maximum number of queued requests are exceeded, it will return
// ErrTooManyQueued. If the provided context is cancelled, it will return the
// result of ctx.Err(). If the returned err is nil, the semaphore has been
// successfully acquired and Release must be called when the operation being
// limited is finished.
func (s *Semaphore) AcquireN(ctx context.Context, n int64) error {
	for {
		c := atomic.LoadInt64(&s.total)
		if c+n > s.max {
			return ErrTooManyQueued
		}
		if !atomic.CompareAndSwapInt64(&s.total, c, c+n) {
			continue
		}
		err := s.sem.Acquire(ctx, n)
		if err != nil {
			atomic.AddInt64(&s.total, -n)
		}
		return err
	}
}

// ReleaseN releases the semaphore from a previously successful call to AcquireN.
// ReleaseN must be called with the same weight 'n' as was acquired. If the
// semaphore count drops below zero, it will panic.
func (s *Semaphore) ReleaseN(n int64) {
	s.sem.Release(n)
	c := atomic.AddInt64(&s.total, -n)
	if c < 0 {
		panic("semaphore: bad release")
	}
}
