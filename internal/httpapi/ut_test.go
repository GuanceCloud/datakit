// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

func TestDisableWhitelist(t *testing.T) {
	// 测试用例：
	// 1. DisableWhitelist=false, PublicAPIs=["/v1/ping"] - 应该只允许/v1/ping访问
	// 2. DisableWhitelist=true, PublicAPIs=["/v1/ping"] - 应该允许所有API访问
	// 3. DisableWhitelist=false, PublicAPIs=[] - 应该只允许本地访问
	// 4. DisableWhitelist=true, PublicAPIs=[] - 应该允许所有API访问

	cases := []struct {
		name              string
		disableWhitelist  bool
		publicAPIs        []string
		requestPath       string
		requestFromLocal  bool
		expectedStatus    int
		expectedBodyMatch string
	}{
		{
			name:             "whitelist_enabled_with_allowed_api",
			disableWhitelist: false,
			publicAPIs:       []string{"/v1/ping"},
			requestPath:      "/v1/ping",
			requestFromLocal: false,
			expectedStatus:   http.StatusOK,
		},
		{
			name:             "whitelist_enabled_with_unregistered_api",
			disableWhitelist: false,
			publicAPIs:       []string{"/v1/ping"},
			requestPath:      "/v1/write/metric",
			requestFromLocal: false,
			expectedStatus:   http.StatusNotFound,
		},
		{
			name:             "whitelist_disabled_with_any_api",
			disableWhitelist: true,
			publicAPIs:       []string{"/v1/ping"},
			requestPath:      "/v1/write/metric",
			requestFromLocal: false,
			expectedStatus:   http.StatusNotFound, // 404因为路由未注册，但不是403
		},
		{
			name:             "whitelist_enabled_empty_list_external",
			disableWhitelist: false,
			publicAPIs:       []string{},
			requestPath:      "/v1/ping",
			requestFromLocal: false,
			expectedStatus:   http.StatusForbidden,
		},
		{
			name:             "whitelist_enabled_empty_list_local",
			disableWhitelist: false,
			publicAPIs:       []string{},
			requestPath:      "/v1/ping",
			requestFromLocal: true,
			expectedStatus:   http.StatusOK,
		},
		{
			name:             "whitelist_disabled_empty_list",
			disableWhitelist: true,
			publicAPIs:       []string{},
			requestPath:      "/v1/ping",
			requestFromLocal: false,
			expectedStatus:   http.StatusOK,
		},
	}

	t.Logf("开始执行测试: TestDisableWhitelist")
	t.Logf("测试用例数量: %d", len(cases))

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("开始执行测试用例: %s", tc.name)
			t.Logf("配置: DisableWhitelist=%v, PublicAPIs=%v, RequestPath=%s, FromLocal=%v",
				tc.disableWhitelist, tc.publicAPIs, tc.requestPath, tc.requestFromLocal)
			t.Logf("期望状态码: %d", tc.expectedStatus)

			// 创建HTTP服务器配置
			hs := defaultHTTPServerConf()
			hs.apiConfig.DisableWhitelist = tc.disableWhitelist
			hs.apiConfig.PublicAPIs = tc.publicAPIs

			t.Logf("HTTP服务器配置创建完成")

			// 设置路由器
			router := setupRouter(hs)
			t.Logf("路由器设置完成")

			// 创建测试请求
			req := httptest.NewRequest("GET", tc.requestPath, nil)

			// 如果是从本地访问，设置相应的头部
			if tc.requestFromLocal {
				req.RemoteAddr = "127.0.0.1:12345"
				t.Logf("设置为本地访问: %s", req.RemoteAddr)
			} else {
				req.RemoteAddr = "192.168.1.100:12345"
				t.Logf("设置为外部访问: %s", req.RemoteAddr)
			}

			// 创建响应记录器
			w := httptest.NewRecorder()
			t.Logf("响应记录器创建完成")

			// 执行请求
			t.Logf("开始执行HTTP请求")
			router.ServeHTTP(w, req)
			t.Logf("HTTP请求执行完成，状态码: %d", w.Code)
			t.Logf("响应体长度: %d 字节", len(w.Body.String()))

			// 验证结果
			assert.Equal(t, tc.expectedStatus, w.Code, "Status code should match expected")
			t.Logf("状态码验证通过: 期望=%d, 实际=%d", tc.expectedStatus, w.Code)

			// 如果期望特定的错误消息，验证响应体
			if tc.expectedBodyMatch != "" {
				assert.Contains(t, w.Body.String(), tc.expectedBodyMatch, "Response body should contain expected text")
				t.Logf("响应体内容验证通过")
			}

			t.Logf("测试用例 %s 执行完成", tc.name)
		})
	}

	t.Logf("所有测试用例执行完成")
}

