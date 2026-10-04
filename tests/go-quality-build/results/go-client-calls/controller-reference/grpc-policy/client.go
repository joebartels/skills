package client

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"time"
)

func Dial(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	base := []grpc.DialOption{grpc.WithDefaultServiceConfig(`{"methodConfig":[{"name":[{"service":"grpc.health.v1.Health","method":"Check"}],"retryPolicy":{"maxAttempts":3,"initialBackoff":"0.010s","maxBackoff":"0.020s","backoffMultiplier":2,"retryableStatusCodes":["UNAVAILABLE"]}}]}`)}
	return grpc.Dial(target, append(base, opts...)...)
}
func Probe(ctx context.Context, conn *grpc.ClientConn, service string) error {
	ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()
	r, err := grpc_health_v1.NewHealthClient(conn).Check(ctx, &grpc_health_v1.HealthCheckRequest{Service: service})
	if err != nil {
		return fmt.Errorf("check: %w", err)
	}
	if r.Status != grpc_health_v1.HealthCheckResponse_SERVING {
		return fmt.Errorf("not serving")
	}
	return nil
}
