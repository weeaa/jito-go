package shredstream_client

import (
	"github.com/MintyFinance/jito-go/pkg"
	"github.com/MintyFinance/jito-go/proto"
	"github.com/MintyFinance/solana-go-custom/rpc"
	"google.golang.org/grpc"
)

type client struct {
	GrpcConn *grpc.ClientConn
	RpcConn  *rpc.Client

	ShredstreamClient proto.ShredstreamClient

	Auth *pkg.AuthenticationService
}
