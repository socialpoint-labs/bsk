package grpcx_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/socialpoint-labs/bsk/grpcx"
	"github.com/socialpoint-labs/bsk/logx"
	"github.com/socialpoint-labs/bsk/metrics"
	"github.com/socialpoint-labs/bsk/recovery"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	method   = "method"
	userID   = "user-id"
	okResult = "ok"
)

var exampleRequest = struct {
	UserID string
}{
	UserID: userID,
}
var exampleResponse = struct {
	Result string
}{
	Result: okResult,
}

func TestWithMetrics(t *testing.T) {
	a := assert.New(t)
	t.Parallel()

	m := metrics.NewRecorder()
	ctx := context.Background()
	req := "my-request"
	expected := "my-response"
	info := &grpc.UnaryServerInfo{FullMethod: method}
	handler := grpc.UnaryHandler(func(ctx context.Context, req interface{}) (interface{}, error) {
		return expected, nil
	})

	interceptor := grpcx.WithMetrics(m)
	resp, err := interceptor(ctx, req, info, handler)

	a.NoError(err)
	a.Equal(expected, resp)

	timer := m.Get("grpc.request_duration")
	a.NotNil(timer)
	a.Contains(timer.Tags(), metrics.NewTag("rpc_method", method))
	a.Contains(timer.Tags(), metrics.NewTag("success", true))
}

func TestWithRequestResponseLogs(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	t.Run("logs request and response", func(t *testing.T) {
		w := bytes.NewBufferString("")
		l := logx.New(logx.WriterOpt(w))

		ctx := context.Background()
		req := exampleRequest
		expectedResponse := exampleResponse
		info := &grpc.UnaryServerInfo{FullMethod: method}
		handler := grpc.UnaryHandler(func(ctx context.Context, req interface{}) (interface{}, error) {
			return expectedResponse, nil
		})

		interceptor := grpcx.WithRequestResponseLogs(l)
		resp, err := interceptor(ctx, req, info, handler)

		a.NoError(err)
		a.Equal(expectedResponse, resp)
		a.Contains(w.String(), fmt.Sprintf(`INFO gRPC Message FIELDS ctx_full_method=%s ctx_request_content={"UserID":"%s"} ctx_response_content={"Result":"%s"}`, method, userID, okResult))
	})

	t.Run("logs the response error", func(t *testing.T) {
		w := bytes.NewBufferString("")
		l := logx.New(logx.WriterOpt(w))

		ctx := context.Background()
		req := exampleRequest
		expectedResponse := exampleResponse
		expectedErr := fmt.Errorf("some error")

		info := &grpc.UnaryServerInfo{FullMethod: method}
		handler := grpc.UnaryHandler(func(ctx context.Context, req interface{}) (interface{}, error) {
			return expectedResponse, expectedErr
		})

		interceptor := grpcx.WithRequestResponseLogs(l)
		resp, err := interceptor(ctx, req, info, handler)

		a.Error(err)
		a.Equal(expectedResponse, resp)
		a.Contains(w.String(), fmt.Sprintf(`INFO gRPC Message FIELDS ctx_full_method=%s ctx_request_content={"UserID":"%s"} ctx_response_content={"Result":"%s"} ctx_response_error=%s`, method, userID, okResult, expectedErr.Error()))
	})
}

func TestWithStructuredPanicLogs(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	t.Run("panic handler works as expected", func (t *testing.T) {
		reached := false
		spyExitFunc := func() { reached = true }
		interceptor := grpcx.WithStructuredPanicLogs(logx.NewDummy(), recovery.WithExitFunction(spyExitFunc))

		ctx := context.Background()
		info := &grpc.UnaryServerInfo{FullMethod: method}
		panickingHandler := grpc.UnaryHandler(func(ctx context.Context, req interface{}) (interface{}, error) {
			panic("test panicking")
		})

		_, err := interceptor(ctx, exampleRequest, info, panickingHandler)

		a.NoError(err)
		a.True(reached)
	})
}

