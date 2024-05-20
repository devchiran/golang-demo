# Queue

The `queue` package provides interfaces that allow for the caller to defined a `Queue` to consume off and a `Handler` to trigger for processing the message.

## Installation

To install this package, use the following command:

```
go get github.com/twitsprout/tools/

```

## Usage

To use this package, first import it into your project:

```
import "github.com/twitsprout/tools/queue"

```

Next, implement a `Queue`:

A `Queue` defines how to interact with your queue. Our definition of a queue defines a visibility timeout that is able to extend the time the message would be not be visible on the queue. It could also be called acknowledgment deadline.

```go
type YourQueueClient struct {}

func (c *YourQueueClient) GetMessages(ctx context.Context, msg queue.GetMessagesRequest) ([]queue.Message, error) {
	var res []queue.Message
	// get messages from your queue
	return res, nil
}

func (c *YourQueueClient) AckMessage(ctx context.Context, msg queue.AckMessageRequest) error {
	// acknowledge messages on your queue
	return nil
}

func (c *YourQueueClient) UpdateVisibility(ctx context.Context, visibility queue.UpdateVisibilityRequest) error {
	// update visibility timeout for the message on your queue
	return nil
}

```

Then, implement `Handler` :

```go
type Processor struct {}

func (p *Processor) Handle(ctx context.Context, msg queue.Message) queue.HandleResult {
	// process message
	return queue.NewHandleResult().AckMessage(true)
}

```

Finally, start the consumer:

```
c := queue.NewConsumer(queueID, Queue).WithNumWorkers(5)
c.Consume(ctx, Handler)

```

In this example, the `Consume` function starts five workers to process items from the queue using the `Handle` function.

## How it works

### Consume

When consume is called a `sync.WaitGroup` is initialized with a unbuffered channel and for each worker

1. The wait group is incremented by 1
2. The worker is started in a Go routine being passed the wait group and channel

### Worker

A worker is started in its own Go routine. A worker is responsible for receiving a `PollMessage` and processing the message. The workers block until a message is passed to them where it is then processed or the channel is closed.

### Polling

Polling is done on a separate Go routine than the workers. Polling is done by attempting to retrieve a certain number of messages (currently set to the number of workers). If the service can long poll this is defined in the `waitTime`.

Once messages are pulled from the `Queue` , the message is wrapped inside a `PollMessage`. PollMessage will keep track of the context and visibility timeout of the individual message.

Each `PollMessage` is then sent to the provided unbuffered channel where it will block until all messages are consumed. The polling then repeats.

### PollMessage

PollMessage is a struct that contains a queue `Message` paired with a context of that Message's lifetime. The context is contained as part of this struct because it is passed from the polling goroutine to a worker, with each message maintaining their own unique context. Any processing of the Message should respect the context's cancellation. In the background, message visibility is extended periodically based on the visibility timeout of the message. Currently this is set to extend at 50% of the current visibility timeout. If the visibility time expires before it can be extended, the pollMessage context is cancelled. It is expected that the cleanup method of the pollMessage is called when processing is complete.