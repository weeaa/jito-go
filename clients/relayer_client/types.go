package relayer_client

import (
	"github.com/weeaa/jito-go/pb"
	"github.com/weeaa/jito-go/pkg"
	"google.golang.org/grpc"
)

type Client struct {
	GrpcConn *grpc.ClientConn

	Relayer jito_pb.RelayerClient

	Auth *pkg.AuthenticationService

	ErrChan <-chan error // ErrChan is used for dispatching errors from functions executed within goroutines.
}
