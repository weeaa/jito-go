package pkg

import (
	"github.com/gagliardetto/solana-go"
)

type Keypair struct {
	PublicKey  solana.PublicKey
	PrivateKey solana.PrivateKey
}

func NewKeyPair(privateKey solana.PrivateKey) *Keypair {
	return &Keypair{PrivateKey: privateKey, PublicKey: privateKey.PublicKey()}
}
