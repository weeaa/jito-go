package relayer_client

import (
	"github.com/gagliardetto/solana-go"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	jito_go "github.com/weeaa/jito-go"
	"github.com/weeaa/jito-go/proto"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestMain(m *testing.M) {
	_, filename, _, _ := runtime.Caller(0)
	godotenv.Load(filepath.Join(filepath.Dir(filename), "..", "..", "..", "jito-go", ".env"))
	os.Exit(m.Run())
}

func Test_RelayerClient(t *testing.T) {
	privKey, ok := os.LookupEnv("PRIVATE_KEY")
	if !assert.True(t, ok, "getting PRIVATE_KEY from .env") {
		t.FailNow()
	}

	client, err := New(jito_go.Amsterdam.BlockEngineURL, solana.MustPrivateKeyFromBase58(privKey), nil)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	t.Run("GetTpuConfig", func(t *testing.T) {
		var resp *proto.GetTpuConfigsResponse
		resp, err = client.GetTpuConfigs()
		assert.NoError(t, err)
		_ = resp
	})

	t.Run("SubscribePacket", func(t *testing.T) {
		var resp proto.Relayer_SubscribePacketsClient
		resp, err = client.SubscribePackets()
		assert.NoError(t, err)

		var recv *proto.SubscribePacketsResponse
		recv, err = resp.Recv()
		assert.NoError(t, err)

		recv.Header.String()
	})
}
