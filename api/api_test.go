package api

import (
	"context"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestApi(t *testing.T) {
	ctx := context.Background()
	client := New(ctx, &http.Client{})

	t.Run("RetrieveBundleIDfromTransactionSignature", func(t *testing.T) {
		bundleID, err := client.RetrieveBundleIDfromTransactionSignature("25buYcoQm4BtjPPyk1VTdjFr6fKVBRSBwg5zFLcocLUkJYrd1HQTwmWCgNGtfJ2p5Gu1LF3rmwLaWySFtJJ7jALd")
		assert.NoError(t, err)
		assert.NotEmpty(t, bundleID)
	})

	t.Run("RetrieveRecentBundles", func(t *testing.T) {
		t.Skip()
		resp, err := client.RetrieveRecentBundles(10, Day)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("GetBundleInfo", func(t *testing.T) {
		resp, err := client.GetBundleInfo("ee86fdca6817fc51992681e2093284e4c429874eec1513dc1a5b7830fdbc783b")
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("GetDailyMevRewards", func(t *testing.T) {
		resp, err := client.GetDailyMevRewards()
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})
}
