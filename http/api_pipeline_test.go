package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
)

// go test -v -timeout 30s -run ^TestGetPipelineLines$ gitlab.jiagouyun.com/cloudcare-tools/datakit/http
func TestGetPipelineLines(t *testing.T) {
	cases := []struct {
		pattern, name string
		in            string
		out           []string
	}{
		{
			name:    "normal",
			pattern: "",
			in: `
127.0.0.1 - - [10/Feb/2022:18:45:09 +0800] "GET /server_status HTTP/1.1" 200 100 "-" "Go-http-client/1.1" "-"
127.0.0.1 - - [10/Feb/2022:18:45:19 +0800] "GET /server_status HTTP/1.1" 200 100 "-" "Go-http-client/1.1" "-"
2021/11/10 16:59:53 [error] 16393#0: *17 open() "/usr/local/Cellar/nginx/1.21.3/html/server_status" failed (2: No such file or directory), client: 127.0.0.1, server: localhost, request: "GET /server_status HTTP/1.1", host: "localhost:8080"
2021/11/10 17:00:03 [error] 16393#0: *18 open() "/usr/local/Cellar/nginx/1.21.3/html/server_status" failed (2: No such file or directory), client: 127.0.0.1, server: localhost, request: "GET /server_status HTTP/1.1", host: "localhost:8080"
2021/11/10 17:00:13 [error] 16393#0: *19 open() "/usr/local/Cellar/nginx/1.21.3/html/server_status" failed (2: No such file or directory), client: 127.0.0.1, server: localhost, request: "GET /server_status HTTP/1.1", host: "localhost:8080"
2021/11/10 17:00:23 [error] 16393#0: *20 open() "/usr/local/Cellar/nginx/1.21.3/html/server_status" failed (2: No such file or directory), client: 127.0.0.1, server: localhost, request: "GET /server_status HTTP/1.1", host: "localhost:8080"
2021/11/10 17:00:30 [notice] 16633#0: signal process started
2021/11/11 11:48:36 [error] 612#0: *3 open() "/usr/local/Cellar/nginx/1.21.3/html/favicon.ico" failed (2: No such file or directory), client: 127.0.0.1, server: localhost, request: "GET /favicon.ico HTTP/1.1", host: "localhost:8080", referrer: "http://localhost:8080/"
2021/11/29 11:09:35 [error] 621#0: *2 open() "/usr/local/Cellar/nginx/1.21.3/html/favicon.ico" failed (2: No such file or directory), client: 127.0.0.1, server: localhost, request: "GET /favicon.ico HTTP/1.1", host: "localhost:8080", referrer: "http://localhost:8080/server_status"
2021/11/30 19:07:29 [error] 596#0: *20 open() "/usr/local/Cellar/nginx/1.21.3/html/favicon.ico" failed (2: No such file or directory), client: 127.0.0.1, server: localhost, request: "GET /favicon.ico HTTP/1.1", host: "localhost:8080", referrer: "http://localhost:8080/"
2021/12/01 11:28:09 [error] 601#0: *2 open() "/usr/local/Cellar/nginx/1.21.3/html/favicon.ico" failed (2: No such file or directory), client: 10.100.65.39, server: localhost, request: "GET /favicon.ico HTTP/1.1", host: "10.100.65.39:8080", referrer: "http://10.100.65.39:8080/"
2022/02/10 18:17:44 [error] 616#0: *8 open() "/usr/local/Cellar/nginx/1.21.3/html/favicon.ico" failed (2: No such file or directory), client: 127.0.0.1, server: localhost, request: "GET /favicon.ico HTTP/1.1", host: "127.0.0.1:8080", referrer: "http://127.0.0.1:8080/"
			`,
			out: []string{
				`127.0.0.1 - - [10/Feb/2022:18:45:09 +0800] "GET /server_status HTTP/1.1" 200 100 "-" "Go-http-client/1.1" "-"`,
				`127.0.0.1 - - [10/Feb/2022:18:45:19 +0800] "GET /server_status HTTP/1.1" 200 100 "-" "Go-http-client/1.1" "-"`,
				`2021/11/10 16:59:53 [error] 16393#0: *17 open() "/usr/local/Cellar/nginx/1.21.3/html/server_status" failed (2: No such file or directory), client: 127.0.0.1, server: localhost, request: "GET /server_status HTTP/1.1", host: "localhost:8080"`,
				`2021/11/10 17:00:03 [error] 16393#0: *18 open() "/usr/local/Cellar/nginx/1.21.3/html/server_status" failed (2: No such file or directory), client: 127.0.0.1, server: localhost, request: "GET /server_status HTTP/1.1", host: "localhost:8080"`,
				`2021/11/10 17:00:13 [error] 16393#0: *19 open() "/usr/local/Cellar/nginx/1.21.3/html/server_status" failed (2: No such file or directory), client: 127.0.0.1, server: localhost, request: "GET /server_status HTTP/1.1", host: "localhost:8080"`,
				`2021/11/10 17:00:23 [error] 16393#0: *20 open() "/usr/local/Cellar/nginx/1.21.3/html/server_status" failed (2: No such file or directory), client: 127.0.0.1, server: localhost, request: "GET /server_status HTTP/1.1", host: "localhost:8080"`,
				`2021/11/10 17:00:30 [notice] 16633#0: signal process started`,
				`2021/11/11 11:48:36 [error] 612#0: *3 open() "/usr/local/Cellar/nginx/1.21.3/html/favicon.ico" failed (2: No such file or directory), client: 127.0.0.1, server: localhost, request: "GET /favicon.ico HTTP/1.1", host: "localhost:8080", referrer: "http://localhost:8080/"`,
				`2021/11/29 11:09:35 [error] 621#0: *2 open() "/usr/local/Cellar/nginx/1.21.3/html/favicon.ico" failed (2: No such file or directory), client: 127.0.0.1, server: localhost, request: "GET /favicon.ico HTTP/1.1", host: "localhost:8080", referrer: "http://localhost:8080/server_status"`,
				`2021/11/30 19:07:29 [error] 596#0: *20 open() "/usr/local/Cellar/nginx/1.21.3/html/favicon.ico" failed (2: No such file or directory), client: 127.0.0.1, server: localhost, request: "GET /favicon.ico HTTP/1.1", host: "localhost:8080", referrer: "http://localhost:8080/"`,
				`2021/12/01 11:28:09 [error] 601#0: *2 open() "/usr/local/Cellar/nginx/1.21.3/html/favicon.ico" failed (2: No such file or directory), client: 10.100.65.39, server: localhost, request: "GET /favicon.ico HTTP/1.1", host: "10.100.65.39:8080", referrer: "http://10.100.65.39:8080/"`,
				`2022/02/10 18:17:44 [error] 616#0: *8 open() "/usr/local/Cellar/nginx/1.21.3/html/favicon.ico" failed (2: No such file or directory), client: 127.0.0.1, server: localhost, request: "GET /favicon.ico HTTP/1.1", host: "127.0.0.1:8080", referrer: "http://127.0.0.1:8080/"`,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := getDataLines([]byte(tc.in), tc.pattern)
			assert.NoError(t, err, "getDataLines failed")
			for k, v := range out {
				s1 := strings.TrimRight(v, "\t")
				s2 := strings.TrimRight(s1, "\n")
				assert.Equal(t, tc.out[k], s2)
			}
		})
	}
}

