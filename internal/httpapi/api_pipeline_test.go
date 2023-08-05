// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpapi

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	uhttp "github.com/GuanceCloud/cliutils/network/http"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
)

// go test -v -timeout 30s -run ^TestGetDecodeData$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi
func TestGetDecodeData(t *testing.T) {
	cases := []struct {
		pattern, name string
		in            *pipelineDebugRequest
		expectError   error
		expectData    []string
	}{
		{
			name: "normal",
			in: &pipelineDebugRequest{
				Data: []string{"aGVsbG8gd29ybGQ="},
			},
			expectData: []string{"hello world"},
		},
		{
			name: "gb18030",
			in: &pipelineDebugRequest{
				Data:   []string{"1tDOxA=="},
				Encode: "gb18030",
			},
			expectData: []string{"中文"},
		},
		{
			name: "gbk",
			in: &pipelineDebugRequest{
				Data:   []string{"1tDOxA=="},
				Encode: "gbk",
			},
			expectData: []string{"中文"},
		},
		{
			name: "UTF-8",
			in: &pipelineDebugRequest{
				Data:   []string{"aGVsbG8gd29ybGQ="},
				Encode: "UTF8",
			},
			expectData: []string{"hello world"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pts, err := decodeDataAndConv2Point(point.Logging, "a", tc.in.Encode, tc.in.Data)

			var r []string
			for _, pt := range pts {
				fields, err := pt.Fields()
				if err != nil {
					t.Error(err)
					return
				}

				r = append(r, fields["message"].(string))
			}
			assert.Equal(t, tc.expectError, err, "getDecodeData found error: %v", err)
			assert.Equal(t, tc.expectData, r, "getDecodeData not equal")
		})
	}
}

//------------------------------------------------------------------------------

