package relayer_client

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/gagliardetto/solana-go"
	"github.com/weeaa/jito-go/pb"
	"github.com/weeaa/jito-go/pkg"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func New(ctx context.Context, grpcDialURL string, privateKey solana.PrivateKey, tlsConfig *tls.Config, opts ...grpc.DialOption) (*Client, error) {
	if tlsConfig != nil {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	}

	chErr := make(chan error)
	conn, err := pkg.CreateAndObserveGRPCConn(ctx, chErr, grpcDialURL, opts...)
	if err != nil {
		return nil, err
	}

	relayerClient := jito_pb.NewRelayerClient(conn)
	authService := pkg.NewAuthenticationService(conn, privateKey)
	if err = authService.AuthenticateAndRefresh(jito_pb.Role_RELAYER); err != nil {
		return nil, err
	}

	return &Client{
		GrpcConn: conn,
		Relayer:  relayerClient,
		Auth:     authService,
		ErrChan:  chErr,
	}, nil
}

func (c *Client) GetTpuConfigs(opts ...grpc.CallOption) (*jito_pb.GetTpuConfigsResponse, error) {
	return c.Relayer.GetTpuConfigs(c.Auth.GrpcCtx, &jito_pb.GetTpuConfigsRequest{}, opts...)
}

func (c *Client) NewPacketsSubscription(opts ...grpc.CallOption) (jito_pb.Relayer_SubscribePacketsClient, error) {
	return c.Relayer.SubscribePackets(c.Auth.GrpcCtx, &jito_pb.SubscribePacketsRequest{}, opts...)
}

// SubscribePackets is a wrapper around NewPacketsSubscription.
func (c *Client) SubscribePackets(ctx context.Context) (<-chan []*solana.Transaction, <-chan error, error) {
	chTx := make(chan []*solana.Transaction)
	chErr := make(chan error)

	sub, err := c.NewPacketsSubscription()
	if err != nil {
		return nil, nil, err
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-c.Auth.GrpcCtx.Done():
				return
			default:
				packet, err := sub.Recv()
				if err != nil {
					chErr <- fmt.Errorf("SubscribePackets: failed to receive packet information: %w", err)
				}

				var txns = make([]*solana.Transaction, 0, len(packet.Batch.GetPackets()))
				txns, err = pkg.ConvertBatchProtobufPacketToTransaction(packet.Batch.GetPackets())
				if err != nil {
					chErr <- fmt.Errorf("SubscribePackets: failed to convert protobuf packet to transaction: %w", err)
				}

				chTx <- txns
			}
		}
	}()

	return chTx, chErr, nil
}
