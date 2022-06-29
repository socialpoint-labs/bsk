package sqsx

import (
	"context"

	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"

	"github.com/socialpoint-labs/bsk/contextx"
)

// OnMessage is the type that callback functions must satisfy
type OnMessage func(msg *sqs.Message) error

// OnError is the type that errors callback function must satisfy
type OnError func(error)

// WatchRunner returns a runner that watch a queue, using the runner's context
func WatchRunner(cli sqsiface.SQSAPI, url string, f OnMessage, e OnError) contextx.Runner {
	return contextx.RunnerFunc(func(ctx context.Context) {
		Watch(ctx, cli, url, f, e)
	})
}

// Watch watches a queue a call the callback function on messages or errors
func Watch(ctx context.Context, cli sqsiface.SQSAPI, url string, f OnMessage, e OnError, opts ...Option) {
	for {
		select {
		case <-ctx.Done():
			return

		default:
			msg, err := ReceiveMessage(ctx, cli, url, opts...)
			if err != nil {
				// There was an error receiving the message
				e(err)
				continue
			}

			if msg == nil {
				// There were no messages in the queue, let's try again
				// No need to sleep, because internally the SDK does long-polling
				continue
			}

			err = f(msg)

			if err != nil {
				// If the callback function returns an error, leave the message in the queue
				e(err)

				// return the message back to the queue by reseting the visibility timeout
				if _, errChange := ChangeMsgVisibilityTimeout(ctx, cli, url, msg.ReceiptHandle, 0); errChange != nil {
					e(errChange)
				}

				continue
			}

			err = DeleteMessage(ctx, msg.ReceiptHandle, cli, url)
			if err != nil {
				// There was an error removing the message from the queue, so probably the message
				// is still in the queue and will receive it again (although we will never know),
				// so be prepared to process the message again without side effects.
				e(err)
				continue
			}
		}
	}
}