// go test -v -timeout 30s -run ^TestApiDebugPipelineHandler$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi
func TestApiDebugPipelineHandler(t *testing.T) {
	cases := []struct {
		name             string
		in               *pipelineDebugRequest
		expectStatusCode int
		expectHeader     http.Header
		hasResult        bool
		expect           *pipelineDebugResponse
	}{
		{
			name: "normal",
			in: &pipelineDebugRequest{
				Pipeline: map[string]map[string]string{
					"logging": scriptsForTest(),
				},
				Category:   "logging",
				ScriptName: "nginx",
				Data: []string{base64.StdEncoding.EncodeToString([]byte(
					`2021/11/10 16:59:53 [error] 16393#0: *17 open() "/usr/local/Cellar/nginx/1.21.3/html/server_status" failed` +
						` (2: No such file or directory), client: 127.0.0.1, server: localhost, request:` +
						` "GET /server_status HTTP/1.1", host: "localhost:8080"`))},
				Multiline: "",
				Encode:    "",
				Benchmark: true,
			},
			expectStatusCode: http.StatusOK,
			expectHeader: map[string][]string{
				"Content-Type": {"application/json"},
			},
			hasResult: true,
			expect: &pipelineDebugResponse{
				PLResults: []pipelineResult{
					{
						Point: &PlRetPoint{
							Dropped: false,
							Name:    "nginx",
							Fields: map[string]interface{}{
								"client_ip":    "127.0.0.1",
								"http_method":  "GET",
								"http_url":     "/server_status",
								"http_version": "1.1",
								"ip_or_host":   "localhost:8080",
								"message":      "2021/11/10 16:59:53 [error] 16393#0: *17 open() \"/usr/local/Cellar/nginx/1.21.3/html/server_status\" failed (2: No such file or directory), client: 127.0.0.1, server: localhost, request: \"GET /server_status HTTP/1.1\", host: \"localhost:8080\"",
								"msg":          "16393#0: *17 open() \"/usr/local/Cellar/nginx/1.21.3/html/server_status\" failed (2: No such file or directory), client: 127.0.0.1, server: localhost, request: \"GET /server_status HTTP/1.1\", host: \"localhost:8080\"",
								"server":       "localhost",
								"status":       "error",
								"b_p":          true,
							},
							Time:   time.Date(2021, 11, 10, 16, 59, 53, 0, time.Local).Unix(),
							TimeNS: 0,
						},
					},
				},
			},
		},

		{
			name: "normal-create-pts",
			in: &pipelineDebugRequest{
				Pipeline: map[string]map[string]string{
					"logging": scriptsForTestCreate(),
				},
				Category:   "logging",
				ScriptName: "nginx",
				Data: []string{base64.StdEncoding.EncodeToString([]byte(
					`2021/11/10 16:59:53 [error] 16393#0: *17 open() "/usr/local/Cellar/nginx/1.21.3/html/server_status" failed` +
						` (2: No such file or directory), client: 127.0.0.1, server: localhost, request:` +
						` "GET /server_status HTTP/1.1", host: "localhost:8080"`))},
				Multiline: "",
				Encode:    "",
				Benchmark: true,
			},
			expectStatusCode: http.StatusOK,
			expectHeader: map[string][]string{
				"Content-Type": {"application/json"},
			},
			hasResult: true,
			expect: &pipelineDebugResponse{
				PLResults: []pipelineResult{
					{
						Point: &PlRetPoint{
							Name: "nginx",
							Fields: map[string]interface{}{
								"client_ip":    "127.0.0.1",
								"http_method":  "GET",
								"http_url":     "/server_status",
								"http_version": "1.1",
								"ip_or_host":   "localhost:8080",
								"message":      "2021/11/10 16:59:53 [error] 16393#0: *17 open() \"/usr/local/Cellar/nginx/1.21.3/html/server_status\" failed (2: No such file or directory), client: 127.0.0.1, server: localhost, request: \"GET /server_status HTTP/1.1\", host: \"localhost:8080\"",
								"msg":          "16393#0: *17 open() \"/usr/local/Cellar/nginx/1.21.3/html/server_status\" failed (2: No such file or directory), client: 127.0.0.1, server: localhost, request: \"GET /server_status HTTP/1.1\", host: \"localhost:8080\"",
								"server":       "localhost",
								"status":       "error",
								"b_p":          true,
							},
							Time:   time.Date(2021, 11, 10, 16, 59, 53, 0, time.Local).Unix(),
							TimeNS: 0,
						},
						CreatePoint: []*PlRetPoint{
							{
								Dropped: false,
								Name:    "nginx",
								Tags:    map[string]string{},
								Fields: map[string]interface{}{
									"client_ip":    "127.0.0.1",
									"http_method":  "GET",
									"http_url":     "/server_status",
									"http_version": "1.1",
									"ip_or_host":   "localhost:8080",
									"message":      "2021/11/10 16:59:53 [error] 16393#0: *17 open() \"/usr/local/Cellar/nginx/1.21.3/html/server_status\" failed (2: No such file or directory), client: 127.0.0.1, server: localhost, request: \"GET /server_status HTTP/1.1\", host: \"localhost:8080\"",
									"msg":          "16393#0: *17 open() \"/usr/local/Cellar/nginx/1.21.3/html/server_status\" failed (2: No such file or directory), client: 127.0.0.1, server: localhost, request: \"GET /server_status HTTP/1.1\", host: \"localhost:8080\"",
									"server":       "localhost",
									"status":       "error",
									"b_p":          true,
								},
								Time:   time.Date(2021, 11, 10, 16, 59, 53, 0, time.Local).Unix(),
								TimeNS: 0,
							},
						},
					},
				},
			},
		},

		{
			name: "invalid category",
			in: &pipelineDebugRequest{
				Category: "else",
			},
			expectStatusCode: http.StatusBadRequest,
			expectHeader: map[string][]string{
				"Content-Type": {"application/json"},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			handler := func(w http.ResponseWriter, req *http.Request) {
				data, err := apiPipelineDebugHandler(w, req)
				var dd interface{}

				if err != nil {
					statusCode, contentType, errData := HttpErr(err)
					w.Header().Add("Content-Type", contentType)
					w.WriteHeader(statusCode)

					_, err = w.Write(errData)
					assert.NoError(t, err, "Write error failed!")
				} else {
					w.Header().Set("Content-Type", "application/json")

					switch d := data.(type) {
					case *pipelineDebugResponse:
						dd = d
					default:
						panic(reflect.TypeOf(d))
					}

					output, err := json.Marshal(dd)
					assert.NoError(t, err, "json.Marshal error")

					_, err = w.Write(output)
					assert.NoError(t, err, "Write output failed!")
				}
			}
			svr := httptest.NewServer(http.HandlerFunc(handler))
			defer func() {
				svr.Close()
			}()

			c := NewClient(svr.URL)
			statusCode, header, body, err := c.PipelineDebug(tc.in)
			assert.NoError(t, err, "PipelineDebug error")

			assert.Equal(t, tc.expectStatusCode, statusCode, "statusCode not equal")

			for k, v := range tc.expectHeader {
				assert.Equal(t, header[k], v,
					fmt.Sprintf("header not equal, left = %s, right = %s",
						header[k], v))
			}

			var resp pipelineDebugResponse
			err = json.Unmarshal(body, &resp)
			assert.NoError(t, err, "json.Unmarshal error")

			if tc.hasResult {
				assert.Equal(t, tc.expect.PLResults[0].Point.Name, strings.TrimSpace(resp.PLResults[0].Point.Name))
				assert.Equal(t, tc.expect.PLResults[0].Point.Time, resp.PLResults[0].Point.Time)
				assert.Equal(t, tc.expect.PLResults[0].Point.TimeNS, resp.PLResults[0].Point.TimeNS)
				assert.Equal(t, tc.expect.PLResults[0].Point.Dropped, resp.PLResults[0].Point.Dropped)
				for k := range resp.PLResults[0].Point.Fields {
					assert.Equal(t, tc.expect.PLResults[0].Point.Fields[k], resp.PLResults[0].Point.Fields[k], k)
				}
				assert.Equal(t, tc.expect.PLResults[0].CreatePoint, resp.PLResults[0].CreatePoint)
			}
		})
	}
}

