package client

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func Dial(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return grpc.Dial(target, opts...)
}
func Probe(ctx context.Context, conn *grpc.ClientConn, service string) error {
	r, err := grpc_health_v1.NewHealthClient(conn).Check(ctx, &grpc_health_v1.HealthCheckRequest{Service: service})
	if err != nil {
		return err
	}
	if r.Status != grpc_health_v1.HealthCheckResponse_SERVING {
		return fmt.Errorf("not serving")
	}
	return nil
}
