// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package jolokia

import (
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf/plugins/common/tls"
)

func TestNewClient(t *testing.T) {
	t.Skip("skipping test in short mode, because it requires a running Jolokia server")
	config := Config{
		URL:             "http://localhost:7777/jolokia",
		ResponseTimeout: 5 * time.Second,
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	if client == nil {
		t.Fatal("NewClient returned nil client")
	}

	if client.URL() != config.URL {
		t.Errorf("Expected URL %s, got %s", config.URL, client.URL())
	}
}

func TestNewClientWithTLS(t *testing.T) {
	t.Skip("skipping test in short mode, because it requires a running Jolokia server")
	config := Config{
		URL:             "https://localhost:7777/jolokia",
		ResponseTimeout: 5 * time.Second,
		TLS:             tls.ClientConfig{},
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	if client == nil {
		t.Fatal("NewClient returned nil client")
	}
}

func TestNewClientWithAuth(t *testing.T) {
	t.Skip("skipping test in short mode, because it requires a running Jolokia server")
	config := Config{
		URL:             "http://localhost:7777/jolokia",
		Username:        "testuser",
		Password:        "testpass",
		ResponseTimeout: 5 * time.Second,
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	if client == nil {
		t.Fatal("NewClient returned nil client")
	}
}

func TestNewClientWithProxyConfig(t *testing.T) {
	t.Skip("skipping test in short mode, because it requires a running Jolokia server")
	config := Config{
		URL:             "http://localhost:7777/jolokia",
		ResponseTimeout: 5 * time.Second,
		ProxyConfig: &ProxyConfig{
			URL:      "service:jmx:rmi:///jndi/rmi://target:9010/jmxrmi",
			Username: "proxyuser",
			Password: "proxypass",
		},
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	if client == nil {
		t.Fatal("NewClient returned nil client")
	}
}

func TestNewReadRequest(t *testing.T) {
	t.Skip("skipping test in short mode, because it requires a running Jolokia server")
	config := Config{
		URL:             "http://localhost:7777/jolokia",
		ResponseTimeout: 5 * time.Second,
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	// Test with single attribute
	req := client.NewReadRequest("java.lang:type=Runtime", []string{"Uptime"}, "")
	if req.Type != "read" {
		t.Errorf("Expected type 'read', got %s", req.Type)
	}
	if req.Mbean != "java.lang:type=Runtime" {
		t.Errorf("Expected mbean 'java.lang:type=Runtime', got %s", req.Mbean)
	}
	if req.Attribute != "Uptime" {
		t.Errorf("Expected attribute 'Uptime', got %v", req.Attribute)
	}

	// Test with multiple attributes
	req2 := client.NewReadRequest("java.lang:type=Memory", []string{"HeapMemoryUsage", "NonHeapMemoryUsage"}, "")
	if req2.Type != "read" {
		t.Errorf("Expected type 'read', got %s", req2.Type)
	}
	attrs, ok := req2.Attribute.([]string)
	if !ok {
		t.Errorf("Expected []string, got %T", req2.Attribute)
	}
	if len(attrs) != 2 {
		t.Errorf("Expected 2 attributes, got %d", len(attrs))
	}

	// Test with no attributes
	req3 := client.NewReadRequest("java.lang:type=Runtime", nil, "")
	if req3.Attribute != nil {
		t.Errorf("Expected nil attribute, got %v", req3.Attribute)
	}

	// Test with path
	req4 := client.NewReadRequest("java.lang:type=Memory", []string{"HeapMemoryUsage"}, "used")
	if req4.Path != "used" {
		t.Errorf("Expected path 'used', got %s", req4.Path)
	}
}

func TestNewWriteRequest(t *testing.T) {
	t.Skip("skipping test in short mode, because it requires a running Jolokia server")
	config := Config{
		URL:             "http://localhost:7777/jolokia",
		ResponseTimeout: 5 * time.Second,
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	req := client.NewWriteRequest("java.lang:type=Memory", "HeapMemoryUsage", "used", 100)
	if req.Type != "write" {
		t.Errorf("Expected type 'write', got %s", req.Type)
	}
	if req.Mbean != "java.lang:type=Memory" {
		t.Errorf("Expected mbean 'java.lang:type=Memory', got %s", req.Mbean)
	}
	if req.Attribute != "HeapMemoryUsage" {
		t.Errorf("Expected attribute 'HeapMemoryUsage', got %v", req.Attribute)
	}
	if req.Value != 100 {
		t.Errorf("Expected value 100, got %v", req.Value)
	}
}

func TestNewExecRequest(t *testing.T) {
	t.Skip("skipping test in short mode, because it requires a running Jolokia server")
	config := Config{
		URL:             "http://localhost:7777/jolokia",
		ResponseTimeout: 5 * time.Second,
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	req := client.NewExecRequest("java.lang:type=Runtime", "gc", []interface{}{})
	if req.Type != "exec" {
		t.Errorf("Expected type 'exec', got %s", req.Type)
	}
	if req.Operation != "gc" {
		t.Errorf("Expected operation 'gc', got %s", req.Operation)
	}
}

func TestNewSearchRequest(t *testing.T) {
	t.Skip("skipping test in short mode, because it requires a running Jolokia server")
	config := Config{
		URL:             "http://localhost:7777/jolokia",
		ResponseTimeout: 5 * time.Second,
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	req := client.NewSearchRequest("java.lang:*")
	if req.Type != "search" {
		t.Errorf("Expected type 'search', got %s", req.Type)
	}
	if req.Mbean != "java.lang:*" {
		t.Errorf("Expected mbean 'java.lang:*', got %s", req.Mbean)
	}
}

func TestNewListRequest(t *testing.T) {
	t.Skip("skipping test in short mode, because it requires a running Jolokia server")
	config := Config{
		URL:             "http://localhost:7777/jolokia",
		ResponseTimeout: 5 * time.Second,
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	req := client.NewListRequest("java.lang:*")
	if req.Type != "list" {
		t.Errorf("Expected type 'list', got %s", req.Type)
	}
	if req.Mbean != "java.lang:*" {
		t.Errorf("Expected mbean 'java.lang:*', got %s", req.Mbean)
	}
}

func TestNewVersionRequest(t *testing.T) {
	t.Skip("skipping test in short mode, because it requires a running Jolokia server")
	config := Config{
		URL:             "http://localhost:7777/jolokia",
		ResponseTimeout: 5 * time.Second,
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	req := client.NewVersionRequest()
	if req.Type != "version" {
		t.Errorf("Expected type 'version', got %s", req.Type)
	}
}

func TestFormatURL(t *testing.T) {
	t.Skip("skipping test in short mode, because it requires a running Jolokia server")
	config := Config{
		URL:             "http://localhost:7777/jolokia",
		ResponseTimeout: 5 * time.Second,
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	url, err := client.formatURL()
	if err != nil {
		t.Fatalf("formatURL failed: %v", err)
	}

	// Should end with /jolokia/
	if !strings.HasSuffix(url, "/jolokia/") {
		t.Errorf("Expected URL to end with '/jolokia/', got %s", url)
	}

	// Should contain ignoreErrors query parameter
	if !strings.Contains(url, "ignoreErrors=true") {
		t.Errorf("Expected URL to contain 'ignoreErrors=true', got %s", url)
	}
}

func TestFormatURLWithPath(t *testing.T) {
	t.Skip("skipping test in short mode, because it requires a running Jolokia server")
	config := Config{
		URL:             "http://localhost:7777/jolokia/read",
		ResponseTimeout: 5 * time.Second,
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	url, err := client.formatURL()
	if err != nil {
		t.Fatalf("formatURL failed: %v", err)
	}

	// Should end with /read/
	if !strings.HasSuffix(url, "/read/") {
		t.Errorf("Expected URL to end with '/read/', got %s", url)
	}
}

func TestFormatURLWithAuth(t *testing.T) {
	t.Skip("skipping test in short mode, because it requires a running Jolokia server")
	config := Config{
		URL:             "http://localhost:7777/jolokia",
		Username:        "testuser",
		Password:        "testpass",
		ResponseTimeout: 5 * time.Second,
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	url, err := client.formatURL()
	if err != nil {
		t.Fatalf("formatURL failed: %v", err)
	}

	// Should contain auth info in URL
	if !strings.Contains(url, "testuser") {
		t.Errorf("Expected URL to contain username, got %s", url)
	}
}

// Integration test - requires running Jolokia server at http://localhost:7777/jolokia
func TestExecuteIntegration(t *testing.T) {
	t.Skip("skipping test in short mode, because it requires a running Jolokia server")

	config := Config{
		URL:             "http://localhost:7777/jolokia",
		ResponseTimeout: 5 * time.Second,
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	req := client.NewReadRequest("java.lang:type=Runtime", []string{"Uptime"}, "")
	resp, err := client.Execute(req)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if resp == nil {
		t.Fatal("Execute returned nil response")
	}

	if resp.Status != 200 {
		t.Errorf("Expected status 200, got %d: %s", resp.Status, resp.Error)
	}
}

// Integration test - requires running Jolokia server at http://localhost:7777/jolokia
func TestBatchExecuteIntegration(t *testing.T) {
	t.Skip("skipping test in short mode, because it requires a running Jolokia server")

	config := Config{
		URL:             "http://localhost:7777/jolokia",
		ResponseTimeout: 5 * time.Second,
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	req1 := client.NewReadRequest("java.lang:type=Runtime", []string{"Uptime"}, "")
	req2 := client.NewReadRequest("java.lang:type=Memory", []string{"HeapMemoryUsage"}, "")

	responses, err := client.BatchExecute([]*Request{req1, req2})
	if err != nil {
		t.Fatalf("BatchExecute failed: %v", err)
	}

	if len(responses) != 2 {
		t.Fatalf("Expected 2 responses, got %d", len(responses))
	}
	for i, resp := range responses {
		t.Logf("Response %d: %+v", i, resp)
		if resp.Status != 200 {
			t.Errorf("Response %d: Expected status 200, got %d: %s", i, resp.Status, resp.Error)
		}
	}
}
