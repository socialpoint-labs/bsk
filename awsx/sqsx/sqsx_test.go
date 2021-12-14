package sqsx_test

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/socialpoint-labs/bsk/awsx/sqsx"
	"github.com/stretchr/testify/assert"
)

func TestReceiveMessage_WithOptions(t *testing.T) {
	const url = "the-url"

	t.Parallel()
	a := assert.New(t)

	// arrange
	spy := &spyClient{}

	// act
	msg, err := sqsx.ReceiveMessage(
		context.Background(),
		spy,
		url,
		sqsx.WithWaitTime(20*time.Second),
		sqsx.WithVisibilityTimeout(time.Minute),
	)

	// assert
	a.Nil(msg)
	a.NoError(err)
	expectedInput := &sqs.ReceiveMessageInput{
		AttributeNames:          nil,
		MaxNumberOfMessages:     aws.Int64(1),
		MessageAttributeNames:   nil,
		QueueUrl:                aws.String(url),
		ReceiveRequestAttemptId: nil,
		VisibilityTimeout:       aws.Int64(60),
		WaitTimeSeconds:         aws.Int64(20),
	}
	a.Equal(expectedInput, spy.input)
}

type spyClient struct {
	sqsiface.SQSAPI

	input *sqs.ReceiveMessageInput
}

func (s *spyClient) ReceiveMessageWithContext(_ aws.Context, input *sqs.ReceiveMessageInput, _ ...request.Option) (*sqs.ReceiveMessageOutput, error) {
	s.input = input

	return &sqs.ReceiveMessageOutput{}, nil
}
