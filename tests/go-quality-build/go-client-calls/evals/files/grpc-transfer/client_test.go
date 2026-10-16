package client

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "google.golang.org/grpc/interop/grpc_testing"
	"google.golang.org/grpc/test/bufconn"
	"net"
	"testing"
)

type existingServer struct {
	pb.UnimplementedTestServiceServer
}

func (existingServer) UnaryCall(ctx context.Context, r *pb.SimpleRequest) (*pb.SimpleResponse, error) {
	return &pb.SimpleResponse{Payload: r.Payload}, nil
}
func TestExistingCommit(t *testing.T) {
	l := bufconn.Listen(1 << 20)
	s := grpc.NewServer()
	pb.RegisterTestServiceServer(s, existingServer{})
	go s.Serve(l)
	defer s.Stop()
	defer l.Close()
	c, err := grpc.Dial("passthrough:///local", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return l.Dial() }))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	got, err := Commit(context.Background(), c, "op-1", []byte("value"))
	if err != nil || string(got) != "value" {
		t.Fatalf("got %q %v", got, err)
	}
}
