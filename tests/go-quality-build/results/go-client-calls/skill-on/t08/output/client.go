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
			"initialBackoff": "0.010s",
			"maxBackoff": "0.020s",
			"backoffMultiplier": 2,
			"retryableStatusCodes": ["UNAVAILABLE"]
		}
	}]
}`

// Dial connects to target with opts and a fallback retry policy for health checks.
// Valid resolver service configurations take precedence. The caller owns the connection.
func Dial(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	dialOpts := make([]grpc.DialOption, 0, len(opts)+1)
	dialOpts = append(dialOpts, grpc.WithDefaultServiceConfig(healthServiceConfig))
	dialOpts = append(dialOpts, opts...)
	return grpc.Dial(target, dialOpts...)
}

// Probe checks whether service is serving within 500ms or the caller's earlier deadline.
// It leaves conn under the caller's ownership.
func Probe(ctx context.Context, conn *grpc.ClientConn, service string) error {
	ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()

	r, err := grpc_health_v1.NewHealthClient(conn).Check(ctx, &grpc_health_v1.HealthCheckRequest{Service: service})
	if err != nil {
		return fmt.Errorf("check health: %w", err)
	}
	if r.Status != grpc_health_v1.HealthCheckResponse_SERVING {
		return fmt.Errorf("not serving: %s", r.Status)
	}
	return nil
}
