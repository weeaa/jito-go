package shredstream_client

import (
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/weeaa/jito-go/pkg"
	"google.golang.org/grpc"
)

type client struct {
	GrpcConn *grpc.ClientConn
	RpcConn  *rpc.Client

	ShredstreamClient pb.ShredstreamClient

	Auth *pkg.AuthenticationService
}
