// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package election

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type recordingPuller struct {
	electionCalls  int
	heartbeatCalls int
}

func (p *recordingPuller) Election(_, _ string, _ io.Reader) ([]byte, error) {
	p.electionCalls++
	return nil, nil
}

func (p *recordingPuller) ElectionHeartbeat(_, _ string, _ io.Reader) ([]byte, error) {
	p.heartbeatCalls++
	return nil, nil
}

func TestNormalizeOperatorURL(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr bool
	}{
		{name: "empty", raw: "  ", want: ""},
		{name: "https", raw: " https://operator.example:443/ ", want: "https://operator.example:443"},
		{name: "service address", raw: "datakit-operator.datakit.svc:443/", want: "https://datakit-operator.datakit.svc:443"},
		{name: "base path", raw: "https://operator.example/base/", want: "https://operator.example/base"},
		{name: "loopback http", raw: "http://127.0.0.1:8080/", want: "http://127.0.0.1:8080"},
		{name: "remote http", raw: "http://operator.example", wantErr: true},
		{name: "userinfo", raw: "https://user:secret@operator.example", wantErr: true},
		{name: "query", raw: "https://operator.example?token=secret", wantErr: true},
		{name: "fragment", raw: "https://operator.example/#secret", wantErr: true},
		{name: "unsupported scheme", raw: "file:///tmp/operator.sock", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeOperatorURL(tt.raw)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected an error")
				}
				return
			}
			if err != nil {
				t.Fatalf("NormalizeOperatorURL() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("NormalizeOperatorURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNormalizeOperatorURLErrorDoesNotEchoInput(t *testing.T) {
	const secret = "tkn-secret-in-malformed-url"
	_, err := NormalizeOperatorURL("https://operator.example/%zz?token=" + secret)
	if err == nil {
		t.Fatal("expected malformed URL error")
	}
	if strings.Contains(err.Error(), secret) {
		t.Fatalf("URL validation error exposes input: %q", err)
	}
}

func TestOperatorPullerUsesCompatibleContractAndReusesConnection(t *testing.T) {
	var newConnections atomic.Int32
	requestPaths := make(chan string, 4)

	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", req.Method)
		}
		if got := req.URL.Query().Get("token"); got != "tkn-secret" {
			t.Errorf("token = %q, want %q", got, "tkn-secret")
		}
		if got := req.URL.Query().Get("namespace"); got != "production/a" {
			t.Errorf("namespace = %q, want %q", got, "production/a")
		}
		id := req.URL.Query().Get("id")
		status := StatusSuccess.String()
		holder := id
		if id == "datakit b" {
			status = StatusFail.String()
			holder = "datakit a"
		}
		requestPaths <- req.URL.Path
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]any{
			"content": map[string]any{
				"status":         status,
				"namespace":      req.URL.Query().Get("namespace"),
				"id":             id,
				"incumbency_id":  holder,
				"error_msg":      "",
				"interval":       3,
				"lease_duration": 30,
				"epoch":          7,
			},
		}); err != nil {
			t.Errorf("encode response: %v", err)
		}
	}))
	server.Config.ConnState = func(_ net.Conn, state http.ConnState) {
		if state == http.StateNew {
			newConnections.Add(1)
		}
	}
	server.StartTLS()
	t.Cleanup(server.Close)

	puller, err := NewOperatorPuller(server.URL+"/", "tkn-secret", 200*time.Millisecond)
	if err != nil {
		t.Fatalf("NewOperatorPuller() error = %v", err)
	}

	tests := []struct {
		call       func(string, string, io.Reader) ([]byte, error)
		id         string
		path       string
		wantStatus string
		wantHolder string
	}{
		{call: puller.Election, id: "datakit a", path: "/v1/dk-election", wantStatus: StatusSuccess.String(), wantHolder: "datakit a"},
		{call: puller.Election, id: "datakit b", path: "/v1/dk-election", wantStatus: StatusFail.String(), wantHolder: "datakit a"},
		{call: puller.ElectionHeartbeat, id: "datakit a", path: "/v1/dk-election/heartbeat", wantStatus: StatusSuccess.String(), wantHolder: "datakit a"},
		{call: puller.ElectionHeartbeat, id: "datakit b", path: "/v1/dk-election/heartbeat", wantStatus: StatusFail.String(), wantHolder: "datakit a"},
	}
	for _, tt := range tests {
		body, err := tt.call("production/a", tt.id, nil)
		if err != nil {
			t.Fatalf("operator request error = %v", err)
		}
		var response leaderElectionResult
		if err := json.Unmarshal(body, &response); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if response.Content.Status != tt.wantStatus || response.Content.Namespace != "production/a" ||
			response.Content.ID != tt.id || response.Content.IncumbencyID != tt.wantHolder ||
			response.Content.Interval != 3 || response.Content.LeaseDuration != 30 || response.Content.Epoch != 7 {
			t.Fatalf("response content = %+v", response.Content)
		}
		if got := <-requestPaths; got != tt.path {
			t.Fatalf("request path = %q, want %q", got, tt.path)
		}
	}
	if got := newConnections.Load(); got != 1 {
		t.Fatalf("new connections = %d, want 1", got)
	}
}

