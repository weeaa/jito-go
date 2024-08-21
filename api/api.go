package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/graphql-go/graphql"
	"net/http"
	"net/url"
)

type method string

var (
	getBundleHistory method = "getBundleHistory"
)

var schemas = map[method]graphql.Schema{
	getBundleHistory: graphql.Schema{},
}

type Client struct {
	c *http.Client
}

func New() *Client {
	return &Client{c: &http.Client{}}
}

func (api *Client) RetrieveBundleIDfromTransaction() error {
	return errors.New("not impl yet")
}

// RetrieveRecentBundles fetches a list of recent bundles from the Jito API within a specified timeframe and limit.
func (api *Client) RetrieveRecentBundles(limit int, timeFrame Timeframe) (*RecentBundlesResponse, error) {
	params := &url.Values{
		"limit":     {fmt.Sprint(limit)},
		"sort":      {"Time"},
		"asc":       {"false"},
		"timeFrame": {string(timeFrame)},
	}

	req := &http.Request{
		Method: http.MethodGet,
		URL: &url.URL{
			Scheme:   "https",
			Host:     baseApi,
			Path:     recentBundlesPath,
			RawQuery: params.Encode(),
		},
		Header: http.Header{
			"Referer":    {"https://explorer.jito.wtf/"},
			"User-Agent": {"jito-golang :)"},
		},
	}

	resp, err := api.c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response status retrieving recent bundles, got %s", resp.Status)
	}

	var bundles *RecentBundlesResponse
	if err = json.NewDecoder(resp.Body).Decode(&bundles); err != nil {
		return nil, err
	}

	return bundles, nil
}
