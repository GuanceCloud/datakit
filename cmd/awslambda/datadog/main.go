// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	lambdaRuntimeAPI  = "AWS_LAMBDA_RUNTIME_API"
	extensionName     = "datadog-agent"
	extensionEvents   = `["SHUTDOWN"]`
	apiRegisterPath   = "/2020-01-01/extension/register"
	apiNextEventPath  = "/2020-01-01/extension/event/next"
	eventTypeShutdown = "SHUTDOWN"
)

type RuntimeAPI struct {
	BaseURL     string
	ExtensionID string

	client *http.Client
}

func NewRuntimeAPI() *RuntimeAPI {
	return &RuntimeAPI{
		BaseURL: os.Getenv(lambdaRuntimeAPI),
		client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:    10,
				IdleConnTimeout: 30 * time.Second,
			},
		},
	}
}

func (r *RuntimeAPI) registerExtension() error {
	if r.BaseURL == "" {
		return fmt.Errorf("AWS_LAMBDA_RUNTIME_API environment variable not set")
	}

	url := fmt.Sprintf("http://%s%s", r.BaseURL, apiRegisterPath)

	reqBody := []byte(fmt.Sprintf(`{"events": %s}`, extensionEvents))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("create register request failed: %w", err)
	}

	req.Header.Set("Lambda-Extension-Name", extensionName)
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("register extension failed: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("register extension failed with status code %d: %s", resp.StatusCode, body)
	}

	r.ExtensionID = resp.Header.Get("Lambda-Extension-Identifier")

	if r.ExtensionID == "" {
		return fmt.Errorf("extension identifier not found in response header")
	}

	return nil
}

func (r *RuntimeAPI) nextEvent() (string, error) {
	if r.ExtensionID == "" {
		return "", fmt.Errorf("extension not registered")
	}

	url := fmt.Sprintf("http://%s%s", r.BaseURL, apiNextEventPath)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("create next event request failed: %w", err)
	}

	req.Header.Set("Lambda-Extension-Identifier", r.ExtensionID)

	// No timeout for nextEvent since it's a blocking request that can wait indefinitely
	// This is appropriate for Lambda extensions as they should wait for shutdown events
	nextEventClient := &http.Client{
		Timeout:   0,                  // No timeout - wait indefinitely for blocking request
		Transport: r.client.Transport, // reuse existing Transport configuration
	}

	resp, err := nextEventClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("get next event failed: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read next event response failed: %w", err)
	}

	return string(body), nil
}

type LambdaEvent struct {
	EventType          string `json:"eventType"`
	DeadlineMS         int64  `json:"deadlineMs"`
	RequestID          string `json:"requestId"`
	InvokedFunctionArn string `json:"invokedFunctionArn"`
}

func main() {
	runtimeAPI := NewRuntimeAPI()

	if err := runtimeAPI.registerExtension(); err != nil {
		fmt.Printf("failed to register extension: %v\n", err)
		os.Exit(1)
	}

	maxRetries := 5
	retryCount := 0
	sleepDuration := 10 * time.Second

	for {
		eventJSON, err := runtimeAPI.nextEvent()
		if err != nil {
			retryCount++
			fmt.Printf("failed to get next event: %v\n", err)
			if retryCount >= maxRetries {
				fmt.Printf("failed to get next event after %d retries: %v\n", maxRetries, err)
				return
			}
			time.Sleep(sleepDuration)
			continue
		} else {
			retryCount = 0
		}

		var event LambdaEvent
		if err := json.Unmarshal([]byte(eventJSON), &event); err != nil {
			fmt.Printf("failed to unmarshal event: %v, eventJSON: %s\n", err, eventJSON)
			continue
		}

		switch event.EventType {
		case eventTypeShutdown:
			return
		default:
			time.Sleep(sleepDuration)
		}
	}
}
