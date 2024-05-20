package queue

import (
	"context"
	"sync"
	"time"
)

// pollMessage contains a queue Message paired with a context of that Message's
// lifetime. The context is contained as part of this struct because it is
// passed from the polling goroutine to a worker, with each message maintaining
// their own unique context. Any processing of the Message should respect the
// context's cancellation. In the background, message visibility is extended
// periodically based on the visibility timeout of the message. If the
// visibility time expires before it can be extended, the pollMessage context is
// cancelled. It is expected that the cleanup method of the pollMessage is
// called when processing is complete.
type pollMessage struct {
	ctx    context.Context
	cancel context.CancelFunc
	msg    Message
	c      *Consumer

	mu          sync.Mutex
	expiryTimer *time.Timer
	extendTimer *time.Timer
}

// registerTimers
func (pm *pollMessage) registerTimers() {
	if pm.c.visibilityTimeout <= 0 {
		return
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	select {
	case <-pm.ctx.Done():
		return
	default:
	}

	// First, cleanup any existing timers to avoid leaking memory.
	pm.unsafeCleanupTimers()

	// Cancel the message context if a certain percentage of the visibility
	// timeout is reached.
	// TODO (fowler): Make this configurable?
	const expiryPct = 0.9
	expiryDur := time.Duration(expiryPct * float64(pm.c.visibilityTimeout))
	pm.expiryTimer = time.AfterFunc(expiryDur, func() { pm.cancel() })

	// Wait for a certain percentage of the visibility timeout to be reached
	// before attempting to extend the visibility.
	// TODO (fowler): Make this configurable?
	const extendPct = 0.5
	extendDur := time.Duration((extendPct * float64(pm.c.visibilityTimeout)))
	pm.extendTimer = time.AfterFunc(extendDur, pm.extend)
}

func (pm *pollMessage) extend() {
	defer pm.registerTimers()

	select {
	case <-pm.ctx.Done():
		return
	default:

	}

	// Attempt to extend the visibility timeout, backing off and retrying in
	// the case of an error.
	var retries int
	for {
		err := pm.updateVisibility()
		if err == nil {
			return
		}
		pm.c.handleError(err)
		retries++
		select {
		case <-pm.ctx.Done():
			return
		case <-time.After(time.Duration(retries) * time.Second):
		}
	}
}

func (pm *pollMessage) updateVisibility() error {
	const requestTimeout = 10 * time.Second
	ctx, cancel := context.WithTimeout(pm.ctx, requestTimeout)
	defer cancel()
	return pm.c.queue.UpdateVisibility(ctx, UpdateVisibilityRequest{
		QueueID:           pm.c.queueID,
		ReceiptHandle:     pm.msg.ReceiptHandle,
		VisibilityTimeout: pm.c.visibilityTimeout,
	})
}

func (pm *pollMessage) cleanup() {
	pm.cancel()
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.unsafeCleanupTimers()
}

func (pm *pollMessage) unsafeCleanupTimers() {
	if pm.expiryTimer != nil {
		pm.expiryTimer.Stop()
		pm.expiryTimer = nil
	}
	if pm.extendTimer != nil {
		pm.extendTimer.Stop()
		pm.extendTimer = nil
	}
}
