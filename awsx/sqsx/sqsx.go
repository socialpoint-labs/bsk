package sqsx

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
)

// ReceiveMessage get a message from the queue
func ReceiveMessage(ctx context.Context, cli sqsiface.SQSAPI, url string, visibilityTimeout, waitTime time.Duration) (*sqs.Message, error) {
	input := &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(url),
		VisibilityTimeout:   aws.Int64(int64(visibilityTimeout)),
		WaitTimeSeconds:     aws.Int64(int64(waitTime)),
		MaxNumberOfMessages: aws.Int64(1),
	}

	output, err := cli.ReceiveMessageWithContext(ctx, input)

	if err != nil {
		return nil, err
	}

	if len(output.Messages) == 0 {
		return nil, nil
	}

	// if there are more than one message, then don't consume the message,
	// we are interested only in one message
	if len(output.Messages) > 1 {
		return nil, errors.New("too many messages received")
	}

	return output.Messages[0], nil
}

// SendMessage delivers a message to the specified queue
func SendMessage(ctx context.Context, cli sqsiface.SQSAPI, url string, body string) (*sqs.SendMessageOutput, error) {
	input := &sqs.SendMessageInput{
		QueueUrl:    aws.String(url),
		MessageBody: aws.String(body),
	}

	return cli.SendMessageWithContext(ctx, input)
}

// SendFIFOMessage delivers a message to the specified FIFO queue
func SendFIFOMessage(ctx context.Context, cli sqsiface.SQSAPI, url string, body string, group string, deduplicationID string) (*sqs.SendMessageOutput, error) {
	input := &sqs.SendMessageInput{
		QueueUrl:               aws.String(url),
		MessageBody:            aws.String(body),
		MessageGroupId:         aws.String(group),
		MessageDeduplicationId: aws.String(deduplicationID),
	}

	return cli.SendMessageWithContext(ctx, input)
}

// ChangeMsgVisibilityTimeout change visibility timeout of a message in the specified queue
func ChangeMsgVisibilityTimeout(ctx context.Context, cli sqsiface.SQSAPI, url string, receiptHandle *string, visibilityTimeout int64) (*sqs.ChangeMessageVisibilityOutput, error) {
	input := &sqs.ChangeMessageVisibilityInput{
		QueueUrl:          aws.String(url),
		ReceiptHandle:     receiptHandle,
		VisibilityTimeout: aws.Int64(visibilityTimeout),
	}

	return cli.ChangeMessageVisibilityWithContext(ctx, input)
}

// GetQueueURL returns the URL of an existing queue.
// This action provides a simple way to retrieve the URL of an Amazon SQS queue
func GetQueueURL(ctx context.Context, cli sqsiface.SQSAPI, queue *string) (string, error) {
	o, err := cli.GetQueueUrlWithContext(ctx, &sqs.GetQueueUrlInput{QueueName: queue})

	if err != nil {
		return "", err
	}

	return aws.StringValue(o.QueueUrl), nil
}

// DeleteMessage deletes a message from SQS queue
func DeleteMessage(ctx context.Context, receiptHandle *string, cli sqsiface.SQSAPI, url string) error {
	input := &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(url),
		ReceiptHandle: receiptHandle,
	}

	_, err := cli.DeleteMessageWithContext(ctx, input)

	return err
}
