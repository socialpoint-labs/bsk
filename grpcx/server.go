package grpcx

import (
	"context"

	"github.com/socialpoint-labs/bsk/logx"
	"github.com/socialpoint-labs/bsk/metrics"
	"google.golang.org/grpc"
)

// WithMetrics returns a gRPC interceptor for unary calls that instrument requests
// with a metric for the request duration.
func WithMetrics(m metrics.Metrics) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		timer := m.Timer("grpc.request_duration")
		timer.Start()

		resp, err := handler(ctx, req)
		timer.WithTag("success", err == nil).Stop()

		timer.Stop()

		return resp, err
	}
}

// WithRequestResponseLogs returns a gRPC interceptor for unary calls that instrument
// requests with logs for the request and response.
func WithRequestResponseLogs(l logx.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		resp, err := handler(ctx, req)

		l.Info("grpc request", logx.Field{Key: "request", Value: req})
		l.Info("grpc response", logx.Field{Key: "response", Value: resp})

		return resp, err
	}
}
