package requestid

import (
	"context"
	"encoding/base64"
	"fmt"
	"sync/atomic"

	"github.com/twitsprout/tools/crypto"
)

type contextKeyType int

const requestIDKey contextKeyType = 0

var (
	counter uint64
	session string
)

func init() {
	// Create random session string.
	buf := make([]byte, 8)
	crypto.ReadRandUnsafe(buf)
	session = base64.URLEncoding.EncodeToString(buf)
	if len(session) > 8 {
		session = session[:8]
	}
}

// Get returns the request ID stored in the provided context. If no
// request ID exists, an empty string is returned.
func Get(ctx context.Context) string {
	reqID, _ := ctx.Value(requestIDKey).(string)
	return reqID
}

// New returns a new request ID as a string.
func New() string {
	cnt := atomic.AddUint64(&counter, 1)
	return fmt.Sprintf("%s-%010d", session, cnt)
}

// WithRequestID reutrns a new child context with a request ID set as a value.
func WithRequestID(ctx context.Context) context.Context {
	return context.WithValue(ctx, requestIDKey, New())
}
