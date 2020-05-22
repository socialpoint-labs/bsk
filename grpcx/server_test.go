package grpcx_test

import (
	"bytes"
	"context"
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
	info := &grpc.UnaryServerInfo{FullMethod: "method"}
	handler := grpc.UnaryHandler(func(ctx context.Context, req interface{}) (interface{}, error) {
		return expected, nil
	})

	interceptor := grpcx.WithMetrics(m)
	resp, err := interceptor(ctx, req, info, handler)

	a.NoError(err)
	a.Equal(expected, resp)

	timer := m.Timer("grpc_request")
	a.NotNil(timer)
}

func TestWithRequestResponseLogs(t *testing.T) {
	a := assert.New(t)
	t.Parallel()

	w := bytes.NewBufferString("")
	l := logx.New(logx.WriterOpt(w))

	ctx := context.Background()
	req := "my-request"
	expected := "my-response"
	info := &grpc.UnaryServerInfo{FullMethod: "method"}
	handler := grpc.UnaryHandler(func(ctx context.Context, req interface{}) (interface{}, error) {
		return expected, nil
	})

	interceptor := grpcx.WithRequestResponseLogs(l)
	resp, err := interceptor(ctx, req, info, handler)

	a.NoError(err)
	a.Equal(expected, resp)
	a.Contains(w.String(), "request=my-request")
	a.Contains(w.String(), "response=my-response")
}
