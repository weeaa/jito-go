package searcher_client

import (
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/weeaa/jito-go"
	"os"
	"testing"
)

func TestSendBundle(t *testing.T) {
	client, err := NewSearcherClient(
		go_jito.NewYork.BlockEngineURL,
		rpc.TestNet_RPC,
		solana.MustPrivateKeyFromBase58(os.Getenv("PRIVATE_KEY")),
	)
	_, _ = client, err
}
