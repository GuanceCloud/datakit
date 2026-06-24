// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package metadata

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	dkhttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpcli"
)

const DefaultBaseURL = "http://metadata.google.internal/computeMetadata/v1"

const (
	projectIDPath       = "/project/project-id"
	clusterNamePath     = "/instance/attributes/cluster-name"
	clusterLocationPath = "/instance/attributes/cluster-location"
)

type Config struct {
	ProjectID       string
	ClusterName     string
	ClusterLocation string
}

type Client struct {
	baseURL string
	client  *http.Client
}

func NewClient() *Client {
	return NewClientWithHTTP(metadataBaseURL(), &http.Client{
		Transport: dkhttp.DefTransport(),
		Timeout:   5 * time.Second,
	})
}

func NewClientWithHTTP(baseURL string, client *http.Client) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	if client == nil {
		client = &http.Client{Transport: dkhttp.DefTransport(), Timeout: 5 * time.Second}
	}
	return &Client{baseURL: baseURL, client: client}
}

func (c *Client) Resolve(ctx context.Context, cfg Config) (Config, error) {
	if cfg.ProjectID == "" {
		value, err := c.get(ctx, projectIDPath)
		if err != nil {
			return cfg, fmt.Errorf("discover gcp project id: %w", err)
		}
		cfg.ProjectID = value
	}
	if cfg.ClusterName == "" {
		value, err := c.get(ctx, clusterNamePath)
		if err != nil {
			return cfg, fmt.Errorf("discover gke cluster name: %w", err)
		}
		cfg.ClusterName = value
	}
	if cfg.ClusterLocation == "" {
		value, err := c.get(ctx, clusterLocationPath)
		if err != nil {
			return cfg, fmt.Errorf("discover gke cluster location: %w", err)
		}
		cfg.ClusterLocation = value
	}
	return cfg, nil
}

func (c *Client) get(ctx context.Context, path string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return "", fmt.Errorf("create gcp metadata request: %w", err)
	}
	req.Header.Set("Metadata-Flavor", "Google")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request gcp metadata %s: %w", path, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
		return "", fmt.Errorf("gcp metadata %s returned %s: %s",
			path, resp.Status, strings.TrimSpace(string(body)))
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return "", fmt.Errorf("read gcp metadata %s: %w", path, err)
	}
	value := strings.TrimSpace(string(body))
	if value == "" {
		return "", fmt.Errorf("gcp metadata %s returned an empty value", path)
	}
	return value, nil
}

func metadataBaseURL() string {
	if host := os.Getenv("GCE_METADATA_HOST"); host != "" {
		return "http://" + strings.TrimRight(host, "/") + "/computeMetadata/v1"
	}
	return DefaultBaseURL
}