// go test -v -timeout 30s -run ^TestGetDecodeData$ gitlab.jiagouyun.com/cloudcare-tools/datakit/http
func TestGetDecodeData(t *testing.T) {
	cases := []struct {
		pattern, name string
		in            *pipelineDebugRequest
		expectError   error
		expectData    string
	}{
		{
			name: "normal",
			in: &pipelineDebugRequest{
				Data: "aGVsbG8gd29ybGQ=",
			},
			expectData: "hello world",
		},
		{
			name: "gb18030",
			in: &pipelineDebugRequest{
				Data:   "1tDOxA==",
				Encode: "gb18030",
			},
			expectData: "中文",
		},
		{
			name: "gbk",
			in: &pipelineDebugRequest{
				Data:   "1tDOxA==",
				Encode: "gbk",
			},
			expectData: "中文",
		},
		{
			name: "UTF-8",
			in: &pipelineDebugRequest{
				Data:   "aGVsbG8gd29ybGQ=",
				Encode: "UTF8",
			},
			expectData: "hello world",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			bys, err := getDecodeData(tc.in)
			assert.Equal(t, tc.expectError, err, "getDecodeData found error: %v", err)
			assert.Equal(t, tc.expectData, string(bys), "getDecodeData not equal")
		})
	}
}

//------------------------------------------------------------------------------

