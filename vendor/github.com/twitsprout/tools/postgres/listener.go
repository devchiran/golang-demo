package postgres

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/lib/pq"
)

// Message represents the message received from a PostgreSQL listener.
type Message struct {
	Channel string
	Payload string
}

// Listener represents a pubsub connection to a PostgreSQL database.
type Listener struct {
	lr        *pq.Listener
	chMessage chan Message

	mu      sync.Mutex
	closed  bool
	chClose chan struct{}
}

// NewListener returns a new Listener using the provided connection options and
// optional ping interval.
func NewListener(pingInterval time.Duration, ops Options) *Listener {
	// Format options.
	urlStr := urlFromOptions(ops)
	dialer := &dialer{
		d: &net.Dialer{Timeout: 30 * time.Second},
	}
	minDur := 100 * time.Millisecond
	maxDur := 30 * time.Second

	// Create listener.
	lr := pq.NewDialListener(dialer, urlStr, minDur, maxDur, nil)
	chClose := make(chan struct{})
	chMessage := make(chan Message, 80)
	l := &Listener{
		lr:        lr,
		chMessage: chMessage,
		chClose:   chClose,
	}

	// Start listener in a new goroutine.
	go l.listener(pingInterval)
	return l
}

// Close closes the underlying connection to the PostgreSQL database, returning
// any error encountered.
func (l *Listener) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return errors.New("postgres listener already closed")
	}
	l.closed = true
	close(l.chClose)
	return l.lr.Close()
}

// Listen causes the Listener to begin receiving messages for the provided
// channel.
func (l *Listener) Listen(channel string) error {
	return l.lr.Listen(channel)
}

// Messages returns the channel that messages received from PostgreSQL will be
// sent to.
func (l *Listener) Messages() <-chan Message {
	return l.chMessage
}

func (l *Listener) ping() error {
	return l.lr.Ping()
}

func (l *Listener) listener(pingInterval time.Duration) {
	if pingInterval > 0 {
		go l.pinger(pingInterval)
	}
	for {
		select {
		case <-l.chClose:
			return
		case n := <-l.lr.Notify:
			var msg Message
			if n == nil {
				msg.Channel = "unstable"
			} else {
				msg.Channel = n.Channel
				msg.Payload = n.Extra
			}
			select {
			case l.chMessage <- msg:
			case <-l.chClose:
				return
			}
		}

	}
}

func (l *Listener) pinger(dur time.Duration) {
	t := time.NewTicker(dur)
	defer t.Stop()
	for {
		select {
		case <-l.chClose:
			return
		case <-t.C:
		}
		_ = l.ping()
	}
}

type dialer struct {
	d *net.Dialer
}

func (d *dialer) Dial(network, address string) (net.Conn, error) {
	return d.d.Dial(network, address)
}

func (d *dialer) DialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return d.d.DialContext(ctx, network, address)
}
