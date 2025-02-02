package searcher_client

import (
	"context"
	"fmt"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/weeaa/jito-go"
	"github.com/weeaa/jito-go/pb"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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
	assert.True(t, ok, "getting PRIVATE_KEY from .env")

	rpcAddr, ok := os.LookupEnv("JITO_RPC")
	assert.True(t, ok, "getting JITO_RPC from .env")

	client, err := New(
		ctx,
		jito_go.NewYork.BlockEngineURL,
		rpc.New(rpcAddr),
		rpc.New(rpc.MainNetBeta_RPC),
		solana.MustPrivateKeyFromBase58(privKey),
		nil,
	)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	defer client.Close()

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
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
		var resp *jito_pb.GetRegionsResponse
		resp, err = client.GetRegions()
		assert.NoError(t, err)
		assert.Equal(t, jito_go.NewYork.Region, resp.CurrentRegion)
	})

	t.Run("GetConnectedLeaders", func(t *testing.T) {
		leaders, err := client.GetConnectedLeaders()
		assert.NoError(t, err)
		assert.NotNil(t, leaders)
	})

	t.Run("GetConnectedLeadersRegioned", func(t *testing.T) {
		leadersRegioned, err := client.GetConnectedLeadersRegioned(regions)
		assert.NoError(t, err)
		assert.NotNil(t, leadersRegioned)
	})

	t.Run("GetTipAccounts", func(t *testing.T) {
		tipAccounts, err := client.GetTipAccounts()
		assert.NoError(t, err)
		assert.NotNil(t, tipAccounts)
	})

	t.Run("GetNextScheduledLeader", func(t *testing.T) {
		scheduledLeader, err := client.GetNextScheduledLeader(regions)
		assert.NoError(t, err)
		assert.NotNil(t, scheduledLeader)
	})

	t.Run("SubscribeMempoolAccount", func(t *testing.T) {
		t.Skip("skipping test due to rpc method being disabled")

		/*
			accounts := []string{
				"GSE6vfr6vws493G22jfwCU6Zawh3dfvSYXYQqKhFsBwe",
				"xxxxxxqSkjrSvY1igNYjwcw5f9QskeLRKYEmJ1MezhB",
				"93WbteZH6nrWeHd5JT6mJE5VHbWcCTDZzzqqnpZ98V9G",
				"FSHZdx73rEGcS5JXUXvW6h8i4AtsrfPTHcgcbXLVUD3A",
				"Ez2U27TRScksd6q7xoVgX44gX9HAjviN2cdKAL3cFBFE",
				"EAHJNfFDtivTMzKMNXzwAF9RTAeTd4aEYVwLjCiQWY1E",
				"AeF3qRpn7DDuRjHhWmyqmZuZGguHXjsNYCzmNv2ZcuMQ",
				"4CX53LQNwFs3tyRFfwkMxsPV8daao1zCiGQjMMAkKSqx",
				"EePnRqV4Q2VEHp5nPADqeGKcfPMFmnhDW1Ln9LKsTzWQ",
				"6WkVGG2vaKcpgsf5dEYHunZRQHHjZWEUfkWiGxtBGnNg",
				"EMtQTumnZYnv7NSZNGr9WpSSMahkvmNeo9hjhbB9gqFR",
				"4SkEmhCEdLbJxKk6iFzCJ4eR1rLQGHRTs3q8i2PHLbq8",
				"HhuJCViqUewRNXrhwNuXCC7gqp2o1cUhx9a3nqEGkkqt",
				"364kNi4LbCh8iDuNvmbHbPML4N3xbg6msZnaj5dFSJbL",
				"nJMeypdLT1FfSkzNrdCZnVk5HXKiMNRcgCB9Hj554zu",
				"HGEj9nJHdAWJKMHGGHRnhvb3i1XakELSRTn5B4otmAhU",
				"DcpYXJsWBgkV6kck4a7cWBg6B4epPeFRCMZJjxudGKh4",
				"6nsC3UXTCpHq4tXZ1GEeVPg28B9NF8UV4G14dpm9WCUb",
				"B4WGtoLYuF4bF5QUjsnSLFYJVrRhNs8N2NqKqojxxKs8",
				"4sUfdLGg4txSZJdfTkihKLNbM7Xx8WyCmmNqmXc65fjY",
			}

			txCh, errCh, err := client.SubscribeAccountsMempoolTransactions(ctx, accounts, regions)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			for {
				select {
				case <-ctx.Done():
					t.Fatal(ctx.Err())
				case <-errCh:
					t.Fatal(err)
				default:
					tx := <-txCh
					assert.NotNil(t, tx)
					break
				}
			}
		*/
	})

	t.Run("SubscribeMempoolProgram", func(t *testing.T) {
		t.Skip("skipping test due to rpc method being disabled")

		/*
			USDC := "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"
			PENG := "675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8"

			programs := []string{USDC, PENG}

			txCh, errCh, err := client.SubscribeProgramsMempoolTransactions(ctx, programs, regions)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			for {
				select {
				case <-ctx.Done():
					t.Fatal(ctx.Err())
				case <-errCh:
					t.Fatal(err)
				default:
					tx := <-txCh
					assert.NotNil(t, tx)
					break
				}
			}
		*/
	})

	t.Run("SimulateBundle", func(t *testing.T) {
		var pkey string
		pkey, ok = os.LookupEnv("PRIVATE_KEY_WITH_FUNDS")
		assert.True(t, ok, "getting PRIVATE_KEY_WITH_FUNDS from .env")

		var fundedWallet solana.PrivateKey
		fundedWallet, err = solana.PrivateKeyFromBase58(pkey)
		assert.NoError(t, err, "converting private key with funds to type solana.PrivateKey")

		var blockHash *rpc.GetRecentBlockhashResult
		var tx *solana.Transaction

		blockHash, err = client.RpcConn.GetRecentBlockhash(ctx, rpc.CommitmentConfirmed)
		assert.NoError(t, err, "getting recent block hash from RPC")

		tipInst, err := client.GenerateTipRandomAccountInstruction(MINIMUM_TIP, fundedWallet.PublicKey())
		assert.NoError(t, err)
		//assert.IsType(t, solana.Instruction(), tipInst)

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
		assert.NoError(t, err, "creating solana transaction")
		assert.NotNil(t, tx)

		_, err = tx.Sign(
			func(key solana.PublicKey) *solana.PrivateKey {
				if fundedWallet.PublicKey().Equals(key) {
					return &fundedWallet
				}
				return nil
			},
		)
		assert.NoError(t, err, "signing transaction")

		simulateBundleResp, err := client.SimulateBundle(
			ctx,
			SimulateBundleParams{
				EncodedTransactions: []string{tx.MustToBase64()},
			},
			SimulateBundleConfig{
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
			},
		)
		assert.NoError(t, err, "simulating bundle")
		assert.NotNil(t, simulateBundleResp)
	})

	t.Run("GetBundleStatuses_Client", func(t *testing.T) {
		bundleStatuses, err := client.GetBundleStatuses(ctx, []string{bundles[0]})
		assert.NoError(t, err)
		assert.NotNil(t, bundleStatuses)
	})

	t.Run("BatchGetBundleStatuses_Client", func(t *testing.T) {
		bundleStatuses, err := client.BatchGetBundleStatuses(ctx, bundles...)
		assert.NoError(t, err)
		assert.NotNil(t, bundleStatuses)
	})

	t.Run("GetBundleStatuses_Http", func(t *testing.T) {
		bundleStatuses, err := GetBundleStatuses(httpClient, []string{bundles[0]})
		assert.NoError(t, err)
		assert.NotNil(t, bundleStatuses)
	})

	t.Run("BatchGetBundleStatuses_Http", func(t *testing.T) {
		bundleStatuses, err := BatchGetBundleStatuses(httpClient, bundles...)
		assert.NoError(t, err)
		assert.NotNil(t, bundleStatuses)
	})
}

