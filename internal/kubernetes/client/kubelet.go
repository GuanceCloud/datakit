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
	GetMetrics() (io.ReadCloser, error)
	GetMetricsCadvisor() (io.ReadCloser, error)
	GetMetricsResource() (io.ReadCloser, error)
}

func NewDefaultKubeletClient(config *rest.Config) (KubeletClient, error) {
	var (
		scheme  = "https"
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

	return NewKubeletClient(config, scheme, address, port)
}

func NewKubeletClient(config *rest.Config, scheme, address, port string) (KubeletClient, error) {
	base, err := NewBaseClient(config)
	if err != nil {
		return nil, err
	}
	return &kubeletClient{
		client: base,
		scheme: scheme,
		host:   net.JoinHostPort(address, port),
	}, nil
}

type kubeletClient struct {
	client *BaseClient
	scheme string
	host   string
}

func (kc *kubeletClient) GetStatsSummary() (*statsv1alpha1.Summary, error) {
	req, err := kc.newRequest("/stats/summary")
	if err != nil {
		return nil, err
	}

	b, err := kc.client.DoRaw(req)
	if err != nil {
		return nil, err
	}

	var ms statsv1alpha1.Summary
	if err := json.Unmarshal(b, &ms); err != nil {
		return nil, fmt.Errorf("failed to parse summary - %w", err)
	}

	return &ms, nil
}

func (kc *kubeletClient) GetMetrics() (io.ReadCloser, error) {
	path := "/metrics"
	return kc.do(path)
}

func (kc *kubeletClient) GetMetricsCadvisor() (io.ReadCloser, error) {
	path := "/metrics/cadvisor"
	return kc.do(path)
}

func (kc *kubeletClient) GetMetricsResource() (io.ReadCloser, error) {
	path := "/metrics/resource"
	return kc.do(path)
}

func (kc *kubeletClient) newRequest(path string) (*http.Request, error) {
	u := url.URL{
		Scheme: kc.scheme,
		Host:   kc.host,
		Path:   path,
	}
	return http.NewRequest("GET", u.String(), nil)
}

func (kc *kubeletClient) do(path string) (io.ReadCloser, error) {
	req, err := kc.newRequest(path)
	if err != nil {
		return nil, err
	}
	return kc.client.Stream(req)
}
