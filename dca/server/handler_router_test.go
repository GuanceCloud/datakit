// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package server

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	tollbooth "github.com/didip/tollbooth/v6"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	ws "gitlab.jiagouyun.com/cloudcare-tools/datakit/dca/websocket"
)

func newGinTestContext(method, target string, body io.Reader) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest(method, target, body)
	ctx.Request = req
	return ctx, rec
}

func decodeDCAResponse(t *testing.T, body string) ws.DCAResponse {
	t.Helper()
	var resp ws.DCAResponse
	require.NoError(t, json.Unmarshal([]byte(body), &resp))
	return resp
}

func TestHandlerResponsesAndCookies(t *testing.T) {
	ctx, rec := newGinTestContext(http.MethodGet, "/api/test", nil)
	ctx.Request.AddCookie(&http.Cookie{Name: cookieWorkspaceUUID, Value: "workspace-1"})
	h := newHandler(ctx)

	require.Equal(t, "workspace-1", h.getCookie(cookieWorkspaceUUID))
	require.Empty(t, h.getCookie("missing"))

	h.success(map[string]string{"ok": "yes"})
	resp := decodeDCAResponse(t, rec.Body.String())
	require.True(t, resp.Success)
	require.EqualValues(t, 200, resp.Code)

	ctx, rec = newGinTestContext(http.MethodGet, "/api/test", nil)
	h = newHandler(ctx)
	h.fail(403, "", "denied")
	resp = decodeDCAResponse(t, rec.Body.String())
	require.False(t, resp.Success)
	require.Equal(t, errorCodeServerError, resp.ErrorCode)
	require.EqualValues(t, 403, resp.Code)

	ctx, rec = newGinTestContext(http.MethodGet, "/api/test", nil)
	h = newHandler(ctx)
	h.send(nil)
	resp = decodeDCAResponse(t, rec.Body.String())
	require.False(t, resp.Success)
	require.EqualValues(t, 500, resp.Code)
}

func TestHandlerDoRequestAndPipe(t *testing.T) {
	oldConsoleAPIURL := consoleAPIURL
	oldConsoleClient := consoleClient
	defer func() {
		consoleAPIURL = oldConsoleAPIURL
		consoleClient = oldConsoleClient
	}()

	requests := []http.Request{}
	console := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, *r)
		require.Equal(t, "front-token", r.Header.Get("X-FT-Auth-Token"))
		require.Equal(t, "workspace-1", r.Header.Get("X-Workspace-Uuid"))
		require.Equal(t, "application/json;charset=UTF-8", r.Header.Get("Content-Type"))
		_, _ = w.Write([]byte(`{"success":true}`))
	}))
	defer console.Close()

	consoleAPIURL = console.URL
	consoleClient = http.Client{Transport: console.Client().Transport}

	ctx, _ := newGinTestContext(http.MethodPost, "/api/test", strings.NewReader(`{"hello":"world"}`))
	ctx.Request.AddCookie(&http.Cookie{Name: cookieFrontToken, Value: "front-token"})
	ctx.Request.AddCookie(&http.Cookie{Name: cookieWorkspaceUUID, Value: "workspace-1"})
	h := newHandler(ctx)

	req, err := http.NewRequest(http.MethodPost, getConsoleAPIURL("/do"), strings.NewReader(`{}`))
	require.NoError(t, err)
	body, err := h.doRequest(req)
	require.NoError(t, err)
	require.JSONEq(t, `{"success":true}`, string(body))

	res, err := h.pipe(http.MethodPost, "/pipe", strings.NewReader(`{"x":1}`))
	require.NoError(t, err)
	require.NoError(t, res.Body.Close())
	require.Len(t, requests, 2)
	require.Equal(t, "/api/v1/do", requests[0].URL.Path)
	require.Equal(t, "/api/v1/pipe", requests[1].URL.Path)
}

func withConsoleServer(t *testing.T, handler http.HandlerFunc) string {
	t.Helper()
	oldConsoleAPIURL := consoleAPIURL
	oldConsoleClient := consoleClient
	server := httptest.NewServer(handler)
	consoleAPIURL = server.URL
	consoleClient = http.Client{Transport: server.Client().Transport}
	t.Cleanup(func() {
		consoleAPIURL = oldConsoleAPIURL
		consoleClient = oldConsoleClient
		server.Close()
	})
	return server.URL
}

func TestGetConsoleAPIURL(t *testing.T) {
	oldConsoleAPIURL := consoleAPIURL
	defer func() { consoleAPIURL = oldConsoleAPIURL }()

	consoleAPIURL = "https://console-api.example.com"
	require.Equal(t, "https://console-api.example.com/api/v1/path", getConsoleAPIURL("/path"))
	require.Equal(t, "https://console-api.example.com/api/v1path", getConsoleAPIURL("path"))
}

