package searcher_client

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/rs/zerolog"
	"github.com/weeaa/jito-go/pkg"
	"github.com/weeaa/jito-go/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"os"
	"sync"
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

type Authentication struct {
	BearerToken string
	ExpiresAt   int64 // seconds
	ErrChan     chan error
	mu          sync.Mutex
}

// NewSearcherClient is a function that creates a new instance of a SearcherClient.
func NewSearcherClient(grpcDialURL string, rpcClient *rpc.Client, privateKey solana.PrivateKey) (*Client, error) {
	conn, err := grpc.Dial(grpcDialURL, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
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

// NewMemPoolStream creates a new MemPool subscription.
func (c *Client) NewMemPoolStream(accounts, regions []string) (proto.SearcherService_SubscribeMempoolClient, error) {
	return c.SearcherService.SubscribeMempool(c.Auth.GrpcCtx, &proto.MempoolSubscription{Msg: &proto.MempoolSubscription_WlaV0Sub{
		WlaV0Sub: &proto.WriteLockedAccountSubscriptionV0{
			Accounts: accounts,
		},
	}, Regions: regions})
}

// SubscribeMemPoolAccounts creates a new MemPool subscription and sends transactions to the provided channel.
func (c *Client) SubscribeMemPoolAccounts(ctx context.Context, accounts, regions []string, ch chan *solana.Transaction) error {
	sub, err := c.NewMemPoolStream(accounts, regions)
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

func (c *Client) GetRegions() (*proto.GetRegionsResponse, error) {
	return c.SearcherService.GetRegions(c.Auth.GrpcCtx, &proto.GetRegionsRequest{})
}

func (c *Client) GetConnectedLeaders() (*proto.ConnectedLeadersResponse, error) {
	return c.SearcherService.GetConnectedLeaders(c.Auth.GrpcCtx, &proto.ConnectedLeadersRequest{})
}

func (c *Client) GetConnectedLeadersRegioned(regions ...string) (*proto.ConnectedLeadersRegionedResponse, error) {
	return c.SearcherService.GetConnectedLeadersRegioned(c.Auth.GrpcCtx, &proto.ConnectedLeadersRegionedRequest{Regions: regions})
}

func (c *Client) GetTipAccounts() (*proto.GetTipAccountsResponse, error) {
	return c.SearcherService.GetTipAccounts(c.Auth.GrpcCtx, &proto.GetTipAccountsRequest{})
}

func (c *Client) GetNextScheduledLeader(regions ...string) (*proto.NextScheduledLeaderResponse, error) {
	return c.SearcherService.GetNextScheduledLeader(c.Auth.GrpcCtx, &proto.NextScheduledLeaderRequest{Regions: regions})
}

func (c *Client) SubscribeBundleResults() (proto.SearcherService_SubscribeBundleResultsClient, error) {
	return c.SearcherService.SubscribeBundleResults(c.Auth.GrpcCtx, &proto.SubscribeBundleResultsRequest{})
}

// BroadcastBundle is a function that sends a bundle of packets to the SearcherService.
func (c *Client) BroadcastBundle(transactions []*solana.Transaction) (*proto.SendBundleResponse, error) {
	packets, err := assemblePackets(transactions)
	if err != nil {
		return nil, err
	}

	return c.SearcherService.SendBundle(c.Auth.GrpcCtx, &proto.SendBundleRequest{Bundle: &proto.Bundle{Packets: packets, Header: nil}})
}

// BroadcastBundleWithConfirmation is a function that sends a bundle of packets to the SearcherService and subscribes to the results.
func (c *Client) BroadcastBundleWithConfirmation(transactions []*solana.Transaction, timeout uint64) (*proto.SendBundleResponse, error) {
	bundleSignatures := pkg.BatchExtractSigFromTx(transactions)

	resp, err := c.BroadcastBundle(transactions)
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
