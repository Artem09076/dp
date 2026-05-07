package authgrpc

import (
	"context"
	"time"

	"github.com/Artem09076/dp/backend/auth_service/internal/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func MetricsUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		duration := time.Since(start).Seconds()

		method := info.FullMethod
		statusCode := codes.OK
		if err != nil {
			if s, ok := status.FromError(err); ok {
				statusCode = s.Code()
			} else {
				statusCode = codes.Unknown
			}
		}

		metrics.GrpcRequestsTotal.WithLabelValues(method, statusCode.String()).Inc()
		metrics.GrpcRequestDuration.WithLabelValues(method).Observe(duration)

		return resp, err
	}
}
