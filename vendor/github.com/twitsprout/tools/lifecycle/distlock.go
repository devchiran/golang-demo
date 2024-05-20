package lifecycle

import (
	"context"
	"fmt"
	"time"

	"github.com/twitsprout/tools/crypto"
	"github.com/twitsprout/tools/distlock"
)

// StartDistLock starts a distributed lock process that is run every [spinMin,
// spinMax) time duration, using the provided function.
func (lc *LifeCycle) StartDistLock(dl *distlock.DistributedLock, spinMin, spinMax time.Duration, fn func(context.Context)) {
	name := fmt.Sprintf("distlock '%s'", dl.LockID)
	lc.Start(name, func() error {
		// Start initial distlock attempt betweem [0, spinMin).
		dur := crypto.PRandInt64(0, int64(spinMin))
		select {
		case <-lc.ctx.Done():
			return lc.ctx.Err()
		case <-time.After(time.Duration(dur)):
			lc.doDistlock(dl, fn)
		}
		// Run distlock every [spinMin, spinMax).
		for {
			dur = crypto.PRandInt64(int64(spinMin), int64(spinMax))
			select {
			case <-lc.ctx.Done():
				return lc.ctx.Err()
			case <-time.After(time.Duration(dur)):
				lc.doDistlock(dl, fn)
			}
		}
	})
}

func (lc *LifeCycle) doDistlock(dl *distlock.DistributedLock, fn func(context.Context)) {
	var executed bool
	err := dl.Do(lc.ctx, func(ctx context.Context) {
		executed = true
		fn(ctx)
	})
	switch err {
	case distlock.ErrLockNotHeld:
		lc.logger.Debug("distlock held by another process",
			"lock_id", dl.LockID,
		)
	case nil:
		lc.logger.Info("distlock completed successfully",
			"lock_id", dl.LockID,
		)
	default:
		select {
		case <-lc.ctx.Done():
		default:
			lc.logger.Warn("distlock failed",
				"lock_id", dl.LockID,
				"executed", executed,
				"details", err.Error(),
			)
		}
	}
}