func TestAPIWhiteListMiddlewareRouteExistence(t *testing.T) {
	cases := []struct {
		name           string
		publicAPIs     []string
		requestMethod  string
		requestPath    string
		expectedStatus int
	}{
		{
			name:           "unknown_route_not_public",
			publicAPIs:     []string{"/v1/ping"},
			requestMethod:  http.MethodGet,
			requestPath:    "/not-exist-api",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "unknown_route_public_exact_match",
			publicAPIs:     []string{"/not-exist-api"},
			requestMethod:  http.MethodGet,
			requestPath:    "/not-exist-api",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "wrong_method_on_existing_route",
			publicAPIs:     []string{"/v1/ping"},
			requestMethod:  http.MethodPost,
			requestPath:    "/v1/ping",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "registered_route_not_public",
			publicAPIs:     []string{"/v1/ping"},
			requestMethod:  http.MethodGet,
			requestPath:    "/metrics",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "registered_route_empty_whitelist",
			publicAPIs:     []string{},
			requestMethod:  http.MethodGet,
			requestPath:    "/metrics",
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			hs := defaultHTTPServerConf()
			hs.apiConfig.DisableWhitelist = false
			hs.apiConfig.PublicAPIs = tc.publicAPIs

			router := setupRouter(hs)

			req := httptest.NewRequest(tc.requestMethod, tc.requestPath, nil)
			req.RemoteAddr = "192.168.1.100:12345"

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}

func TestAPIWhiteListMiddlewareMatchesRequestPath(t *testing.T) {
	cases := []struct {
		name           string
		publicAPIs     []string
		expectedStatus int
	}{
		{
			name:           "match_raw_path",
			publicAPIs:     []string{"/v1/write/logging"},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "route_pattern_is_not_public_api_path",
			publicAPIs:     []string{"/v1/write/:category"},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "registered_route_not_matched",
			publicAPIs:     []string{"/v1/write/tracing"},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			router := gin.New()
			router.Use(apiWhiteListMiddleware(tc.publicAPIs))
			router.POST("/v1/write/:category", func(c *gin.Context) {
				c.Status(http.StatusNoContent)
			})

			req := httptest.NewRequest(http.MethodPost, "/v1/write/logging", nil)
			req.RemoteAddr = "192.168.1.100:12345"

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}

func TestInputHTTPRouteMatcherNoRoute(t *testing.T) {
	t.Cleanup(func() {
		cleanInputHTTPRouteMatchers()
		CleanHTTPHandler()
	})

	promtailMatcher := func() {
		RegInputHTTPRouteMatcher(func(method, path string) (string, bool) {
			if strings.EqualFold(method, http.MethodPost) &&
				path == "/v1/write/promtail" {
				return "promtail", true
			}
			return "", false
		})
	}

	cases := []struct {
		name           string
		declare        func()
		requestMethod  string
		requestPath    string
		publicAPIs     []string
		disable404Page bool
		expectedStatus int
		expectedBody   string
		unexpectedBody string
	}{
		{
			name:           "unknown_route_keeps_404",
			requestMethod:  http.MethodPost,
			requestPath:    "/not-exist-api",
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "exact_matcher_returns_input_not_enabled",
			declare: func() {
				RegInputHTTPRouteMatcher(func(method, path string) (string, bool) {
					if strings.EqualFold(method, http.MethodPost) &&
						path == "/profiling/v1/input" {
						return "profile", true
					}
					return "", false
				})
			},
			requestMethod:  http.MethodPost,
			requestPath:    "/profiling/v1/input",
			expectedStatus: http.StatusServiceUnavailable,
			expectedBody:   `input \"profile\" is not enabled for API \"/profiling/v1/input\"`,
			unexpectedBody: "Welcome to use DataKit",
		},
		{
			name: "exact_matcher_returns_input_not_enabled_when_404_page_disabled",
			declare: func() {
				RegInputHTTPRouteMatcher(func(method, path string) (string, bool) {
					if strings.EqualFold(method, http.MethodPost) &&
						path == "/profiling/v1/input" {
						return "profile", true
					}
					return "", false
				})
			},
			requestMethod:  http.MethodPost,
			requestPath:    "/profiling/v1/input",
			disable404Page: true,
			expectedStatus: http.StatusServiceUnavailable,
			expectedBody:   `input \"profile\" is not enabled for API \"/profiling/v1/input\"`,
		},
		{
			name: "wrong_method_on_input_route_keeps_404",
			declare: func() {
				RegInputHTTPRouteMatcher(func(method, path string) (string, bool) {
					if strings.EqualFold(method, http.MethodPost) &&
						path == "/profiling/v1/input" {
						return "profile", true
					}
					return "", false
				})
			},
			requestMethod:  http.MethodGet,
			requestPath:    "/profiling/v1/input",
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "dynamic_matcher_returns_input_not_enabled",
			declare: func() {
				RegInputHTTPRouteMatcher(func(method, path string) (string, bool) {
					if strings.EqualFold(method, http.MethodPost) &&
						strings.HasPrefix(path, "/v1/write/ploffload/") {
						return "ploffload", true
					}
					return "", false
				})
			},
			requestMethod:  http.MethodPost,
			requestPath:    "/v1/write/ploffload/logging",
			expectedStatus: http.StatusServiceUnavailable,
			expectedBody:   `input \"ploffload\" is not enabled for API \"/v1/write/ploffload/logging\"`,
		},
		{
			name:           "matcher_under_builtin_pattern_returns_input_not_enabled",
			declare:        promtailMatcher,
			requestMethod:  http.MethodPost,
			requestPath:    "/v1/write/promtail",
			publicAPIs:     []string{"/v1/write/promtail"},
			expectedStatus: http.StatusServiceUnavailable,
			expectedBody:   `input \"promtail\" is not enabled for API \"/v1/write/promtail\"`,
		},
		{
			name:           "matcher_under_builtin_pattern_returns_input_not_enabled_before_empty_whitelist_403",
			declare:        promtailMatcher,
			requestMethod:  http.MethodPost,
			requestPath:    "/v1/write/promtail",
			publicAPIs:     []string{},
			expectedStatus: http.StatusServiceUnavailable,
			expectedBody:   `input \"promtail\" is not enabled for API \"/v1/write/promtail\"`,
		},
		{
			name:           "matcher_under_builtin_pattern_returns_input_not_enabled_before_whitelist_miss_403",
			declare:        promtailMatcher,
			requestMethod:  http.MethodPost,
			requestPath:    "/v1/write/promtail",
			publicAPIs:     []string{"/v1/ping"},
			expectedStatus: http.StatusServiceUnavailable,
			expectedBody:   `input \"promtail\" is not enabled for API \"/v1/write/promtail\"`,
		},
		{
			name: "registered_route_wins_over_matcher",
			declare: func() {
				RegInputHTTPRouteMatcher(func(method, path string) (string, bool) {
					if strings.EqualFold(method, http.MethodPost) &&
						path == "/profiling/v1/input" {
						return "profile", true
					}
					return "", false
				})
				RegHTTPHandler(http.MethodPost, "/profiling/v1/input", func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusNoContent)
				})
			},
			requestMethod:  http.MethodPost,
			requestPath:    "/profiling/v1/input",
			expectedStatus: http.StatusNoContent,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cleanInputHTTPRouteMatchers()
			CleanHTTPHandler()

			if tc.declare != nil {
				tc.declare()
			}

			hs := defaultHTTPServerConf()
			if tc.publicAPIs != nil {
				hs.apiConfig.PublicAPIs = tc.publicAPIs
			}
			hs.apiConfig.Disable404Page = tc.disable404Page
			router := setupRouter(hs)

			req := httptest.NewRequest(tc.requestMethod, tc.requestPath, nil)
			req.RemoteAddr = "192.168.1.100:12345"

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			if tc.expectedBody != "" {
				assert.Contains(t, w.Body.String(), tc.expectedBody)
			}
			if tc.unexpectedBody != "" {
				assert.NotContains(t, w.Body.String(), tc.unexpectedBody)
			}
		})
	}
}

func TestNewWhiteListItemWithRegex(t *testing.T) {
	t.Logf("开始执行测试: TestNewWhiteListItemWithRegex")
	// 测试正则表达式白名单条目
	item := datakit.NewWhiteListItem("reg:^/v1/write/.*$")
	t.Logf("创建正则表达式白名单条目: reg:^/v1/write/.*$")

	assert.True(t, item.Match("/v1/write/metric"), "Should match regex pattern")
	t.Logf("匹配测试通过: /v1/write/metric")

	assert.True(t, item.Match("/v1/write/logging"), "Should match regex pattern")
	t.Logf("匹配测试通过: /v1/write/logging")

	assert.False(t, item.Match("/v1/ping"), "Should not match regex pattern")
	t.Logf("不匹配测试通过: /v1/ping")

	t.Logf("TestNewWhiteListItemWithRegex测试完成")
}

func TestNewWhiteListItemWithString(t *testing.T) {
	t.Logf("开始执行测试: TestNewWhiteListItemWithString")

	// 测试普通字符串白名单条目
	item := datakit.NewWhiteListItem("/v1/ping")
	t.Logf("创建普通字符串白名单条目: /v1/ping")

	assert.True(t, item.Match("/v1/ping"), "Should match exact path")
	t.Logf("匹配测试通过: /v1/ping")

	assert.False(t, item.Match("/v1/write/metric"), "Should not match different path")
	t.Logf("不匹配测试通过: /v1/write/metric")

	t.Logf("TestNewWhiteListItemWithString测试完成")
}
