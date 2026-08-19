// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package server

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

func TestSSOLoginHandler(t *testing.T) {
	withConsoleServer(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/v1/auth-token/get", r.URL.Path)
		require.Equal(t, "code-1", r.URL.Query().Get("auth_code"))
		_, _ = w.Write([]byte(`{"success":true,"content":{"token":"front-token","workspaceUUID":"workspace-1"}}`))
	})

	ctx, rec := newGinTestContext(http.MethodGet, "/sso/login?code=code-1", nil)
	ssoLoginHandler(ctx)
	require.Equal(t, "/dashboard", rec.Header().Get("Location"))
	require.Contains(t, rec.Header().Values("Set-Cookie")[0], cookieFrontToken)

	ctx, rec = newGinTestContext(http.MethodGet, "/sso/login", nil)
	ssoLoginHandler(ctx)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestSSOLoginHandlerConsoleErrors(t *testing.T) {
	withConsoleServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{bad-json`))
	})
	ctx, rec := newGinTestContext(http.MethodGet, "/sso/login?code=bad-json", nil)
	ssoLoginHandler(ctx)
	require.Equal(t, http.StatusInternalServerError, rec.Code)

	withConsoleServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"success":false,"errorCode":"bad.code"}`))
	})
	ctx, rec = newGinTestContext(http.MethodGet, "/sso/login?code=bad-code", nil)
	ssoLoginHandler(ctx)
	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestConsoleHandlers(t *testing.T) {
	withConsoleServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/workspace/query_list":
			_, _ = w.Write([]byte(`{"success":true,"content":["w1"]}`))
		case "/api/v1/workspace/change":
			_, _ = w.Write([]byte(`{"success":true}`))
		default:
			t.Fatalf("unexpected console path %s", r.URL.Path)
		}
	})

	ctx, rec := newGinTestContext(http.MethodGet, "/api/console/workspaceList", nil)
	ctx.Request.AddCookie(&http.Cookie{Name: cookieFrontToken, Value: "front-token"})
	ctx.Request.AddCookie(&http.Cookie{Name: cookieWorkspaceUUID, Value: "workspace-1"})
	consoleHandler(workspaceListPath)(ctx)
	require.Equal(t, http.StatusOK, rec.Code)
	require.JSONEq(t, `{"success":true,"content":["w1"]}`, rec.Body.String())

	ctx, rec = newGinTestContext(http.MethodGet, "/api/console/unknown", nil)
	consoleHandler("unknown")(ctx)
	require.False(t, decodeDCAResponse(t, rec.Body.String()).Success)

	ctx, rec = newGinTestContext(http.MethodPost, "/api/console/changeWorkspace", strings.NewReader(`{}`))
	ctx.Request.Header.Set("X-Workspace-Uuid", "workspace-2")
	ctx.Request.AddCookie(&http.Cookie{Name: cookieFrontToken, Value: "front-token"})
	ctx.Request.AddCookie(&http.Cookie{Name: cookieWorkspaceUUID, Value: "workspace-1"})
	consoleChangeWorkspaceHandler(ctx)
	resp := decodeDCAResponse(t, rec.Body.String())
	require.True(t, resp.Success)
	require.Contains(t, rec.Header().Values("Set-Cookie")[0], cookieWorkspaceUUID)

	ctx, rec = newGinTestContext(http.MethodPost, "/api/console/logout", nil)
	logoutHandler(ctx)
	resp = decodeDCAResponse(t, rec.Body.String())
	require.True(t, resp.Success)
	require.Len(t, rec.Header().Values("Set-Cookie"), 2)
}

func TestConsoleChangeWorkspaceRejectsConsoleFailure(t *testing.T) {
	withConsoleServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"success":false,"message":"denied"}`))
	})

	ctx, rec := newGinTestContext(http.MethodPost, "/api/console/changeWorkspace", strings.NewReader(`{}`))
	ctx.Request.Header.Set("X-Workspace-Uuid", "workspace-2")
	consoleChangeWorkspaceHandler(ctx)
	resp := decodeDCAResponse(t, rec.Body.String())
	require.False(t, resp.Success)
}

func TestConsoleChangeWorkspaceBadResponses(t *testing.T) {
	withConsoleServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{bad-json`))
	})
	ctx, rec := newGinTestContext(http.MethodPost, "/api/console/changeWorkspace", strings.NewReader(`{}`))
	ctx.Request.Header.Set("X-Workspace-Uuid", "workspace-2")
	consoleChangeWorkspaceHandler(ctx)
	resp := decodeDCAResponse(t, rec.Body.String())
	require.False(t, resp.Success)

	oldConsoleAPIURL := consoleAPIURL
	oldConsoleClient := consoleClient
	consoleAPIURL = "http://127.0.0.1:1"
	consoleClient = http.Client{}
	t.Cleanup(func() {
		consoleAPIURL = oldConsoleAPIURL
		consoleClient = oldConsoleClient
	})
	ctx, rec = newGinTestContext(http.MethodPost, "/api/console/changeWorkspace", strings.NewReader(`{}`))
	ctx.Request.Header.Set("X-Workspace-Uuid", "workspace-2")
	consoleChangeWorkspaceHandler(ctx)
	resp = decodeDCAResponse(t, rec.Body.String())
	require.False(t, resp.Success)
}

