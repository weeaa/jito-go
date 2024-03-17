package geyser_client

import (
	"context"
	"crypto/tls"
	"github.com/weeaa/jito-go/pkg"
	"github.com/weeaa/jito-go/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Client struct {
	GrpcConn  *grpc.ClientConn
	GrpcErrCh chan error
	Ctx       context.Context

	Geyser proto.GeyserClient

	BlockUpdateChannel          chan *proto.BlockUpdate
	SlotUpdateChannel           chan *proto.SlotUpdate
	TransactionUpdateChannel    chan *proto.TransactionUpdate
	AccountUpdateChannel        chan *proto.AccountUpdate
	PartialAccountUpdateChannel chan *proto.PartialAccountUpdate

	ErrCh chan error
}

func NewGeyserClient(ctx context.Context, grpcDialURL string, tlsConfig *tls.Config, opts ...grpc.DialOption) (*Client, error) {
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

	geyserClient := proto.NewGeyserClient(conn)

	return &Client{
		GrpcConn:                 conn,
		Geyser:                   geyserClient,
		BlockUpdateChannel:       make(chan *proto.BlockUpdate),
		SlotUpdateChannel:        make(chan *proto.SlotUpdate),
		TransactionUpdateChannel: make(chan *proto.TransactionUpdate),
		AccountUpdateChannel:     make(chan *proto.AccountUpdate),
		ErrCh:                    make(chan error),
		GrpcErrCh:                grpcErrChan,
		Ctx:                      ctx,
	}, nil
}

func (c *Client) SubscribePartialAccountUpdates(opts ...grpc.CallOption) (proto.Geyser_SubscribePartialAccountUpdatesClient, error) {
	return c.Geyser.SubscribePartialAccountUpdates(c.Ctx, &proto.SubscribePartialAccountUpdatesRequest{}, opts...)
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
					c.ErrCh <- err
					continue
				}
				c.PartialAccountUpdateChannel <- subInfo.GetPartialAccountUpdate()
			}
		}
	}()
}

func (c *Client) SubscribeBlockUpdates(opts ...grpc.CallOption) (proto.Geyser_SubscribeBlockUpdatesClient, error) {
	return c.Geyser.SubscribeBlockUpdates(c.Ctx, &proto.SubscribeBlockUpdatesRequest{}, opts...)
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
					c.ErrCh <- err
					continue
				}
				c.BlockUpdateChannel <- subInfo.BlockUpdate
			}
		}
	}()
}

func (c *Client) SubscribeAccountUpdates(accounts []string, opts ...grpc.CallOption) (proto.Geyser_SubscribeAccountUpdatesClient, error) {
	return c.Geyser.SubscribeAccountUpdates(c.Ctx, &proto.SubscribeAccountUpdatesRequest{Accounts: strSliceToByteSlice(accounts)}, opts...)
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
					c.ErrCh <- err
					continue
				}
				c.AccountUpdateChannel <- subInfo.AccountUpdate
			}
		}
	}()
}

func (c *Client) SubscribeProgramUpdates(programs []string, opts ...grpc.CallOption) (proto.Geyser_SubscribeProgramUpdatesClient, error) {
	return c.Geyser.SubscribeProgramUpdates(c.Ctx, &proto.SubscribeProgramsUpdatesRequest{Programs: strSliceToByteSlice(programs)}, opts...)
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
					c.ErrCh <- err
					continue
				}
				c.AccountUpdateChannel <- subInfo.AccountUpdate
			}
		}
	}()
}

func (c *Client) SubscribeTransactionUpdates(opts ...grpc.CallOption) (proto.Geyser_SubscribeTransactionUpdatesClient, error) {
	return c.Geyser.SubscribeTransactionUpdates(c.Ctx, &proto.SubscribeTransactionUpdatesRequest{}, opts...)
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
					c.ErrCh <- err
					continue
				}
				c.TransactionUpdateChannel <- subInfo.Transaction
			}
		}
	}()
}

func (c *Client) SubscribeSlotUpdates(opts ...grpc.CallOption) (proto.Geyser_SubscribeSlotUpdatesClient, error) {
	return c.Geyser.SubscribeSlotUpdates(c.Ctx, &proto.SubscribeSlotUpdateRequest{}, opts...)
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
					c.ErrCh <- err
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
