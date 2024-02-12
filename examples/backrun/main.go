package main

import (
	"github.com/gagliardetto/solana-go"
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

	resp, err := client.GetRegions()
	if err != nil {
		log.Fatal(err)
	}

	log.Println(resp)
}