func TestConsoleRedirectAndLastVersion(t *testing.T) {
	oldConsoleWebURL := consoleWebURL
	consoleWebURL = "https://console.example.com"
	t.Cleanup(func() { consoleWebURL = oldConsoleWebURL })

	ctx, rec := newGinTestContext(http.MethodGet, "/console/dca", nil)
	consoleRedirectHandler(ctx)
	require.Equal(t, http.StatusTemporaryRedirect, rec.Code)
	require.Equal(t, "https://console.example.com/integration/dca", rec.Header().Get("Location"))

	oldStaticBaseURL := staticBaseURL
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/datakit/version", r.URL.Path)
		_, _ = w.Write([]byte(`{"version":"1.2.3","commit":"abc"}`))
	}))
	staticBaseURL = server.URL
	t.Cleanup(func() {
		staticBaseURL = oldStaticBaseURL
		server.Close()
	})

	ctx, rec = newGinTestContext(http.MethodGet, "/api/lastDatakitVersion", nil)
	getLastDatakitVersionHandler(ctx)
	resp := decodeDCAResponse(t, rec.Body.String())
	require.True(t, resp.Success)

	server.Close()
	ctx, rec = newGinTestContext(http.MethodGet, "/api/lastDatakitVersion", nil)
	getLastDatakitVersionHandler(ctx)
	resp = decodeDCAResponse(t, rec.Body.String())
	require.False(t, resp.Success)

	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{bad-json`))
	}))
	staticBaseURL = server.URL
	ctx, rec = newGinTestContext(http.MethodGet, "/api/lastDatakitVersion", nil)
	getLastDatakitVersionHandler(ctx)
	resp = decodeDCAResponse(t, rec.Body.String())
	require.False(t, resp.Success)
	server.Close()
}

func TestLastDatakitVersions(t *testing.T) {
	oldStaticBaseURL := staticBaseURL
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/datakit/version":
			_, _ = w.Write([]byte(`{"version":"1.94.1","commit":"v1-commit"}`))
		case "/datakit-v2/version":
			_, _ = w.Write([]byte(`{"version":"2.9.0","commit":"v2-commit"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	staticBaseURL = server.URL + "/"
	t.Cleanup(func() {
		staticBaseURL = oldStaticBaseURL
		server.Close()
	})

	ctx, rec := newGinTestContext(http.MethodGet, "/api/lastDatakitVersions", nil)
	getLastDatakitVersionsHandler(ctx)
	resp := decodeDCAResponse(t, rec.Body.String())
	require.True(t, resp.Success)
	content, ok := resp.Content.(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, "1.94.1", content["v1"].(map[string]interface{})["version"])
	require.Equal(t, "2.9.0", content["v2"].(map[string]interface{})["version"])
}

func TestLastDatakitVersionsAllowsPartialResponse(t *testing.T) {
	oldStaticBaseURL := staticBaseURL
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/datakit-v2/version" {
			_, _ = w.Write([]byte(`{"version":"2.9.0","commit":"v2-commit"}`))
			return
		}
		http.NotFound(w, r)
	}))
	staticBaseURL = server.URL
	t.Cleanup(func() {
		staticBaseURL = oldStaticBaseURL
		server.Close()
	})

	ctx, rec := newGinTestContext(http.MethodGet, "/api/lastDatakitVersions", nil)
	getLastDatakitVersionsHandler(ctx)
	resp := decodeDCAResponse(t, rec.Body.String())
	require.True(t, resp.Success)
	content, ok := resp.Content.(map[string]interface{})
	require.True(t, ok)
	require.NotContains(t, content, "v1")
	require.Equal(t, "2.9.0", content["v2"].(map[string]interface{})["version"])
}

