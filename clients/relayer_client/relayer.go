package relayer_client

import (
	"crypto/tls"
	"github.com/gagliardetto/solana-go"
	"github.com/rs/zerolog"
	"github.com/weeaa/jito-go/pkg"
	"github.com/weeaa/jito-go/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"os"
)

type Client struct {
	GrpcConn *grpc.ClientConn
	logger   zerolog.Logger

	Relayer proto.RelayerClient

	Auth *pkg.AuthenticationService
}

// NewRelayerClient is a function that creates a new instance of a Client.
func NewRelayerClient(grpcDialURL string, privateKey solana.PrivateKey) (*Client, error) {
	conn, err := grpc.Dial(grpcDialURL, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	if err != nil {
		return nil, err
	}

	relayerClient := proto.NewRelayerClient(conn)
	authService := pkg.NewAuthenticationService(conn, privateKey)
	if err = authService.AuthenticateAndRefresh(proto.Role_RELAYER); err != nil {
		return nil, err
	}

	return &Client{
		GrpcConn: conn,
		Relayer:  relayerClient,
		Auth:     authService,
		logger:   zerolog.New(os.Stdout).With().Timestamp().Str("service", "relayer").Logger(),
	}, nil
}

func (c *Client) GetTpuConfigs() (*proto.GetTpuConfigsResponse, error) {
	return c.Relayer.GetTpuConfigs(c.Auth.GrpcCtx, &proto.GetTpuConfigsRequest{})
}

func (c *Client) SubscribePackets() (proto.Relayer_SubscribePacketsClient, error) {
	return c.Relayer.SubscribePackets(c.Auth.GrpcCtx, &proto.SubscribePacketsRequest{})
}
