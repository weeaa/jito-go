package searcher_client

import (
	"context"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/weeaa/jito-go"
	"github.com/weeaa/jito-go/pkg"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func init() {
	_, filename, _, _ := runtime.Caller(0)
	godotenv.Load(filepath.Join(filepath.Dir(filename), "..", "..", "..", "jito-go", ".env"))
}

func TestSearcherClient(t *testing.T) {

	privKey, ok := os.LookupEnv("PRIVATE_KEY")
	assert.True(t, ok, "getting PRIVATE_KEY from .env")

	rpcAddr, ok := os.LookupEnv("JITO_RPC")
	assert.True(t, ok, "getting JITO_RPC from .env")

	client, err := NewSearcherClient(
		jito_go.NewYork.BlockEngineURL,
		rpc.New(rpcAddr),
		solana.MustPrivateKeyFromBase58(privKey),
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}

	regions := []string{
		jito_go.Amsterdam.Region,
		jito_go.NewYork.Region,
		jito_go.Frankfurt.Region,
		jito_go.Tokyo.Region,
	}

	t.Run("GetRegions", func(t *testing.T) {
		_, err = client.GetRegions()
		assert.NoError(t, err)
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

		ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
		defer cancel()

		USDC := ("EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v")
		PENG := ("675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8")

		payload := &SubscribeAccountsMempoolTransactionsPayload{
			Ctx:      context.TODO(),
			Accounts: []string{USDC, PENG},
			Regions:  regions,
			TxCh:     make(chan *solana.Transaction),
			ErrCh:    make(chan error),
		}

		err = client.SubscribeAccountsMempoolTransactions(payload)
		assert.NoError(t, err)

		for {
			select {
			case <-ctx.Done():
				t.Fatal()
			default:
				tx := <-payload.TxCh
				assert.Contains(t, pkg.ExtractSigFromTx(tx).String(), "")
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

		blockHash, err = client.RpcConn.GetRecentBlockhash(context.TODO(), rpc.CommitmentConfirmed)
		assert.NoError(t, err, "getting recebt blockhash from RPC")

		tx, err = solana.NewTransaction(
			[]solana.Instruction{
				system.NewTransferInstruction(
					10,
					fundedWallet.PublicKey(),
					solana.MustPublicKeyFromBase58("A6njahNqC6qKde6YtbHdr1MZsB5KY9aKfzTY1cj8jU3v"),
				).Build(),
			},
			blockHash.Value.Blockhash,
			solana.TransactionPayer(fundedWallet.PublicKey()),
		)
		assert.NoError(t, err, "creating solana transaction")

		_, err = tx.Sign(
			func(key solana.PublicKey) *solana.PrivateKey {
				if fundedWallet.PublicKey().Equals(key) {
					return &fundedWallet
				}
				return nil
			},
		)
		assert.NoError(t, err, "signing transaction")

		_, err = client.SimulateBundle(
			context.TODO(),
			SimulateBundleParams{
				EncodedTransactions: []string{tx.MustToBase64()},
			},
			config,
		)
		assert.NoError(t, err, "simulating bundle")
	})
}
