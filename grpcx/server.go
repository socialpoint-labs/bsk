package grpcx

import (
	"context"
	"encoding/json"

	"github.com/socialpoint-labs/bsk/logx"
	"github.com/socialpoint-labs/bsk/metrics"
	"github.com/socialpoint-labs/bsk/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

		fields := []logx.Field{
			{Key: "ctx_full_method", Value: info.FullMethod},
			{Key: "ctx_request_content", Value: string(reqMsg)},
			{Key: "ctx_response_content", Value: string(respMsg)},
		}

		if err != nil {
			fields = append(fields, logx.Field{
				Key:   "ctx_response_error",
				Value: err.Error(),
			})
		}

		l.Debug("gRPC Message", fields...)

		return resp, err
	}
}

// WithErrorLogs returns a gRPC interceptor for unary calls that instrument requests
// with logs for the errors.
func WithErrorLogs(l logx.Logger, options ...ErrorLogsOption) grpc.UnaryServerInterceptor {
	var logOptions = &errorLogsOptions{}
	for _, option := range options {
		option(logOptions)
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		resp, err := handler(ctx, req)

		if err != nil {
			errCode := status.Code(err)
			if !inCodeList(errCode, logOptions.discardedCodes) {
				reqMsg, _ := json.Marshal(req)
				respMsg, _ := json.Marshal(resp)
				fields := []logx.Field{
					{Key: "ctx_full_method", Value: info.FullMethod},
					{Key: "ctx_request_content", Value: string(reqMsg)},
					{Key: "ctx_response_content", Value: string(respMsg)},
					{Key: "ctx_response_error_code", Value: errCode},
					{Key: "ctx_response_error_message", Value: err.Error()},
				}
				if inCodeList(errCode, logOptions.debugLevelCodes) {
					l.Debug("gRPC Error", fields...)
				} else {
					l.Error("gRPC Error", fields...)
				}
			}
		}

		return resp, err
	}
}

// ErrorLogsOption is the common type of functions that set errorLogsOptions
type ErrorLogsOption func(*errorLogsOptions)

type errorLogsOptions struct {
	debugLevelCodes []codes.Code
	discardedCodes  []codes.Code
}

func inCodeList(needle codes.Code, haystack []codes.Code) bool {
	for _, elem := range haystack {
		if elem == needle {
			return true
		}
	}
	return false
}

func WithDebugLevelCodes(codes ...codes.Code) func(*errorLogsOptions) {
	return func(logsConfig *errorLogsOptions) {
		logsConfig.debugLevelCodes = append(logsConfig.debugLevelCodes, codes...)
	}
}

func WithDiscardedCodes(codes ...codes.Code) func(*errorLogsOptions) {
	return func(logsConfig *errorLogsOptions) {
		logsConfig.discardedCodes = append(logsConfig.discardedCodes, codes...)
	}
}

func WithStructuredPanicLogs(l logx.Logger, options ...recovery.Options) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		recoveryHandler := recovery.Handler(l, options...)
		defer recoveryHandler()
		return handler(ctx, req)
	}
}
