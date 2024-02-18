package searcher_client

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/rs/zerolog"
	"github.com/weeaa/jito-go/pkg"
	"github.com/weeaa/jito-go/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"math/rand"
	"os"
)

type BundleRejectionError struct {
	Message string
}

func (e BundleRejectionError) Error() string {
	return e.Message
}

func NewStateAuctionBidRejectedError(auction string, tip uint64) error {
	return BundleRejectionError{
		Message: fmt.Sprintf("bundle lost state auction, auction: %s, tip %d lamports", auction, tip),
	}
}

func NewWinningBatchBidRejectedError(auction string, tip uint64) error {
	return BundleRejectionError{
		Message: fmt.Sprintf("bundle won state auction but failed global auction, auction %s, tip %d lamports", auction, tip),
	}
}

func NewSimulationFailureError(tx string, message string) error {
	return BundleRejectionError{
		Message: fmt.Sprintf("bundle simulation failure on tx %s, message: %s", tx, message),
	}
}

func NewInternalError(message string) error {
	return BundleRejectionError{
		Message: fmt.Sprintf("internal error %s", message),
	}
}

func NewDroppedBundle(message string) error {
	return BundleRejectionError{
		Message: fmt.Sprintf("bundle dropped %s", message),
	}
}

type Client struct {
	GrpcConn *grpc.ClientConn
	RpcConn  *rpc.Client
	logger   zerolog.Logger

	SearcherService proto.SearcherServiceClient

	Auth *pkg.AuthenticationService
}

func NewSearcherClient(grpcDialURL string, rpcClient *rpc.Client, privateKey solana.PrivateKey, tlsConfig *tls.Config, opts ...grpc.DialOption) (*Client, error) {

	if tlsConfig != nil {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	}

	conn, err := grpc.Dial(grpcDialURL, opts...)
	if err != nil {
		return nil, err
	}

	searcherService := proto.NewSearcherServiceClient(conn)
	authService := pkg.NewAuthenticationService(conn, privateKey)
	if err = authService.AuthenticateAndRefresh(proto.Role_SEARCHER); err != nil {
		return nil, err
	}

	return &Client{
		GrpcConn:        conn,
		RpcConn:         rpcClient,
		SearcherService: searcherService,
		Auth:            authService,
		logger:          zerolog.New(os.Stdout).With().Timestamp().Str("service", "searcher-client").Logger(),
	}, nil
}

// NewMempoolStream creates a new mempool subscription.
func (c *Client) NewMempoolStream(accounts, regions []string) (proto.SearcherService_SubscribeMempoolClient, error) {
	return c.SearcherService.SubscribeMempool(c.Auth.GrpcCtx, &proto.MempoolSubscription{Msg: &proto.MempoolSubscription_WlaV0Sub{
		WlaV0Sub: &proto.WriteLockedAccountSubscriptionV0{
			Accounts: accounts,
		},
	}, Regions: regions})
}

// SubscribeAccountsMempoolTransactions subscribes to the mempool transactions of the provided accounts.
func (c *Client) SubscribeAccountsMempoolTransactions(ctx context.Context, accounts, regions []string, ch chan *solana.Transaction) error {
	sub, err := c.NewMempoolStream(accounts, regions)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				var receipt *proto.PendingTxNotification
				receipt, err = sub.Recv()
				if err != nil {
					c.logger.Error().Err(err).Msg("failed to receive mempool notification")
					continue
				}

				for _, transaction := range receipt.Transactions {
					go func(transaction *proto.Packet) {
						var tx *solana.Transaction
						tx, err = pkg.ConvertProtobufPacketToTransaction(transaction)
						if err != nil {
							c.logger.Error().Err(err).Msg("failed to convert protobuf packet to transaction")
							return
						}

						ch <- tx
					}(transaction)
				}
			}
		}
	}()

	return nil
}

func (c *Client) GetRegions(opts ...grpc.CallOption) (*proto.GetRegionsResponse, error) {
	return c.SearcherService.GetRegions(c.Auth.GrpcCtx, &proto.GetRegionsRequest{}, opts...)
}

func (c *Client) GetConnectedLeaders(opts ...grpc.CallOption) (*proto.ConnectedLeadersResponse, error) {
	return c.SearcherService.GetConnectedLeaders(c.Auth.GrpcCtx, &proto.ConnectedLeadersRequest{}, opts...)
}

func (c *Client) GetConnectedLeadersRegioned(regions []string, opts ...grpc.CallOption) (*proto.ConnectedLeadersRegionedResponse, error) {
	return c.SearcherService.GetConnectedLeadersRegioned(c.Auth.GrpcCtx, &proto.ConnectedLeadersRegionedRequest{Regions: regions}, opts...)
}

func (c *Client) GetTipAccounts(opts ...grpc.CallOption) (*proto.GetTipAccountsResponse, error) {
	return c.SearcherService.GetTipAccounts(c.Auth.GrpcCtx, &proto.GetTipAccountsRequest{}, opts...)
}

func (c *Client) GetRandomTipAccount(opts ...grpc.CallOption) (string, error) {
	resp, err := c.GetTipAccounts(opts...)
	if err != nil {
		return "", err
	}

	return resp.Accounts[rand.Intn(len(resp.Accounts))], nil
}

