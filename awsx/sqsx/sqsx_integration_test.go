//go:build integration
// +build integration

package sqsx_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/socialpoint-labs/bsk/awsx/awstest"
	"github.com/socialpoint-labs/bsk/awsx/sqsx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const payload = "random-payload:life-is-to-short-to-generate-a-really-random-payload-when-it-should-not-be-random-at-all"

func Test_Send_And_Receive(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	// arrange
	url := arrangeQueue(t)
	cli := arrangeClient()
	ctx := context.Background()

	// act
	_, err1 := sqsx.SendMessage(ctx, cli, url, payload)
	msg, err2 := sqsx.ReceiveMessage(ctx, cli, url, sqsx.WithVisibilityTimeout(2), sqsx.WithWaitTime(2))

	// assert
	a.NoError(err1)
	a.NoError(err2)
	a.NotNil(msg)
	a.Equal(payload, *msg.Body)
}

func Test_Delete(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	// arrange
	url := arrangeQueue(t)
	cli := arrangeClient()
	ctx := context.Background()

	// act
	_, err1 := sqsx.SendMessage(ctx, cli, url, payload)
	msg, err2 := sqsx.ReceiveMessage(ctx, cli, url, sqsx.WithVisibilityTimeout(2), sqsx.WithWaitTime(2))
	err3 := sqsx.DeleteMessage(ctx, msg.ReceiptHandle, cli, url)

	// assert
	a.NoError(err1)
	a.NoError(err2)
	a.NoError(err3)
	a.NotNil(msg)
	a.Equal(payload, *msg.Body)

}

func Test_ChangeMsgVisibilityTimeout(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	// arrange
	url := arrangeQueue(t)
	cli := arrangeClient()
	ctx := context.Background()

	// act
	_, err1 := sqsx.SendMessage(ctx, cli, url, payload)
	msg1, err2 := sqsx.ReceiveMessage(ctx, cli, url, sqsx.WithVisibilityTimeout(300), sqsx.WithWaitTime(2))
	_, err3 := sqsx.ChangeMsgVisibilityTimeout(ctx, cli, url, msg1.ReceiptHandle, 0)
	msg2, err4 := sqsx.ReceiveMessage(ctx, cli, url, sqsx.WithVisibilityTimeout(300), sqsx.WithWaitTime(2))

	// assert
	a.NoError(err1)
	a.NoError(err2)
	a.NoError(err3)
	a.NoError(err4)
	a.NotNil(msg1)
	a.NotNil(msg2)
	a.Equal(payload, *msg1.Body)
	a.Equal(payload, *msg2.Body)
}

func Test_Send_And_Receive_From_FIFO(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	url := arrangeFIFOQueue(t)
	cli := arrangeClient()
	ctx := context.Background()

	payloadA := "payloadA"
	payloadB := "payloadB"
	group := "group"
	id1 := "1"
	id2 := "2"

	_, err := sqsx.SendFIFOMessage(ctx, cli, url, payloadA, group, id1)
	a.NoError(err)

	_, err = sqsx.SendFIFOMessage(ctx, cli, url, payloadB, group, id2)
	a.NoError(err)

	// Publish with the same deduplicationID -> message is not stored in the queue
	_, err = sqsx.SendFIFOMessage(ctx, cli, url, payloadB, group, id2)
	a.NoError(err)

	msg, err := sqsx.ReceiveMessage(ctx, cli, url, sqsx.WithVisibilityTimeout(2), sqsx.WithWaitTime(2))
	a.NoError(err)
	a.NotNil(msg)

	a.Equal(payloadA, *msg.Body)
	err = sqsx.DeleteMessage(ctx, msg.ReceiptHandle, cli, url)
	a.NoError(err)

	msg, err = sqsx.ReceiveMessage(ctx, cli, url, sqsx.WithVisibilityTimeout(2), sqsx.WithWaitTime(2))
	a.NoError(err)
	a.NotNil(msg)

	a.Equal(payloadB, *msg.Body)
	err = sqsx.DeleteMessage(ctx, msg.ReceiptHandle, cli, url)
	a.NoError(err)

	// Duplicated message is not found
	msg, err = sqsx.ReceiveMessage(ctx, cli, url, sqsx.WithVisibilityTimeout(2), sqsx.WithWaitTime(2))
	a.NoError(err)
	a.Nil(msg)
}

func arrangeQueue(t *testing.T) string {
	queue := awstest.CreateResource(sqs.ServiceName)
	awstest.AssertResourceExists(t, queue, sqs.ServiceName)

	sess := awstest.NewSession()
	cli := sqs.New(sess)

	url, err := sqsx.GetQueueURL(context.Background(), cli, queue)
	require.NoError(t, err)

	return url
}

func arrangeFIFOQueue(t *testing.T) string {
	queue := awstest.CreateResource(awstest.SQSFifoServiceName)
	awstest.AssertResourceExists(t, queue, awstest.SQSFifoServiceName)

	sess := awstest.NewSession()
	cli := sqs.New(sess)

	url, err := sqsx.GetQueueURL(context.Background(), cli, queue)
	require.NoError(t, err)

	return url
}

func arrangeClient() *sqs.SQS {
	sess := awstest.NewSession()
	cli := sqs.New(sess)

	return cli
}
