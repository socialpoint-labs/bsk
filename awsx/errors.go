package awsx

import "github.com/aws/aws-sdk-go/aws/awserr"

const (
	// Errors from http://docs.aws.amazon.com/kinesis/latest/APIReference/CommonErrors.html

	// ErrCodeInternalFailure means failure because of an unknown error, exception or failure
	ErrCodeInternalFailure = "InternalFailure"

	// ErrCodeServiceUnavailable means failure due to a temporary failure of the server
	ErrCodeServiceUnavailable = "ServiceUnavailable"
)

// IsErrorCode returns whether an error is an aws error with the specified code
func IsErrorCode(err error, expectedCode string) bool {
	awsErr, ok := err.(awserr.Error)
	if ok && awsErr.Code() == expectedCode {
		return true
	}
	return false
}
