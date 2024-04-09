package searcher_client

import (
	"context"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/weeaa/jito-go"
	"github.com/weeaa/jito-go/proto"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	_, filename, _, _ := runtime.Caller(0)
	godotenv.Load(filepath.Join(filepath.Dir(filename), "..", "..", "..", "jito-go", ".env"))
	os.Exit(m.Run())
}

func Test_SearcherClient(t *testing.T) {
	privKey, ok := os.LookupEnv("PRIVATE_KEY")
	if !assert.True(t, ok, "getting PRIVATE_KEY from .env") {
		t.FailNow()
	}

	rpcAddr, ok := os.LookupEnv("JITO_RPC")
	if !assert.True(t, ok, "getting JITO_RPC from .env") {
		t.FailNow()
	}

	client, err := New(
		jito_go.NewYork.BlockEngineURL,
		rpc.New(rpcAddr),
		solana.MustPrivateKeyFromBase58(privKey),
		nil,
	)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	regions := []string{
		jito_go.Amsterdam.Region,
		jito_go.NewYork.Region,
		jito_go.Frankfurt.Region,
		jito_go.Tokyo.Region,
	}

	t.Run("GetRegions", func(t *testing.T) {
		var resp *proto.GetRegionsResponse
		resp, err = client.GetRegions()
		assert.NoError(t, err)
		assert.Equal(t, jito_go.NewYork.Region, resp.CurrentRegion)
	})

	t.Run("GetConnectedLeaders", func(t *testing.T) {
		_, err = client.GetConnectedLeaders()
		assert.NoError(t, err)
	})

	t.Run("GetConnectedLeadersRegioned", func(t *testing.T) {
		_, err = client.GetConnectedLeadersRegioned(regions)
		assert.NoError(t, err)
	})

	t.Run("GetTipAccounts", func(t *testing.T) {
		_, err = client.GetTipAccounts()
		assert.NoError(t, err)
	})

	t.Run("GetNextScheduledLeader", func(t *testing.T) {
		_, err = client.GetNextScheduledLeader(regions)
		assert.NoError(t, err)
	})

	t.Run("SubscribeMempoolAccount", func(t *testing.T) {
		//t.Skip("skipping test due to rpc method being disabled")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		_ = ctx

	})

	t.Run("SubscribeMempoolProgram", func(t *testing.T) {
		//t.Skip("skipping test due to rpc method being disabled")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		USDC := ("EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v")
		PENG := ("675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8")

		payload := &SubscribeAccountsMempoolTransactionsPayload{
			Ctx:      context.Background(),
			Accounts: []string{USDC, PENG},
			Regions:  regions,
			TxCh:     make(chan *solana.Transaction),
			ErrCh:    make(chan error),
		}

		err = client.SubscribeAccountsMempoolTransactions(payload)
		if !assert.NoError(t, err) {
			t.FailNow()
		}

		for {
			select {
			case <-ctx.Done():
				t.Fatal()
			default:
				tx := <-payload.TxCh
				assert.NotNil(t, tx)
				break
			}
		}
	})

	t.Run("SimulateBundle", func(t *testing.T) {
		var pkey string
		pkey, ok = os.LookupEnv("PRIVATE_KEY_WITH_FUNDS")
		assert.True(t, ok, "getting PRIVATE_KEY_WITH_FUNDS from .env")

		var fundedWallet solana.PrivateKey
		fundedWallet, err = solana.PrivateKeyFromBase58(pkey)
		assert.NoError(t, err, "converting private key with funds to type solana.PrivateKey")

		config := SimulateBundleConfig{
			PreExecutionAccountsConfigs: []ExecutionAccounts{
				{
					Encoding:  "base64",
					Addresses: []string{"3vjULHsUbX4J2nXZJQQSHkTHoBqhedvHQPDNaAgT9dwG"},
				},
			},
			PostExecutionAccountsConfigs: []ExecutionAccounts{
				{
					Encoding:  "base64",
					Addresses: []string{"3vjULHsUbX4J2nXZJQQSHkTHoBqhedvHQPDNaAgT9dwG"},
				},
			},
		}

		var blockHash *rpc.GetRecentBlockhashResult
		var tx *solana.Transaction

		blockHash, err = client.RpcConn.GetRecentBlockhash(context.Background(), rpc.CommitmentConfirmed)
		if !assert.NoError(t, err, "getting recent block hash from RPC") {
			t.FailNow()
		}

		tx, err = solana.NewTransaction(
			[]solana.Instruction{
				system.NewTransferInstruction(
					10000,
					fundedWallet.PublicKey(),
					solana.MustPublicKeyFromBase58("A6njahNqC6qKde6YtbHdr1MZsB5KY9aKfzTY1cj8jU3v"),
				).Build(),
			},
			blockHash.Value.Blockhash,
			solana.TransactionPayer(fundedWallet.PublicKey()),
		)
		if !assert.NoError(t, err, "creating solana transaction") {
			t.FailNow()
		}

		_, err = tx.Sign(
			func(key solana.PublicKey) *solana.PrivateKey {
				if fundedWallet.PublicKey().Equals(key) {
					return &fundedWallet
				}
				return nil
			},
		)
		if !assert.NoError(t, err, "signing transaction") {
			t.FailNow()
		}

		_, err = client.SimulateBundle(
			context.Background(),
			SimulateBundleParams{
				EncodedTransactions: []string{tx.MustToBase64()},
			},
			config,
		)
		assert.NoError(t, err, "simulating bundle")
	})
}