func TestLimiterCollectMetricsAndAuthPublicBranches(t *testing.T) {
	gin.SetMode(gin.TestMode)

	routerWithRoutes := gin.New()
	require.NoError(t, setupRouter(routerWithRoutes))

	ctx, rec := newGinTestContext(http.MethodOptions, "/api/datakit/list", nil)
	auth(ctx)
	require.Equal(t, http.StatusNoContent, rec.Code)
	require.True(t, ctx.IsAborted())

	ctx, rec = newGinTestContext(http.MethodGet, "/public/app.js", nil)
	called := false
	ctx.Set("called", &called)
	auth(ctx)
	require.False(t, ctx.IsAborted())
	require.Equal(t, http.StatusOK, rec.Code)

	ctx, rec = newGinTestContext(http.MethodGet, "/ws", nil)
	auth(ctx)
	require.False(t, ctx.IsAborted())
	require.Equal(t, http.StatusOK, rec.Code)

	router := gin.New()
	router.GET("/limited", limiterHandler(nlimiter), func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/limited", nil)
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	router = gin.New()
	router.GET("/metrics-test", collectRequestMetricsHandler(), func(c *gin.Context) { c.String(http.StatusCreated, "created") })
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/metrics-test", nil)
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	router = gin.New()
	tinyLimiter := tollbooth.NewLimiter(0.0001, nil)
	router.GET("/limited-deny", limiterHandler(tinyLimiter), func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/limited-deny", nil)
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/limited-deny", nil)
	router.ServeHTTP(w, req)
	resp := decodeDCAResponse(t, w.Body.String())
	require.False(t, resp.Success)
	require.EqualValues(t, http.StatusTooManyRequests, resp.Code)
}

func TestAuthPermissionBranches(t *testing.T) {
	db := withTestDatakitDB(t)
	dk := newTestDataKit("auth-conn")
	dk.RunInContainer = true
	require.NoError(t, db.Insert(dk))
	rows := []ws.DataKit{}
	require.NoError(t, db.Select("select * from datakit where conn_id=?", &rows, dk.ConnID))
	require.Len(t, rows, 1)

	withConsoleServer(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/v1/workspace/account/permissions", r.URL.Path)
		_, _ = w.Write([]byte(`{"success":true,"content":{"permissions":[]}}`))
	})

	router := gin.New()
	router.POST("/api/datakit/reload", auth, func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/datakit/reload?datakit_id="+rows[0].ID, nil)
	req.AddCookie(&http.Cookie{Name: cookieFrontToken, Value: "token"})
	req.AddCookie(&http.Cookie{Name: cookieWorkspaceUUID, Value: "workspace-1"})
	router.ServeHTTP(w, req)
	resp := decodeDCAResponse(t, w.Body.String())
	require.False(t, resp.Success)
	require.Equal(t, "permission.denied.container", resp.ErrorCode)

	dk.RunInContainer = false
	require.NoError(t, db.Update(dk))
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/datakit/reload?datakit_id="+rows[0].ID, nil)
	req.AddCookie(&http.Cookie{Name: cookieFrontToken, Value: "token"})
	req.AddCookie(&http.Cookie{Name: cookieWorkspaceUUID, Value: "workspace-1"})
	router.ServeHTTP(w, req)
	resp = decodeDCAResponse(t, w.Body.String())
	require.False(t, resp.Success)
	require.Equal(t, "permission.denied", resp.ErrorCode)
}

func TestAuthAllowsDcaPermissionAndRejectsConsoleFailure(t *testing.T) {
	db := withTestDatakitDB(t)
	dk := newTestDataKit("auth-allowed")
	require.NoError(t, db.Insert(dk))
	rows := []ws.DataKit{}
	require.NoError(t, db.Select("select * from datakit where conn_id=?", &rows, dk.ConnID))
	require.Len(t, rows, 1)

	withConsoleServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"success":true,"content":{"permissions":["dca.*"]}}`))
	})

	router := gin.New()
	router.POST("/api/datakit/reload", auth, func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/datakit/reload?datakit_id="+rows[0].ID, nil)
	req.AddCookie(&http.Cookie{Name: cookieFrontToken, Value: "token"})
	req.AddCookie(&http.Cookie{Name: cookieWorkspaceUUID, Value: "workspace-1"})
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "ok", w.Body.String())

	withConsoleServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"success":false,"message":"bad token","errorCode":"bad.token"}`))
	})
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/datakit/reload?datakit_id="+rows[0].ID, nil)
	req.AddCookie(&http.Cookie{Name: cookieFrontToken, Value: "token"})
	req.AddCookie(&http.Cookie{Name: cookieWorkspaceUUID, Value: "workspace-1"})
	router.ServeHTTP(w, req)
	resp := decodeDCAResponse(t, w.Body.String())
	require.False(t, resp.Success)
	require.EqualValues(t, 401, resp.Code)
}

func TestAuthRejectsPermissionServiceErrors(t *testing.T) {
	router := gin.New()
	router.POST("/api/datakit/reload", auth, func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	withConsoleServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{bad-json`))
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/datakit/reload", nil)
	req.AddCookie(&http.Cookie{Name: cookieFrontToken, Value: "token"})
	req.AddCookie(&http.Cookie{Name: cookieWorkspaceUUID, Value: "workspace-1"})
	router.ServeHTTP(w, req)
	resp := decodeDCAResponse(t, w.Body.String())
	require.False(t, resp.Success)
	require.Equal(t, "auth.failed", resp.ErrorCode)

	oldConsoleAPIURL := consoleAPIURL
	oldConsoleClient := consoleClient
	consoleAPIURL = "http://127.0.0.1:1"
	consoleClient = http.Client{}
	t.Cleanup(func() {
		consoleAPIURL = oldConsoleAPIURL
		consoleClient = oldConsoleClient
	})
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/datakit/reload", nil)
	req.AddCookie(&http.Cookie{Name: cookieFrontToken, Value: "token"})
	req.AddCookie(&http.Cookie{Name: cookieWorkspaceUUID, Value: "workspace-1"})
	router.ServeHTTP(w, req)
	resp = decodeDCAResponse(t, w.Body.String())
	require.False(t, resp.Success)
	require.Equal(t, "auth.failed", resp.ErrorCode)
}
