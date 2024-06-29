package blockengine_client

import (
	"github.com/weeaa/jito-go/pb"
	"github.com/weeaa/jito-go/pkg"
	"google.golang.org/grpc"
)

type Relayer struct {
	GrpcConn *grpc.ClientConn

	Client jito_pb.BlockEngineRelayerClient

	Auth *pkg.AuthenticationService

	ErrChan <-chan error // ErrChan is used for dispatching errors from functions executed within goroutines.
}

type Validator struct {
	GrpcConn *grpc.ClientConn

	Client jito_pb.BlockEngineValidatorClient

	Auth *pkg.AuthenticationService

	ErrChan <-chan error // ErrChan is used for dispatching errors from functions executed within goroutines.
}
