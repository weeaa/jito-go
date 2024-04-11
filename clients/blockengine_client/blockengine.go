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

type Relayer struct {
	GrpcConn *grpc.ClientConn
	RpcConn  *rpc.Client

	Client proto.BlockEngineRelayerClient

	Auth *pkg.AuthenticationService

	ErrChan chan error
}

type Validator struct {
	GrpcConn *grpc.ClientConn
	RpcConn  *rpc.Client

	Client proto.BlockEngineValidatorClient

	Auth *pkg.AuthenticationService

	ErrChan chan error
}

func NewRelayer(grpcDialURL string, rpcClient *rpc.Client, privateKey solana.PrivateKey, tlsConfig *tls.Config, opts ...grpc.DialOption) (*Relayer, error) {
	if tlsConfig != nil {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	}

	conn, err := pkg.CreateAndObserveGRPCConn(context.TODO(), grpcDialURL, opts...)
	if err != nil {
		return nil, err
	}

	blockEngineRelayerClient := proto.NewBlockEngineRelayerClient(conn)
	authService := pkg.NewAuthenticationService(conn, privateKey)
	if err = authService.AuthenticateAndRefresh(proto.Role_RELAYER); err != nil {
		return nil, err
	}

	return &Relayer{
		GrpcConn: conn,
		RpcConn:  rpcClient,
		Client:   blockEngineRelayerClient,
		Auth:     pkg.NewAuthenticationService(conn, privateKey),
	}, nil
}

func NewValidator(grpcDialURL string, rpcClient *rpc.Client, privateKey solana.PrivateKey, tlsConfig *tls.Config, opts ...grpc.DialOption) (*Validator, error) {
	if tlsConfig != nil {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	}

	conn, err := pkg.CreateAndObserveGRPCConn(context.Background(), grpcDialURL, opts...)
	if err != nil {
		return nil, err
	}

	blockEngineValidatorClient := proto.NewBlockEngineValidatorClient(conn)
	authService := pkg.NewAuthenticationService(conn, privateKey)
	if err = authService.AuthenticateAndRefresh(proto.Role_VALIDATOR); err != nil {
		return nil, err
	}

	return &Validator{
		GrpcConn: conn,
		RpcConn:  rpcClient,
		Client:   blockEngineValidatorClient,
		Auth:     authService,
	}, nil
}

func (c *Validator) SubscribePackets() (proto.BlockEngineValidator_SubscribePacketsClient, error) {
	return c.Client.SubscribePackets(c.Auth.GrpcCtx, &proto.SubscribePacketsRequest{})
}

func (c *Validator) HandlePacketSubscription(ctx context.Context, ch chan *proto.PacketBatch) (proto.BlockEngineValidator_SubscribePacketsClient, error) {
	sub, err := c.SubscribePackets()
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

func (c *Validator) SubscribeBundles() (proto.BlockEngineValidator_SubscribeBundlesClient, error) {
	return c.Client.SubscribeBundles(c.Auth.GrpcCtx, &proto.SubscribeBundlesRequest{})
}

func (c *Validator) HandleBundleSubscription(ctx context.Context, ch chan []*proto.BundleUuid) (proto.BlockEngineValidator_SubscribeBundlesClient, error) {
	sub, err := c.SubscribeBundles()
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

func (c *Validator) GetBlockBuilderFeeInfo(opts ...grpc.CallOption) (*proto.BlockBuilderFeeInfoResponse, error) {
	return c.Client.GetBlockBuilderFeeInfo(c.Auth.GrpcCtx, &proto.BlockBuilderFeeInfoRequest{}, opts...)
}

func (c *Relayer) SubscribeAccountsOfInterest(opts ...grpc.CallOption) (proto.BlockEngineRelayer_SubscribeAccountsOfInterestClient, error) {
	return c.Client.SubscribeAccountsOfInterest(c.Auth.GrpcCtx, &proto.AccountsOfInterestRequest{}, opts...)
}

func (c *Relayer) SubscribeProgramsOfInterest(opts ...grpc.CallOption) (proto.BlockEngineRelayer_SubscribeProgramsOfInterestClient, error) {
	return c.Client.SubscribeProgramsOfInterest(c.Auth.GrpcCtx, &proto.ProgramsOfInterestRequest{}, opts...)
}

func (c *Relayer) StartExpiringPacketStream(opts ...grpc.CallOption) (proto.BlockEngineRelayer_StartExpiringPacketStreamClient, error) {
	return c.Client.StartExpiringPacketStream(c.Auth.GrpcCtx, opts...)
}
