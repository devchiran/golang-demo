package queue

import "time"

// Message represents a single message received from a queue.
type Message struct {
	// Attempts represents the total number of attempts made to process this
	//  message. If below 1, this field may not be supported by the queue.
	Attempts int
	// Body contains the raw bytes of the message.
	Body []byte
	// ID represents the message's internal identifier.
	ID string
	// ReceiptHandle is the identifier used to interact with the message via
	// the queue.
	ReceiptHandle string
}

// GetMessagesRequest contains the parameters for a request to fetch messages
// from a queue. Only the QueueID field is required.
type GetMessagesRequest struct {
	MessageCount      int
	QueueID           string
	VisibilityTimeout time.Duration
	WaitTime          time.Duration
}

// AckMessageRequest contains the parameters for a request to acknowledge a
// message from a queue. Both the QueueID and ReceiptHandle must be provided.
type AckMessageRequest struct {
	QueueID       string
	ReceiptHandle string
}

// UpdateVisibilityRequest contains the parameters for a request to update the
// visibility timeout of a message in a queue. All fields on this struct are
// required.
type UpdateVisibilityRequest struct {
	QueueID           string
	ReceiptHandle     string
	VisibilityTimeout time.Duration
}

// PostMessagesRequest contains the parameters for a request to send messages
// to a queue. All fields on this struct are required.
type PostMessagesRequest struct {
	Messages []PostMessage
	QueueID  string
}

// PostMessage represents the information for a single message being sent to a
// queue.
type PostMessage struct {
	Body []byte
}

// PostMessageResult represents the result for attempting to send a single
// message to a queue. If Err is non-nil, the message was not sent successfully
// and it should contain the reason why.
type PostMessageResult struct {
	ID  string
	Err error
}
