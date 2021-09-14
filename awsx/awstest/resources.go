package awstest

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/socialpoint-labs/bsk/awsx"
	"github.com/socialpoint-labs/bsk/logx"
	"github.com/socialpoint-labs/bsk/uuid"
)

const (
	fifoSuffix = ".fifo"
	// SQSFifoServiceName is the service name for SQS FIFO queues
	SQSFifoServiceName = sqs.ServiceName + fifoSuffix
)

// Supported resource types
type s3ResourceService struct{}
type sqsResourceService struct{}
type sqsFifoResourceService struct{}
type kmsResourceService struct{}
type dynamodbResourceService struct{}
type kinesisResourceService struct{}

// ResourceService isolates creation and checking existence by session for each supported resource
type ResourceService interface {
	// CreateResourceForSession creates a resource for testing purpouses
	CreateResourceForSession(t *testing.T, sess *session.Session, name *string) (*string, error)
	// AssertResourceExistsForSession check if a resource exists for a given session
	AssertResourceExistsForSession(t *testing.T, sess *session.Session, name *string) (bool, error)
}

// Map for getting ResourceService by name
var serviceNameToResources = map[string]ResourceService{
	s3.ServiceName:       s3ResourceService{},
	sqs.ServiceName:      sqsResourceService{},
	SQSFifoServiceName:   sqsFifoResourceService{},
	kms.ServiceName:      kmsResourceService{},
	dynamodb.ServiceName: dynamodbResourceService{},
	kinesis.ServiceName:  kinesisResourceService{},
}

// GetResourceServiceByName returns a ResourceService if it's supported
func GetResourceServiceByName(r string) ResourceService {
	resource, ok := serviceNameToResources[r]
	if !ok {
		log.Panicf("Resource %q not supported", r)
	}

	return resource
}

// NewSession creates a new AWS session, suited for testing
func NewSession() *session.Session {
	return NewSessionWithRegion("us-east-1")
}

// NewSessionWithRegion creates a new AWS session configured for specific region, suited for testing
func NewSessionWithRegion(region string) *session.Session {
	cfg := aws.NewConfig().
		WithRegion(region).
		WithCredentialsChainVerboseErrors(true).
		WithS3ForcePathStyle(true)

	awsEndpoint := Endpoint()
	if awsEndpoint != "" {
		cfg.WithEndpoint(awsEndpoint)
	}

	awsSecretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	awsAccessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	if awsSecretAccessKey != "" || awsAccessKeyID != "" {
		cfg.WithCredentials(credentials.NewStaticCredentials(awsAccessKeyID, awsSecretAccessKey, ""))
	}

	return session.Must(session.NewSession(cfg))
}

func Endpoint() string {
	return os.Getenv("SP_BSK_AWS_ENDPOINT")
}

// CreateResource creates an AWS resource for testing purposes
func CreateResource(t *testing.T, r string) *string {
	name := aws.String("integration-test-" + uuid.New())

	resource := GetResourceServiceByName(r)
	sess := NewSession()
	name, err := resource.CreateResourceForSession(t, sess, name)
	if err != nil {
		log.Panicf("Error %v creating resource %q", err, r)
	}

	return name
}

// AssertResourceExists asserts if the resource with the given name exists
func AssertResourceExists(t *testing.T, name *string, r string) bool {
	resource := GetResourceServiceByName(r)
	sess := NewSession()
	response, err := resource.AssertResourceExistsForSession(t, sess, name)
	if err != nil {
		t.Errorf("Error %v checking resource %v", err, r)
	}

	return response
}

func (s3ResourceService) CreateResourceForSession(t *testing.T, sess *session.Session, name *string) (*string, error) {
	svc := s3.New(sess)

	input := &s3.CreateBucketInput{Bucket: name}
	_, err := svc.CreateBucket(input)

	return name, err
}

func (s3ResourceService) AssertResourceExistsForSession(t *testing.T, sess *session.Session, name *string) (bool, error) {
	svc := s3.New(sess)

	input := &s3.HeadBucketInput{Bucket: name}
	_, err := svc.HeadBucket(input)

	return err == nil, err
}

func (sqsResourceService) CreateResourceForSession(t *testing.T, sess *session.Session, name *string) (*string, error) {
	svc := sqs.New(sess)

	input := &sqs.CreateQueueInput{QueueName: name}
	_, err := svc.CreateQueue(input)

	return name, err
}

