// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package client

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"k8s.io/client-go/rest"
)

type BaseClient struct {
	client *http.Client
}

func NewBaseClient(config *rest.Config) (*BaseClient, error) {
	transport, err := rest.TransportFor(config)
	if err != nil {
		return nil, fmt.Errorf("unable to construct transport: %w", err)
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   config.Timeout,
	}

	return &BaseClient{client}, nil
}

func (b *BaseClient) Get(path string) (*http.Response, error) {
	u, err := url.Parse(path)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	return b.Do(req)
}

func (b *BaseClient) Do(req *http.Request) (*http.Response, error) {
	return b.client.Do(req)
}

func (b *BaseClient) DoRaw(req *http.Request) ([]byte, error) {
	body, err := b.Stream(req)
	if err != nil {
		return nil, err
	}
	// nolint:errcheck
	defer body.Close()
	data, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (b *BaseClient) Stream(req *http.Request) (io.ReadCloser, error) {
	resp, err := b.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed, status: %q", resp.Status)
	}
	return resp.Body, nil
}
