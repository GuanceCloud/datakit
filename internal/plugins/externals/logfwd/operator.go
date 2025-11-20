// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux
// +build linux

package logfwd

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type operatorAPIResponse struct {
	Name            string   `json:"name,omitempty"`
	PodTargetLabels []string `json:"pod_target_labels,omitempty"`
	Configs         string   `json:"configs,omitempty"`
	Error           string   `json:"error,omitempty"`
}

// operatorClient datakit-operator HTTP 客户端.
type operatorClient struct {
	baseURL string // 完整的 base URL，如 https://datakit-operator.datakit.svc:443
	client  *http.Client
}

// newOperatorClient 创建新的 operator 客户端.
func newOperatorClient(baseURL string) *operatorClient {
	if strings.TrimSpace(baseURL) == "" {
		return nil
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // nolint:gosec
		},
	}

	return &operatorClient{
		baseURL: strings.TrimSuffix(baseURL, "/"), // 移除末尾的斜杠
		client: &http.Client{
			Transport: transport,
			Timeout:   10 * time.Second,
		},
	}
}

// fetchOperatorResponse 从 datakit-operator 获取响应.
func (oc *operatorClient) fetchOperatorResponse(namespace, podName, podLabels string) (*operatorAPIResponse, error) {
	if oc.baseURL == "" {
		return nil, fmt.Errorf("operator base URL not configured")
	}

	reqURL := fmt.Sprintf("%s/v1/logging/configs?namespace=%s&pod_name=%s",
		oc.baseURL, url.QueryEscape(namespace), url.QueryEscape(podName))

	if podLabels != "" {
		reqURL += "&pod_labels=" + url.QueryEscape(podLabels)
	}

	resp, err := oc.client.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch operator response: %w", err)
	}
	defer resp.Body.Close() // nolint:errcheck

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		// 没有找到配置是正常的，返回空的响应
		return &operatorAPIResponse{}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var apiResp operatorAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if apiResp.Error != "" {
		return nil, fmt.Errorf("API error: %s", apiResp.Error)
	}

	return &apiResp, nil
}
