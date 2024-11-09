package api

import (
	"context"
	"log"
	"net/http"
	"testing"
)

// todo add more tests
func TestApi(t *testing.T) {
	ctx := context.Background()
	client := New(ctx, &http.Client{})

	t.Run("RetrieveBundleIDfromTransactionSignature", func(t *testing.T) {
		bundleID, err := client.RetrieveBundleIDfromTransactionSignature("25buYcoQm4BtjPPyk1VTdjFr6fKVBRSBwg5zFLcocLUkJYrd1HQTwmWCgNGtfJ2p5Gu1LF3rmwLaWySFtJJ7jALd")
		if err != nil {
			t.Fatal(err)
		}

		log.Printf("bundleID from signature: %s", bundleID)
	})

	t.Run("RetrieveRecentBundles", func(t *testing.T) {
		resp, err := client.RetrieveRecentBundles(10, Day)
		if err != nil {
			t.Fatal(err)
		}

		log.Printf("recent bundles: %+v", resp)
	})
}