// go test -v -timeout 30s -run ^TestApiDebugPipelineHandler$ gitlab.jiagouyun.com/cloudcare-tools/datakit/http
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
				Pipeline:  "Iy0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLSAgIOitpuWRiiAgIC0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0KIyDkuI3opoHkv67mlLnmnKzmlofku7bvvIzlpoLmnpzopoHmm7TmlrDvvIzor7fmi7fotJ3oh7PlhbblroPmlofku7bvvIzmnIDlpb3ku6Xmn5Dnp43liY3nvIDljLrliIbvvIzpgb/lhY3ph43lkK/lkI7ooqvopobnm5YKIy0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tCgphZGRfcGF0dGVybigiZGF0ZTIiLCAiJXtZRUFSfVsuL10le01PTlRITlVNfVsuL10le01PTlRIREFZfSAle1RJTUV9IikKCiMgYWNjZXNzIGxvZwpncm9rKF8sICIle05PVFNQQUNFOmNsaWVudF9pcH0gJXtOT1RTUEFDRTpodHRwX2lkZW50fSAle05PVFNQQUNFOmh0dHBfYXV0aH0gXFxbJXtIVFRQREFURTp0aW1lfVxcXSBcIiV7REFUQTpodHRwX21ldGhvZH0gJXtHUkVFRFlEQVRBOmh0dHBfdXJsfSBIVFRQLyV7TlVNQkVSOmh0dHBfdmVyc2lvbn1cIiAle0lOVDpzdGF0dXNfY29kZX0gJXtJTlQ6Ynl0ZXN9IikKCiMgYWNjZXNzIGxvZwphZGRfcGF0dGVybigiYWNjZXNzX2NvbW1vbiIsICIle05PVFNQQUNFOmNsaWVudF9pcH0gJXtOT1RTUEFDRTpodHRwX2lkZW50fSAle05PVFNQQUNFOmh0dHBfYXV0aH0gXFxbJXtIVFRQREFURTp0aW1lfVxcXSBcIiV7REFUQTpodHRwX21ldGhvZH0gJXtHUkVFRFlEQVRBOmh0dHBfdXJsfSBIVFRQLyV7TlVNQkVSOmh0dHBfdmVyc2lvbn1cIiAle0lOVDpzdGF0dXNfY29kZX0gJXtJTlQ6Ynl0ZXN9IikKZ3JvayhfLCAnJXthY2Nlc3NfY29tbW9ufSAiJXtOT1RTUEFDRTpyZWZlcnJlcn0iICIle0dSRUVEWURBVEE6YWdlbnR9JykKdXNlcl9hZ2VudChhZ2VudCkKCiMgZXJyb3IgbG9nCmdyb2soXywgIiV7ZGF0ZTI6dGltZX0gXFxbJXtMT0dMRVZFTDpzdGF0dXN9XFxdICV7R1JFRURZREFUQTptc2d9LCBjbGllbnQ6ICV7Tk9UU1BBQ0U6Y2xpZW50X2lwfSwgc2VydmVyOiAle05PVFNQQUNFOnNlcnZlcn0sIHJlcXVlc3Q6IFwiJXtEQVRBOmh0dHBfbWV0aG9kfSAle0dSRUVEWURBVEE6aHR0cF91cmx9IEhUVFAvJXtOVU1CRVI6aHR0cF92ZXJzaW9ufVwiLCAodXBzdHJlYW06IFwiJXtHUkVFRFlEQVRBOnVwc3RyZWFtfVwiLCApP2hvc3Q6IFwiJXtOT1RTUEFDRTppcF9vcl9ob3N0fVwiIikKZ3JvayhfLCAiJXtkYXRlMjp0aW1lfSBcXFsle0xPR0xFVkVMOnN0YXR1c31cXF0gJXtHUkVFRFlEQVRBOm1zZ30sIGNsaWVudDogJXtOT1RTUEFDRTpjbGllbnRfaXB9LCBzZXJ2ZXI6ICV7Tk9UU1BBQ0U6c2VydmVyfSwgcmVxdWVzdDogXCIle0dSRUVEWURBVEE6aHR0cF9tZXRob2R9ICV7R1JFRURZREFUQTpodHRwX3VybH0gSFRUUC8le05VTUJFUjpodHRwX3ZlcnNpb259XCIsIGhvc3Q6IFwiJXtOT1RTUEFDRTppcF9vcl9ob3N0fVwiIikKZ3JvayhfLCIle2RhdGUyOnRpbWV9IFxcWyV7TE9HTEVWRUw6c3RhdHVzfVxcXSAle0dSRUVEWURBVEE6bXNnfSIpCgpncm91cF9pbihzdGF0dXMsIFsid2FybiIsICJub3RpY2UiXSwgIndhcm5pbmciKQpncm91cF9pbihzdGF0dXMsIFsiZXJyb3IiLCAiY3JpdCIsICJhbGVydCIsICJlbWVyZyJdLCAiZXJyb3IiKQoKY2FzdChzdGF0dXNfY29kZSwgImludCIpCmNhc3QoYnl0ZXMsICJpbnQiKQoKZ3JvdXBfYmV0d2VlbihzdGF0dXNfY29kZSwgWzIwMCwyOTldLCAiT0siLCBzdGF0dXMpCmdyb3VwX2JldHdlZW4oc3RhdHVzX2NvZGUsIFszMDAsMzk5XSwgIm5vdGljZSIsIHN0YXR1cykKZ3JvdXBfYmV0d2VlbihzdGF0dXNfY29kZSwgWzQwMCw0OTldLCAid2FybmluZyIsIHN0YXR1cykKZ3JvdXBfYmV0d2VlbihzdGF0dXNfY29kZSwgWzUwMCw1OTldLCAiZXJyb3IiLCBzdGF0dXMpCgoKbnVsbGlmKGh0dHBfaWRlbnQsICItIikKbnVsbGlmKGh0dHBfYXV0aCwgIi0iKQpudWxsaWYodXBzdHJlYW0sICIiKQpkZWZhdWx0X3RpbWUodGltZSk=",
				Source:    "nginx",
				Service:   "",
				Category:  "logging",
				Data:      "MjAyMS8xMS8xMCAxNjo1OTo1MyBbZXJyb3JdIDE2MzkzIzA6ICoxNyBvcGVuKCkgIi91c3IvbG9jYWwvQ2VsbGFyL25naW54LzEuMjEuMy9odG1sL3NlcnZlcl9zdGF0dXMiIGZhaWxlZCAoMjogTm8gc3VjaCBmaWxlIG9yIGRpcmVjdG9yeSksIGNsaWVudDogMTI3LjAuMC4xLCBzZXJ2ZXI6IGxvY2FsaG9zdCwgcmVxdWVzdDogIkdFVCAvc2VydmVyX3N0YXR1cyBIVFRQLzEuMSIsIGhvc3Q6ICJsb2NhbGhvc3Q6ODA4MCI=",
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
				PLResults: []*pipelineDebugResult{
					{
						Measurement: "nginx",
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
						},
						Time:    1636534793,
						TimeNS:  0,
						Dropped: false,
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
				data, err := apiDebugPipelineHandler(w, req)
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
				assert.Equal(t, tc.expect.PLResults[0].Measurement, strings.TrimSpace(resp.PLResults[0].Measurement))
				assert.Equal(t, tc.expect.PLResults[0].Time, resp.PLResults[0].Time)
				assert.Equal(t, tc.expect.PLResults[0].TimeNS, resp.PLResults[0].TimeNS)
				assert.Equal(t, tc.expect.PLResults[0].Dropped, resp.PLResults[0].Dropped)

				for k := range resp.PLResults[0].Fields {
					assert.Equal(t, tc.expect.PLResults[0].Fields[k], resp.PLResults[0].Fields[k])
				}
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
// modified from: vendor/gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http/err.go

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
