package pkg

import (
	"context"
	"fmt"
	"github.com/gagliardetto/solana-go"
	"github.com/gorilla/websocket"
	"time"
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

// NewKeyPair creates a Keypair from a private key.
func NewKeyPair(privateKey solana.PrivateKey) *Keypair {
	return &Keypair{PrivateKey: privateKey, PublicKey: privateKey.PublicKey()}
}

const tipStreamURL = "ws://bundles-api-rest.jito.wtf/api/v1/bundles/tip_stream"

type TipStreamInfo struct {
	Time                        time.Time `json:"time"`
	LandedTips25ThPercentile    float64   `json:"landed_tips_25th_percentile"`
	LandedTips50ThPercentile    float64   `json:"landed_tips_50th_percentile"`
	LandedTips75ThPercentile    float64   `json:"landed_tips_75th_percentile"`
	LandedTips95ThPercentile    float64   `json:"landed_tips_95th_percentile"`
	LandedTips99ThPercentile    float64   `json:"landed_tips_99th_percentile"`
	EmaLandedTips50ThPercentile float64   `json:"ema_landed_tips_50th_percentile"`
}

// SubscribeTipStream establishes a connection to the Jito websocket and receives TipStreamInfo.
func SubscribeTipStream(ctx context.Context) (chan *TipStreamInfo, error) {
	dialer := websocket.Dialer{}
	ch := make(chan *TipStreamInfo)

	conn, _, err := dialer.Dial(tipStreamURL, nil)
	if err != nil {
		return nil, err
	}

	for {
		select {
		case <-ctx.Done():
			return nil, nil
		default:
			var r *TipStreamInfo
			if err = conn.ReadJSON(r); err != nil {
				continue
			}

			ch <- r
		}
	}
}

// GenerateKeypair creates a new Solana Keypair.
func GenerateKeypair() *Keypair {
	wallet := solana.NewWallet()
	return &Keypair{PublicKey: wallet.PublicKey(), PrivateKey: wallet.PrivateKey}
}
