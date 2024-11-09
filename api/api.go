package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

type Client struct {
	ctx context.Context
	c   *http.Client
}

func New(ctx context.Context, client *http.Client) *Client {
	return &Client{ctx: ctx, c: client}
}

// RetrieveBundleIDfromTransactionSignature returns the bundleID associated with the transaction signature provided, if existing.
func (api *Client) RetrieveBundleIDfromTransactionSignature(signature string) (string, error) {
	req := &http.Request{
		Method: http.MethodGet,
		URL: &url.URL{
			Scheme: "https",
			Host:   "bundles.jito.wtf",
			Path:   fmt.Sprintf("/api/v1/bundles/transaction/%s", signature),
		},
		Header: headers.Clone(),
	}

	resp, err := api.c.Do(req)
	if err != nil {
		return "", fmt.Errorf("error retrieving bundle ID from transaction: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error getting bundle id: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var out []map[string]string
	if err = json.Unmarshal(body, &out); err != nil {
		return "", err
	}

	return out[0]["bundle_id"], nil
}

type tmp struct{}

/*
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
*/

// RetrieveRecentBundles fetches a list of recent bundles from the Jito API within a specified timeframe and limit.
func (api *Client) RetrieveRecentBundles(limit int, timeFrame Timeframe) (*RecentBundlesResponse, error) {
	return nil, errors.New("i'm broken pls wait ser till im fixed")

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
	log.Println(req)

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

// GetBundleInfo returns information associated with a bundle ID.
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

func (api *Client) GetLargestBundlesBySigner() error {
	return errors.New("pls impl me")
}
