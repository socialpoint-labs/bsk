package grpcx

import (
	"context"
	"encoding/json"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/socialpoint-labs/bsk/logx"
	"github.com/socialpoint-labs/bsk/metrics"
	"github.com/socialpoint-labs/bsk/recovery"
)

// Deprecated: use WithMetricsUnary instead
func WithMetrics(m metrics.Metrics) grpc.UnaryServerInterceptor {
	return WithMetricsUnary(m)
}

// WithMetricsUnary returns a gRPC interceptor for UNARY calls that instrument requests with a metric for the request duration.
func WithMetricsUnary(m metrics.Metrics) grpc.UnaryServerInterceptor {
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

// WithMetricsStream returns a gRPC interceptor for STREAM calls that instrument requests with a metric for the request duration.
func WithMetricsStream(m metrics.Metrics) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		timer := m.Timer("grpc.request_duration")
		timer.Start()

		err := handler(srv, ss)
		timer.
			WithTag("rpc_method", info.FullMethod).
			WithTag("rpc_error_code", status.Code(err)).
			WithTag("success", err == nil).
			Stop()

		return err
	}
}

// Deprecated: use WithRequestResponseLogsUnary instead
func WithRequestResponseLogs(l logx.Logger) grpc.UnaryServerInterceptor {
	return WithRequestResponseLogsUnary(l)
}

// WithRequestResponseLogsUnary returns a gRPC interceptor for UNARY calls that instrument requests with logs for the request and response.
func WithRequestResponseLogsUnary(l logx.Logger) grpc.UnaryServerInterceptor {
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

		l.Info("gRPC Message", fields...)

		return resp, err
	}
}

// Deprecated: use WithErrorLogsUnary instead
func WithErrorLogs(l logx.Logger, options ...ErrorLogsOption) grpc.UnaryServerInterceptor {
	return WithErrorLogsUnary(l, options...)
}

// WithErrorLogsUnary returns a gRPC interceptor for UNARY calls that instrument requests with logs for the errors.
func WithErrorLogsUnary(l logx.Logger, options ...ErrorLogsOption) grpc.UnaryServerInterceptor {
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
					l.Info("gRPC Error: "+err.Error(), fields...)
				} else {
					l.Error("gRPC Error: "+err.Error(), fields...)
				}
			}
		}

		return resp, err
	}
}

// WithErrorLogsStream returns a gRPC interceptor for STREAM calls that instrument requests with logs for the errors.
func WithErrorLogsStream(l logx.Logger, options ...ErrorLogsOption) grpc.StreamServerInterceptor {
	var logOptions = &errorLogsOptions{}
	for _, option := range options {
		option(logOptions)
	}

	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		err := handler(srv, ss)

		if err != nil {
			errCode := status.Code(err)
			if !inCodeList(errCode, logOptions.discardedCodes) {
				fields := []logx.Field{
					{Key: "ctx_full_method", Value: info.FullMethod},
					{Key: "ctx_response_error_code", Value: errCode},
					{Key: "ctx_response_error_message", Value: err.Error()},
					{Key: "ctx_error_code", Value: status.Code(err)},
				}
				if inCodeList(errCode, logOptions.debugLevelCodes) {
					l.Info("gRPC Error", fields...)
				} else {
					l.Error("gRPC Error", fields...)
				}
			}
		}

		return err
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

// Deprecated: use WithStructuredPanicLogsUnary instead
func WithStructuredPanicLogs(l logx.Logger, options ...recovery.Options) grpc.UnaryServerInterceptor {
	return WithStructuredPanicLogsUnary(l, options...)
}

func WithStructuredPanicLogsUnary(l logx.Logger, options ...recovery.Options) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		recoveryHandler := recovery.Handler(l, options...)
		defer recoveryHandler()
		return handler(ctx, req)
	}
}