type Client struct {
	url string
}

func NewClient(url string) Client {
	return Client{url}
}

func (c Client) PipelineDebug(in *pipelineDebugRequest) (int, http.Header, []byte, error) {
	reqURL := c.url + "/v1/pipeline/debug"
	bys, err := json.Marshal(in)
	if err != nil {
		return 0, nil, nil, err
	}

	res, err := http.Post(reqURL, "application/json", bytes.NewReader(bys))
	if err != nil {
		return 0, nil, nil, err
	}
	defer res.Body.Close() //nolint:errcheck
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return 0, nil, nil, err
	}

	return res.StatusCode, res.Header, body, nil
}

//------------------------------------------------------------------------------
// modified from: vendor/github.com/GuanceCloud/cliutils/network/http/err.go

type HttpError struct {
	ErrCode  string `json:"error_code,omitempty"`
	Err      error  `json:"-"`
	HttpCode int    `json:"-"`
}

func (he *HttpError) Error() string {
	if he.Err == nil {
		return ""
	} else {
		return he.Err.Error()
	}
}

func (he *HttpError) httpResp(format string, args ...interface{}) (int, string, []byte) {
	resp := &bodyResp{
		HttpError: he,
	}

	if args != nil {
		resp.Message = fmt.Sprintf(format, args...)
	}

	j, err := json.Marshal(&resp)
	if err != nil {
		panic(err)
	}

	return he.HttpCode, "application/json", j
}

type MsgError struct {
	*HttpError
	Fmt  string
	Args []interface{}
}

type bodyResp struct {
	*HttpError
	Message string      `json:"message,omitempty"`
	Content interface{} `json:"content,omitempty"`
}

func HttpErr(err error) (int, string, []byte) {
	switch me := err.(type) { //nolint:errorlint
	case *uhttp.HttpError:
		he := &HttpError{ErrCode: me.ErrCode, Err: me.Err, HttpCode: me.HttpCode}
		return he.httpResp("")
	case *uhttp.MsgError:
		he := &MsgError{HttpError: (*HttpError)(me.HttpError), Fmt: me.Fmt, Args: me.Args}
		if he.Args != nil {
			return he.HttpError.httpResp(he.Fmt, he.Args...)
		} else {
			panic("missing args")
		}
	default:
		panic("unknown error type")
	}
}

