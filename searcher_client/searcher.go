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
	"strings"
	"sync"
	"time"
)

var (
	ErrTransportError    = errors.New("transport error")
	ErrClientError       = errors.New("client error")
	ErrBundleRejection   = errors.New("bundle rejection error")
	ErrInternalError     = errors.New("internal error")
	ErrSimulationFailure = errors.New("simulation failure")
)

type Client struct {
	GrpcConn *grpc.ClientConn
	GrpcCtx  context.Context
	Keypair  *Keypair
	RpcConn  *rpc.Client
	logger   zerolog.Logger

	SearcherService       proto.SearcherServiceClient
	AuthenticationService proto.AuthServiceClient

	Auth Authentication
}

type Authentication struct {
	BearerToken string
	ExpiresAt   int64 // seconds
	ErrChan     chan error
	mu          sync.Mutex
}

type Keypair struct {
	PublicKey  solana.PublicKey
	PrivateKey solana.PrivateKey
}

// NewSearcherClient is a function that creates a new instance of a SearcherClient.
func NewSearcherClient(grpcDialURL string, nodeURL string, privateKey solana.PrivateKey) (*Client, error) {
	if !strings.Contains(nodeURL, "http") {
		return nil, fmt.Errorf("invalid node URL: %s", nodeURL)
	}

	conn, err := grpc.Dial(grpcDialURL, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	if err != nil {
		return nil, err
	}

	authService := proto.NewAuthServiceClient(conn)
	searcherService := proto.NewSearcherServiceClient(conn)

	return &Client{
		GrpcConn:              conn,
		RpcConn:               rpc.New(nodeURL),
		SearcherService:       searcherService,
		AuthenticationService: authService,
		Keypair:               &Keypair{PrivateKey: privateKey, PublicKey: privateKey.PublicKey()},
		logger:                zerolog.New(os.Stdout).With().Timestamp().Str("service", "searcher-client").Logger(),
	}, nil
}

func (c *Client) NewMemPoolStream(accounts, regions []string) (proto.SearcherService_SubscribeMempoolClient, error) {
	return c.SearcherService.SubscribeMempool(c.GrpcCtx, &proto.MempoolSubscription{Msg: &proto.MempoolSubscription_WlaV0Sub{
		WlaV0Sub: &proto.WriteLockedAccountSubscriptionV0{
			Accounts: accounts,
		},
	}, Regions: regions})
}

func (c *Client) SubscribeMemPoolAccounts(ctx context.Context, accounts, regions []string) (chan *solana.Transaction, error) {
	sub, err := c.NewMemPoolStream(accounts, regions)
	if err != nil {
		return nil, err
	}

	notifications := make(chan *solana.Transaction)

	go func(ch chan *solana.Transaction) {
		time.Sleep(200 * time.Millisecond)
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

						notifications <- tx
					}(transaction)
				}
			}
		}
	}(notifications)

	return notifications, nil
}

func (c *Client) GetRegions() (*proto.GetRegionsResponse, error) {
	return c.SearcherService.GetRegions(c.GrpcCtx, &proto.GetRegionsRequest{})
}

func (c *Client) GetConnectedLeaders() (*proto.ConnectedLeadersResponse, error) {
	return c.SearcherService.GetConnectedLeaders(c.GrpcCtx, &proto.ConnectedLeadersRequest{})
}

func (c *Client) GetConnectedLeadersRegioned(regions ...string) (*proto.ConnectedLeadersRegionedResponse, error) {
	return c.SearcherService.GetConnectedLeadersRegioned(c.GrpcCtx, &proto.ConnectedLeadersRegionedRequest{Regions: regions})
}

func (c *Client) GetTipAccounts() (*proto.GetTipAccountsResponse, error) {
	return c.SearcherService.GetTipAccounts(c.GrpcCtx, &proto.GetTipAccountsRequest{})
}

func (c *Client) GetNextScheduledLeader(regions ...string) (*proto.NextScheduledLeaderResponse, error) {
	return c.SearcherService.GetNextScheduledLeader(c.GrpcCtx, &proto.NextScheduledLeaderRequest{Regions: regions})
}

// BroadcastBundle is a function that sends a bundle of packets to the SearcherService.
func (c *Client) BroadcastBundle(transactions []*solana.Transaction) (*proto.SendBundleResponse, error) {
	packets, err := assemblePackets(transactions)
	if err != nil {
		return nil, err
	}

	return c.SearcherService.SendBundle(c.GrpcCtx, &proto.SendBundleRequest{Bundle: &proto.Bundle{Packets: packets, Header: nil}})
}

// BroadcastBundleWithConfirmation is a function that sends a bundle of packets to the SearcherService and subscribes to the results.
func (c *Client) BroadcastBundleWithConfirmation(transactions []*solana.Transaction) (*proto.SendBundleResponse, error) {
	bundleSignatures := make([]solana.Signature, 0, len(transactions))

	for _, tx := range transactions {
		bundleSignatures = append(bundleSignatures, tx.Signatures[0])
	}

	packets, err := assemblePackets(transactions)
	if err != nil {
		return nil, err
	}

	resp, err := c.SearcherService.SendBundle(c.GrpcCtx, &proto.SendBundleRequest{Bundle: &proto.Bundle{Packets: packets, Header: nil}})
	if err != nil {
		return nil, err
	}

	subResult, err := c.SearcherService.SubscribeBundleResults(c.GrpcCtx, &proto.SubscribeBundleResultsRequest{})
	if err != nil {
		return nil, err
	}

	for {
		select {
		case <-c.GrpcCtx.Done():
			return nil, c.GrpcCtx.Err()
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
				case *proto.Rejected_WinningBatchBidRejected:
					return nil, ErrClientError
				case *proto.Rejected_DroppedBundle:
					return nil, ErrTransportError
				case *proto.Rejected_SimulationFailure:
					return nil, ErrSimulationFailure
				case *proto.Rejected_StateAuctionBidRejected:
					return nil, ErrBundleRejection
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
