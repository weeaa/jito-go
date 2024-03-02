# Jito Go SDK
[![GoDoc](https://pkg.go.dev/badge/github.com/weeaa/jito-go?status.svg)](https://pkg.go.dev/github.com/weeaa/jito-go?tab=doc)
[![Go Report Card](https://goreportcard.com/badge/github.com/weeaa/jito-go)](https://goreportcard.com/report/github.com/weeaa/jito-go)
[![License](https://img.shields.io/badge/license-Apache_2.0-crimson)](https://opensource.org/license/apache-2-0)

This library contains tooling to interact with **[Jito Labs](https://www.jito.wtf/)** MEV software. ⚠️ Work in progress. ⚠️

<p align="center">
  <img src="https://media.discordapp.net/attachments/1180285583273246720/1206572835372269598/image.png?ex=65dc7f84&is=65ca0a84&hm=2793d93eb12ef7becff685b3d56d2e64d9ef61f892751412222a98b8d5fc135d&=&format=webp&quality=lossless&width=2204&height=1028" />
</p>

## ❇️ Contents
- [Features](#features)
- [Installing](#installing)
- [RPC Methods](#rpc-methods)
- [Keypair Authentication](#keypair-authentication)
- [Examples](#examples)
- [Disclaimer](#disclaimer)
- [License](#-license)

## ✨ Features
- [x] Searcher
- [x] Block Engine
- [x] Relayer
- [ ] ShredStream (WIP, help welcome 😊)
- [x] Geyser

## 📡 RPC Methods
- [x] **Searcher**
  - `SubscribeMempoolAccounts`
  - `SubscribeMempoolPrograms`
  - `GetNextScheduledLeader`
  - `GetRegions`
  - `GetConnectedLeaders`
  - `GetConnectedLeadersRegioned`
  - `GetTipAccounts`
  - `SimulateBundle`
  - `SendBundle`
  - `SendBundleWithConfirmation`
  - `SubscribeBundleResults`
- [x] **Block Engine**
  - Validator
    - `SubscribePackets`
    - `SubscribeBundles`
    - `GetBlockBuilderFeeInfo`
  - Relayer
    - `SubscribeAccountsOfInterest`
    - `SubscribeProgramsOfInterest`
    - `StartExpiringPacketStream`
- [x] **Geyser**
  - `SubscribePartialAccountUpdates`
  - `SubscribeBlockUpdates`
  - `SubscribeAccountUpdates`
  - `SubscribeProgramUpdates`
  - `SubscribeTransactionUpdates`
  - `SubscribeSlotUpdates`
- [ ] **ShredStream**

## 💾 Installing

```shell
go get github.com/weeaa/jito-go@latest
```

## 🔑 Keypair Authentication

To utilize the features of Jito MEV, you are required to generate a new Solana KeyPair and submit the Public Key [here](https://web.miniextensions.com/WV3gZjFwqNqITsMufIEp).
You can create a new KeyPair by following the instructions provided in the code snippet below.
```go
package main

import (
	"encoding/json"
	"github.com/gagliardetto/solana-go"
	"github.com/weeaa/jito-go/pkg"
	"log"
	"os"
)

func main() {
	wallet := solana.NewWallet()

	keypair := pkg.Keypair{
		PrivateKey: wallet.PrivateKey,
		PublicKey:  wallet.PublicKey(),
	}

	data, err := json.Marshal(keypair)
	if err != nil {
		log.Fatalf("failed to encode keypair as JSON: %v", err)
	}

	if err = os.WriteFile("keypair.json", data, 0600); err != nil {
		log.Fatal(err)
	}

	log.Println("Successfully generated and saved new keypair.")
}
```

## 💻 Examples

### `Send Bundle`
```go
package main

import (
  "context"
  "github.com/gagliardetto/solana-go"
  "github.com/gagliardetto/solana-go/programs/system"
  "github.com/gagliardetto/solana-go/rpc"
  "github.com/weeaa/jito-go"
  "github.com/weeaa/jito-go/clients/searcher_client"
  "log"
  "os"
)

func main() {
  client, err := searcher_client.NewSearcherClient(
    jito_go.NewYork.BlockEngineURL,
    rpc.New(rpc.MainNetBeta_RPC),
    solana.MustPrivateKeyFromBase58(os.Getenv("PRIVATE_KEY")),
	nil,
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
```
### `Subscribe to MemPool Transactions [Accounts]`
```go
package main

import (
	"context"
    "github.com/gagliardetto/solana-go"
    "github.com/gagliardetto/solana-go/rpc"
    "github.com/weeaa/jito-go"
    "github.com/weeaa/jito-go/clients/searcher_client"
    "log"
    "os"
)

func main() {
  client, err := searcher_client.NewSearcherClient(
    jito_go.NewYork.BlockEngineURL,
    rpc.New(rpc.MainNetBeta_RPC),
    solana.MustPrivateKeyFromBase58(os.Getenv("PRIVATE_KEY")),
    nil,
  )
  if err != nil {
    log.Fatal(err)
  }

  txSub := make(chan *solana.Transaction)
  regions := []string{jito_go.NewYork.Region}
  accounts := []string{
    "GuHvDyajPfQpHrg2oCWmArYHrZn2ynxAkSxAPFn9ht1g",
    "4EKP9SRfykwQxDvrPq7jUwdkkc93Wd4JGCbBgwapeJhs",
    "Hn98nGFGfZwJPjd4bk3uAX5pYHJe5VqtrtMhU54LNNhe",
    "MuUEAu5tFfEMhaFGoz66jYTFBUHZrwfn3KWimXLNft2",
    "CSGeQFoSuN56QZqf9WLqEEkWhRFt6QksTjMDLm68PZKA",
  }

  payload := &searcher_client.SubscribeAccountsMempoolTransactionsPayload{
    Ctx:      context.TODO(),
    Accounts: accounts,
    Regions:  regions,
    TxCh:     make(chan *solana.Transaction),
    ErrCh:    make(chan error),
  }
  
  if err = client.SubscribeAccountsMempoolTransactions(payload); err != nil {
    log.Fatal(err)
  }

  for tx := range txSub {
    log.Println(tx)
  }
}
```

### `Get Regions`
```go
package main

import (
    "github.com/gagliardetto/solana-go"
    "github.com/gagliardetto/solana-go/rpc"
    "github.com/weeaa/jito-go"
    "github.com/weeaa/jito-go/clients/searcher_client"
    "log"
    "os"
)

func main() {
  client, err := searcher_client.NewSearcherClient(
    jito_go.NewYork.BlockEngineURL,
    rpc.New(rpc.MainNetBeta_RPC),
    solana.MustPrivateKeyFromBase58(os.Getenv("PRIVATE_KEY")),
	nil,
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

## 🚨 Disclaimer

**This library is not affiliated with Jito Labs**. It is a community project and is not officially supported by Jito Labs. Use at your own risk.

## 📃 License

[Apache-2.0 License](https://github.com/weeaa/jito-go/blob/main/LICENSE).
