// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package client wrap kubernetes client functions
package client

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// DefaultOperatorBaseURL 默认的 operator 服务地址.
	DefaultOperatorBaseURL = "https://datakit-operator.datakit.svc:443"
	// DefaultHTTPTimeout 默认 HTTP 请求超时时间.
	DefaultHTTPTimeout = 10 * time.Second
)

// OperatorPodInterface 定义 operator pod 接口，模仿 k8s client-go 的 PodInterface.
type OperatorPodInterface interface {
	// List 获取 pod 列表
	// 注意：opts 参数会被忽略，operator API 不支持 ListOptions（如 labelSelector、fieldSelector 等）
	List(ctx context.Context, opts metav1.ListOptions) (*corev1.PodList, error)
	// Get 获取指定的 pod
	// 注意：opts 参数会被忽略，operator API 不支持 GetOptions.
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*corev1.Pod, error)
}

type OperatorClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewKubernetesClientForOperator 创建一个新的 operator 客户端。
func NewKubernetesClientForOperator(baseURL string) (*OperatorClient, error) {
	if baseURL == "" {
		baseURL = DefaultOperatorBaseURL
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // nolint:gosec
		},
	}

	client := &OperatorClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   DefaultHTTPTimeout,
		},
	}

	return client, nil
}

func (c *OperatorClient) GetPods(namespace string) OperatorPodInterface {
	return &operatorPods{
		client:    c,
		namespace: namespace,
	}
}

type operatorPods struct {
	client    *OperatorClient
	namespace string
}

func (p *operatorPods) List(ctx context.Context, opts metav1.ListOptions) (*corev1.PodList, error) {
	var url string
	if p.namespace == "" {
		url = fmt.Sprintf("%s/v1/cluster/api/v1/pods", p.client.baseURL)
	} else {
		url = fmt.Sprintf("%s/v1/cluster/api/v1/namespaces/%s/pods", p.client.baseURL, p.namespace)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := p.client.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pods: %w", err)
	}
	defer resp.Body.Close() // nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var pods []corev1.Pod
	if err := json.Unmarshal(body, &pods); err != nil {
		return nil, fmt.Errorf("failed to unmarshal pods: %w", err)
	}

	podList := &corev1.PodList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PodList",
			APIVersion: "v1",
		},
		Items: pods,
	}

	return podList, nil
}

func (p *operatorPods) Get(ctx context.Context, name string, opts metav1.GetOptions) (*corev1.Pod, error) {
	if p.namespace == "" {
		return nil, fmt.Errorf("namespace cannot be empty")
	}
	if name == "" {
		return nil, fmt.Errorf("pod name cannot be empty")
	}

	url := fmt.Sprintf("%s/v1/cluster/api/v1/namespaces/%s/pods/%s", p.client.baseURL, p.namespace, name)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := p.client.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pod: %w", err)
	}
	defer resp.Body.Close() // nolint:errcheck

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("pod not found")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var pod corev1.Pod
	if err := json.Unmarshal(body, &pod); err != nil {
		return nil, fmt.Errorf("failed to unmarshal pod: %w", err)
	}

	return &pod, nil
}