func TestOperatorPullerLimitsInsecureTLSCompatibilityToClusterHosts(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		wantInsecure bool
	}{
		{name: "cluster service", url: "https://datakit-operator.datakit.svc:443", wantInsecure: true},
		{name: "cluster local service", url: "https://datakit-operator.datakit.svc.cluster.local:443", wantInsecure: true},
		{name: "loopback test server", url: "https://127.0.0.1:443", wantInsecure: true},
		{name: "external endpoint", url: "https://operator.example:443", wantInsecure: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			puller, err := NewOperatorPuller(tt.url, "tkn-secret", 100*time.Millisecond)
			if err != nil {
				t.Fatalf("NewOperatorPuller() error = %v", err)
			}
			transport, ok := puller.client.Transport.(*http.Transport)
			if !ok {
				t.Fatalf("transport type = %T", puller.client.Transport)
			}
			if got := transport.TLSClientConfig.InsecureSkipVerify; got != tt.wantInsecure {
				t.Fatalf("InsecureSkipVerify = %t, want %t", got, tt.wantInsecure)
			}
		})
	}
}

func TestOperatorPullerErrorsAreBoundedAndRedacted(t *testing.T) {
	const token = "tkn-super-secret"

	t.Run("http status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "server accidentally echoed "+token, http.StatusServiceUnavailable)
		}))
		t.Cleanup(server.Close)

		puller, err := NewOperatorPuller(server.URL, token, 200*time.Millisecond)
		if err != nil {
			t.Fatalf("NewOperatorPuller() error = %v", err)
		}
		_, err = puller.Election("default", "datakit-a", nil)
		if err == nil {
			t.Fatal("expected an HTTP error")
		}
		if strings.Contains(err.Error(), token) || strings.Contains(err.Error(), server.URL) {
			t.Fatalf("error exposes credentials or URL: %q", err)
		}
		if !strings.Contains(err.Error(), "503") {
			t.Fatalf("error %q does not contain bounded status", err)
		}
	})

	t.Run("timeout", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		t.Cleanup(server.Close)

		puller, err := NewOperatorPuller(server.URL, token, 10*time.Millisecond)
		if err != nil {
			t.Fatalf("NewOperatorPuller() error = %v", err)
		}
		_, err = puller.Election("default", "datakit-a", nil)
		if err == nil {
			t.Fatal("expected a timeout")
		}
		if strings.Contains(err.Error(), token) || strings.Contains(err.Error(), server.URL) {
			t.Fatalf("error exposes credentials or URL: %q", err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.NotFoundHandler())
		t.Cleanup(server.Close)

		puller, err := NewOperatorPuller(server.URL, token, 200*time.Millisecond)
		if err != nil {
			t.Fatalf("NewOperatorPuller() error = %v", err)
		}
		_, err = puller.Election("default", "datakit-a", nil)
		if err == nil || !strings.Contains(err.Error(), "404") {
			t.Fatalf("error = %v, want bounded 404 error", err)
		}
	})

	t.Run("connection failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
		serverURL := server.URL
		server.Close()

		puller, err := NewOperatorPuller(serverURL, token, 200*time.Millisecond)
		if err != nil {
			t.Fatalf("NewOperatorPuller() error = %v", err)
		}
		_, err = puller.Election("default", "datakit-a", nil)
		if err == nil {
			t.Fatal("expected connection failure")
		}
		if strings.Contains(err.Error(), token) || strings.Contains(err.Error(), serverURL) {
			t.Fatalf("error exposes credentials or URL: %q", err)
		}
	})
}

