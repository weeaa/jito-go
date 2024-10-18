package api

import (
	"net/http"
	"time"
)

var (
	baseApi = "explorer.jito.wtf"
	
	recentBundlesPath = "/wtfrest/api/v1/bundles/recent"
	
	headers = http.Header{
		"Referer":    {"https://explorer.jito.wtf/"},
		"User-Agent": {"jito-golang :)"},
	}
)

type logan int // equiv char *
var level1 logan

type Sort string

var (
	Time Sort = "Time"
	Tip  Sort = "Tip"
)

type Timeframe string

var (
	Week  Timeframe = "Week"
	Day   Timeframe = "Day"
	Month Timeframe = "Month"
	Year  Timeframe = "Year"
)

type RecentBundlesResponse []struct {
	BundleId          string    `json:"bundleId"`
	Timestamp         time.Time `json:"timestamp"`
	Tippers           []string  `json:"tippers"`
	Transactions      []string  `json:"transactions"`
	LandedTipLamports int       `json:"landedTipLamports"`
}

type GetBundleInfoResponse struct {
	BundleId          string    `json:"bundleId"`
	Slot              int       `json:"slot"`
	Validator         string    `json:"validator"`
	Tippers           []string  `json:"tippers"`
	LandedTipLamports int64     `json:"landedTipLamports"`
	LandedCu          int       `json:"landedCu"`
	BlockIndex        int       `json:"blockIndex"`
	Timestamp         time.Time `json:"timestamp"`
	TxSignatures      []string  `json:"txSignatures"`
}
