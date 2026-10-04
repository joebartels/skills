package client

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/test/bufconn"
	"net"
	"testing"
)

func TestExistingServing(t *testing.T) {
	l := bufconn.Listen(1 << 20)
	s := grpc.NewServer()
	h := health.NewServer()
	h.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(s, h)
	go s.Serve(l)
	defer s.Stop()
	defer l.Close()
	c, err := Dial("passthrough:///local", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return l.Dial() }))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if err := Probe(context.Background(), c, ""); err != nil {
		t.Fatal(err)
	}
}
