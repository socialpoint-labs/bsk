//go:build integration
// +build integration

package sqsx_test

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/socialpoint-labs/bsk/awsx/awstest"
	"github.com/socialpoint-labs/bsk/awsx/sqsx"
	"github.com/stretchr/testify/assert"
)

func TestWatch(t *testing.T) {
	a := assert.New(t)

	payload := "test1"
	url := getTestQueue(t)
	sendMessage(t, url, payload)

	messages := make(chan *sqs.Message)

	sess := awstest.NewSession()
	cli := sqs.New(sess)

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
	a := assert.New(t)

	payload := "test3"
	url := getTestQueue(t)
	sendMessage(t, url, payload)

	messages := make(chan *sqs.Message)

	sess := awstest.NewSession()
	cli := sqs.New(sess)

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

func getTestQueue(t *testing.T) string {
	a := assert.New(t)

	queue := awstest.CreateResource(t, sqs.ServiceName)
	awstest.AssertResourceExists(t, queue, sqs.ServiceName)

	sess := awstest.NewSession()
	cli := sqs.New(sess)

	url, err := sqsx.GetQueueURL(context.Background(), cli, queue)
	a.NoError(err)

	return url
}

func sendMessage(t *testing.T, url string, payload string) {
	a := assert.New(t)

	svc := sqs.New(awstest.NewSession())

	input := &sqs.SendMessageInput{
		QueueUrl:    aws.String(url),
		MessageBody: aws.String(payload),
	}

	_, err := svc.SendMessage(input)
	a.NoError(err)
}
