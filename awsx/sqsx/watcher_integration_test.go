//go:build integration
// +build integration

package sqsx_test

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/socialpoint-labs/bsk/awsx/awstest"
	"github.com/socialpoint-labs/bsk/awsx/sqsx"
)

func TestWatch(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	url := arrangeQueue(t)
	cli := arrangeClient()
	sendMessage(t, url, payload)

	messages := make(chan *sqs.Message)

	f := func(msg *sqs.Message) error {
		messages <- msg
		return nil
	}

	e := func(err error) {}

	runner := sqsx.WatchRunner(cli, url, f, e)

	ctx, cancel := context.WithCancel(context.Background())
	go runner.Run(ctx)

	received := <-messages

	a.Equal(payload, *received.Body)

	cancel()
}

func TestWatchError(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	url := arrangeQueue(t)
	cli := arrangeClient()
	sendMessage(t, url, payload)

	messages := make(chan *sqs.Message)

	spyCall := 0
	f := func(msg *sqs.Message) error {
		spyCall++
		if spyCall == 1 {
			return errors.New("error on first call in order to execute ChangeMsgVisibilityTimeout")
		}
		messages <- msg
		return nil
	}

	e := func(err error) {}

	runner := sqsx.WatchRunner(cli, url, f, e)

	ctx, cancel := context.WithCancel(context.Background())
	go runner.Run(ctx)

	received := <-messages

	a.Equal(payload, *received.Body)
	a.Equal(2, spyCall)

	cancel()
}

func sendMessage(t *testing.T, url string, payload string) {
	svc := sqs.New(awstest.NewSession())

	input := &sqs.SendMessageInput{
		QueueUrl:    aws.String(url),
		MessageBody: aws.String(payload),
	}

	_, err := svc.SendMessage(input)
	require.NoError(t, err)
}
