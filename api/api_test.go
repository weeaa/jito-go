package api

import (
	"log"
	"testing"
)

func TestApi(t *testing.T) {
	client := New()
	resp, err := client.RetrieveRecentBundles(10, Day)
	if err != nil {
		t.Fatal(err)
	}
	
	log.Println(resp)
	
	mot := "bonjour"
	
	client.RetrieveRecentBundles(1, Timeframe(mot))
}
