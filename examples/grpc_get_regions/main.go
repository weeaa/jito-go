package main

import (
	"context"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/joho/godotenv"
	"github.com/weeaa/jito-go"
	"github.com/weeaa/jito-go/clients/searcher_client"
	"log"
	"os"
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

	resp, err := client.GetRegions()
	if err != nil {
		log.Fatal(err)
	}

	log.Println(resp)
}
