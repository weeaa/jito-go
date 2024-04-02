package geyser_client

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/weeaa/jito-go/proto"
	"os"
	"testing"
)

func Test_GeyserClient(t *testing.T) {
	rpcAddr, ok := os.LookupEnv("GEYSER_RPC")
	assert.True(t, ok, "getting JITO_RPC from .env")

	client, err := NewGeyserClient(
		context.Background(),
		rpcAddr,
		nil,
	)
	assert.NoError(t, err)

	ctx := context.Background()

	t.Run("SubscribeBlockUpdates", func(t *testing.T) {
		sub, err := client.SubscribeBlockUpdates()
		assert.NoError(t, err)

		ch := make(chan *proto.BlockUpdate)
		client.OnBlockUpdates(ctx, sub, ch)

		block := <-ch
		assert.NotNil(t, block.BlockHeight)
	})
}
