package client

import (
	"context"
	"errors"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

type reviewHealthServer struct {
	grpc_health_v1.UnimplementedHealthServer
	calls atomic.Int32
}

func (s *reviewHealthServer) Check(_ context.Context, r *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	s.calls.Add(1)
	if r.Service == "serving" {
		return &grpc_health_v1.HealthCheckResponse{Status: grpc_health_v1.HealthCheckResponse_SERVING}, nil
	}
	return nil, status.Error(codes.Unavailable, "review failure")
}

func reviewConnection(t *testing.T, dial func(string, ...grpc.DialOption) (*grpc.ClientConn, error), srv *reviewHealthServer, opts ...grpc.DialOption) *grpc.ClientConn {
	t.Helper()
	l := bufconn.Listen(1 << 20)
	s := grpc.NewServer(grpc.WaitForHandlers(true))
	grpc_health_v1.RegisterHealthServer(s, srv)
	done := make(chan error, 1)
	go func() { done <- s.Serve(l) }()
	t.Cleanup(func() {
		s.Stop()
		l.Close()
		select {
		case err := <-done:
			if err != nil && !errors.Is(err, grpc.ErrServerStopped) {
				t.Errorf("Serve: %v", err)
			}
		case <-time.After(3 * time.Second):
			t.Error("Serve did not return")
		}
	})
	base := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return l.DialContext(ctx) }),
		grpc.WithBlock(),
		grpc.WithTimeout(3 * time.Second),
	}
	c, err := dial("passthrough:///review", append(base, opts...)...)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.Close() })
	return c
}

func TestReviewBorrowedConnectionKeepsPolicyAndOwnership(t *testing.T) {
	srv := &reviewHealthServer{}
	c := reviewConnection(t, grpc.Dial, srv)
	if p := c.GetMethodConfig(grpc_health_v1.Health_Check_FullMethodName).RetryPolicy; p != nil {
		t.Fatalf("borrowed connection unexpectedly starts with policy: %+v", p)
	}
	err := Probe(context.Background(), c, "failing")
	if status.Code(err) != codes.Unavailable || errors.Unwrap(err) == nil {
		t.Fatalf("Probe = %v; want wrapped Unavailable", err)
	}
	if got := srv.calls.Load(); got != 1 {
		t.Fatalf("borrowed connection calls = %d; want 1", got)
	}
	if p := c.GetMethodConfig(grpc_health_v1.Health_Check_FullMethodName).RetryPolicy; p != nil {
		t.Fatalf("Probe added a retry policy: %+v", p)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := grpc_health_v1.NewHealthClient(c).Check(ctx, &grpc_health_v1.HealthCheckRequest{Service: "serving"})
	if err != nil || r.GetStatus() != grpc_health_v1.HealthCheckResponse_SERVING {
		t.Fatalf("borrowed connection reuse = %v, %v", r, err)
	}
}

func TestReviewHostDialOptionsRemainEffective(t *testing.T) {
	for _, tc := range []struct {
		name string
		option grpc.DialOption
	}{
		{"disable retries", grpc.WithDisableRetry()},
		{"host empty default", grpc.WithDefaultServiceConfig(`{}`)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := &reviewHealthServer{}
			var intercepted atomic.Int32
			c := reviewConnection(t, Dial, srv, tc.option, grpc.WithUnaryInterceptor(func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoke grpc.UnaryInvoker, opts ...grpc.CallOption) error {
				intercepted.Add(1)
				return invoke(ctx, method, req, reply, cc, opts...)
			}))
			if err := Probe(context.Background(), c, "failing"); status.Code(err) != codes.Unavailable {
				t.Fatalf("Probe = %v; want Unavailable", err)
			}
			if got := srv.calls.Load(); got != 1 {
				t.Errorf("server calls = %d; want host option to limit calls to 1", got)
			}
			if got := intercepted.Load(); got != 1 {
				t.Errorf("host interceptor calls = %d; want 1", got)
			}
		})
	}
	c, err := Dial("passthrough:///unused")
	if c != nil {
		c.Close()
	}
	if err == nil {
		t.Fatal("Dial unexpectedly supplied credentials")
	}
}
