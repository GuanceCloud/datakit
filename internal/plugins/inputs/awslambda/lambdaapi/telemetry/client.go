// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2024-present Guance, Inc.

package telemetry

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/awslambda/lambdaapi/consts"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/awslambda/lambdaapi/model"
)

// BufferingCfg Configuration for receiving telemetry from the Telemetry API.
// Telemetry will be sent to your listener when one of the conditions below is met.
type BufferingCfg struct {
	// Maximum number of log events to be buffered in memory. (default: 10000, minimum: 1000, maximum: 10000).
	MaxItems uint32 `json:"maxItems"`
	// Maximum size in bytes of the log events to be buffered in memory. (default: 262144, minimum: 262144, maximum: 1048576).
	MaxBytes uint32 `json:"maxBytes"`
	// Maximum time (in milliseconds) for a batch to be buffered. (default: 1000, minimum: 100, maximum: 30000).
	TimeoutMS uint32 `json:"timeoutMs"`
}

// Destination Configuration for listeners that would like to receive telemetry via HTTP.
type Destination struct {
	Protocol   string `json:"protocol"`
	URI        string `json:"URI"`
	HTTPMethod string `json:"method"`
	Encoding   string `json:"encoding"`
}

// SubscribeRequest Request body that is sent to the Telemetry API on subscribe.
type SubscribeRequest struct {
	SchemaVersion string                        `json:"schemaVersion"`
	EventTypes    []model.SubscriptionEventType `json:"types"`
	BufferingCfg  BufferingCfg                  `json:"buffering"`
	Destination   Destination                   `json:"destination"`
}

// Client struct defines the structure for the telemetry API client.
type Client struct {
	BaseURL     string
	ExtensionID string
	Port        string
	Path        string
}

// NewTelemetryClient creates a new instance of TelemetryAPIClient.
func NewTelemetryClient(awsLambdaRuntimeAPI, extensionID, port, path string) *Client {
	baseURL := getBaseURL(awsLambdaRuntimeAPI)
	return &Client{
		BaseURL:     baseURL,
		ExtensionID: extensionID,
		Port:        port,
		Path:        path,
	}
}

// Subscribe sends a subscription request to the telemetry API.
func (c *Client) Subscribe(ctx context.Context) error {
	// Subscription request payload
	payload := &SubscribeRequest{
		SchemaVersion: "2022-12-13",
		Destination: Destination{
			Protocol:   "HTTP",
			URI:        fmt.Sprintf("http://sandbox:%s/%s", c.Port, c.Path),
			HTTPMethod: "POST",
			Encoding:   "JSON",
		},
		EventTypes: []model.SubscriptionEventType{model.SubscriptionEventFunction, model.SubscriptionEventPlatform},
		BufferingCfg: BufferingCfg{
			MaxItems:  1000,
			MaxBytes:  256 * 1024,
			TimeoutMS: 25,
		},
	}

	l.Infof("subscribe conf: %+#v", payload)

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		l.Errorf("[client:Subscribe] Failed to marshal SubscribeRequest. err: %s", err)
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.BaseURL, bytes.NewReader(payloadBytes))
	if err != nil {
		return fmt.Errorf("new subscribe request(PUT: %q) failed: %w", c.BaseURL, err)
	}

	req.Header.Set(consts.ExtensionIdentifierHeader, c.ExtensionID)
	req.Header.Set("Content-Type", "application/json")

	l.Info("[client:Subscribe] Subscribing using baseUrl:", c.BaseURL)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		l.Error("[client:Subscribe] Subscription failed:", err)
		return err
	}

	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode == http.StatusAccepted {
		l.Error("[client:Subscribe] Subscription failed. Logs API is not supported! Is this extension running in a local sandbox?")
	} else if resp.StatusCode != http.StatusOK {
		l.Error("[client:Subscribe] Subscription failed")

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			l.Errorf("[client:Subscribe] %s failed: %d[%s]", c.BaseURL, resp.StatusCode, resp.Status)
			return err
		}

		l.Errorf("[client:Subscribe] %s failed: %d[%s] %s", c.BaseURL, resp.StatusCode, resp.Status, string(body))
		return err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		l.Errorf("[client:Subscribe] io.ReadAll: %s", err)
		return err
	}

	l.Infof("[client:Subscribe] Subscription success: %s, Subscribed to telemetry: %+v", string(body), resp)
	return nil
}
