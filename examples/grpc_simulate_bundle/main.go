package main

import (
	"context"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/joho/godotenv"
	"github.com/weeaa/jito-go"
	"github.com/weeaa/jito-go/clients/searcher_client"
	"log"
	"os"
	"time"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	rpcAddr, ok := os.LookupEnv("JITO_RPC")
	if !ok {
		log.Fatal("JITO_RPC could not be found in .env")
	}

	privateKey, ok := os.LookupEnv("PRIVATE_KEY")
	if !ok {
		log.Fatal("PRIVATE_KEY could not be found in .env")
	}

	key, err := solana.PrivateKeyFromBase58(privateKey)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	client, err := searcher_client.New(
		ctx,
		jito_go.NewYork.BlockEngineURL,
		rpc.New(rpcAddr),
		rpc.New(rpc.MainNetBeta_RPC),
		key,
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	var pkey string
	pkey, ok = os.LookupEnv("PRIVATE_KEY_WITH_FUNDS")
	if !ok {
		log.Fatal("could not get PRIVATE_KEY from .env")
	}

	var fundedWallet solana.PrivateKey
	fundedWallet, err = solana.PrivateKeyFromBase58(pkey)
	if err != nil {
		log.Fatal(err)
	}

	var blockHash *rpc.GetRecentBlockhashResult
	var tx *solana.Transaction

	blockHash, err = client.RpcConn.GetRecentBlockhash(ctx, rpc.CommitmentConfirmed)
	if err != nil {
		log.Fatal(err)
	}

	var tipInst solana.Instruction
	tipInst, err = client.GenerateTipRandomAccountInstruction(1000000, fundedWallet.PublicKey())
	if err != nil {
		log.Fatal(err)
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
	if err != nil {
		log.Fatal(err)
	}

	_, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if fundedWallet.PublicKey().Equals(key) {
				return &fundedWallet
			}
			return nil
		},
	)

	resp, err := client.SimulateBundle(
		ctx,
		searcher_client.SimulateBundleParams{
			EncodedTransactions: []string{tx.MustToBase64()},
		},
		searcher_client.SimulateBundleConfig{
			PreExecutionAccountsConfigs: []searcher_client.ExecutionAccounts{
				{
					Encoding:  "base64",
					Addresses: []string{"3vjULHsUbX4J2nXZJQQSHkTHoBqhedvHQPDNaAgT9dwG"},
				},
			},
			PostExecutionAccountsConfigs: []searcher_client.ExecutionAccounts{
				{
					Encoding:  "base64",
					Addresses: []string{"3vjULHsUbX4J2nXZJQQSHkTHoBqhedvHQPDNaAgT9dwG"},
				},
			},
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(resp)
}
