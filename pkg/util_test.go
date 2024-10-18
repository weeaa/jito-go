package pkg

import (
	"context"
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
