package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gagliardetto/solana-go"
	"github.com/gorilla/websocket"
	"io"
	"math/big"
	"net/http"
	"net/url"
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

// BuildTransactionLinks generates a list of URLs for the provided transactions,
// linking each transaction signature to the appropriate blockchain explorer based on the platform.
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

// SubscribeTipStream establishes a connection to the Jito websocket and receives TipStreamInfo.
func SubscribeTipStream(ctx context.Context) (<-chan []*TipStreamInfo, <-chan error, error) {
	dialer := websocket.Dialer{}
	ch := make(chan []*TipStreamInfo)
	chErr := make(chan error)

	conn, _, err := dialer.Dial(tipStreamURL, nil)
	if err != nil {
		return nil, nil, err
	}

	go func() {
		defer conn.Close()
		defer close(ch)
		defer close(chErr)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				var r []*TipStreamInfo
				if err = conn.ReadJSON(&r); err != nil {
					chErr <- err
					time.Sleep(500 * time.Millisecond)
					continue
				}

				ch <- r
			}
		}
	}()

	return ch, chErr, nil
}

// GenerateKeypair creates a new Solana Keypair.
func GenerateKeypair() *Keypair {
	wallet := solana.NewWallet()
	return &Keypair{PublicKey: wallet.PublicKey(), PrivateKey: wallet.PrivateKey}
}

// LamportsToSol converts the given amount of lamports to SOL by dividing by the
// number of lamports per SOL.
func LamportsToSol(lamports *big.Float) *big.Float {
	return new(big.Float).Quo(lamports, new(big.Float).SetUint64(solana.LAMPORTS_PER_SOL))
}

// StrSliceToByteSlice converts a string array to a byte array.
func StrSliceToByteSlice(s []string) [][]byte {
	byteSlice := make([][]byte, 0, len(s))
	for _, b := range s {
		byteSlice = append(byteSlice, []byte(b))
	}
	return byteSlice
}

// TxToStr converts type solana.Transaction to string.
func TxToStr(txns []*solana.Transaction) []string {
	txnsStr := make([]string, len(txns))
	for _, tx := range txns {
		txnsStr = append(txnsStr, tx.String())
	}
	return txnsStr
}

func GetTipInformation(client *http.Client) (*[]TipStreamInfo, error) {
	if client == nil {
		client = &http.Client{}
	}

	req := &http.Request{
		Method: http.MethodGet,
		URL:    &url.URL{Scheme: "https", Host: "bundles.jito.wtf", Path: "/api/v1/bundles/tip_floor"},
		Header: http.Header{
			"User-Agent": {"jito-go"},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response status, got %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tipInfo []TipStreamInfo
	err = json.Unmarshal(body, &tipInfo)
	return &tipInfo, err
}
