// This dynamic sempahore implementation is based on the original work by the
// Go Authors in "golang.org/x/sync/semaphore", with major modifications made.
//
// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package semaphore

import (
	"container/list"
	"context"
	"sync"
)

// NewDynamic creates a new weighted semaphore with the given maximum combined
// weight for concurrent access.
func NewDynamic(n int64) *Dynamic {
	w := &Dynamic{size: n}
	return w
}

// Dynamic provides a way to bound concurrent access to a resource.
// The callers can request access with a given weight.
type Dynamic struct {
	mu      sync.Mutex
	cur     int64
	size    int64
	waiters list.List
}

// Acquire is an alias for AcquireN(ctx, 1).
func (d *Dynamic) Acquire(ctx context.Context) error {
	return d.AcquireN(ctx, 1)
}

// TryAcquire is an alias for TryAcquireN(1).
func (d *Dynamic) TryAcquire() bool {
	return d.TryAcquireN(1)
}

// Release is an alias for ReleaseN(1).
func (d *Dynamic) Release() {
	d.ReleaseN(1)
}

// AcquireN acquires the semaphore with a weight of n, blocking only until ctx
// is done. On success, returns nil. On failure, returns ctx.Err() and leaves
// the semaphore unchanged.
//
// If ctx is already done, Acquire may still succeed without blocking.
func (d *Dynamic) AcquireN(ctx context.Context, n int64) error {
	d.mu.Lock()
	if d.size-d.cur >= n && d.waiters.Len() == 0 {
		d.cur += n
		d.mu.Unlock()
		return nil
	}

	if n > d.size && d.size > 0 {
		// Don't make other Acquire calls block on one that's doomed to fail.
		d.mu.Unlock()
		<-ctx.Done()
		return ctx.Err()
	}

	ready := make(chan struct{})
	w := waiter{n: n, ready: ready}
	elem := d.waiters.PushBack(w)
	d.mu.Unlock()

	select {
	case <-ctx.Done():
		err := ctx.Err()
		d.mu.Lock()
		select {
		case <-ready:
			// Acquired the semaphore after we were canceled.  Rather than trying to
			// fix up the queue, just pretend we didn't notice the cancelation.
			err = nil
		default:
			d.waiters.Remove(elem)
		}
		d.mu.Unlock()
		return err
	case <-ready:
		return nil
	}
}

// TryAcquireN acquires the semaphore with a weight of n without blocking.
// On success, returns true. On failure, returns false and leaves the semaphore unchanged.
func (d *Dynamic) TryAcquireN(n int64) bool {
	d.mu.Lock()
	success := d.size-d.cur >= n && d.waiters.Len() == 0
	if success {
		d.cur += n
	}
	d.mu.Unlock()
	return success
}

// ReleaseN releases the semaphore with a weight of n.
func (d *Dynamic) ReleaseN(n int64) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.cur -= n
	if d.cur < 0 {
		panic("semaphore: bad release")
	}
	d.lockedRelease()
}

// SetMaxWeight safely updates the maximum combined weight for concurrent
// access to the semaphore, making it "dynamic".
func (d *Dynamic) SetMaxWeight(n int64) {
	d.mu.Lock()
	defer d.mu.Unlock()
	oldSize := d.size
	d.size = n
	if n > oldSize {
		d.lockedRelease()
	}
}

type waiter struct {
	n     int64
	ready chan<- struct{} // Closed when semaphore acquired.
}

func (d *Dynamic) lockedRelease() {
	for {
		next := d.waiters.Front()
		if next == nil {
			return // No more waiters blocked.
		}

		w := next.Value.(waiter)
		if d.size-d.cur < w.n {
			// Not enough tokens for the next waiter.  We could keep going (to try to
			// find a waiter with a smaller request), but under load that could cause
			// starvation for large requests; instead, we leave all remaining waiters
			// blocked.
			//
			// Consider a semaphore used as a read-write lock, with N tokens, N
			// readers, and one writer.  Each reader can Acquire(1) to obtain a read
			// lock.  The writer can Acquire(N) to obtain a write lock, excluding all
			// of the readers.  If we allow the readers to jump ahead in the queue,
			// the writer will starve — there is always one token available for every
			// reader.
			return
		}

		d.cur += w.n
		d.waiters.Remove(next)
		close(w.ready)
	}
}