// GetNextScheduledLeader returns the next scheduled leader for the provided regions.
func (c *Client) GetNextScheduledLeader(regions []string, opts ...grpc.CallOption) (*proto.NextScheduledLeaderResponse, error) {
	return c.SearcherService.GetNextScheduledLeader(c.Auth.GrpcCtx, &proto.NextScheduledLeaderRequest{Regions: regions}, opts...)
}

// SubscribeBundleResults subscribes to the results of a bundle.
func (c *Client) SubscribeBundleResults(opts ...grpc.CallOption) (proto.SearcherService_SubscribeBundleResultsClient, error) {
	return c.SearcherService.SubscribeBundleResults(c.Auth.GrpcCtx, &proto.SubscribeBundleResultsRequest{}, opts...)
}

// BroadcastBundle is a function that sends a bundle of packets to the SearcherService.
func (c *Client) BroadcastBundle(transactions []*solana.Transaction, opts ...grpc.CallOption) (*proto.SendBundleResponse, error) {
	packets, err := assemblePackets(transactions)
	if err != nil {
		return nil, err
	}

	return c.SearcherService.SendBundle(c.Auth.GrpcCtx, &proto.SendBundleRequest{Bundle: &proto.Bundle{Packets: packets, Header: nil}}, opts...)
}

// BroadcastBundleWithConfirmation is a function that sends a bundle of packets to the SearcherService and subscribes to the results.
func (c *Client) BroadcastBundleWithConfirmation(transactions []*solana.Transaction, opts ...grpc.CallOption) (*proto.SendBundleResponse, error) {
	bundleSignatures := pkg.BatchExtractSigFromTx(transactions)

	resp, err := c.BroadcastBundle(transactions, opts...)
	if err != nil {
		return nil, err
	}

	subResult, err := c.SubscribeBundleResults()
	if err != nil {
		return nil, err
	}

	for {
		select {
		case <-c.Auth.GrpcCtx.Done():
			return nil, c.Auth.GrpcCtx.Err()
		default:
			var bundleResult *proto.BundleResult
			bundleResult, err = subResult.Recv()
			if err != nil {
				continue
			}

			switch bundleResult.Result.(type) {
			case *proto.BundleResult_Accepted:
				break
			case *proto.BundleResult_Rejected:
				rejected := bundleResult.Result.(*proto.BundleResult_Rejected)
				switch rejected.Rejected.Reason.(type) {
				case *proto.Rejected_SimulationFailure:
					rejection := rejected.Rejected.GetSimulationFailure()
					return nil, NewSimulationFailureError(rejection.TxSignature, rejection.GetMsg())
				case *proto.Rejected_StateAuctionBidRejected:
					rejection := rejected.Rejected.GetStateAuctionBidRejected()
					return nil, NewStateAuctionBidRejectedError(rejection.AuctionId, rejection.SimulatedBidLamports)
				case *proto.Rejected_WinningBatchBidRejected:
					rejection := rejected.Rejected.GetWinningBatchBidRejected()
					return nil, NewWinningBatchBidRejectedError(rejection.AuctionId, rejection.SimulatedBidLamports)
				case *proto.Rejected_InternalError:
					rejection := rejected.Rejected.GetInternalError()
					return nil, NewInternalError(rejection.Msg)
				case *proto.Rejected_DroppedBundle:
					rejection := rejected.Rejected.GetDroppedBundle()
					return nil, NewDroppedBundle(rejection.Msg)
				}
			}
		}

		var statuses *rpc.GetSignatureStatusesResult
		statuses, err = c.RpcConn.GetSignatureStatuses(context.TODO(), false, bundleSignatures...)
		if err != nil {
			return nil, err
		}

		for _, status := range statuses.Value {
			if status.ConfirmationStatus != rpc.ConfirmationStatusProcessed {
				return nil, errors.New("searcher service did not provide bundle status in time")
			}
		}

		return resp, nil
	}
}

func (c *Client) GenerateTipInstruction(tipAmount uint64, from, tipAccount solana.PublicKey) solana.Instruction {
	return system.NewTransferInstruction(tipAmount, from, tipAccount).Build()
}

func (c *Client) GenerateTipRandomAccountInstruction(tipAmount uint64, from solana.PublicKey) (solana.Instruction, error) {
	tipAccount, err := c.GetRandomTipAccount()
	if err != nil {
		return nil, err
	}

	return system.NewTransferInstruction(tipAmount, from, solana.MustPublicKeyFromBase58(tipAccount)).Build(), nil
}

// assemblePackets is a function that converts a slice of transactions to a slice of protobuf packets.
func assemblePackets(transactions []*solana.Transaction) ([]*proto.Packet, error) {
	packets := make([]*proto.Packet, 0, len(transactions))

	for i, tx := range transactions {
		packet, err := pkg.ConvertTransactionToProtobufPacket(tx)
		if err != nil {
			return nil, fmt.Errorf("%d: error converting tx to proto packet [%w]", i, err)
		}

		packets = append(packets, &packet)
	}

	return packets, nil
}
