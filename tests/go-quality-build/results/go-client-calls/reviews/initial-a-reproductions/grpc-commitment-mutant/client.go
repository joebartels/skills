package client

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
)

const healthServiceConfig = `{
	"methodConfig": [{
		"name": [{"service": "grpc.health.v1.Health", "method": "Check"}],
		"retryPolicy": {
			"maxAttempts": 3,
			"initialBackoff": "0.01s",
			"maxBackoff": "0.02s",
			"backoffMultiplier": 2,
			"retryableStatusCodes": ["UNAVAILABLE"]
		}
	}]
}`

// Dial creates a connection with a fallback retry policy for Health/Check.
func Dial(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	opts = append([]grpc.DialOption{grpc.WithDefaultServiceConfig(healthServiceConfig)}, opts...)
	return grpc.Dial(target, opts...)
}

// Probe checks for SERVING within the caller's deadline or 500ms, whichever is earlier.
func Probe(ctx context.Context, conn *grpc.ClientConn, service string) error {
	ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()
	r, err := grpc_health_v1.NewHealthClient(conn).Check(ctx, &grpc_health_v1.HealthCheckRequest{Service: service})
	for attempt := 0; attempt < 2 && err != nil; attempt++ {
		r, err = grpc_health_v1.NewHealthClient(conn).Check(ctx, &grpc_health_v1.HealthCheckRequest{Service: service})
	}
	if err != nil {
		return fmt.Errorf("health check %q: %w", service, err)
	}
	if r.Status != grpc_health_v1.HealthCheckResponse_SERVING {
		return fmt.Errorf("health check %q: status %s", service, r.Status)
	}
	return nil
}
