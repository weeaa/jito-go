package pkg

import (
	"fmt"
	"github.com/gagliardetto/solana-go"
)

type Platform string

var (
	SolanaFM       Platform = "https://solana.fm/tx/"
	Solscan        Platform = "https://solscan.io/tx/"
	SolanaExplorer Platform = "https://explorer.solana.com/tx/"
	SolanaBeach    Platform = "https://solanabeach.io/transaction/"
	XRAY           Platform = "https://xray.helius.xyz/tx/"
)

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
