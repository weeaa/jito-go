# Jito Go SDK
[![GoDoc](https://pkg.go.dev/badge/github.com/weeaa/jito-go?status.svg)](https://pkg.go.dev/github.com/weeaa/jito-go?tab=doc)
[![Go Report Card](https://goreportcard.com/badge/github.com/weeaa/jito-go)](https://goreportcard.com/report/github.com/weeaa/jito-go)
[![License](https://img.shields.io/badge/license-Apache_2.0-crimson)](https://opensource.org/license/apache-2-0)

This library contains tooling to interact with **[Jito Labs](https://www.jito.wtf/)** MEV software. ⚠️ Work in progress. ⚠️

We currently use [gagliardetto/solana-go](https://github.com/gagliardetto/solana-go) to interact with Solana.  PRs and contributions are welcome.

![jitolabs](https://github.com/weeaa/jito-go/assets/108926252/5751416c-333b-412e-8f3f-f26b2839be98)

## ❇️ Contents
- [Features](#-features)
- [RPC Methods](#-rpc-methods)
- [Installing](#-installing)
- [Keypair Authentication](#-keypair-authentication)
- [Examples](#-examples)
- [Disclaimer](#-disclaimer)
- [Support](#-support)
- [License](#-license)

## ✨ Features
- [x] Searcher
- [x] Block Engine
- [x] Relayer
- [ ] ShredStream (WIP, help welcome 😊)
- [x] Geyser

## 📡 RPC Methods
`🤡* methods which are deprecated by Jito due to malicious use`
- [x] **Searcher**
  - `SubscribeMempoolAccounts` 🤡
  - `SubscribeMempoolPrograms` 🤡
  - `GetNextScheduledLeader`
  - `GetRegions`
  - `GetConnectedLeaders`
  - `GetConnectedLeadersRegioned`
  - `GetTipAccounts`
  - `SimulateBundle`
  - `SendBundle`
  - `SendBundleWithConfirmation`
  - `SubscribeBundleResults`
  - `GetBundleStatuses` (gRPC & HTTP)
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
- [x] **Others** (pkg/util.go)
  - `SubscribeTipStream`

## 💾 Installing

Go 1.21 or higher
```shell
go get github.com/weeaa/jito-go@latest
```

If you want to run tests:

1. Install [Task](https://taskfile.dev/installation/).
2. Initialize your `.env` file by running `task install:<os>` (darwin/linux/windows).
3. Run tests with `task test`.

## 🔑 Keypair Authentication
To access Jito MEV functionalities, you'll need a whitelisted Public Key obtained from a fresh KeyPair; submit your Public Key [here](https://web.miniextensions.com/WV3gZjFwqNqITsMufIEp).
In order to generate a new KeyPair, you can use the following function `GenerateKeypair()` from the `/pkg` package.

## 💻 Examples

### `Send Bundle`
```go
package main

import (
  "context"
  "github.com/davecgh/go-spew/spew"
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

  client, err := searcher_client.New(
    jito_go.NewYork.BlockEngineURL,
    rpc.New(rpcAddr),
    rpc.New(rpc.MainNetBeta_RPC),
    key,
    nil,
  )
  if err != nil {
    log.Fatal(err)
  }

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
```
### `Subscribe to MemPool Transactions [Accounts]`
```go
package main

import (
  "context"
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

  client, err := searcher_client.New(
    jito_go.NewYork.BlockEngineURL,
    rpc.New(rpcAddr),
    rpc.New(rpc.MainNetBeta_RPC),
    key,
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
    Ctx:      context.Background(),
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

  client, err := searcher_client.New(
    jito_go.NewYork.BlockEngineURL,
    rpc.New(rpcAddr),
    rpc.New(rpc.MainNetBeta_RPC),
    key,
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

### `Subscribe Block Updates [Geyser]`
```go
package main

import (
	"context"
	"github.com/weeaa/jito-go/clients/geyser_client"
	"github.com/weeaa/jito-go/proto"
	"log"
)

func main() {
	rpcAddr := "myGeyserRpcNodeURL"
	
	// establish conn to geyser node...
	client, err := geyser_client.New(context.Background(), rpcAddr, nil)
	if err != nil {
		log.Fatal(err)
	}
	
	// create block update sub
	sub, err := client.SubscribeBlockUpdates()
	if err != nil {
		log.Fatal(err)
	}
	// create chan to receive block updates
	ch := make(chan *proto.BlockUpdate)
	// feed the chan with block updates 
	client.OnBlockUpdates(context.Background(), sub, ch)
	
	// loop to read new block updates from chan
	for {
		block := <-ch
		log.Println(block)
	}
}
```
## 🚨 Disclaimer

**This library is not affiliated with Jito Labs**. It is a community project and is not officially supported by Jito Labs. Use at your own risk.

## 🛟 Support
If my work has been useful in building your for-profit services/infra/bots/etc, consider donating at
`EcrHvqa5Vh4NhR3bitRZVrdcUGr1Z3o6bXHz7xgBU2FB` (SOL).

## 📃 License

[Apache-2.0 License](https://github.com/weeaa/jito-go/blob/main/LICENSE).
