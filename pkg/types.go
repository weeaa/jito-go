package pkg

import (
	"github.com/gagliardetto/solana-go"
)

type Keypair struct {
	PublicKey  solana.PublicKey
	PrivateKey solana.PrivateKey
}

type Platform string

var (
	SolanaFM       Platform = "https://solana.fm/tx/"
	Solscan        Platform = "https://solscan.io/tx/"
	SolanaExplorer Platform = "https://explorer.solana.com/tx/"
	SolanaBeach    Platform = "https://solanabeach.io/transaction/"
	XRAY           Platform = "https://xray.helius.xyz/tx/"
)
