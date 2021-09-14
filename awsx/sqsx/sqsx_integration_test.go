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
)

const payload = "random-payload:life-is-to-short-to-generate-a-really-random-payload-when-it-should-not-be-random-at-all"

func Test_Send_And_Receive(t *testing.T) {
	a := assert.New(t)
	url := setupQueue(t, a)
	sess := awstest.NewSession()
	cli := sqs.New(sess)

	_, err := sqsx.SendMessage(context.Background(), cli, url, payload)
	a.NoError(err)

	msg, err := sqsx.ReceiveMessage(context.Background(), cli, url, 2, 2)
	a.NoError(err)

	a.NotNil(msg)
	a.Equal(payload, *msg.Body)
}

func Test_Delete(t *testing.T) {
	a := assert.New(t)
	url := setupQueue(t, a)
	sess := awstest.NewSession()
	cli := sqs.New(sess)

	_, err := sqsx.SendMessage(context.Background(), cli, url, payload)
	a.NoError(err)

	msg, err := sqsx.ReceiveMessage(context.Background(), cli, url, 2, 2)
	a.NoError(err)

	a.NotNil(msg)
	a.Equal(payload, *msg.Body)

	err = sqsx.DeleteMessage(context.Background(), msg.ReceiptHandle, cli, url)
	a.NoError(err)
}

func Test_ChangeMsgVisibilityTimeout(t *testing.T) {
	a := assert.New(t)
	url := setupQueue(t, a)
	sess := awstest.NewSession()
	cli := sqs.New(sess)

	_, err := sqsx.SendMessage(context.Background(), cli, url, payload)
	a.NoError(err)

	msg, err := sqsx.ReceiveMessage(context.Background(), cli, url, 300, 2)
	a.NoError(err)
	a.NotNil(msg)
	a.Equal(payload, *msg.Body)

	_, err = sqsx.ChangeMsgVisibilityTimeout(context.Background(), cli, url, msg.ReceiptHandle, 0)
	a.NoError(err)
}

func setupQueue(t *testing.T, assert *assert.Assertions) (url string) {
	queue := awstest.CreateResource(t, sqs.ServiceName)
	awstest.AssertResourceExists(t, queue, sqs.ServiceName)

	sess := awstest.NewSession()
	cli := sqs.New(sess)

	url, err := sqsx.GetQueueURL(context.Background(), cli, queue)
	assert.NoError(err)

	return
}

func Test_Send_And_Receive_From_FIFO(t *testing.T) {
	a := assert.New(t)
	url := setupFIFOQueue(t, a)
	sess := awstest.NewSession()
	cli := sqs.New(sess)

	payloadA := "payloadA"
	payloadB := "payloadB"
	group := "group"
	id1 := "1"
	id2 := "2"

	_, err := sqsx.SendFIFOMessage(context.Background(), cli, url, payloadA, group, id1)
	a.NoError(err)

	_, err = sqsx.SendFIFOMessage(context.Background(), cli, url, payloadB, group, id2)
	a.NoError(err)

	// Publish with the same deduplicationID -> message is not stored in the queue
	_, err = sqsx.SendFIFOMessage(context.Background(), cli, url, payloadB, group, id2)
	a.NoError(err)

	msg, err := sqsx.ReceiveMessage(context.Background(), cli, url, 2, 2)
	a.NoError(err)
	a.NotNil(msg)

	a.Equal(payloadA, *msg.Body)
	err = sqsx.DeleteMessage(context.Background(), msg.ReceiptHandle, cli, url)
	a.NoError(err)

	msg, err = sqsx.ReceiveMessage(context.Background(), cli, url, 2, 2)
	a.NoError(err)
	a.NotNil(msg)

	a.Equal(payloadB, *msg.Body)
	err = sqsx.DeleteMessage(context.Background(), msg.ReceiptHandle, cli, url)
	a.NoError(err)

	// Duplicated message is not found
	msg, err = sqsx.ReceiveMessage(context.Background(), cli, url, 2, 2)
	a.NoError(err)
	a.Nil(msg)
}

func setupFIFOQueue(t *testing.T, assert *assert.Assertions) string {
	queue := awstest.CreateResource(t, awstest.SQSFifoServiceName)
	awstest.AssertResourceExists(t, queue, awstest.SQSFifoServiceName)

	sess := awstest.NewSession()
	cli := sqs.New(sess)

	url, err := sqsx.GetQueueURL(context.Background(), cli, queue)
	assert.NoError(err)

	return url
}