//------------------------------------------------------------------------------

func scriptsForTest() map[string]string {
	return map[string]string{
		"nginx": base64.StdEncoding.EncodeToString(
			[]byte(
				`#------------------------------------   警告   -------------------------------------
# 不要修改本文件，如果要更新，请拷贝至其它文件，最好以某种前缀区分，避免重启后被覆盖
#-----------------------------------------------------------------------------------

add_pattern("date2", "%{YEAR}[./]%{MONTHNUM}[./]%{MONTHDAY} %{TIME}")

# access log
grok(_, "%{NOTSPACE:client_ip} %{NOTSPACE:http_ident} %{NOTSPACE:http_auth} \\[%{HTTPDATE:time}\\] \"%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\" %{INT:status_code} %{INT:bytes}")

# access log
add_pattern("access_common", "%{NOTSPACE:client_ip} %{NOTSPACE:http_ident} %{NOTSPACE:http_auth} \\[%{HTTPDATE:time}\\] \"%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\" %{INT:status_code} %{INT:bytes}")
grok(_, '%{access_common} "%{NOTSPACE:referrer}" "%{GREEDYDATA:agent}')
user_agent(agent)

# error log
grok(_, "%{date2:time} \\[%{LOGLEVEL:status}\\] %{GREEDYDATA:msg}, client: %{NOTSPACE:client_ip}, server: %{NOTSPACE:server}, request: \"%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\", (upstream: \"%{GREEDYDATA:upstream}\", )?host: \"%{NOTSPACE:ip_or_host}\"")
grok(_, "%{date2:time} \\[%{LOGLEVEL:status}\\] %{GREEDYDATA:msg}, client: %{NOTSPACE:client_ip}, server: %{NOTSPACE:server}, request: \"%{GREEDYDATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\", host: \"%{NOTSPACE:ip_or_host}\"")
grok(_,"%{date2:time} \\[%{LOGLEVEL:status}\\] %{GREEDYDATA:msg}")

group_in(status, ["warn", "notice"], "warning")
group_in(status, ["error", "crit", "alert", "emerg"], "error")

cast(status_code, "int")
cast(bytes, "int")

group_between(status_code, [200,299], "OK", status)
group_between(status_code, [300,399], "notice", status)
group_between(status_code, [400,499], "warning", status)
group_between(status_code, [500,599], "error", status)


nullif(http_ident, "-")
nullif(http_auth, "-")
nullif(upstream, "")
default_time(time)
use("b.p")
		`)),
		"b": base64.StdEncoding.EncodeToString(
			[]byte(` add_key(b_p, true)
			for ;; {

			}
			add_key(b_p, false)
`)),
		"c": base64.StdEncoding.EncodeToString(
			[]byte(`use("b.p")`)),
	}
}

