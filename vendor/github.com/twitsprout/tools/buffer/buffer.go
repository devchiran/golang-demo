package buffer

import (
	"bytes"
	"sync"
)

// globalPool represents a global buffer pool used by the functions Get and Put.
var globalPool = New()

// Pool represents a pool of buffers that can be concurrently retrieved and
// returned.
type Pool struct {
	pool sync.Pool
}

// New returns a new buffer pool that can be concurrently accessed.
func New() *Pool {
	return &Pool{
		pool: sync.Pool{
			New: func() interface{} {
				return &bytes.Buffer{}
			},
		},
	}
}

// Get returns a buffer ready for use. After, and ONLY after use, the buffer
// should be returned to the pool by calling the Put method.
func (p *Pool) Get() *bytes.Buffer {
	b := p.pool.Get().(*bytes.Buffer)
	b.Reset()
	return b
}

// Put returns the provided buffer to the pool in order to be reused by another
// goroutine. Do NOT return a buffer that is still being used in the current
// goroutine!
func (p *Pool) Put(b *bytes.Buffer) {
	if b != nil {
		p.pool.Put(b)
	}
}

// Get retrieves a fresh buffer from the shared pool.
func Get() *bytes.Buffer {
	return globalPool.Get()
}

// Put places a buffer back onto the shared pool. The buffer should no longer be
// used or memory retained.
func Put(b *bytes.Buffer) {
	globalPool.Put(b)
}
