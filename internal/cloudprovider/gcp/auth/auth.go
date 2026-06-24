// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	dkhttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpcli"
)

const DefaultMetadataTokenURL = "http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/token"

type MetadataTokenTransport struct {
	base     http.RoundTripper
	tokenURL string
	now      func() time.Time

	mu          sync.Mutex
	accessToken string
	expiry      time.Time
}

type metadataToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

func NewHTTPClient() *http.Client {
	return &http.Client{
		Transport: NewMetadataTokenTransport(dkhttp.DefTransport(), metadataTokenURL(), time.Now),
		Timeout:   30 * time.Second,
	}
}

func NewMetadataTokenTransport(base http.RoundTripper, tokenURL string, now func() time.Time) *MetadataTokenTransport {
	if base == nil {
		base = http.DefaultTransport
	}
	if tokenURL == "" {
		tokenURL = DefaultMetadataTokenURL
	}
	if now == nil {
		now = time.Now
	}
	return &MetadataTokenTransport{
		base:     base,
		tokenURL: tokenURL,
		now:      now,
	}
}

func (t *MetadataTokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	token, err := t.token(req.Context())
	if err != nil {
		return nil, err
	}

	cloned := req.Clone(req.Context())
	cloned.Header = req.Header.Clone()
	cloned.Header.Set("Authorization", "Bearer "+token)
	return t.base.RoundTrip(cloned)
}

func (t *MetadataTokenTransport) token(ctx context.Context) (string, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := t.now()
	if t.accessToken != "" && now.Add(3*time.Minute+45*time.Second).Before(t.expiry) {
		return t.accessToken, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, t.tokenURL, nil)
	if err != nil {
		return "", fmt.Errorf("create gcp metadata token request: %w", err)
	}
	req.Header.Set("Metadata-Flavor", "Google")

	resp, err := t.base.RoundTrip(req)
	if err != nil {
		return "", fmt.Errorf("request gcp metadata token: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
		return "", fmt.Errorf("gcp metadata token returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var token metadataToken
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return "", fmt.Errorf("decode gcp metadata token: %w", err)
	}
	if token.AccessToken == "" || token.ExpiresIn <= 0 {
		return "", fmt.Errorf("gcp metadata token response is incomplete")
	}

	t.accessToken = token.AccessToken
	t.expiry = now.Add(time.Duration(token.ExpiresIn) * time.Second)
	return t.accessToken, nil
}

func metadataTokenURL() string {
	if host := os.Getenv("GCE_METADATA_HOST"); host != "" {
		return "http://" + strings.TrimRight(host, "/") +
			"/computeMetadata/v1/instance/service-accounts/default/token"
	}
	return DefaultMetadataTokenURL
}