func TestLastDatakitVersionsFailsWhenNoVersionIsAvailable(t *testing.T) {
	oldStaticBaseURL := staticBaseURL
	server := httptest.NewServer(http.NotFoundHandler())
	staticBaseURL = server.URL
	t.Cleanup(func() {
		staticBaseURL = oldStaticBaseURL
		server.Close()
	})

	ctx, rec := newGinTestContext(http.MethodGet, "/api/lastDatakitVersions", nil)
	getLastDatakitVersionsHandler(ctx)
	resp := decodeDCAResponse(t, rec.Body.String())
	require.False(t, resp.Success)
}

func TestStartReturnsDBInitError(t *testing.T) {
	oldDBPath := dbPath
	oldDatakitDB := datakitDB
	oldHTTPPort := dcaHTTPPort
	oldConsoleWebURL := consoleWebURL
	oldConsoleAPIURL := consoleAPIURL
	oldStaticBaseURL := staticBaseURL
	oldEnableTLS := enableTLS
	oldTLSCertFile := tlsCertFile
	oldTLSKeyFile := tlsKeyFile
	oldTransport := consoleClient.Transport
	t.Cleanup(func() {
		dbPath = oldDBPath
		datakitDB = oldDatakitDB
		dcaHTTPPort = oldHTTPPort
		consoleWebURL = oldConsoleWebURL
		consoleAPIURL = oldConsoleAPIURL
		staticBaseURL = oldStaticBaseURL
		enableTLS = oldEnableTLS
		tlsCertFile = oldTLSCertFile
		tlsKeyFile = oldTLSKeyFile
		consoleClient.Transport = oldTransport
	})

	datakitDB = NewDB()
	err := Start(&ServerOptions{
		HTTPPort:        "0",
		ConsoleWebURL:   "https://console.example.com/",
		ConsoleAPIURL:   "https://console-api.example.com/",
		StaticBaseURL:   "https://static.example.com/",
		ConsoleAPIProxy: "://bad-proxy",
		DBPath:          filepath.Join(t.TempDir(), "missing", "dca.db"),
		TLSEnable:       true,
		TLSCertFile:     "cert.pem",
		TLSKeyFile:      "key.pem",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to init DB")
}

func TestStartReturnsTLSRunError(t *testing.T) {
	oldDBPath := dbPath
	oldDatakitDB := datakitDB
	oldHTTPPort := dcaHTTPPort
	oldConsoleWebURL := consoleWebURL
	oldConsoleAPIURL := consoleAPIURL
	oldStaticBaseURL := staticBaseURL
	oldEnableTLS := enableTLS
	oldTLSCertFile := tlsCertFile
	oldTLSKeyFile := tlsKeyFile
	oldRegister := Manager.Register
	oldUnregister := Manager.Unregister
	oldClients := Manager.Clients
	oldWebsocketConns := Manager.WebsocketConns
	t.Cleanup(func() {
		dbPath = oldDBPath
		datakitDB = oldDatakitDB
		dcaHTTPPort = oldHTTPPort
		consoleWebURL = oldConsoleWebURL
		consoleAPIURL = oldConsoleAPIURL
		staticBaseURL = oldStaticBaseURL
		enableTLS = oldEnableTLS
		tlsCertFile = oldTLSCertFile
		tlsKeyFile = oldTLSKeyFile
		Manager.Register = oldRegister
		Manager.Unregister = oldUnregister
		Manager.Clients = oldClients
		Manager.WebsocketConns = oldWebsocketConns
	})

	datakitDB = NewDB()
	Manager.Register = make(chan *Client, 1)
	Manager.Unregister = make(chan *Client, 1)
	Manager.Clients = map[string]*Client{}
	Manager.WebsocketConns = map[string]chan *websocket.Conn{}
	err := Start(&ServerOptions{
		HTTPPort:                 "0",
		PromListen:               "127.0.0.1:0",
		ConsoleWebURL:            "https://console.example.com",
		ConsoleAPIURL:            "https://console-api.example.com",
		StaticBaseURL:            "https://static.example.com",
		UploadHostStatus:         true,
		UploadHostStatusInterval: time.Hour,
		DBPath:                   filepath.Join(t.TempDir(), "dca.db"),
		TLSEnable:                true,
		TLSCertFile:              filepath.Join(t.TempDir(), "missing-cert.pem"),
		TLSKeyFile:               filepath.Join(t.TempDir(), "missing-key.pem"),
	})
	require.Error(t, err)
}
