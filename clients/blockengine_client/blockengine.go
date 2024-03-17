package blockengine_client

import (
	"context"
	"crypto/tls"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/weeaa/jito-go/pkg"
	"github.com/weeaa/jito-go/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type RelayerBlockEngineClient struct {
	GrpcConn  *grpc.ClientConn
	GrpcErrCh chan error
	RpcConn   *rpc.Client
	ErrCh     chan error

	BlockEngineRelayerClient proto.BlockEngineRelayerClient

	Auth *pkg.AuthenticationService
}

type ValidatorBlockEngineClient struct {
	GrpcConn  *grpc.ClientConn
	GrpcErrCh chan error
	RpcConn   *rpc.Client
	ErrCh     chan error

	BlockEngineValidatorClient proto.BlockEngineValidatorClient

	Auth *pkg.AuthenticationService
}

func NewRelayerBlockEngineClient(grpcDialURL string, rpcClient *rpc.Client, privateKey solana.PrivateKey, tlsConfig *tls.Config, opts ...grpc.DialOption) (*RelayerBlockEngineClient, error) {
	if tlsConfig != nil {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	}

	grpcErrChan := make(chan error)
	conn, err := pkg.CreateAndObserveGRPCConn(context.TODO(), grpcErrChan, grpcDialURL, opts...)
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
		GrpcErrCh:                grpcErrChan,
	}, nil
}

func NewValidatorBlockEngineClient(grpcDialURL string, rpcClient *rpc.Client, privateKey solana.PrivateKey, tlsConfig *tls.Config, opts ...grpc.DialOption) (*ValidatorBlockEngineClient, error) {
	if tlsConfig != nil {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	}

	grpcErrChan := make(chan error)
	conn, err := pkg.CreateAndObserveGRPCConn(context.TODO(), grpcErrChan, grpcDialURL, opts...)
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
		GrpcErrCh:                  grpcErrChan,
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

func (c *ValidatorBlockEngineClient) BlockEngineValidatorClientGetBlockBuilderFeeInfo(opts ...grpc.CallOption) (*proto.BlockBuilderFeeInfoResponse, error) {
	return c.BlockEngineValidatorClient.GetBlockBuilderFeeInfo(c.Auth.GrpcCtx, &proto.BlockBuilderFeeInfoRequest{}, opts...)
}

func (c *RelayerBlockEngineClient) BlockEngineRelayerClientSubscribeAccountsOfInterest(opts ...grpc.CallOption) (proto.BlockEngineRelayer_SubscribeAccountsOfInterestClient, error) {
	return c.BlockEngineRelayerClient.SubscribeAccountsOfInterest(c.Auth.GrpcCtx, &proto.AccountsOfInterestRequest{}, opts...)
}

func (c *RelayerBlockEngineClient) BlockEngineRelayerClientSubscribeProgramsOfInterest(opts ...grpc.CallOption) (proto.BlockEngineRelayer_SubscribeProgramsOfInterestClient, error) {
	return c.BlockEngineRelayerClient.SubscribeProgramsOfInterest(c.Auth.GrpcCtx, &proto.ProgramsOfInterestRequest{}, opts...)
}

func (c *RelayerBlockEngineClient) BlockEngineRelayerClientStartExpiringPacketStream(opts ...grpc.CallOption) (proto.BlockEngineRelayer_StartExpiringPacketStreamClient, error) {
	return c.BlockEngineRelayerClient.StartExpiringPacketStream(c.Auth.GrpcCtx, opts...)
}
