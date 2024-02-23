package geyser_client

import (
	"context"
	"crypto/tls"
	"github.com/rs/zerolog"
	"github.com/weeaa/jito-go/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"os"
)

type Client struct {
	GrpcConn *grpc.ClientConn
	logger   zerolog.Logger
	Ctx      context.Context

	Geyser proto.GeyserClient

	BlockUpdateChannel          chan *proto.BlockUpdate
	SlotUpdateChannel           chan *proto.SlotUpdate
	TransactionUpdateChannel    chan *proto.TransactionUpdate
	AccountUpdateChannel        chan *proto.AccountUpdate
	PartialAccountUpdateChannel chan *proto.PartialAccountUpdate
	ErrChan                     chan error
}

func NewGeyserClient(ctx context.Context, grpcDialURL string, tlsConfig *tls.Config, opts ...grpc.DialOption) (*Client, error) {
	if tlsConfig != nil {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	}

	conn, err := grpc.Dial(grpcDialURL, opts...)
	if err != nil {
		return nil, err
	}

	geyserClient := proto.NewGeyserClient(conn)

	return &Client{
		GrpcConn:                 conn,
		Geyser:                   geyserClient,
		BlockUpdateChannel:       make(chan *proto.BlockUpdate),
		SlotUpdateChannel:        make(chan *proto.SlotUpdate),
		TransactionUpdateChannel: make(chan *proto.TransactionUpdate),
		AccountUpdateChannel:     make(chan *proto.AccountUpdate),
		ErrChan:                  make(chan error),
		logger:                   zerolog.New(os.Stdout).With().Timestamp().Str("service", "geyser-client").Logger(),
		Ctx:                      ctx,
	}, nil
}

func (c *Client) SubscribePartialAccountUpdates() (proto.Geyser_SubscribePartialAccountUpdatesClient, error) {
	return c.Geyser.SubscribePartialAccountUpdates(c.Ctx, &proto.SubscribePartialAccountUpdatesRequest{})
}

func (c *Client) HandlePartialAccountUpdates(ctx context.Context, sub proto.Geyser_SubscribePartialAccountUpdatesClient) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				subInfo, err := sub.Recv()
				if err != nil {
					c.ErrChan <- err
					continue
				}
				c.PartialAccountUpdateChannel <- subInfo.GetPartialAccountUpdate()
			}
		}
	}()
}

func (c *Client) SubscribeBlockUpdates() (proto.Geyser_SubscribeBlockUpdatesClient, error) {
	return c.Geyser.SubscribeBlockUpdates(c.Ctx, &proto.SubscribeBlockUpdatesRequest{})
}

func (c *Client) HandleBlockUpdates(ctx context.Context, sub proto.Geyser_SubscribeBlockUpdatesClient) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				subInfo, err := sub.Recv()
				if err != nil {
					c.ErrChan <- err
					continue
				}
				c.BlockUpdateChannel <- subInfo.BlockUpdate
			}
		}
	}()
}

func (c *Client) SubscribeAccountUpdates(accounts []string) (proto.Geyser_SubscribeAccountUpdatesClient, error) {
	return c.Geyser.SubscribeAccountUpdates(c.Ctx, &proto.SubscribeAccountUpdatesRequest{Accounts: strSliceToByteSlice(accounts)})
}

func (c *Client) OnAccountUpdates(ctx context.Context, sub proto.Geyser_SubscribeAccountUpdatesClient) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				subInfo, err := sub.Recv()
				if err != nil {
					c.ErrChan <- err
					continue
				}
				c.AccountUpdateChannel <- subInfo.AccountUpdate
			}
		}
	}()
}

func (c *Client) SubscribeProgramUpdates(programs []string) (proto.Geyser_SubscribeProgramUpdatesClient, error) {
	return c.Geyser.SubscribeProgramUpdates(c.Ctx, &proto.SubscribeProgramsUpdatesRequest{Programs: strSliceToByteSlice(programs)})
}

func (c *Client) OnProgramUpdate(ctx context.Context, sub proto.Geyser_SubscribeProgramUpdatesClient) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				subInfo, err := sub.Recv()
				if err != nil {
					c.ErrChan <- err
					continue
				}
				c.AccountUpdateChannel <- subInfo.AccountUpdate
			}
		}
	}()
}

func (c *Client) SubscribeTransactionUpdates() (proto.Geyser_SubscribeTransactionUpdatesClient, error) {
	return c.Geyser.SubscribeTransactionUpdates(c.Ctx, &proto.SubscribeTransactionUpdatesRequest{})
}

func (c *Client) HandleTransactionUpdates(ctx context.Context, sub proto.Geyser_SubscribeTransactionUpdatesClient) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				subInfo, err := sub.Recv()
				if err != nil {
					c.ErrChan <- err
					continue
				}
				c.TransactionUpdateChannel <- subInfo.Transaction
			}
		}
	}()
}

func (c *Client) SubscribeSlotUpdates() (proto.Geyser_SubscribeSlotUpdatesClient, error) {
	return c.Geyser.SubscribeSlotUpdates(c.Ctx, &proto.SubscribeSlotUpdateRequest{})
}

func (c *Client) HandleSlotUpdates(ctx context.Context, sub proto.Geyser_SubscribeSlotUpdatesClient) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				subInfo, err := sub.Recv()
				if err != nil {
					c.ErrChan <- err
					continue
				}
				c.SlotUpdateChannel <- subInfo.SlotUpdate
			}
		}
	}()
}

func strSliceToByteSlice(s []string) [][]byte {
	byteSlice := make([][]byte, 0, len(s))
	for _, b := range s {
		byteSlice = append(byteSlice, []byte(b))
	}
	return byteSlice
}
