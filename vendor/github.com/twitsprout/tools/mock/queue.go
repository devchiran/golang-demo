package mock

import (
	"context"

	"github.com/twitsprout/tools/queue"
)

var _ queue.Queue = (*Queue)(nil)

// Queue implements the queue package's Queue interface for mocking purposes.
type Queue struct {
	AckMessageFn       func(context.Context, queue.AckMessageRequest) error
	GetMessagesFn      func(context.Context, queue.GetMessagesRequest) ([]queue.Message, error)
	UpdateVisibilityFn func(context.Context, queue.UpdateVisibilityRequest) error
}

// AckMessage calls the Queue's AckMessageFn.
func (q *Queue) AckMessage(ctx context.Context, r queue.AckMessageRequest) error {
	return q.AckMessageFn(ctx, r)
}

// GetMessages calls the Queue's GetMessagesFn.
func (q *Queue) GetMessages(ctx context.Context, r queue.GetMessagesRequest) ([]queue.Message, error) {
	return q.GetMessagesFn(ctx, r)
}

// UpdateVisibility calls the Queue's UpdateVisibilityFn.
func (q *Queue) UpdateVisibility(ctx context.Context, r queue.UpdateVisibilityRequest) error {
	return q.UpdateVisibilityFn(ctx, r)
}
