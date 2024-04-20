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
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

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
		rpc.New(rpc.MainNetBeta_RPC),
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

	bundles := []string{
		"fc5f0b63c8a2e75193394311d7063e904ce4cf31a63ad6c5809277c1a68ec935",
		"d3925b4f9c6dc5112b89d07c854d55f6c52c2a528ff85a42bc3be9e91a80f290",
		"bdf8b0e5cad979ed71c40fe1a6b8d2589cc046d96fc43ff020a98c6b732b9ae9",
		"3821951e175b8186dfe0fbf05b15837edd1538659692721cea8c27368670859b",
		"e41a99b23554f26bc1c85552e027fe7ad95133c6a741a4920b6442f1e6f98ca7",
		"a20fc66cf0155f2c803d1f37bab54775a9110840e21d50369d8196886df97798",
		"52c0375ca1acee04d697faf2abe5d7e499856fcc15457673cba8c5e0673f398b",
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
		t.Skip("skipping test due to rpc method being disabled")
	})

	t.Run("SubscribeMempoolProgram", func(t *testing.T) {
		t.Skip("skipping test due to rpc method being disabled")
		USDC := ("EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v")
		PENG := ("675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8")

		payload := &SubscribeAccountsMempoolTransactionsPayload{
			Ctx:      ctx,
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

		blockHash, err = client.RpcConn.GetRecentBlockhash(ctx, rpc.CommitmentConfirmed)
		if !assert.NoError(t, err, "getting recent block hash from RPC") {
			t.FailNow()
		}

		var tipInst solana.Instruction
		tipInst, err = client.GenerateTipRandomAccountInstruction(1000000, fundedWallet.PublicKey())
		if !assert.NoError(t, err) {
			t.FailNow()
		}

		tx, err = solana.NewTransaction(
			[]solana.Instruction{
				system.NewTransferInstruction(
					10000000,
					fundedWallet.PublicKey(),
					solana.MustPublicKeyFromBase58("A6njahNqC6qKde6YtbHdr1MZsB5KY9aKfzTY1cj8jU3v"),
				).Build(),
				tipInst,
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
			ctx,
			SimulateBundleParams{
				EncodedTransactions: []string{tx.MustToBase64()},
			},
			config,
		)
		assert.NoError(t, err, "simulating bundle")
	})

	t.Run("GetBundleStatuses_Client", func(t *testing.T) {
		_, err := client.GetBundleStatuses(ctx, []string{bundles[0]})
		if !assert.NoError(t, err) {
			t.FailNow()
		}
	})

	t.Run("BatchGetBundleStatuses_Client", func(t *testing.T) {
		_, err := client.BatchGetBundleStatuses(ctx, bundles...)
		if !assert.NoError(t, err) {
			t.FailNow()
		}
	})

	t.Run("GetBundleStatuses_Http", func(t *testing.T) {
		_, err := GetBundleStatuses([]string{bundles[0]})
		if !assert.NoError(t, err) {
			t.FailNow()
		}
	})

	t.Run("BatchGetBundleStatuses_Http", func(t *testing.T) {
		_, err := BatchGetBundleStatuses(bundles...)
		if !assert.NoError(t, err) {
			t.FailNow()
		}
	})
}
