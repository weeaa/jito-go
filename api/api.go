package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

type method string

var (
	getBundleHistory method = "getBundleHistory"
)

type Client struct {
	ctx context.Context
	c   *http.Client
}

func New(ctx context.Context, client *http.Client) *Client {
	return &Client{ctx: ctx, c: client}
}

func (api *Client) RetrieveBundleIDfromTransaction() error {
	return errors.New("not impl yet")
}

type tmp struct{}

func (api *Client) DoYeet() (*tmp, error) {
	panic("impl me pls")
	//https://explorer.jito.wtf/wtfrest/api/v1/bundles/recent?limit=1000&sort=Tip&asc=false&timeframe=Week

	req := &http.Request{
		URL: &url.URL{
			Scheme: "https",
			Host:   baseApi,
			Path:   fmt.Sprintf(""),
		},
		Header: headers.Clone(),
	}

	resp, err := api.c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error %w")
	}

	return nil, nil
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
		Header: headers.Clone(),
	}

	resp, err := api.c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response status retrieving recent bundles, got %s", resp.Status)
	}

	var bundles RecentBundlesResponse
	if err = json.NewDecoder(resp.Body).Decode(&bundles); err != nil {
		return nil, err
	}

	return &bundles, nil
}

func (api *Client) GetBundleInfo(bundleID string) (*GetBundleInfoResponse, error) {

	req := &http.Request{
		Method: http.MethodPost,
		URL: &url.URL{
			Scheme: "https",
			Host:   baseApi,
			Path:   fmt.Sprintf("/wtfrest/api/v1/bundles/bundle/%s", bundleID),
		},
		Header: headers.Clone(),
	}

	resp, err := api.c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response status getting bundle [%s], got %s", bundleID, resp.Status)
	}

	var bundleInfo GetBundleInfoResponse
	if err = json.NewDecoder(resp.Body).Decode(&bundleInfo); err != nil {
		return nil, err
	}

	return &bundleInfo, nil
}

func GetLargestBundlesBySigner() {}