func TestWithErrorLogs(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	t.Run("logs nothing if no error", func(t *testing.T) {
		w := bytes.NewBufferString("")
		l := logx.New(logx.WriterOpt(w))

		ctx := context.Background()
		req := exampleRequest
		expectedResponse := exampleResponse

		info := &grpc.UnaryServerInfo{FullMethod: method}
		handler := grpc.UnaryHandler(func(ctx context.Context, req interface{}) (interface{}, error) {
			return expectedResponse, nil
		})

		interceptor := grpcx.WithErrorLogs(l)
		resp, err := interceptor(ctx, req, info, handler)

		a.NoError(err)
		a.Equal(expectedResponse, resp)
		a.Equal("", w.String())
	})

	t.Run("logs error on info level with default options", func(t *testing.T) {
		w := bytes.NewBufferString("")
		l := logx.New(logx.WriterOpt(w))

		ctx := context.Background()
		req := exampleRequest
		expectedResponse := exampleResponse
		expectedCode := codes.Unknown
		expectedErr := status.Error(expectedCode, "some error")

		info := &grpc.UnaryServerInfo{FullMethod: method}
		handler := grpc.UnaryHandler(func(ctx context.Context, req interface{}) (interface{}, error) {
			return expectedResponse, expectedErr
		})

		interceptor := grpcx.WithErrorLogs(l)
		resp, err := interceptor(ctx, req, info, handler)

		a.Error(err)
		a.Equal(expectedResponse, resp)
		a.Contains(w.String(), fmt.Sprintf(`INFO gRPC Error FIELDS ctx_full_method=%s ctx_request_content={"UserID":"%s"} ctx_response_content={"Result":"%s"} ctx_response_error_code=%s ctx_response_error_message=%s`, method, userID, okResult, expectedCode, expectedErr.Error()))
	})

	t.Run("logs error on debug level if custom options added", func(t *testing.T) {
		w := bytes.NewBufferString("")
		l := logx.New(logx.WriterOpt(w))

		ctx := context.Background()
		req := exampleRequest
		expectedResponse := exampleResponse
		expectedCode := codes.NotFound
		expectedErr := status.Error(expectedCode, "some error")

		info := &grpc.UnaryServerInfo{FullMethod: method}
		handler := grpc.UnaryHandler(func(ctx context.Context, req interface{}) (interface{}, error) {
			return expectedResponse, expectedErr
		})

		interceptor := grpcx.WithErrorLogs(l, grpcx.WithDebugLevelCodes(expectedCode))
		resp, err := interceptor(ctx, req, info, handler)

		a.Error(err)
		a.Equal(expectedResponse, resp)
		a.Contains(w.String(), fmt.Sprintf(`DEBU gRPC Error FIELDS ctx_full_method=%s ctx_request_content={"UserID":"%s"} ctx_response_content={"Result":"%s"} ctx_response_error_code=%s ctx_response_error_message=%s`, method, userID, okResult, expectedCode, expectedErr.Error()))
	})

	t.Run("do not log error if discarded options added", func(t *testing.T) {
		w := bytes.NewBufferString("")
		l := logx.New(logx.WriterOpt(w))

		ctx := context.Background()
		req := exampleRequest
		expectedResponse := exampleResponse
		expectedCode := codes.NotFound
		expectedErr := status.Error(expectedCode, "some error")

		info := &grpc.UnaryServerInfo{FullMethod: method}
		handler := grpc.UnaryHandler(func(ctx context.Context, req interface{}) (interface{}, error) {
			return expectedResponse, expectedErr
		})

		interceptor := grpcx.WithErrorLogs(l, grpcx.WithDiscardedCodes(expectedCode, codes.InvalidArgument, codes.AlreadyExists))
		resp, err := interceptor(ctx, req, info, handler)

		a.Error(err)
		a.Equal(expectedResponse, resp)
		a.Equal("", w.String())
	})
}
