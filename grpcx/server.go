package grpcx

import (
	"context"
	"encoding/json"

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
		timer.
			WithTag("rpc_method", info.FullMethod).
			WithTag("success", err == nil).
			Stop()

		return resp, err
	}
}

// WithRequestResponseLogs returns a gRPC interceptor for unary calls that instrument
// requests with logs for the request and response.
func WithRequestResponseLogs(l logx.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		resp, err := handler(ctx, req)

		reqMsg, _ := json.Marshal(req)
		respMsg, _ := json.Marshal(resp)

		l.Info("gRPC Message",
			logx.Field{Key: "ctx_request_content", Value: string(reqMsg)},
			logx.Field{Key: "ctx_response_content", Value: string(respMsg)},
		)

		return resp, err
	}
}
