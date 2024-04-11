package blockengine_client

import (
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
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

func Test_BlockEngineRelayerClient(t *testing.T) {
	privKey, ok := os.LookupEnv("PRIVATE_KEY")
	if !assert.True(t, ok, "getting PRIVATE_KEY from .env") {
		t.FailNow()
	}

	rpcAddr, ok := os.LookupEnv("JITO_RPC")
	if !assert.True(t, ok, "getting JITO_RPC from .env") {
		t.FailNow()
	}

	if assert.Equal(t, "", rpcAddr) {
		t.FailNow()
	}

	rpcClient := rpc.New(rpcAddr)
	privateKey, err := solana.PrivateKeyFromBase58(privKey)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	validator, err := NewValidator(jito_go.Amsterdam.BlockEngineURL, rpcClient, privateKey, nil)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	t.Run("Validator_SubscribePackets", func(t *testing.T) {
		var sub proto.BlockEngineValidator_SubscribePacketsClient
		sub, err = validator.SubscribePackets()
		if !assert.NoError(t, err) {
			t.FailNow()
		}

		var resp *proto.SubscribePacketsResponse
		resp, err = sub.Recv()
		assert.NoError(t, err)

		assert.NotNil(t, resp.Batch)
	})

	t.Run("Validator_SubscribeBundles", func(t *testing.T) {
		var sub proto.BlockEngineValidator_SubscribeBundlesClient
		sub, err = validator.SubscribeBundles()
		if !assert.NoError(t, err) {
			t.FailNow()
		}

		var resp *proto.SubscribeBundlesResponse
		resp, err = sub.Recv()
		if !assert.NoError(t, err) {
			t.FailNow()
		}

		for _, bundle := range resp.Bundles {
			assert.NotNil(t, bundle.Bundle)
		}
	})

	t.Run("Validator_GetBlockBuilderFeeInfo", func(t *testing.T) {
		var resp *proto.BlockBuilderFeeInfoResponse
		resp, err = validator.GetBlockBuilderFeeInfo()
		if !assert.NoError(t, err) {
			t.FailNow()
		}

		resp.String()
	})
}

func Test_BlockEngineValidatorClient(t *testing.T) {
	privKey, ok := os.LookupEnv("PRIVATE_KEY")
	if !assert.True(t, ok, "getting PRIVATE_KEY from .env") {
		t.FailNow()
	}

	rpcAddr, ok := os.LookupEnv("JITO_RPC")
	if !assert.True(t, ok, "getting JITO_RPC from .env") {
		t.FailNow()
	}

	rpcClient := rpc.New(rpcAddr)
	privateKey, err := solana.PrivateKeyFromBase58(privKey)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	relayer, err := NewRelayer(jito_go.NewYork.RelayerURL, rpcClient, privateKey, nil)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	t.Run("SubscribeAccountsOfInterest", func(t *testing.T) {
		var sub proto.BlockEngineRelayer_SubscribeAccountsOfInterestClient
		sub, err = relayer.SubscribeAccountsOfInterest()
		if !assert.NoError(t, err) {
			t.FailNow()
		}

		var resp *proto.AccountsOfInterestUpdate
		resp, err = sub.Recv()
		if !assert.NoError(t, err) {
			t.FailNow()
		}

		for _, account := range resp.Accounts {
			assert.NotEqual(t, "", account)
		}
	})

	t.Run("StartExpiringPacketStream", func(t *testing.T) {
		var sub proto.BlockEngineRelayer_StartExpiringPacketStreamClient
		sub, err = relayer.StartExpiringPacketStream()
		if !assert.NoError(t, err) {
			t.FailNow()
		}

		var resp *proto.StartExpiringPacketStreamResponse
		resp, err = sub.Recv()
		if !assert.NoError(t, err) {
			t.FailNow()
		}

		assert.NotNil(t, resp.Heartbeat)
	})

	t.Run("", func(t *testing.T) {
		sub, err := relayer.SubscribeProgramsOfInterest()
		if !assert.NoError(t, err) {
			t.FailNow()
		}

		resp, err := sub.Recv()
		if !assert.NoError(t, err) {
			t.FailNow()
		}

		for _, program := range resp.Programs {
			assert.NotEqual(t, "", program)
		}
	})
}
