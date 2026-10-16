package client

import (
	"context"
	"google.golang.org/grpc"
	pb "google.golang.org/grpc/interop/grpc_testing"
)

func Commit(ctx context.Context, conn *grpc.ClientConn, operationID string, payload []byte) ([]byte, error) {
	r, err := pb.NewTestServiceClient(conn).UnaryCall(ctx, &pb.SimpleRequest{Payload: &pb.Payload{Body: payload}})
	if err != nil {
		return nil, err
	}
	return r.GetPayload().GetBody(), nil
}
