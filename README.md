![image](https://media.discordapp.net/attachments/689063280358064158/1206006337226539008/image.png?ex=65da6fed&is=65c7faed&hm=02f9daef5065a40bb2718d1073684d4db431378b4bcaf5fd19ee677b84f3f832&=&format=webp&quality=lossless&width=1224&height=440)

# Jito SDK library for Go

## About
This library contains tooling to interact with **[Jito Labs](https://www.jito.wtf/)**.

## Contents
- [About](#about)
- [Features](#features)
- [Installing](#installing)
- [Keypair Authentication](#keypair-authentication)
- [Disclaimer](#disclaimer)

## Features
- [x] Send Bundles
- [x] Subscribe to MemPool

## Installing
```bash
go get github.com/weeaa/jito-go@latest
```

## Keypair Authentication

To get access to the block engine, please generate a new solana keypair and submit the public key [here](https://web.miniextensions.com/WV3gZjFwqNqITsMufIEp).

## Example

```go
package main

import (
	"context"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	go_jito "github.com/weeaa/jito-go"
	"github.com/weeaa/jito-go/proto"
	"github.com/weeaa/jito-go/searcher_client"
	"log"
	"os"
)

func main() {
	client, err := searcher_client.NewSearcherClient(
		go_jito.NewYork.BlockEngineURL,
		rpc.TestNet_RPC,
		solana.MustPrivateKeyFromBase58(os.Getenv("PRIVATE_KEY")),
	)
	if err != nil {
		log.Fatal(err)
	}
	if err = client.AuthenticateAndRefresh(context.Background(), proto.Role_SEARCHER); err != nil {
		log.Fatal(err)
	}
	resp, err := client.SearcherService.GetRegions(client.GrpcCtx, &proto.GetRegionsRequest{})
	if err != nil {
		log.Fatal(err)
	}
	log.Println(resp)
}
```

## Disclaimer

This library is not affiliated with Jito Labs. It is a community project and is not officially supported by Jito Labs. Use at your own risk.