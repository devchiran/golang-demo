package clock

import "time"

// Clock is the interface for working with time.
type Clock interface {
	Now() time.Time
}

// Default is an implementation of Clock that uses the real time.
type Default struct{}

// Now returns the current time.
func (d *Default) Now() time.Time {
	return time.Now()
}
