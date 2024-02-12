package main

import (
	"context"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/weeaa/jito-go"
	"github.com/weeaa/jito-go/searcher_client"
	"log"
	"os"
)

func main() {
	client, err := searcher_client.NewSearcherClient(
		jito_go.NewYork.BlockEngineURL,
		rpc.New(rpc.MainNetBeta_RPC),
		solana.MustPrivateKeyFromBase58(os.Getenv("PRIVATE_KEY")),
	)
	if err != nil {
		log.Fatal(err)
	}

	// max per bundle is 5 transactions
	txns := make([]*solana.Transaction, 0, 5)

	block, err := client.RpcConn.GetRecentBlockhash(context.TODO(), rpc.CommitmentFinalized)
	if err != nil {
		log.Fatal(err)
	}

	from := solana.MustPrivateKeyFromBase58("Tq5gFBU4QG6b6aUYAwi87CUx64iy5tZT1J6nuphN4FXov3UZahMYGSbxLGhb8a9UZ1VvxWB4NzDavSzTorqKCio")
	to := solana.MustPublicKeyFromBase58("BLrQPbKruZgFkNhpdGGrJcZdt1HnfrBLojLYYgnrwNrz")

	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			system.NewTransferInstruction(
				10000,
				from.PublicKey(),
				to,
			).Build(),
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

	txns = append(txns, tx)

	resp, err := client.BroadcastBundleWithConfirmation(txns, 100)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(resp)
}