func (sqsResourceService) AssertResourceExistsForSession(t *testing.T, sess *session.Session, name *string) (bool, error) {
	svc := sqs.New(sess)

	input := &sqs.GetQueueUrlInput{QueueName: name}
	_, err := svc.GetQueueUrl(input)

	return err == nil, err
}

func (sqsFifoResourceService) CreateResourceForSession(t *testing.T, sess *session.Session, name *string) (*string, error) {
	svc := sqs.New(sess)

	// Add FIFO suffix
	name = aws.String(fmt.Sprintf("%s%s", *name, fifoSuffix))

	input := &sqs.CreateQueueInput{
		QueueName: name,
		Attributes: map[string]*string{
			"FifoQueue":                 aws.String("true"),
			"ContentBasedDeduplication": aws.String("true"),
		},
	}
	_, err := svc.CreateQueue(input)

	return name, err
}

func (sqsFifoResourceService) AssertResourceExistsForSession(t *testing.T, sess *session.Session, name *string) (bool, error) {
	svc := sqs.New(sess)

	input := &sqs.GetQueueUrlInput{QueueName: name}
	_, err := svc.GetQueueUrl(input)

	return err == nil, err
}

func (kmsResourceService) CreateResourceForSession(t *testing.T, sess *session.Session, name *string) (*string, error) {
	svc := kms.New(sess)

	input := &kms.CreateKeyInput{
		Description: name,
		KeyUsage:    aws.String("ENCRYPT_DECRYPT"),
	}
	res, err := svc.CreateKey(input)

	if err == nil {
		_, err = svc.CreateAlias(&kms.CreateAliasInput{
			TargetKeyId: res.KeyMetadata.KeyId,
			AliasName:   aws.String("alias/" + aws.StringValue(name)),
		})

		name = res.KeyMetadata.KeyId
	}
	return name, err
}

func (kmsResourceService) AssertResourceExistsForSession(t *testing.T, sess *session.Session, name *string) (bool, error) {
	svc := kms.New(NewSession())

	input := &kms.DescribeKeyInput{KeyId: name}
	_, err := svc.DescribeKey(input)

	return err == nil, err
}

func (dynamodbResourceService) CreateResourceForSession(t *testing.T, sess *session.Session, name *string) (*string, error) {
	svc := dynamodb.New(sess)

	input := &dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("KeySchemaAttributeName"),
				AttributeType: aws.String(dynamodb.ScalarAttributeTypeS),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("KeySchemaAttributeName"),
				KeyType:       aws.String(dynamodb.KeyTypeHash),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(1),
			WriteCapacityUnits: aws.Int64(1),
		},

		TableName: name,
	}

	_, err := svc.CreateTable(input)

	return name, err
}

func (dynamodbResourceService) AssertResourceExistsForSession(t *testing.T, sess *session.Session, name *string) (bool, error) {
	svc := dynamodb.New(sess)

	input := &dynamodb.DescribeTableInput{TableName: name}
	_, err := svc.DescribeTable(input)

	return err == nil, err
}

// CreateResourceForSession always reuses the same stream because creating one is very slow
func (kinesisResourceService) CreateResourceForSession(t *testing.T, sess *session.Session, name *string) (*string, error) {
	kin := kinesis.New(sess)

	// fixed name because aws takes up to 17s to be able to provide the shard info about a new stream
	createStreamInput := kinesis.CreateStreamInput{
		StreamName: name,
		ShardCount: aws.Int64(1)}
	_, err := kin.CreateStream(&createStreamInput)
	if awsx.IsErrorCode(err, kinesis.ErrCodeResourceInUseException) {
		err = nil // already exists
	} else {
		logx.New().Info(fmt.Sprintf(
			"Stream %s does not exist yet. It will take around 15s to create it ", aws.StringValue(name)))

		if err := kin.WaitUntilStreamExists(&kinesis.DescribeStreamInput{StreamName: name}); err != nil {
			panic(err)
		}
	}

	return name, err
}

func (kinesisResourceService) AssertResourceExistsForSession(t *testing.T, sess *session.Session, name *string) (bool, error) {
	kin := kinesis.New(sess)
	input := &kinesis.DescribeStreamInput{StreamName: name}
	err := kin.WaitUntilStreamExists(input)

	return err == nil, err
}
