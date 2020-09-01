package grpcx_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/socialpoint-labs/bsk/grpcx"
	"github.com/socialpoint-labs/bsk/logx"
	"github.com/socialpoint-labs/bsk/metrics"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestWithMetrics(t *testing.T) {
	a := assert.New(t)
	t.Parallel()

	m := metrics.NewRecorder()
	ctx := context.Background()
	req := "my-request"
	expected := "my-response"
	method := "method"
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
	a := assert.New(t)
	t.Parallel()

	w := bytes.NewBufferString("")
	l := logx.New(logx.WriterOpt(w))

	ctx := context.Background()
	userID := "user-id"
	req := struct {
		UserID string
	}{
		UserID: userID,
	}

	result := "ok"
	expected := struct {
		Result string
	}{
		Result: result,
	}

	method := "method"
	info := &grpc.UnaryServerInfo{FullMethod: method}
	handler := grpc.UnaryHandler(func(ctx context.Context, req interface{}) (interface{}, error) {
		return expected, nil
	})

	interceptor := grpcx.WithRequestResponseLogs(l)
	resp, err := interceptor(ctx, req, info, handler)

	a.NoError(err)
	a.Equal(expected, resp)
	a.Contains(w.String(), fmt.Sprintf(`INFO gRPC Message FIELDS ctx_full_method=%s ctx_request_content={"UserID":"%s"} ctx_response_content={"Result":"%s"}`, method, userID, result))
}