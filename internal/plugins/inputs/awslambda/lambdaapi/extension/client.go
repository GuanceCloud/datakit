// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2024-present Guance, Inc.

// Package extension aws extension api related operations.
package extension

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/awslambda/lambdaapi/consts"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/awslambda/lambdaapi/model"
)

// RegisterResponse is the body of the response for /register.
type RegisterResponse struct {
	FunctionName    string `json:"functionName,omitempty"`
	FunctionVersion string `json:"functionVersion,omitempty"`
	Handler         string `json:"handler,omitempty"`
	AccountID       string `json:"accountId,omitempty"`
}

// NextEventResponse is the response for /event/next.
type NextEventResponse struct {
	EventType          model.EventType `json:"eventType,omitempty"`
	DeadlineMs         int64           `json:"deadlineMs,omitempty"`
	RequestID          string          `json:"requestId,omitempty"`
	InvokedFunctionArn string          `json:"invokedFunctionArn,omitempty"`
	Tracing            *model.Tracing  `json:"tracing,omitempty"`
	ShutdownReason     string          `json:"shutdownReason,omitempty"`
}

// StatusResponse is the body of the response for /init/error and /exit/error.
type StatusResponse struct {
	Status string `json:"status,omitempty"`
}

// Client is a simple client for the Lambda Extensions API.
type Client struct {
	baseURL     string
	httpClient  *http.Client
	ExtensionID string
}

// NewClient returns a Lambda Extensions API client.
func NewClient(awsLambdaRuntimeAPI string) *Client {
	baseURL := getBaseURL(awsLambdaRuntimeAPI)
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

// Register will register the extension with the Extensions API.
func (e *Client) Register(ctx context.Context, filename string) (*RegisterResponse, error) {
	const action = "/register"
	url := e.baseURL + action

	l.Info("[client:Register] Registering using baseURL ", url)

	reqBody, err := json.Marshal(map[string]interface{}{
		"events": []model.EventType{model.Invoke, model.Shutdown},
	})
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set(consts.ExtensionNameHeader, filename)
	httpRes, err := e.httpClient.Do(httpReq)
	if err != nil {
		l.Error("[client:Register] Registration failed ", err)
		return nil, err
	}

	if httpRes.StatusCode < 200 && httpRes.StatusCode >= 300 {
		l.Error("[client:Register] Registration failed with statusCode ", httpReq.Response.StatusCode)
		return nil, fmt.Errorf("request failed with status %s", httpRes.Status)
	}

	defer httpRes.Body.Close() // nolint: errcheck

	body, err := io.ReadAll(httpRes.Body)
	if err != nil {
		return nil, err
	}

	if l.Level() <= zap.DebugLevel {
		l.Debug("req body", string(body))
	}

	res := RegisterResponse{}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	e.ExtensionID = httpRes.Header.Get(consts.ExtensionIdentifierHeader)
	l.Info("[client:Register] Registration success with extensionId ", e.ExtensionID)

	return &res, nil
}

// NextEvent blocks while long polling for the next lambda invoke or shutdown.
func (e *Client) NextEvent(ctx context.Context) (*NextEventResponse, error) {
	const action = "/event/next"
	url := e.baseURL + action

	l.Debugf("request GET %q", url)
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set(consts.ExtensionIdentifierHeader, e.ExtensionID)
	httpRes, err := e.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request %q failed: %w", url, err)
	}

	if httpRes.StatusCode < 200 && httpRes.StatusCode >= 300 {
		return nil, fmt.Errorf("request failed with status %s", httpRes.Status)
	}

	defer httpRes.Body.Close() // nolint:errcheck

	body, err := io.ReadAll(httpRes.Body)
	if err != nil {
		return nil, err
	}

	var res NextEventResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("json.Unmarshal %q failed: %w", string(body), err)
	}

	return &res, nil
}

func (e *Client) AsyncNextEventLoop(ctx context.Context, eventDone <-chan struct{}) (<-chan *NextEventResponse, error) {
	resChan := make(chan *NextEventResponse)

	go func() {
		for {
			<-eventDone

			select {
			case <-ctx.Done():
				return
			default:
				l.Info("[client:AsyncNextEventLoop] starting next event")
				event, err := e.NextEvent(ctx)
				if err != nil {
					l.Errorf("[client:AsyncNextEventLoop] failed with err %s", err)
					close(resChan)
					return
				}
				l.Info("[client:AsyncNextEventLoop] end next event", event)
				resChan <- event
			}
		}
	}()

	return resChan, nil
}

// InitError reports an initialization error to the platform. Call it when you registered but failed to initialize.
func (e *Client) InitError(ctx context.Context, errorType string) (*StatusResponse, error) {
	const action = "/init/error"
	url := e.baseURL + action

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set(consts.ExtensionIdentifierHeader, e.ExtensionID)
	httpReq.Header.Set(consts.ExtensionErrorType, errorType)
	httpRes, err := e.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	if httpRes.StatusCode < 200 && httpRes.StatusCode >= 300 {
		return nil, fmt.Errorf("request failed with status %s", httpRes.Status)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(httpRes.Body)
	body, err := io.ReadAll(httpRes.Body)
	if err != nil {
		return nil, err
	}
	res := StatusResponse{}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// ExitError reports an error to the platform before exiting. Call it when you encounter an unexpected failure.
func (e *Client) ExitError(ctx context.Context, errorType string) (*StatusResponse, error) {
	const action = "/exit/error"
	url := e.baseURL + action

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set(consts.ExtensionIdentifierHeader, e.ExtensionID)
	httpReq.Header.Set(consts.ExtensionErrorType, errorType)
	httpRes, err := e.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	if httpRes.StatusCode < 200 && httpRes.StatusCode >= 300 {
		return nil, fmt.Errorf("request failed with status %s", httpRes.Status)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(httpRes.Body)
	body, err := io.ReadAll(httpRes.Body)
	if err != nil {
		return nil, err
	}
	res := StatusResponse{}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}
