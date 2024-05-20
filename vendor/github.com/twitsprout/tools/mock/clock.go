package mock

import (
	"time"

	"github.com/twitsprout/tools/clock"
)

var _ clock.Clock = (*Clock)(nil)

// Clock is a mock implementation of the Clock interface.
type Clock struct {
	NowFn func() time.Time
}

// Now returns the result of calling Clock's NowFn.
func (c *Clock) Now() time.Time {
	return c.NowFn()
}
