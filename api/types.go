package api

import (
	"time"
)

var (
	baseApi = "explorer.jito.wtf"

	recentBundlesPath = "/wtfrest/api/v1/bundles/recent"
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
