package client

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	pb "google.golang.org/grpc/interop/grpc_testing"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Commit submits payload under a nonempty operationID using the borrowed conn.
// It has a 500ms total budget, subject to the caller's earlier deadline.
// A failed acknowledgement may follow an effect; callers can repeat the same
// identity and payload under the server's deduplication contract.
func Commit(ctx context.Context, conn *grpc.ClientConn, operationID string, payload []byte) ([]byte, error) {
	if operationID == "" {
		return nil, status.Error(codes.InvalidArgument, "operationID must not be empty")
	}
	ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()
	md, _ := metadata.FromOutgoingContext(ctx)
	if md == nil {
		md = metadata.MD{}
	}
	md.Set("operation-id", operationID)
	ctx = metadata.NewOutgoingContext(ctx, md)
	r, err := pb.NewTestServiceClient(conn).UnaryCall(ctx, &pb.SimpleRequest{Payload: &pb.Payload{Body: payload}})
	if err != nil {
		return nil, err
	}
	return r.GetPayload().GetBody(), nil
}
