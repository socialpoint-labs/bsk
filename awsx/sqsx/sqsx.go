package sqsx

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/sqs"
)

// ReceiveMessage get a message from the queue
func ReceiveMessage(ctx context.Context, p client.ConfigProvider, url string, visibilityTimeout, waitTime int64) (*sqs.Message, error) {
	svc := sqs.New(p)

	input := &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(url),
		VisibilityTimeout:   aws.Int64(visibilityTimeout),
		WaitTimeSeconds:     aws.Int64(waitTime),
		MaxNumberOfMessages: aws.Int64(1),
	}

	req, output := svc.ReceiveMessageRequest(input)

	req.HTTPRequest = req.HTTPRequest.WithContext(ctx)
	err := req.Send()

	if err != nil {
		return nil, err
	}

	if len(output.Messages) == 0 {
		return nil, nil
	}

	// if there are more than one message, then don't consume the message,
	// we are  interested only in one message
	if len(output.Messages) > 1 {
		return nil, errors.New("too many messages received")
	}

	return output.Messages[0], nil
}

// ReceiveMessageWithRetries get a message from the queue with an exponential backoff for retries.
func ReceiveMessageWithRetries(ctx context.Context, p client.ConfigProvider, url string, visibilityTimeout, waitTime int64, maxRetries int) (*sqs.Message, error) {
	retries := 0

	msg, err := ReceiveMessage(ctx, p, url, visibilityTimeout, waitTime)
	for err != nil && retries < maxRetries {
		sleepTime := time.Duration(math.Pow(2, float64(retries))*100) * time.Millisecond
		time.Sleep(sleepTime)

		retries++
		msg, err = ReceiveMessage(ctx, p, url, visibilityTimeout, waitTime)
	}

	return msg, err
}

// SendMessage delivers a message to the specified queue
func SendMessage(ctx context.Context, p client.ConfigProvider, url string, body string) (*sqs.SendMessageOutput, error) {
	svc := sqs.New(p)

	input := &sqs.SendMessageInput{
		QueueUrl:    aws.String(url),
		MessageBody: aws.String(body),
	}

	return svc.SendMessage(input)
}

// SendFIFOMessage delivers a message to the specified FIFO queue
func SendFIFOMessage(ctx context.Context, p client.ConfigProvider, url string, body string, group string, deduplicationID string) (*sqs.SendMessageOutput, error) {
	svc := sqs.New(p)

	input := &sqs.SendMessageInput{
		QueueUrl:               aws.String(url),
		MessageBody:            aws.String(body),
		MessageGroupId:         aws.String(group),
		MessageDeduplicationId: aws.String(deduplicationID),
	}

	return svc.SendMessage(input)
}

// ChangeMsgVisibilityTimeout change visibility timeout of a message in the specified queue
func ChangeMsgVisibilityTimeout(ctx context.Context, p client.ConfigProvider, url string, receiptHandle *string, visibilityTimeout int64) (*sqs.ChangeMessageVisibilityOutput, error) {
	svc := sqs.New(p)

	input := &sqs.ChangeMessageVisibilityInput{
		QueueUrl:          aws.String(url),
		ReceiptHandle:     receiptHandle,
		VisibilityTimeout: aws.Int64(visibilityTimeout),
	}

	return svc.ChangeMessageVisibility(input)
}

// GetQueueURL returns the URL of an existing queue.
// This action provides a simple way to retrieve the URL of an Amazon SQS queue
func GetQueueURL(ctx context.Context, p client.ConfigProvider, queue *string) (string, error) {
	svc := sqs.New(p)

	o, err := svc.GetQueueUrl(&sqs.GetQueueUrlInput{QueueName: queue})

	if err != nil {
		return "", err
	}

	return aws.StringValue(o.QueueUrl), nil
}

// DeleteMessage deletes a message from SQS queue
func DeleteMessage(receiptHandle *string, p client.ConfigProvider, url string) error {
	svc := sqs.New(p)

	input := &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(url),
		ReceiptHandle: receiptHandle,
	}

	_, err := svc.DeleteMessage(input)

	return err
}
