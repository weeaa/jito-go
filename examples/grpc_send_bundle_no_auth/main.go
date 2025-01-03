package main

import (
	"context"
	"github.com/davecgh/go-spew/spew"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/joho/godotenv"
	jito_go "github.com/weeaa/jito-go"
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

	ctx := context.Background()

	client, err := searcher_client.NewNoAuth(
		ctx,
		jito_go.NewYork.BlockEngineURL,
		rpc.New(rpcAddr),
		rpc.New(rpc.MainNetBeta_RPC),
		"IP:PORT:USERNAME:PASSWORD", // this is a placeholder value showcasing the proxy format you should use, but you may have this field empty
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	// max per bundle is 5 transactions
	txns := make([]*solana.Transaction, 0, 5)

	block, err := client.RpcConn.GetRecentBlockhash(context.Background(), rpc.CommitmentFinalized)
	if err != nil {
		log.Fatal(err)
	}

	// change w ur keys =)
	from := solana.MustPrivateKeyFromBase58("Tq5gFBU4QG6b6aUYAwi87CUx64iy5tZT1J6nuphN4FXov3UZahMYGSbxLGhb8a9UZ1VvxWB4NzDavSzTorqKCio")
	to := solana.MustPublicKeyFromBase58("BLrQPbKruZgFkNhpdGGrJcZdt1HnfrBLojLYYgnrwNrz")

	tipInst, err := client.GenerateTipRandomAccountInstruction(1000000, from.PublicKey())
	if err != nil {
		log.Fatal(err)
	}

	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			system.NewTransferInstruction(
				10000000,
				from.PublicKey(),
				to,
			).Build(),
			tipInst,
		},
		block.Value.Blockhash,
		solana.TransactionPayer(from.PublicKey()),
	)
	if err != nil {
		log.Fatal(err)
	}

	if _, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if from.PublicKey().Equals(key) {
				return &from
			}
			return nil
		},
	); err != nil {
		log.Fatal(err)
	}

	// debug print
	spew.Dump(tx)

	txns = append(txns, tx)

	resp, err := client.BroadcastBundleWithConfirmation(ctx, txns)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(resp)
}