func TestSelectPullerChoosesOnlyAtStartup(t *testing.T) {
	dataway := &recordingPuller{}
	provider, puller, err := SelectPuller("", "", dataway)
	assert.NoError(t, err)
	assert.Equal(t, ProviderDataway, provider)
	assert.Same(t, dataway, puller)

	for _, test := range []struct {
		name string
		code int
		body string
		want Provider
	}{
		{"ready", 200, `{"content":{"status":"ready"}}`, ProviderOperator},
		{"old version", 404, "not found", ProviderDataway},
		{"missing RBAC", 503, `{"content":{"status":"error","error_code":"rbac_forbidden"}}`, ProviderDataway},
		{"cache syncing", 503, `{"content":{"status":"error","error_code":"cache_not_ready"}}`, ProviderDataway},
		{"wrong response", 200, `{"status":"ok"}`, ProviderDataway},
		{"malformed response", 200, "invalid JSON", ProviderDataway},
		{"timeout", 0, "", ProviderDataway},
	} {
		t.Run(test.name, func(t *testing.T) {
			var probes atomic.Int32
			var changed atomic.Bool
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if req.Method != http.MethodGet {
					w.WriteHeader(http.StatusServiceUnavailable)
					return
				}
				probes.Add(1)
				assert.Equal(t, "/base/v1/dk-election/status", req.URL.Path)
				assert.Empty(t, req.URL.RawQuery, "capability probe must not send the token")
				code, body := test.code, test.body
				if changed.Load() {
					code, body = 200, `{"content":{"status":"ready"}}`
					if test.want == ProviderOperator {
						code = 503
					}
				}
				if code == 0 {
					<-req.Context().Done()
					return
				}
				w.WriteHeader(code)
				_, _ = io.WriteString(w, body)
			}))
			defer server.Close()
			dataway := &recordingPuller{}
			provider, puller, err := SelectPuller(server.URL+"/base", "tkn-secret", dataway)
			if !assert.NoError(t, err) {
				return
			}
			assert.Equal(t, test.want, provider)
			assert.EqualValues(t, 1, probes.Load())
			changed.Store(true)
			_, campaignErr := puller.Election("default", "a", nil)
			_, heartbeatErr := puller.ElectionHeartbeat("default", "a", nil)
			if test.want == ProviderOperator {
				assert.Error(t, campaignErr)
				assert.Error(t, heartbeatErr)
				assert.Zero(t, dataway.electionCalls+dataway.heartbeatCalls)
			} else {
				assert.NoError(t, campaignErr)
				assert.NoError(t, heartbeatErr)
				assert.Equal(t, 2, dataway.electionCalls+dataway.heartbeatCalls)
			}
			assert.EqualValues(t, 1, probes.Load(), "running clients must not reselect")
			restarted, _, err := SelectPuller(server.URL+"/base", "tkn-secret", dataway)
			assert.NoError(t, err)
			assert.NotEqual(t, test.want, restarted, "a new process must probe again")
			assert.EqualValues(t, 2, probes.Load())
		})
	}
}
