package awstest_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/stretchr/testify/assert"

	"github.com/socialpoint-labs/bsk/awsx/awstest"
	"github.com/socialpoint-labs/bsk/uuid"
)

var resourceTypes = []string{
	s3.ServiceName,
	sqs.ServiceName,
	kms.ServiceName,
	dynamodb.ServiceName,
	awstest.SQSFifoServiceName,
	// kinesis.ServiceName,
}

func TestCreateResource(t *testing.T) {
	for _, res := range resourceTypes {
		awstest.AssertResourceExists(t, awstest.CreateResource(res), res)
	}
}

func TestKMSAliasCreatedForResource(t *testing.T) {
	a := assert.New(t)

	keyID := awstest.CreateResource(kms.ServiceName)

	svc := kms.New(awstest.NewSession())
	res, err := svc.ListAliases(&kms.ListAliasesInput{Limit: aws.Int64(100)})

	a.NoError(err)
	exists := false
	for _, a := range res.Aliases {
		if a.TargetKeyId != nil && *a.TargetKeyId == *keyID {
			exists = true
			break
		}
	}
	a.True(exists)
}

func TestAssertResourceExists(t *testing.T) {
	mt := new(testing.T)

	for _, res := range resourceTypes {
		exists := awstest.AssertResourceExists(mt, aws.String(uuid.New()), res)
		assert.False(t, exists, res)
	}
}
