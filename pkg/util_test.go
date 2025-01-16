package pkg

import (
	"context"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestTipStream(t *testing.T) {
	ctx := context.Background()

	ch, chErr, err := SubscribeTipStream(ctx)
	if err != nil {
		t.Fatal(err)
	}

loop:
	for {
		select {
		case errStream := <-chErr:
			t.Fatal(errStream)
		case <-ch:
			break loop
		}
	}
}

func TestGetTipInfo(t *testing.T) {
	tipInfo, err := GetTipInformation(&http.Client{})
	assert.NoError(t, err)
	assert.NotNil(t, tipInfo)
}
