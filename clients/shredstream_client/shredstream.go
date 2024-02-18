package shredstream_client

import (
	"crypto/tls"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/rs/zerolog"
	"github.com/weeaa/jito-go/pkg"
	"github.com/weeaa/jito-go/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"os"
)

type Client struct {
	GrpcConn *grpc.ClientConn
	RpcConn  *rpc.Client
	logger   zerolog.Logger

	ShredstreamClient proto.ShredstreamClient

	Auth *pkg.AuthenticationService
}

func NewShredstreamClient(grpcDialURL string, rpcClient *rpc.Client, privateKey solana.PrivateKey) (*Client, error) {
	conn, err := grpc.Dial(grpcDialURL, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	if err != nil {
		return nil, err
	}

	shredstreamService := proto.NewShredstreamClient(conn)
	authService := pkg.NewAuthenticationService(conn, privateKey)
	if err = authService.AuthenticateAndRefresh(proto.Role_SHREDSTREAM_SUBSCRIBER); err != nil {
		return nil, err
	}

	return &Client{
		GrpcConn:          conn,
		RpcConn:           rpcClient,
		ShredstreamClient: shredstreamService,
		Auth:              authService,
		logger:            zerolog.New(os.Stdout).With().Timestamp().Str("service", "shredstream-client").Logger(),
	}, nil
}

func (c *Client) SendHeartbeat(count uint64, opts ...grpc.CallOption) (*proto.HeartbeatResponse, error) {
	return c.ShredstreamClient.SendHeartbeat(c.Auth.GrpcCtx, &proto.Heartbeat{Count: count}, opts...)
}
