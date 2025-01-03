# Jito Go SDK
[![GoDoc](https://pkg.go.dev/badge/github.com/weeaa/jito-go?status.svg)](https://pkg.go.dev/github.com/weeaa/jito-go?tab=doc)
[![Go Report Card](https://goreportcard.com/badge/github.com/weeaa/jito-go)](https://goreportcard.com/report/github.com/weeaa/jito-go)
[![License](https://img.shields.io/badge/license-Apache_2.0-crimson)](https://opensource.org/license/apache-2-0)

This library contains tooling to interact with **[Jito Labs](https://www.jito.wtf/)** software.

We currently use [gagliardetto/solana-go](https://github.com/gagliardetto/solana-go) to interact with Solana.  PRs and contributions are welcome.

![jitolabs](https://github.com/weeaa/jito-go/assets/108926252/5751416c-333b-412e-8f3f-f26b2839be98)

## ❇️ Contents
- [Support](#-support)
- [Features](#-features)
- [ToDo](#-todo)
- [Methods](#-methods)
- [Installing](#-installing)
- [Update Proto](#-update-proto)
- [Keypair Authentication](#-keypair-authentication)
- [Examples](#-examples)
- [Common errors](#-common-errors)
- [Disclaimer](#-disclaimer)
- [License](#-license)

## 🛟 Support
If my work has been useful in building your for-profit services/infra/bots/etc, consider donating at
`ACPc147BD5SE7Rh2HKLED7nLWJyiNNHM8ruyGNHqcE8U` (Solana address).

## ✨ Features
- [x] Searcher
- [x] Block Engine
- [x] Relayer
- [x] [Geyser](https://github.com/weeaa/goyser) 🐳
- [ ] Shredstream
- [x] JSON RPC API
- [x] Others
- [x] API

## 📋 ToDo
- tbd

## 📡 Methods
- `💀* methods which are deprecated by Jito`
- `most methods have wrappers for ease of use`
- `both gRPC and JSON-RPC methods are supported`


- [x] **Searcher**
  - `SubscribeMempoolAccounts` 💀
  - `SubscribeMempoolPrograms` 💀
  - `GetNextScheduledLeader`
  - `GetRegions`
  - `GetConnectedLeaders`
  - `GetConnectedLeadersRegioned`
  - `GetTipAccounts`
  - `SimulateBundle`
  - `SendBundle` (gRPC & JSON-RPC)
  - `SendBundleWithConfirmation` (gRPC & JSON-RPC)
  - `SubscribeBundleResults`
  - `GetBundleStatuses` (gRPC & JSON-RPC)
- [x] **Block Engine**
  - Validator
    - `SubscribePackets`
    - `SubscribeBundles`
    - `GetBlockBuilderFeeInfo`
  - Relayer
    - `SubscribeAccountsOfInterest`
    - `SubscribeProgramsOfInterest`
    - `StartExpiringPacketStream`
- [ ] **ShredStream**
  - `SendHeartbeat`
- [x] **Others** (pkg/util.go & pkg/convert.go)
  - `SubscribeTipStream`
  - Converting functions
- [x] **API** (api/api.go)
  - `RetrieveRecentBundles`
  - `RetrieveBundleIDfromTransactionSignature`

## 💾 Installing

Go 1.22.0 or higher.
```shell
go get github.com/weeaa/jito-go@latest
```

If you want to run tests:

1. Install [Task](https://taskfile.dev/installation/).
2. Initialize your `.env` file by running `task install:<os>` (darwin/linux/windows).
3. Run tests with `task test`.

## 🔄 Update Proto
🚨 **Beware that due to duplicate types from jito proto files it will cause an error when building, this issue has been submitted to the team.** 🚨

In order to get latest proto generated files, make sure you have protobuf & its plugins installed!

Using Homebrew (macOS).
```shell
brew install protobuf
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

Edit perms.
```shell
chmod +x ./scripts/generate_protos.sh
```

Run script.
```shell
./scripts/generate_protos.sh
```

## 🔑 Keypair Authentication
**[The following isn't mandatory anymore for Searcher access](https://docs.google.com/document/d/e/2PACX-1vRZoiYWNvIdX4r6lf-8E5E0l8SEPKeXXRYRcviwQJjmizeJkeQ_YM4IWGQne-C_8_lFFXv-z6yI6y4K/pub)**.

To access Jito MEV functionalities, you'll need a whitelisted Public Key obtained from a fresh KeyPair; submit your Public Key [here](https://web.miniextensions.com/WV3gZjFwqNqITsMufIEp).
In order to generate a new KeyPair, you can use the following function `GenerateKeypair()` from the `/pkg` package.

## 💻 Examples

- [gRPC Get Regions](./examples/grpc_get_regions)
- [gRPC Send Bundle](./examples/grpc_send_bundle)
- [gRPC Send Bundle No Auth](./examples/grpc_send_bundle_no_auth)
- [gRPC Simulate Bundle](./examples/grpc_simulate_bundle)

## 🐥 Common errors

- `The supplied pubkey is not authorized to generate a token.`

The public key you supplied is not whitelisted by Jito, you may either use `NewNoAuth` instead of `New`
to authenticate, or apply for whitelist [here](https://web.miniextensions.com/WV3gZjFwqNqITsMufIEp).

## 🚨 Disclaimer

**This library is not affiliated with Jito Labs**. It is a community project and is not officially supported by Jito Labs. Use at your own risk.

## 📃 License

[Apache-2.0 License](https://github.com/weeaa/jito-go/blob/main/LICENSE).
