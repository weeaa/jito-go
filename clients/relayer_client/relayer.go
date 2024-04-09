package relayer_client

import (
	"context"
	"crypto/tls"
	"github.com/gagliardetto/solana-go"
	"github.com/weeaa/jito-go/pkg"
	"github.com/weeaa/jito-go/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Client struct {
	GrpcConn *grpc.ClientConn
	Auth     *pkg.AuthenticationService

	Relayer proto.RelayerClient
}

func New(grpcDialURL string, privateKey solana.PrivateKey, tlsConfig *tls.Config, opts ...grpc.DialOption) (*Client, error) {
	if tlsConfig != nil {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	}

	conn, err := pkg.CreateAndObserveGRPCConn(context.Background(), grpcDialURL, opts...)
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
	}, nil
}

func (c *Client) GetTpuConfigs(opts ...grpc.CallOption) (*proto.GetTpuConfigsResponse, error) {
	return c.Relayer.GetTpuConfigs(c.Auth.GrpcCtx, &proto.GetTpuConfigsRequest{}, opts...)
}

func (c *Client) SubscribePackets(opts ...grpc.CallOption) (proto.Relayer_SubscribePacketsClient, error) {
	return c.Relayer.SubscribePackets(c.Auth.GrpcCtx, &proto.SubscribePacketsRequest{}, opts...)
}

func (c *Client) HandlePacketsSubscription() (proto.Relayer_SubscribePacketsClient, chan []*solana.Transaction, chan error) {
	chTx := make(chan []*solana.Transaction)
	chErr := make(chan error)
	sub, err := c.SubscribePackets()
	if err != nil {
		chErr <- err
		return nil, nil, chErr
	}

	go func() {
		for {
			select {
			case <-c.Auth.GrpcCtx.Done():
				return
			default:
				var packet *proto.SubscribePacketsResponse
				packet, err = sub.Recv()
				if err != nil {
					chErr <- err
				}

				var txns = make([]*solana.Transaction, 0, len(packet.Batch.GetPackets()))
				txns, err = pkg.ConvertBatchProtobufPacketToTransaction(packet.Batch.GetPackets())
				if err != nil {
					chErr <- err
				}

				chTx <- txns
			}
		}
	}()

	return sub, chTx, chErr
}
