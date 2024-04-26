package geyser_client

import (
	"context"
	"github.com/weeaa/jito-go/proto"
	"google.golang.org/grpc"
)

type Client struct {
	GrpcConn *grpc.ClientConn
	Ctx      context.Context

	Geyser proto.GeyserClient

	ErrChan <-chan error
}