func TestNewNoAuthProxy(t *testing.T) {
	proxyStr, ok := os.LookupEnv("PROXY_URL")
	assert.True(t, ok)
	assert.NotEmpty(t, proxyStr)

	proxylessClient, err := NewNoAuth(context.Background(), jito_go.NewYork.BlockEngineURL, nil, nil, "", nil)
	assert.NoError(t, err)
	assert.NotNil(t, proxylessClient)

	client, err := NewNoAuth(context.Background(), jito_go.NewYork.BlockEngineURL, nil, nil, proxyStr, nil)
	assert.NoError(t, err)
	assert.NotNil(t, client)

	for i := 0; i < 10; i++ { // we use 10 but u get rate limited if u do more than 1 req per sec
		resp, err := client.GetRegions()
		if err != nil {
			if strings.Contains(err.Error(), "Rate limit exceeded") {
				resp, err = proxylessClient.GetRegions()
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				break
			} else {
				t.Errorf("Unexpected error: %v", err)
			}
		} else {
			assert.NotNil(t, resp)
		}
	}
}

func Test_SearcherClientNoAuth(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	rpcAddr, ok := os.LookupEnv("JITO_RPC")
	assert.True(t, ok)

	client, err := NewNoAuth(
		ctx,
		jito_go.NewYork.BlockEngineURL,
		rpc.New(rpcAddr),
		rpc.New(rpc.MainNetBeta_RPC),
		os.Getenv("PROXY_URL"),
		nil,
	)
	assert.NoError(t, err)

	defer client.GrpcConn.Close()

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
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
		var resp *jito_pb.GetRegionsResponse
		resp, err = client.GetRegions()
		assert.NoError(t, err)
		fmt.Println(resp.CurrentRegion)
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

		/*
			accounts := []string{
				"GSE6vfr6vws493G22jfwCU6Zawh3dfvSYXYQqKhFsBwe",
				"xxxxxxqSkjrSvY1igNYjwcw5f9QskeLRKYEmJ1MezhB",
				"93WbteZH6nrWeHd5JT6mJE5VHbWcCTDZzzqqnpZ98V9G",
				"FSHZdx73rEGcS5JXUXvW6h8i4AtsrfPTHcgcbXLVUD3A",
				"Ez2U27TRScksd6q7xoVgX44gX9HAjviN2cdKAL3cFBFE",
				"EAHJNfFDtivTMzKMNXzwAF9RTAeTd4aEYVwLjCiQWY1E",
				"AeF3qRpn7DDuRjHhWmyqmZuZGguHXjsNYCzmNv2ZcuMQ",
				"4CX53LQNwFs3tyRFfwkMxsPV8daao1zCiGQjMMAkKSqx",
				"EePnRqV4Q2VEHp5nPADqeGKcfPMFmnhDW1Ln9LKsTzWQ",
				"6WkVGG2vaKcpgsf5dEYHunZRQHHjZWEUfkWiGxtBGnNg",
				"EMtQTumnZYnv7NSZNGr9WpSSMahkvmNeo9hjhbB9gqFR",
				"4SkEmhCEdLbJxKk6iFzCJ4eR1rLQGHRTs3q8i2PHLbq8",
				"HhuJCViqUewRNXrhwNuXCC7gqp2o1cUhx9a3nqEGkkqt",
				"364kNi4LbCh8iDuNvmbHbPML4N3xbg6msZnaj5dFSJbL",
				"nJMeypdLT1FfSkzNrdCZnVk5HXKiMNRcgCB9Hj554zu",
				"HGEj9nJHdAWJKMHGGHRnhvb3i1XakELSRTn5B4otmAhU",
				"DcpYXJsWBgkV6kck4a7cWBg6B4epPeFRCMZJjxudGKh4",
				"6nsC3UXTCpHq4tXZ1GEeVPg28B9NF8UV4G14dpm9WCUb",
				"B4WGtoLYuF4bF5QUjsnSLFYJVrRhNs8N2NqKqojxxKs8",
				"4sUfdLGg4txSZJdfTkihKLNbM7Xx8WyCmmNqmXc65fjY",
			}

			txCh, errCh, err := client.SubscribeAccountsMempoolTransactions(ctx, accounts, regions)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			for {
				select {
				case <-ctx.Done():
					t.Fatal(ctx.Err())
				case <-errCh:
					t.Fatal(err)
				default:
					tx := <-txCh
					assert.NotNil(t, tx)
					break
				}
			}
		*/
	})

	t.Run("SubscribeMempoolProgram", func(t *testing.T) {
		t.Skip("skipping test due to rpc method being disabled")
		/*
			USDC := "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"
			PENG := "675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8"

			programs := []string{USDC, PENG}

			txCh, errCh, err := client.SubscribeProgramsMempoolTransactions(ctx, programs, regions)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			for {
				select {
				case <-ctx.Done():
					t.Fatal(ctx.Err())
				case <-errCh:
					t.Fatal(err)
				default:
					tx := <-txCh
					assert.NotNil(t, tx)
					break
				}
			}
		*/
	})

	t.Run("SimulateBundle", func(t *testing.T) {
		var pkey string
		pkey, ok = os.LookupEnv("PRIVATE_KEY_WITH_FUNDS")
		assert.True(t, ok, "getting PRIVATE_KEY_WITH_FUNDS from .env")

		var fundedWallet solana.PrivateKey
		fundedWallet, err = solana.PrivateKeyFromBase58(pkey)
		assert.NoError(t, err, "converting private key with funds to type solana.PrivateKey")

		var blockHash *rpc.GetLatestBlockhashResult
		var tx *solana.Transaction

		blockHash, err = client.RpcConn.GetLatestBlockhash(ctx, rpc.CommitmentConfirmed)
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
			SimulateBundleConfig{
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
			},
		)
		assert.NoError(t, err, "simulating bundle")
	})

	t.Run("GetBundleStatuses_Client", func(t *testing.T) {
		_, err := client.GetBundleStatuses(ctx, []string{bundles[0]})
		assert.NoError(t, err)
	})

	t.Run("BatchGetBundleStatuses_Client", func(t *testing.T) {
		_, err = client.BatchGetBundleStatuses(ctx, bundles...)
		assert.NoError(t, err)
	})

	t.Run("GetBundleStatuses_Http", func(t *testing.T) {
		_, err := GetBundleStatuses(httpClient, []string{bundles[0]})
		assert.NoError(t, err)
	})

	t.Run("BatchGetBundleStatuses_Http", func(t *testing.T) {
		_, err := BatchGetBundleStatuses(httpClient, bundles...)
		assert.NoError(t, err)
	})
}

