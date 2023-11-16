// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"

	"k8s.io/client-go/rest"

	statsv1alpha1 "k8s.io/kubelet/pkg/apis/stats/v1alpha1"
)

type KubeletClientConfig struct {
	Client      *rest.Config
	Scheme      string
	DefaultPort int
	// UseNodeStatusPort bool
}

type kubeletClient struct {
	client      *http.Client
	scheme      string
	defaultPort int
}

func NewKubeletClientForConfig(config *KubeletClientConfig) (*kubeletClient, error) {
	transport, err := rest.TransportFor(config.Client)
	if err != nil {
		return nil, fmt.Errorf("unable to construct transport: %w", err)
	}

	c := &http.Client{
		Transport: transport,
		Timeout:   config.Client.Timeout,
	}

	return &kubeletClient{
		client:      c,
		scheme:      config.Scheme,
		defaultPort: config.DefaultPort,
	}, nil
}

func (kc *kubeletClient) GetMetrics() (*statsv1alpha1.Summary, error) {
	addr := "127.0.0.1"
	port := kc.defaultPort
	path := "/stats/summary"

	url := url.URL{
		Scheme: kc.scheme,
		Host:   net.JoinHostPort(addr, strconv.Itoa(port)),
		Path:   path,
	}
	return kc.getMetrics(context.Background(), url.String())
}

func (kc *kubeletClient) getMetrics(ctx context.Context, url string) (*statsv1alpha1.Summary, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	response, err := kc.client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = response.Body.Close()
	}()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed, status: %q", response.Status)
	}

	b, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body - %w", err)
	}

	ms, err := decodeSummary(b)
	if err != nil {
		return nil, err
	}

	return ms, nil
}

func decodeSummary(b []byte) (*statsv1alpha1.Summary, error) {
	var ms statsv1alpha1.Summary
	if err := json.Unmarshal(b, &ms); err != nil {
		return nil, fmt.Errorf("failed to parse summary - %w", err)
	}
	return &ms, nil
}
