<p align="center">
  <img src="https://media.discordapp.net/attachments/1180285583273246720/1206572835372269598/image.png?ex=65dc7f84&is=65ca0a84&hm=2793d93eb12ef7becff685b3d56d2e64d9ef61f892751412222a98b8d5fc135d&=&format=webp&quality=lossless&width=2204&height=1028" />
</p>

# Jito SDK library for Go
[![GoDoc](https://pkg.go.dev/badge/github.com/weeaa/jito-go?status.svg)](https://pkg.go.dev/github.com/weeaa/jito-go?tab=doc)
[![Go Report Card](https://goreportcard.com/badge/github.com/weeaa/jito-go)](https://goreportcard.com/report/github.com/weeaa/jito-go)

## About
This library contains tooling to interact with **[Jito Labs](https://www.jito.wtf/)**.

## Contents
- [Features](#features)
- [Installing](#installing)
- [RPC Methods](#rpc-methods)
- [Keypair Authentication](#keypair-authentication)
- [Example](#example)
- [Disclaimer](#disclaimer)

## Features
- [x] Searcher
- [x] Block Engine
- [ ] Relayer
- [ ] ShredStream
- [ ] Geyser

## RPC Methods
- Searcher
  - SubscribeMempool
  - GetNextScheduledLeader
  - GetRegions
  - SendBundle
  - GetConnectedLeaders
  - GetConnectedLeadersRegioned
  - SubscribeBundleResults
  - GetTipAccounts
- Block Engine
  - Validator
    - SubscribePackets
    - SubscribeBundles
    - GetBlockBuilderFeeInfo
  - Relayer
    - SubscribeAccountsOfInterest
    - SubscribeProgramsOfInterest
    - StartExpiringPacketStream

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
```

## Disclaimer

This library is not affiliated with Jito Labs. It is a community project and is not officially supported by Jito Labs. Use at your own risk.