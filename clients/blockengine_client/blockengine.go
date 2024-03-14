package blockengine_client

import (
	"context"
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

type RelayerBlockEngineClient struct {
	GrpcConn *grpc.ClientConn
	RpcConn  *rpc.Client
	logger   zerolog.Logger

	BlockEngineRelayerClient proto.BlockEngineRelayerClient

	Auth  *pkg.AuthenticationService
	ErrCh chan error
}

type ValidatorBlockEngineClient struct {
	GrpcConn *grpc.ClientConn
	RpcConn  *rpc.Client
	logger   zerolog.Logger

	BlockEngineValidatorClient proto.BlockEngineValidatorClient

	Auth  *pkg.AuthenticationService
	ErrCh chan error
}

func NewRelayerBlockEngineClient(grpcDialURL string, rpcClient *rpc.Client, privateKey solana.PrivateKey) (*RelayerBlockEngineClient, error) {
	conn, err := grpc.Dial(grpcDialURL, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	if err != nil {
		return nil, err
	}

	blockEngineRelayerClient := proto.NewBlockEngineRelayerClient(conn)
	authService := pkg.NewAuthenticationService(conn, privateKey)
	if err = authService.AuthenticateAndRefresh(proto.Role_RELAYER); err != nil {
		return nil, err
	}

	return &RelayerBlockEngineClient{
		GrpcConn:                 conn,
		RpcConn:                  rpcClient,
		BlockEngineRelayerClient: blockEngineRelayerClient,
		Auth:                     pkg.NewAuthenticationService(conn, privateKey),
		logger:                   zerolog.New(os.Stdout).With().Logger(),
	}, nil
}

func NewValidatorBlockEngineClient(grpcDialURL string, rpcClient *rpc.Client, privateKey solana.PrivateKey) (*ValidatorBlockEngineClient, error) {
	conn, err := grpc.Dial(grpcDialURL, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	if err != nil {
		return nil, err
	}

	blockEngineValidatorClient := proto.NewBlockEngineValidatorClient(conn)
	authService := pkg.NewAuthenticationService(conn, privateKey)
	if err = authService.AuthenticateAndRefresh(proto.Role_VALIDATOR); err != nil {
		return nil, err
	}

	return &ValidatorBlockEngineClient{
		GrpcConn:                   conn,
		RpcConn:                    rpcClient,
		BlockEngineValidatorClient: blockEngineValidatorClient,
		Auth:                       authService,
		logger:                     zerolog.New(os.Stdout).With().Logger(),
	}, nil
}

func (c *ValidatorBlockEngineClient) BlockEngineValidatorClientSubscribePackets() (proto.BlockEngineValidator_SubscribePacketsClient, error) {
	return c.BlockEngineValidatorClient.SubscribePackets(c.Auth.GrpcCtx, &proto.SubscribePacketsRequest{})
}

func (c *ValidatorBlockEngineClient) HandlePacketSubscription(ctx context.Context, ch chan *proto.PacketBatch) (proto.BlockEngineValidator_SubscribePacketsClient, error) {
	sub, err := c.BlockEngineValidatorClientSubscribePackets()
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				c.logger.Error().Err(ctx.Err())
				return
			default:
				var subInfo *proto.SubscribePacketsResponse
				subInfo, err = sub.Recv()
				if err != nil {
					continue
				}
				ch <- subInfo.Batch
			}
		}
	}()

	return sub, err
}

func (c *ValidatorBlockEngineClient) BlockEngineValidatorClientSubscribeBundles() (proto.BlockEngineValidator_SubscribeBundlesClient, error) {
	return c.BlockEngineValidatorClient.SubscribeBundles(c.Auth.GrpcCtx, &proto.SubscribeBundlesRequest{})
}

func (c *ValidatorBlockEngineClient) HandleBundleSubscription(ctx context.Context, ch chan []*proto.BundleUuid) (proto.BlockEngineValidator_SubscribeBundlesClient, error) {
	sub, err := c.BlockEngineValidatorClientSubscribeBundles()
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				c.logger.Error().Err(ctx.Err())
				return
			default:
				var subInfo *proto.SubscribeBundlesResponse
				subInfo, err = sub.Recv()
				if err != nil {
					continue
				}
				ch <- subInfo.Bundles
			}
		}
	}()

	return sub, nil
}

func (c *ValidatorBlockEngineClient) BlockEngineValidatorClientGetBlockBuilderFeeInfo() (*proto.BlockBuilderFeeInfoResponse, error) {
	return c.BlockEngineValidatorClient.GetBlockBuilderFeeInfo(c.Auth.GrpcCtx, &proto.BlockBuilderFeeInfoRequest{})
}

func (c *RelayerBlockEngineClient) BlockEngineRelayerClientSubscribeAccountsOfInterest() (proto.BlockEngineRelayer_SubscribeAccountsOfInterestClient, error) {
	return c.BlockEngineRelayerClient.SubscribeAccountsOfInterest(c.Auth.GrpcCtx, &proto.AccountsOfInterestRequest{})
}

func (c *RelayerBlockEngineClient) BlockEngineRelayerClientSubscribeProgramsOfInterest() (proto.BlockEngineRelayer_SubscribeProgramsOfInterestClient, error) {
	return c.BlockEngineRelayerClient.SubscribeProgramsOfInterest(c.Auth.GrpcCtx, &proto.ProgramsOfInterestRequest{})
}

func (c *RelayerBlockEngineClient) BlockEngineRelayerClientStartExpiringPacketStream() (proto.BlockEngineRelayer_StartExpiringPacketStreamClient, error) {
	return c.BlockEngineRelayerClient.StartExpiringPacketStream(c.Auth.GrpcCtx)
}
