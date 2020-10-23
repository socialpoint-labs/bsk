package grpcx

import (
	"context"
	"encoding/json"

	"github.com/socialpoint-labs/bsk/logx"
	"github.com/socialpoint-labs/bsk/metrics"
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

		l.Info("gRPC Message", fields...)

		return resp, err
	}
}

// WithErrorLogs returns a gRPC interceptor for unary calls that instrument requests
// with logs for the errors.
func WithErrorLogs(l logx.Logger, options ...func(*withErrorLogsOptions)) grpc.UnaryServerInterceptor {
	var logOptions = &withErrorLogsOptions{
		debugLevelCodes: []codes.Code{},
		infoLevelCodes:  []codes.Code{codes.Canceled, codes.Unknown, codes.InvalidArgument, codes.DeadlineExceeded, codes.NotFound, codes.AlreadyExists, codes.PermissionDenied, codes.ResourceExhausted, codes.FailedPrecondition, codes.Aborted, codes.OutOfRange, codes.Unimplemented, codes.Internal, codes.Unavailable, codes.DataLoss, codes.Unauthenticated},
	}
	for _, option := range options {
		option(logOptions)
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		resp, err := handler(ctx, req)

		if err != nil {
			errCode := status.Code(err)
			reqMsg, _ := json.Marshal(req)
			respMsg, _ := json.Marshal(resp)
			fields := []logx.Field{
				{Key: "ctx_full_method", Value: info.FullMethod},
				{Key: "ctx_request_content", Value: string(reqMsg)},
				{Key: "ctx_response_content", Value: string(respMsg)},
				{Key: "ctx_response_error_code", Value: errCode},
				{Key: "ctx_response_error_message", Value: err.Error()},
			}
			if logOptions.hasDebugLevel(errCode) {
				l.Debug("gRPC Error", fields...)
			}
			if logOptions.hasInfoLevel(errCode) {
				l.Info("gRPC Error", fields...)
			}
		}

		return resp, err
	}
}

type withErrorLogsOptions struct {
	debugLevelCodes []codes.Code
	infoLevelCodes  []codes.Code
}

func (o *withErrorLogsOptions) hasDebugLevel(code codes.Code) bool {
	return inArray(code, o.debugLevelCodes)
}

func (o *withErrorLogsOptions) hasInfoLevel(code codes.Code) bool {
	return inArray(code, o.infoLevelCodes)
}

func inArray(needle codes.Code, haystack []codes.Code) bool {
	for _, elem := range haystack {
		if elem == needle {
			return true
		}
	}
	return false
}

func SetDebugLevelCodes(codes []codes.Code) func(*withErrorLogsOptions) {
	return func(logsConfig *withErrorLogsOptions) {
		logsConfig.debugLevelCodes = codes
	}
}

func SetInfoLevelCodes(codes []codes.Code) func(*withErrorLogsOptions) {
	return func(logsConfig *withErrorLogsOptions) {
		logsConfig.infoLevelCodes = codes
	}
}