/*
func TestHandleBundleResult(t *testing.T) {
	t.Run("handles jito_pb.BundleResult_Accepted", func(t *testing.T) {
		acceptedBundle := &jito_pb.BundleResult{
			Result: &jito_pb.BundleResult_Accepted{},
		}

		err := handleBundleResult(acceptedBundle)
		require.NoError(t, err)
	})

	t.Run("handles jito_pb.BundleResult_Rejected with SimulationFailure", func(t *testing.T) {
		rejection := &jito_pb.Rejected_SimulationFailure{
			SimulationFailure: &jito_pb.SimulationFailure{
				TxSignature: "signature",
				Msg:         "simulation failed",
			},
		}

		rejectedBundle := &jito_pb.BundleResult{
			Result: &jito_pb.BundleResult_Rejected{
				Rejected: &jito_pb.Rejected{
					Reason: rejection,
				},
			},
		}

		err := handleBundleResult(rejectedBundle)
		require.Error(t, err)
		assert.IsType(t, &SimulationFailureError{}, err)
		assert.Equal(t, "simulation failed", err.Error())
	})

	t.Run("handles jito_pb.BundleResult_Rejected with StateAuctionBidRejected", func(t *testing.T) {
		rejection := &jito_pb.Rejected_StateAuctionBidRejected{
			StateAuctionBidRejected: &jito_pb.StateAuctionBidRejected{
				AuctionId:             "auction123",
				SimulatedBidLamports:  1000,
			},
		}

		rejectedBundle := &jito_pb.BundleResult{
			Result: &jito_pb.BundleResult_Rejected{
				Rejected: &jito_pb.Rejected{
					Reason: rejection,
				},
			},
		}

		err := handleBundleResult(rejectedBundle)
		require.Error(t, err)
		assert.IsType(t, &StateAuctionBidRejectedError{}, err)
	})

	t.Run("handles GetInflightBundlesStatusesResponse with mixed statuses", func(t *testing.T) {
		inflightBundle := &GetInflightBundlesStatusesResponse{
			Result: &InflightBundlesStatusesResult{
				Value: []BundleStatus{
					{Status: "Invalid"},
					{Status: "Pending"},
					{Status: "Failed"},
					{Status: "Landed"},
					{Status: "Unknown"},
				},
			},
		}

		err := handleBundleResult(inflightBundle)
		require.Error(t, err)
		expectedErrMsg := "bundle 0 is invalid; bundle 1 is pending; bundle 2 failed to land; bundle 4 unknown error"
		assert.Equal(t, expectedErrMsg, err.Error())
	})

	t.Run("handles GetInflightBundlesStatusesResponse with no errors", func(t *testing.T) {
		inflightBundle := &GetInflightBundlesStatusesResponse{
			Result: &InflightBundlesStatusesResult{
				Value: []BundleStatus{
					{Status: "Landed"},
				},
			},
		}

		err := handleBundleResult(inflightBundle)
		require.NoError(t, err)
	})
}
*/
