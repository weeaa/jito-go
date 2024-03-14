package pkg

import (
	"fmt"
	"github.com/gagliardetto/solana-go"
)

// ExtractSigFromTx extracts the transaction's signature.
func ExtractSigFromTx(tx *solana.Transaction) solana.Signature {
	return tx.Signatures[0]
}

func BatchExtractSigFromTx(txns []*solana.Transaction) []solana.Signature {
	sigs := make([]solana.Signature, 0, len(txns))
	for _, tx := range txns {
		sigs = append(sigs, tx.Signatures[0])
	}
	return sigs
}

func BuildTransactionLinks(txns []solana.Signature, platform Platform) []string {
	txs := make([]string, 0, len(txns))
	for _, tx := range txns {
		txs = append(txs, fmt.Sprintf("%s%s", platform, tx.String()))
	}
	return txs
}

// NewKeyPair creates a new Solana Wallet.
func NewKeyPair(privateKey solana.PrivateKey) *Keypair {
	return &Keypair{PrivateKey: privateKey, PublicKey: privateKey.PublicKey()}
}
