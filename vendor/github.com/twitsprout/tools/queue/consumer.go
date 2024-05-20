package queue

import (
	"context"
	"runtime"
	"sync"
	"time"

	"github.com/twitsprout/tools/clock"
)

// Queue is the interface for receiving, updating visibility, and acknowledging
// messages from a queue.
type Queue interface {
	AckMessage(context.Context, AckMessageRequest) error
	GetMessages(context.Context, GetMessagesRequest) ([]Message, error)
	UpdateVisibility(context.Context, UpdateVisibilityRequest) error
}

// Handler is the interface that gets invoked by a Consumer for each message
// that it pulls from the queue.
type Handler interface {
	Handle(context.Context, Message) HandleResult
}

// HandlerFunc implements the Handler interface, allowing for the use of a
// function directly.
type HandlerFunc func(context.Context, Message) HandleResult

func (h HandlerFunc) Handle(ctx context.Context, msg Message) HandleResult {
	return h(ctx, msg)
}

// Consumer allows for concurrently consuming messages from a queue.
type Consumer struct {
	clock             clock.Clock
	errHandler        ErrHandler
	numWorkers        int
	queue             Queue
	queueID           string
	visibilityTimeout time.Duration
	waitTime          time.Duration
}

// NewConsumer returns an initialized Consumer with the default settings.
func NewConsumer(queueID string, q Queue) *Consumer {
	return &Consumer{
		clock:             &clock.Default{},
		numWorkers:        runtime.NumCPU(),
		queue:             q,
		queueID:           queueID,
		visibilityTimeout: 30 * time.Second,
		waitTime:          20 * time.Second,
	}
}

// WithErrHandler updates the Consumer to use the provided ErrHandler for when
// errors when interacting with the queue occur.
func (c *Consumer) WithErrHandler(e ErrHandler) *Consumer {
	c.errHandler = e
	return c
}

// WithNumWorkers updates the Consumer to use the provided number of concurrent
// workers.
func (c *Consumer) WithNumWorkers(n int) *Consumer {
	c.numWorkers = n
	return c
}

// WithVisibilityTimeout updates the Consumer to use the provided visibility
// timeout when fetching messages from the queue.
func (c *Consumer) WithVisibilityTimeout(vis time.Duration) *Consumer {
	c.visibilityTimeout = vis
	return c
}

// WithWaitTime updates the Consumer to use the provided wait time when fetching
// messages from the queue.
func (c *Consumer) WithWaitTime(wait time.Duration) *Consumer {
	c.waitTime = wait
	return c
}

// Consume polls the queue and invokes the provided Handler with each Message.
// It blocks until the provided context is cancelled, where it will return the
// result of ctx.Err().
func (c *Consumer) Consume(ctx context.Context, h Handler) error {
	// Consume works by spinning up a number of "worker" goroutines and then
	// polling messages in bulk, distributing individual messages to the
	// workers until the polling channel is closed. Consume then waits for
	// all worker goroutines to exit before returning.

	var wg sync.WaitGroup
	ch := make(chan *pollMessage)
	for i := 0; i < c.numWorkers; i++ {
		wg.Add(1)
		go c.startWorker(ctx, h, &wg, ch)
	}

	err := c.poll(ctx, ch)
	close(ch)

	wg.Wait()
	return err
}

// poll continues to poll messages from the queue, sending them to the provided
// channel until the context is cancelled.
func (c *Consumer) poll(ctx context.Context, ch chan<- *pollMessage) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if err := c.pollMessages(ctx, ch); err != nil {
			c.handleError(err)
		}
	}
}

// pollMessages pulls messages from the queue, sending each message with all of
// its context to the provided channel.
func (c *Consumer) pollMessages(ctx context.Context, ch chan<- *pollMessage) error {
	msgs, err := c.getMessages(ctx)
	if err != nil || len(msgs) == 0 {
		return err
	}

	// For each message received, create a unique context, and register any
	// extending/expiry timers based on the visibility timeout.
	pollMsgs := make([]*pollMessage, 0, len(msgs))
	for _, msg := range msgs {
		ctx, cancel := context.WithCancel(context.Background())
		pm := &pollMessage{
			ctx:    ctx,
			cancel: cancel,
			msg:    msg,
			c:      c,
		}
		pm.registerTimers()
		pollMsgs = append(pollMsgs, pm)
	}

	// Send each message to an available worker process. If the context has
	// been cancelled, cleanup each message (cancelling all timers). If sent
	// to a worker, the receiving worker process is expected to call the
	// cleanup method for each message it accepts.
	for _, pm := range pollMsgs {
		select {
		case <-ctx.Done():
			pm.cleanup()
		case ch <- pm:
		}
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func (c *Consumer) getMessages(ctx context.Context) ([]Message, error) {
	// The Consumer WaitTime is forwarded to the queue's GetMessages method,
	// which means the call will block for up to the specificed time.
	// Therefore, it is used to calculate the requestTimeout below so that
	// the context isn't cancelled while the request is in an intentional
	// blocking status.
	requestTimeout := maxDuration(30*time.Second, c.waitTime+time.Second)
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()
	return c.queue.GetMessages(ctx, GetMessagesRequest{
		// For now, use the number of workers as the number of messages
		// to fetch in a given request. We may want to expose this as a
		// setting in the future.
		MessageCount:      c.numWorkers,
		QueueID:           c.queueID,
		VisibilityTimeout: c.visibilityTimeout,
		WaitTime:          c.waitTime,
	})
}

// startWorker processes messages from the provided channel until it is closed.
func (c *Consumer) startWorker(ctx context.Context, h Handler, wg *sync.WaitGroup, ch <-chan *pollMessage) {
	defer wg.Done()
	for {
		pm, ok := <-ch
		if !ok {
			return
		}
		err := c.consumeMessage(ctx, h, pm)
		if err != nil {
			c.handleError(err)
		}
	}
}

// consumeMessage invokes the Handler with the provided Message, acknowledging
// the Message with the backing queue if instructed to.
func (c *Consumer) consumeMessage(ctx context.Context, h Handler, pm *pollMessage) error {
	res := handleMsg(h, pm)
	if !shouldAckMessage(res) {
		return nil
	}
	return c.ackMessage(ctx, pm.msg.ReceiptHandle)
}

func (c *Consumer) ackMessage(ctx context.Context, receiptHandle string) error {
	const timeout = 10 * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return c.queue.AckMessage(ctx, AckMessageRequest{
		QueueID:       c.queueID,
		ReceiptHandle: receiptHandle,
	})
}

func handleMsg(h Handler, pm *pollMessage) HandleResult {
	defer pm.cleanup()
	return h.Handle(pm.ctx, pm.msg)
}

func (c *Consumer) handleError(err error) {
	if err != nil && c.errHandler != nil {
		c.errHandler(err)
	}
}

// HandleResult represents the result of processing a Message.
type HandleResult struct {
	shouldAck bool
}

// NewHandleResult returns an empty HandleResult, resulting in no action taken
// by the Consumer.
func NewHandleResult() HandleResult {
	return HandleResult{}
}

// AckMessage returns a new HandleResult that indicates to the Consumer that it
// should Acknowledge the message with the Queue.
func (r HandleResult) AckMessage(b bool) HandleResult {
	r.shouldAck = b
	return r
}

func shouldAckMessage(r HandleResult) bool {
	return r.shouldAck
}

// ErrHandler represents a callback function that is envoked by a Consumer upon
// any queue errors.
type ErrHandler func(error)

func maxDuration(a, b time.Duration) time.Duration {
	if a > b {
		return a
	}
	return b
}
