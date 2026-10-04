package client

import (
	"context"
	"errors"
	"fmt"
	"google.golang.org/grpc"
	pb "google.golang.org/grpc/interop/grpc_testing"
	"google.golang.org/grpc/metadata"
	"time"
)

func Commit(ctx context.Context, conn *grpc.ClientConn, operationID string, payload []byte) ([]byte, error) {
	if operationID == "" {
		return nil, errors.New("empty identity")
	}
	ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()
	md, _ := metadata.FromOutgoingContext(ctx)
	md = md.Copy()
	md.Set("operation-id", operationID)
	ctx = metadata.NewOutgoingContext(ctx, md)
	r, err := pb.NewTestServiceClient(conn).UnaryCall(ctx, &pb.SimpleRequest{Payload: &pb.Payload{Body: payload}})
	if err != nil {
		return nil, fmt.Errorf("commit outcome uncertain: %w", err)
	}
	return r.GetPayload().GetBody(), nil
}
