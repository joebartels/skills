package client

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/resolver/manual"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"net"
	"sync/atomic"
	"testing"
)

type probeHealth struct {
	grpc_health_v1.UnimplementedHealthServer
	calls  atomic.Int32
	commit bool
}

func (s *probeHealth) Check(ctx context.Context, r *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	s.calls.Add(1)
	if s.commit {
		grpc.SendHeader(ctx, metadata.Pairs("ack", "yes"))
	}
	return nil, status.Error(codes.Unavailable, "lost reply")
}
func healthConn(t *testing.T, srv *probeHealth, opts ...grpc.DialOption) *grpc.ClientConn {
	t.Helper()
	l := bufconn.Listen(1 << 20)
	s := grpc.NewServer()
	grpc_health_v1.RegisterHealthServer(s, srv)
	go s.Serve(l)
	t.Cleanup(func() { s.Stop(); l.Close() })
	base := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return l.Dial() })}
	base = append(base, opts...)
	c, err := Dial("probe:///local", base...)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.Close() })
	return c
}
func TestDefaultNativePolicy(t *testing.T) {
	s := &probeHealth{}
	c := healthConn(t, s)
	err := Probe(context.Background(), c, "")
	if status.Code(err) != codes.Unavailable || s.calls.Load() != 3 {
		t.Fatalf("status=%v executions=%d want 3", err, s.calls.Load())
	}
}
func TestCommittedRPCFailureIsNotReplayed(t *testing.T) {
	s := &probeHealth{commit: true}
	c := healthConn(t, s)
	err := Probe(context.Background(), c, "")
	if status.Code(err) != codes.Unavailable || s.calls.Load() != 1 {
		t.Fatalf("committed status=%v executions=%d want 1", err, s.calls.Load())
	}
}
func TestResolverPolicyPrecedesDefault(t *testing.T) {
	b := manual.NewBuilderWithScheme("probe")
	b.InitialState(resolver.State{Addresses: []resolver.Address{{Addr: "local"}}})
	s := &probeHealth{}
	c := healthConn(t, s, grpc.WithResolvers(b))
	config := b.CC.ParseServiceConfig(`{"methodConfig":[{"name":[{"service":"grpc.health.v1.Health","method":"Check"}],"retryPolicy":{"maxAttempts":2,"initialBackoff":"0.010s","maxBackoff":"0.020s","backoffMultiplier":2,"retryableStatusCodes":["UNAVAILABLE"]}}]}`)
	b.UpdateState(resolver.State{Addresses: []resolver.Address{{Addr: "local"}}, ServiceConfig: config})
	err := Probe(context.Background(), c, "")
	if status.Code(err) != codes.Unavailable || s.calls.Load() != 2 {
		t.Fatalf("resolver status=%v executions=%d want 2", err, s.calls.Load())
	}
}
