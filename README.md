# Jito Go SDK
[![GoDoc](https://pkg.go.dev/badge/github.com/weeaa/jito-go?status.svg)](https://pkg.go.dev/github.com/weeaa/jito-go?tab=doc)
[![Go Report Card](https://goreportcard.com/badge/github.com/weeaa/jito-go)](https://goreportcard.com/report/github.com/weeaa/jito-go)
[![License](https://img.shields.io/badge/license-Apache_2.0-crimson)](https://opensource.org/license/apache-2-0)

This library contains tooling to interact with **[Jito Labs](https://www.jito.wtf/)** MEV software. ⚠️ Work in progress. ⚠️

PRs and contributions are welcome.

![jitolabs](https://github.com/weeaa/jito-go/assets/108926252/5751416c-333b-412e-8f3f-f26b2839be98)

## ❇️ Contents
- [Features](#-features)
- [Installing](#-installing)
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
`🤡* methods which are disabled by Jito due to malicious use`
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

If you want to run tests:

1. Install [Task](https://taskfile.dev/installation/).
2. Initialize your `.env` file by running `task install:<os>` (darwin/linux/windows).
3. Run tests with `task test`.

## 🔑 Keypair Authentication
To access Jito MEV functionalities, you'll need a whitelisted Public Key obtained from a fresh KeyPair; submit your Public Key [here](https://web.miniextensions.com/WV3gZjFwqNqITsMufIEp).
In order to generate a new KeyPair, you can use the following function `GenerateWallet()` from the `/pkg` package.

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

## 🛟 Support
If my work has been useful in building your for-profit services/infra/bots/etc, consider donating at
`EcrHvqa5Vh4NhR3bitRZVrdcUGr1Z3o6bXHz7xgBU2FB` (SOL).

## 📃 License

[Apache-2.0 License](https://github.com/weeaa/jito-go/blob/main/LICENSE).
