// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"

	"k8s.io/client-go/rest"

	statsv1alpha1 "k8s.io/kubelet/pkg/apis/stats/v1alpha1"
)

type KubeletClient interface {
	GetStatsSummary() (*statsv1alpha1.Summary, error)
}

func DefaultKubeletHostInCluster() string {
	var (
		address = "127.0.0.1"
		port    = "10250"
	)
	// Some of kubelet not bind to 127.0.0.1
	if s := os.Getenv("ENV_K8S_NODE_IP"); s != "" {
		address = s
	}
	// Deprecated: use ENV_K8S_NODE_IP
	if s := os.Getenv("HOST_IP"); s != "" {
		address = s
	}
	return net.JoinHostPort(address, port)
}

func NewDefaultKubeletClient(config *rest.Config) (KubeletClient, error) {
	host := DefaultKubeletHostInCluster()
	return NewKubeletClient(config, "https", host)
}

func NewKubeletClient(config *rest.Config, scheme, host string) (KubeletClient, error) {
	transport, err := rest.TransportFor(config)
	if err != nil {
		return nil, fmt.Errorf("unable to construct transport: %w", err)
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   config.Timeout,
	}
	return &kubeletClient{
		client: client,
		scheme: scheme,
		host:   host,
	}, nil
}

type kubeletClient struct {
	client *http.Client
	scheme string
	host   string
}

func (kc *kubeletClient) GetStatsSummary() (*statsv1alpha1.Summary, error) {
	u := url.URL{
		Scheme: kc.scheme,
		Host:   kc.host,
		Path:   "/stats/summary",
	}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := kc.client.Do(req)
	if err != nil {
		return nil, err
	}
	//nolint:errcheck
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed, status: %q", resp.Status)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body - %w", err)
	}

	var ms statsv1alpha1.Summary
	if err := json.Unmarshal(b, &ms); err != nil {
		return nil, fmt.Errorf("failed to parse summary - %w", err)
	}

	return &ms, nil
}