func scriptsForTestCreate() map[string]string {
	return map[string]string{
		"nginx": base64.StdEncoding.EncodeToString(
			[]byte(
				`#------------------------------------   警告   -------------------------------------
# 不要修改本文件，如果要更新，请拷贝至其它文件，最好以某种前缀区分，避免重启后被覆盖
#-----------------------------------------------------------------------------------

add_pattern("date2", "%{YEAR}[./]%{MONTHNUM}[./]%{MONTHDAY} %{TIME}")

# access log
grok(_, "%{NOTSPACE:client_ip} %{NOTSPACE:http_ident} %{NOTSPACE:http_auth} \\[%{HTTPDATE:time}\\] \"%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\" %{INT:status_code} %{INT:bytes}")

# access log
add_pattern("access_common", "%{NOTSPACE:client_ip} %{NOTSPACE:http_ident} %{NOTSPACE:http_auth} \\[%{HTTPDATE:time}\\] \"%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\" %{INT:status_code} %{INT:bytes}")
grok(_, '%{access_common} "%{NOTSPACE:referrer}" "%{GREEDYDATA:agent}')
user_agent(agent)

# error log
grok(_, "%{date2:time} \\[%{LOGLEVEL:status}\\] %{GREEDYDATA:msg}, client: %{NOTSPACE:client_ip}, server: %{NOTSPACE:server}, request: \"%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\", (upstream: \"%{GREEDYDATA:upstream}\", )?host: \"%{NOTSPACE:ip_or_host}\"")
grok(_, "%{date2:time} \\[%{LOGLEVEL:status}\\] %{GREEDYDATA:msg}, client: %{NOTSPACE:client_ip}, server: %{NOTSPACE:server}, request: \"%{GREEDYDATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\", host: \"%{NOTSPACE:ip_or_host}\"")
grok(_,"%{date2:time} \\[%{LOGLEVEL:status}\\] %{GREEDYDATA:msg}")

group_in(status, ["warn", "notice"], "warning")
group_in(status, ["error", "crit", "alert", "emerg"], "error")

cast(status_code, "int")
cast(bytes, "int")

group_between(status_code, [200,299], "OK", status)
group_between(status_code, [300,399], "notice", status)
group_between(status_code, [400,499], "warning", status)
group_between(status_code, [500,599], "error", status)


nullif(http_ident, "-")
nullif(http_auth, "-")
nullif(upstream, "")
default_time(time)
create_point("nginx", nil, {"message": _}, category="L", ts=1, after_use="d.p")
		`)),
		"b": base64.StdEncoding.EncodeToString(
			[]byte(` add_key(b_p, true)
			for ;; {

			}
			add_key(b_p, false)
`)),
		"d": base64.StdEncoding.EncodeToString(
			[]byte(
				`#------------------------------------   警告   -------------------------------------
# 不要修改本文件，如果要更新，请拷贝至其它文件，最好以某种前缀区分，避免重启后被覆盖
#-----------------------------------------------------------------------------------

add_pattern("date2", "%{YEAR}[./]%{MONTHNUM}[./]%{MONTHDAY} %{TIME}")

# access log
grok(_, "%{NOTSPACE:client_ip} %{NOTSPACE:http_ident} %{NOTSPACE:http_auth} \\[%{HTTPDATE:time}\\] \"%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\" %{INT:status_code} %{INT:bytes}")

# access log
add_pattern("access_common", "%{NOTSPACE:client_ip} %{NOTSPACE:http_ident} %{NOTSPACE:http_auth} \\[%{HTTPDATE:time}\\] \"%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\" %{INT:status_code} %{INT:bytes}")
grok(_, '%{access_common} "%{NOTSPACE:referrer}" "%{GREEDYDATA:agent}')
user_agent(agent)

# error log
grok(_, "%{date2:time} \\[%{LOGLEVEL:status}\\] %{GREEDYDATA:msg}, client: %{NOTSPACE:client_ip}, server: %{NOTSPACE:server}, request: \"%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\", (upstream: \"%{GREEDYDATA:upstream}\", )?host: \"%{NOTSPACE:ip_or_host}\"")
grok(_, "%{date2:time} \\[%{LOGLEVEL:status}\\] %{GREEDYDATA:msg}, client: %{NOTSPACE:client_ip}, server: %{NOTSPACE:server}, request: \"%{GREEDYDATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\", host: \"%{NOTSPACE:ip_or_host}\"")
grok(_,"%{date2:time} \\[%{LOGLEVEL:status}\\] %{GREEDYDATA:msg}")

group_in(status, ["warn", "notice"], "warning")
group_in(status, ["error", "crit", "alert", "emerg"], "error")

cast(status_code, "int")
cast(bytes, "int")

group_between(status_code, [200,299], "OK", status)
group_between(status_code, [300,399], "notice", status)
group_between(status_code, [400,499], "warning", status)
group_between(status_code, [500,599], "error", status)


nullif(http_ident, "-")
nullif(http_auth, "-")
nullif(upstream, "")
default_time(time)
use("b.p")
`)),
	}
}
