package client

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	pb "google.golang.org/grpc/interop/grpc_testing"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"net"
	"sync"
	"testing"
)

type commitProbe struct {
	pb.UnimplementedTestServiceServer
	mu             sync.Mutex
	calls, effects int
	seen           map[string]string
	committed      bool
	t              *testing.T
}

func (s *commitProbe) UnaryCall(ctx context.Context, r *pb.SimpleRequest) (*pb.SimpleResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls++
	md, _ := metadata.FromIncomingContext(ctx)
	key := ""
	if v := md.Get("operation-id"); len(v) > 0 {
		key = v[0]
	}
	body := string(r.GetPayload().GetBody())
	if key != "stable-9" || body != "payload-9" {
		s.t.Errorf("identity/payload %q %q", key, body)
	}
	if len(md.Get("tenant")) != 1 || md.Get("tenant")[0] != "north" {
		s.t.Error("caller metadata lost")
	}
	if old, ok := s.seen[key]; !ok || key == "" {
		s.effects++
		s.seen[key] = body
	} else if old != body {
		s.t.Error("payload changed")
	}
	if s.committed && s.calls == 1 {
		grpc.SendHeader(ctx, metadata.Pairs("ack", "yes"))
		return nil, status.Error(codes.Unavailable, "committed lost ack")
	}
	if !s.committed && s.calls < 3 {
		return nil, status.Error(codes.Unavailable, "lost ack")
	}
	return &pb.SimpleResponse{Payload: &pb.Payload{Body: []byte("receipt")}}, nil
}
func transferConn(t *testing.T, srv *commitProbe) *grpc.ClientConn {
	t.Helper()
	l := bufconn.Listen(1 << 20)
	s := grpc.NewServer()
	pb.RegisterTestServiceServer(s, srv)
	go s.Serve(l)
	t.Cleanup(func() { s.Stop(); l.Close() })
	c, err := grpc.Dial("passthrough:///local", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return l.Dial() }), grpc.WithDefaultServiceConfig(`{"methodConfig":[{"name":[{"service":"grpc.testing.TestService","method":"UnaryCall"}],"retryPolicy":{"maxAttempts":3,"initialBackoff":"0.010s","maxBackoff":"0.020s","backoffMultiplier":2,"retryableStatusCodes":["UNAVAILABLE"]}}]}`))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.Close() })
	return c
}
func TestTransferNativeReplayIdentity(t *testing.T) {
	s := &commitProbe{seen: map[string]string{}, t: t}
	c := transferConn(t, s)
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("tenant", "north"))
	got, err := Commit(ctx, c, "stable-9", []byte("payload-9"))
	if err != nil || string(got) != "receipt" {
		t.Fatalf("result %q %v", got, err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.calls != 3 || s.effects != 1 {
		t.Fatalf("executions=%d effects=%d", s.calls, s.effects)
	}
}
func TestTransferIdentitySurvivesOuterRetry(t *testing.T) {
	s := &commitProbe{seen: map[string]string{}, committed: true, t: t}
	c := transferConn(t, s)
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("tenant", "north"))
	_, err := Commit(ctx, c, "stable-9", []byte("payload-9"))
	if status.Code(err) != codes.Unavailable {
		t.Fatalf("committed failure hidden: %v", err)
	}
	s.mu.Lock()
	first := s.calls
	s.mu.Unlock()
	if first != 1 {
		t.Fatalf("committed execution repeated %d", first)
	}
	got, err := Commit(ctx, c, "stable-9", []byte("payload-9"))
	if err != nil || string(got) != "receipt" {
		t.Fatalf("outer result %q %v", got, err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.calls != 2 || s.effects != 1 {
		t.Fatalf("executions=%d effects=%d", s.calls, s.effects)
	}
}
