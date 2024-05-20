package distlock

import (
	"context"
	"errors"
	"sync"
	"time"
)

// ErrLockNotHeld is the error returned when a Lock, Extend, or Unlock operation
// is attempted and the operation was not successful because another service
// currently holds the lock.
var ErrLockNotHeld = errors.New("lock not held by current process")

// Locker is the interface which manages the lock state.
type Locker interface {
	// Extend extends the TTL of the provided lock. If the lock isn't held
	// by the instanceID, ErrLockNotHeld should be returned.
	Extend(ctx context.Context, instanceID, lockID string, ttlSeconds int) error

	// Lock attempts to obtain the lock for the provided lockID. If the lock
	// is already held by another service, ErrLockNotHeld should be returned.
	Lock(ctx context.Context, instanceID, lockID string, ttlSeconds int) error

	// Unlock releases the lock for provided instance/lock ID. If the lock
	// is not held by the current service, ErrLockNotHeld should be returned.
	Unlock(ctx context.Context, instanceID, lockID string) error
}

// DistributedLock controls the execution of a function for a distributed lock
// implementation.
type DistributedLock struct {
	ErrorFunc         func(error)
	InstanceID        string
	Locker            Locker
	LockID            string
	MaxRetries        int
	RetryBaseDuration time.Duration
	TTLSeconds        int
	UnlockTimeout     time.Duration
}

// Do accepts a context and a DoFunc, that will only be called if the lock is
// acquired successfully. If an error is encountered obtaining the lock, Do will
// will return immediately with (false, err). If the lock was already held by
// another process, (false, nil) is returned.
// If the lock was obtain successfully, the provided DoFunc is invoked with a
// new context who's timeout MUST be respected. Do attempts to extend the lock
// every TTL / 2. If this call fails, the optional ErrorFunc is invoked, and the
// extend call is retried. If the TTL expires before the lock could be extended,
// the context provided to the DoFunc is cancelled, and should immediately
// return. Do will return (true, err), where "err" is the error returned by the
// DoFunc.
func (d *DistributedLock) Do(ctx context.Context, fn func(context.Context)) (err error) {
	// Attempt to obtain the lock.
	err = d.Locker.Lock(ctx, d.InstanceID, d.LockID, d.TTLSeconds)
	if err != nil {
		return
	}

	// Create the context that will be cancelled if unable to extend the
	// lock before the TTL expires.
	ctx, cancel := context.WithCancel(ctx)

	// Defer unlock
	var wg sync.WaitGroup
	defer func() {
		cancel()
		wg.Wait()
		uCtx, uCancel := context.WithTimeout(context.Background(), d.UnlockTimeout)
		err = d.Locker.Unlock(uCtx, d.InstanceID, d.LockID)
		uCancel()
	}()

	// Start the extender process in a new goroutine.
	wg.Add(1)
	go func() {
		d.extender(ctx, cancel)
		cancel()
		wg.Done()
	}()

	fn(ctx)
	return
}

func (d *DistributedLock) extender(ctx context.Context, cancel context.CancelFunc) {
	// Expire context if 90% of the TTL reached.
	renewInterval := time.Duration(d.TTLSeconds) * time.Second * 9 / 10
	chRenewExpiry := make(chan struct{})
	go func() {
		renewExpiry(ctx, renewInterval, chRenewExpiry)
		cancel()
	}()

	// Attempt to extend the TTL on the lock every half TTL.
	extendInterval := time.Duration(d.TTLSeconds) * time.Second / 2
	t := time.NewTimer(extendInterval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
		if !d.extend(ctx) {
			return
		}
		select {
		case chRenewExpiry <- struct{}{}:
		default:
		}
		resetTimer(t, extendInterval)
	}
}

func (d *DistributedLock) extend(ctx context.Context) bool {
	var retries int
	for {
		// Attempt to extend TTL on lock.
		err := d.Locker.Extend(ctx, d.InstanceID, d.LockID, d.TTLSeconds)
		if err == nil {
			return true
		}
		if d.ErrorFunc != nil {
			d.ErrorFunc(err)
		}
		if err == ErrLockNotHeld {
			return false
		}
		retries++
		if retries > d.MaxRetries {
			return false
		}

		dur := time.Duration(retries*retries) * d.RetryBaseDuration
		select {
		case <-ctx.Done():
			return false
		case <-time.After(dur):
		}
	}
}

func renewExpiry(ctx context.Context, dur time.Duration, chRenewExpiry <-chan struct{}) {
	t := time.NewTimer(dur)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			return
		case <-chRenewExpiry:
			resetTimer(t, dur)
		}
	}
}

func resetTimer(t *time.Timer, dur time.Duration) {
	if !t.Stop() {
		select {
		case <-t.C:
		default:
		}
	}
	t.Reset(dur)
}
