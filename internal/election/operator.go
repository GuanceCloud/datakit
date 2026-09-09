// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package election

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	ProviderDataway  Provider = "dataway"
	ProviderOperator Provider = "operator"

	operatorRequestTimeout  = 2 * time.Second
	maxOperatorResponseSize = 1 << 20
)

type Provider string

type operatorRequestError struct {
	kind       string
	statusCode int
	cause      error
}

func (e *operatorRequestError) Error() string {
	if e.statusCode != 0 {
		return fmt.Sprintf("operator election request failed: %s (HTTP status %d)", e.kind, e.statusCode)
	}
	return "operator election request failed: " + e.kind
}

func (e *operatorRequestError) Unwrap() error { return e.cause }

type operatorPuller struct {
	baseURL *url.URL
	token   string
	client  *http.Client
}

func NormalizeOperatorURL(rawURL string) (string, error) {
	normalized := strings.TrimSpace(rawURL)
	if normalized == "" {
		return "", nil
	}
	if !strings.Contains(normalized, "://") {
		normalized = "https://" + normalized
	}

	u, err := url.Parse(normalized)
	if err != nil {
		return "", errors.New("invalid Operator URL")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", errors.New("Operator URL scheme must be http or https")
	}
	if u.Host == "" || u.Hostname() == "" || u.Opaque != "" {
		return "", errors.New("Operator URL host is required")
	}
	if u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return "", errors.New("Operator URL must not contain credentials, query parameters, or a fragment")
	}
	if u.Scheme == "http" && !isLoopbackHost(u.Hostname()) {
		return "", errors.New("Operator URL must use https except on loopback")
	}

	u.Path = strings.TrimRight(u.Path, "/")
	u.RawPath = ""
	return u.String(), nil
}

func NewOperatorPuller(rawURL, token string, timeout time.Duration) (*operatorPuller, error) {
	normalized, err := NormalizeOperatorURL(rawURL)
	if err != nil {
		return nil, err
	}
	if normalized == "" {
		return nil, errors.New("Operator URL is required")
	}
	if token == "" {
		return nil, errors.New("workspace token is required for Operator election scope")
	}
	if timeout <= 0 {
		return nil, errors.New("Operator request timeout must be positive")
	}

	baseURL, err := url.Parse(normalized)
	if err != nil {
		return nil, fmt.Errorf("parse normalized Operator URL: %w", err)
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	// DataKit Operator serves its in-cluster endpoint with an Operator-managed
	// certificate. Keep the existing Operator-client behavior only for cluster
	// service names and loopback; externally addressed endpoints use normal
	// certificate verification.
	transport.TLSClientConfig = &tls.Config{ //nolint:gosec
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: isInClusterOperatorHost(baseURL.Hostname()),
	}

	return &operatorPuller{
		baseURL: baseURL,
		token:   token,
		client: &http.Client{
			Transport: transport,
			Timeout:   timeout,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}, nil
}

// SelectPuller probes once at startup. The returned puller never changes providers.
func SelectPuller(operatorURL, token string, dataway Puller) (Provider, Puller, error) {
	normalized, err := NormalizeOperatorURL(operatorURL)
	if err != nil {
		return "", nil, err
	}
	if dataway == nil {
		return "", nil, errors.New("DataWay election puller is required")
	}
	if normalized == "" {
		return ProviderDataway, dataway, nil
	}

	puller, err := NewOperatorPuller(normalized, token, operatorRequestTimeout)
	if err != nil {
		return "", nil, err
	}
	if err := puller.probe(); err != nil {
		puller.client.CloseIdleConnections()
		log.Warnf("election startup fallback provider=dataway reason=%s", err)
		return ProviderDataway, dataway, nil
	}
	return ProviderOperator, puller, nil
}

type operatorStatusResponse struct {
	Content struct {
		Status    string `json:"status"`
		ErrorCode string `json:"error_code"`
	} `json:"content"`
}

func (p *operatorPuller) probe() error {
	body, err := p.request(http.MethodGet, "/v1/dk-election/status", "", "", nil)
	if err != nil {
		return err
	}
	var status operatorStatusResponse
	if err := json.Unmarshal(body, &status); err != nil || status.Content.Status != "ready" {
		return &operatorRequestError{kind: "invalid_status_response"}
	}
	return nil
}

// ProviderForOperatorURL returns the process provider implied by the Operator URL.
func ProviderForOperatorURL(operatorURL string) Provider {
	if strings.TrimSpace(operatorURL) == "" {
		return ProviderDataway
	}
	return ProviderOperator
}

func (p *operatorPuller) Election(namespace, id string, reqBody io.Reader) ([]byte, error) {
	return p.request(http.MethodPost, "/v1/dk-election", namespace, id, reqBody)
}

func (p *operatorPuller) ElectionHeartbeat(namespace, id string, reqBody io.Reader) ([]byte, error) {
	return p.request(http.MethodPost, "/v1/dk-election/heartbeat", namespace, id, reqBody)
}

func (p *operatorPuller) request(method, endpoint, namespace, id string, reqBody io.Reader) ([]byte, error) {
	u := *p.baseURL
	u.Path = strings.TrimRight(u.Path, "/") + endpoint
	if method == http.MethodPost {
		query := u.Query()
		query.Set("token", p.token)
		query.Set("namespace", namespace)
		query.Set("id", id)
		u.RawQuery = query.Encode()
	}

	req, err := http.NewRequest(method, u.String(), reqBody)
	if err != nil {
		return nil, &operatorRequestError{kind: "invalid_request", cause: err}
	}
	req.Header.Set("Accept", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		cause := err
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			cause = urlErr.Err
		}
		kind := "transport_error"
		var netErr net.Error
		if errors.Is(cause, context.DeadlineExceeded) || (errors.As(cause, &netErr) && netErr.Timeout()) {
			kind = "timeout"
		}
		return nil, &operatorRequestError{kind: kind, cause: cause}
	}
	defer resp.Body.Close() //nolint:errcheck

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxOperatorResponseSize+1))
	if err != nil {
		return nil, &operatorRequestError{kind: "read_error", cause: err}
	}
	if len(body) > maxOperatorResponseSize {
		return nil, &operatorRequestError{kind: "response_too_large"}
	}
	if resp.StatusCode != http.StatusOK {
		requestErr := &operatorRequestError{kind: "http_error", statusCode: resp.StatusCode}
		var status operatorStatusResponse
		if json.Unmarshal(body, &status) == nil {
			// Only known protocol codes may enter logs; never echo arbitrary server text.
			switch status.Content.ErrorCode {
			case "rbac_forbidden", "cache_not_ready", "storage_unavailable":
				requestErr.kind = status.Content.ErrorCode
			}
		}
		return nil, requestErr
	}
	return body, nil
}

func isLoopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func isInClusterOperatorHost(host string) bool {
	host = strings.ToLower(strings.TrimSuffix(host, "."))
	return isLoopbackHost(host) || strings.HasSuffix(host, ".svc") ||
		strings.HasSuffix(host, ".svc.cluster.local")
}
