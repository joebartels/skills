package review

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
	health "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/resolver/manual"
	"google.golang.org/grpc/serviceconfig"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

type healthServer struct {
	health.UnimplementedHealthServer
	calls atomic.Int32
}

func (s *healthServer) Check(context.Context, *health.HealthCheckRequest) (*health.HealthCheckResponse, error) {
	s.calls.Add(1)
	return nil, status.Error(codes.Unavailable, "deliberate dependency failure")
}

func TestResolverServiceConfig(t *testing.T) {
	for _, tc := range []struct {
		name                         string
		empty, invalid, invalidAfter bool
		want                         int32
	}{
		{name: "absent uses default", want: 3},
		{name: "valid empty supersedes default", empty: true, want: 1},
		{name: "invalid initial fails without default", invalid: true, want: 0},
		{name: "invalid update retains previous policy", invalidAfter: true, want: 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			listener := bufconn.Listen(1 << 20)
			defer listener.Close()
			implementation := &healthServer{}
			server := grpc.NewServer()
			health.RegisterHealthServer(server, implementation)
			defer server.Stop()
			go server.Serve(listener)
			builder := manual.NewBuilderWithScheme("review")
			state := resolver.State{Addresses: []resolver.Address{{Addr: "local"}}}
			invalid := &serviceconfig.ParseResult{Err: errors.New("invalid resolver policy")}
			if tc.invalid {
				state.ServiceConfig = invalid
			}
			builder.InitialState(state)
			fallback := `{"methodConfig":[{"name":[{}],"retryPolicy":{"maxAttempts":3,"initialBackoff":"0.001s","maxBackoff":"0.002s","backoffMultiplier":1,"retryableStatusCodes":["UNAVAILABLE"]}}]}`
			cc, err := grpc.DialContext(context.Background(), "review:///target",
				grpc.WithResolvers(builder), grpc.WithTransportCredentials(insecure.NewCredentials()),
				grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return listener.Dial() }),
				grpc.WithDefaultServiceConfig(fallback))
			if err != nil {
				t.Fatal(err)
			}
			defer cc.Close()
			client := health.NewHealthClient(cc)
			call := func(want int32) {
				t.Helper()
				implementation.calls.Store(0)
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				_, err := client.Check(ctx, &health.HealthCheckRequest{})
				if status.Code(err) != codes.Unavailable {
					t.Fatalf("unexpected status %v", err)
				}
				if got := implementation.calls.Load(); got != want {
					t.Fatalf("executions=%d want=%d", got, want)
				}
				t.Logf("executions=%d; effective retry policy=%v", want,
					cc.GetMethodConfig("/grpc.health.v1.Health/Check").RetryPolicy)
			}
			if tc.empty {
				state.ServiceConfig = builder.CC.ParseServiceConfig(`{}`)
				builder.UpdateState(state)
			}
			if tc.invalidAfter {
				call(3)
				state.ServiceConfig = invalid
				builder.UpdateState(state)
			}
			call(tc.want)
		})
	}
}
