package shredstream_client

import (
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/weeaa/jito-go/pkg"
	"github.com/weeaa/jito-go/proto"
	"google.golang.org/grpc"
)

type client struct {
	GrpcConn *grpc.ClientConn
	RpcConn  *rpc.Client

	ShredstreamClient proto.ShredstreamClient

	Auth *pkg.AuthenticationService
}
