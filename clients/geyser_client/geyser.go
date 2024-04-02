package geyser_client

import (
	"context"
	"crypto/tls"
	"fmt"
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

	ErrCh chan error
}

// NewGeyserClient creates a new RPC client and connects to the provided endpoint. A Geyser RPC URL is required.
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
		GrpcConn:  conn,
		Geyser:    geyserClient,
		ErrCh:     make(chan error),
		GrpcErrCh: grpcErrChan,
		Ctx:       ctx,
	}, nil
}

func (c *Client) SubscribePartialAccountUpdates(opts ...grpc.CallOption) (proto.Geyser_SubscribePartialAccountUpdatesClient, error) {
	return c.Geyser.SubscribePartialAccountUpdates(c.Ctx, &proto.SubscribePartialAccountUpdatesRequest{}, opts...)
}

func (c *Client) OnPartialAccountUpdates(ctx context.Context, sub proto.Geyser_SubscribePartialAccountUpdatesClient, ch chan *proto.PartialAccountUpdate) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				subInfo, err := sub.Recv()
				if err != nil {
					c.ErrCh <- fmt.Errorf("error OnPartialAccountUpdates: %w", err)
					continue
				}
				ch <- subInfo.GetPartialAccountUpdate()
			}
		}
	}()
}

func (c *Client) SubscribeBlockUpdates(opts ...grpc.CallOption) (proto.Geyser_SubscribeBlockUpdatesClient, error) {
	return c.Geyser.SubscribeBlockUpdates(c.Ctx, &proto.SubscribeBlockUpdatesRequest{}, opts...)
}

func (c *Client) OnBlockUpdates(ctx context.Context, sub proto.Geyser_SubscribeBlockUpdatesClient, ch chan *proto.BlockUpdate) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				subInfo, err := sub.Recv()
				if err != nil {
					c.ErrCh <- fmt.Errorf("error OnBlockUpdates: %w", err)
					continue
				}
				ch <- subInfo.BlockUpdate
			}
		}
	}()
}

func (c *Client) SubscribeAccountUpdates(accounts []string, opts ...grpc.CallOption) (proto.Geyser_SubscribeAccountUpdatesClient, error) {
	return c.Geyser.SubscribeAccountUpdates(c.Ctx, &proto.SubscribeAccountUpdatesRequest{Accounts: strSliceToByteSlice(accounts)}, opts...)
}

func (c *Client) OnAccountUpdates(ctx context.Context, sub proto.Geyser_SubscribeAccountUpdatesClient, ch chan *proto.AccountUpdate) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				subInfo, err := sub.Recv()
				if err != nil {
					c.ErrCh <- fmt.Errorf("error OnAccountUpdates: %w", err)
					continue
				}
				ch <- subInfo.AccountUpdate
			}
		}
	}()
}

func (c *Client) SubscribeProgramUpdates(programs []string, opts ...grpc.CallOption) (proto.Geyser_SubscribeProgramUpdatesClient, error) {
	return c.Geyser.SubscribeProgramUpdates(c.Ctx, &proto.SubscribeProgramsUpdatesRequest{Programs: strSliceToByteSlice(programs)}, opts...)
}

func (c *Client) OnProgramUpdate(ctx context.Context, sub proto.Geyser_SubscribeProgramUpdatesClient, ch chan *proto.AccountUpdate) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				subInfo, err := sub.Recv()
				if err != nil {
					c.ErrCh <- fmt.Errorf("error OnProgramUpdate: %w", err)
					continue
				}
				ch <- subInfo.AccountUpdate
			}
		}
	}()
}

func (c *Client) SubscribeTransactionUpdates(opts ...grpc.CallOption) (proto.Geyser_SubscribeTransactionUpdatesClient, error) {
	return c.Geyser.SubscribeTransactionUpdates(c.Ctx, &proto.SubscribeTransactionUpdatesRequest{}, opts...)
}

func (c *Client) OnTransactionUpdates(ctx context.Context, sub proto.Geyser_SubscribeTransactionUpdatesClient, ch chan *proto.TransactionUpdate) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				subInfo, err := sub.Recv()
				if err != nil {
					c.ErrCh <- fmt.Errorf("error OnTransactionUpdates: %w", err)
					continue
				}
				ch <- subInfo.Transaction
			}
		}
	}()
}

func (c *Client) SubscribeSlotUpdates(opts ...grpc.CallOption) (proto.Geyser_SubscribeSlotUpdatesClient, error) {
	return c.Geyser.SubscribeSlotUpdates(c.Ctx, &proto.SubscribeSlotUpdateRequest{}, opts...)
}

func (c *Client) OnSlotUpdates(ctx context.Context, sub proto.Geyser_SubscribeSlotUpdatesClient, ch chan *proto.SlotUpdate) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				subInfo, err := sub.Recv()
				if err != nil {
					c.ErrCh <- fmt.Errorf("error OnSlotUpdates: %w", err)
					continue
				}
				ch <- subInfo.SlotUpdate
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
