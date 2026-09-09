// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package pipeline

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/ptinput/funcs"
	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/plval"
)

// TestHTTPRequest, pipeline-go v1.4.3 ptinput/funcs/fn_http_request_test.go.
func TestJITHTTPRequestUpstreamOracle(t *testing.T) {
	var requests atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		body, err := io.ReadAll(io.LimitReader(r.Body, 64*1024))
		if err != nil {
			t.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if len(body) == 0 {
			body = []byte("hello")
		}
		_, _ = w.Write(body)
	}))
	defer server.Close()
	url := strconv.Quote(server.URL)
	runJSONOracle(t, []jsonOracleCase{
		{name: "0-test-post", source: fmt.Sprintf(`resp=http_request("POST",%s,{"extraHeader":"1"},{"a":"1"}); add_key(abc,resp["body"])`, url), input: "[]", key: "abc", want: `{"a":"1"}`},
		{name: "1-test-file", source: `resp=http_request("POST","file:///etc/",{"extraHeader":"1"},{"a":"1"}); add_key(abc,resp)`, input: "[]", key: "abc", want: nil},
		{name: "2-test-put", source: fmt.Sprintf(`resp=http_request("put",%s,{"extraHeader":"1"},{"a":"1"}); add_key(abc,resp["body"])`, url), input: "[]", key: "abc", want: `{"a":"1"}`},
		{name: "3-required-only", source: fmt.Sprintf(`resp=http_request("GET",%s); add_key(status_code,resp["status_code"])`, url), input: "[]", key: "status_code", want: int64(200)},
		{name: "4-prefix", source: fmt.Sprintf(`resp=http_request("POST",%s,{"extraHeader":"1"},{"a":"1"},"hr_"); add_key(abc,resp["hr_body"]); if resp["body"]!=nil { add_key(unexpected_unprefixed_key,true) }`, url), input: "[]", key: "abc", want: `{"a":"1"}`},
		{name: "5-prefix-status", source: fmt.Sprintf(`resp=http_request("GET",%s,{"extraHeader":"1"},nil,"hr_"); add_key(abc,resp["hr_status_code"])`, url), input: "[]", key: "abc", want: int64(200)},
		{name: "6-prefix-variable", source: fmt.Sprintf(`p="hr_"; resp=http_request("POST",%s,{"extraHeader":"1"},{"a":"1"},p); add_key(abc,resp["hr_body"])`, url), input: "[]", key: "abc", want: `{"a":"1"}`},
		{name: "7-named-prefix", source: fmt.Sprintf(`resp=http_request("GET",%s,prefix="hr_"); add_key(abc,resp["hr_body"]); if resp["body"]!=nil { add_key(unexpected_unprefixed_key,true) }`, url), input: "[]", key: "abc", want: "hello"},
	})
	if !t.Failed() && requests.Load() != 350 {
		t.Fatalf("upstream valid cases issued %d requests, want 350", requests.Load())
	}

	literalInvalid := fmt.Sprintf(`resp=http_request("POST",%s,{"extraHeader":"1"},{"a":"1"},123)`, url)
	if _, err := NewPlScriptSimple(point.Logging, "http-upstream.p", literalInvalid); err == nil {
		t.Fatal("pipeline-go accepted constant non-string prefix")
	}
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := pljit.NewRunner(path, "pipeline-go-1.4.3-datakit", 1)
	if err != nil {
		t.Fatal(err)
	}
	check := runner.Check(literalInvalid)
	_ = runner.Close()
	if check.Route != pljit.RoutePipelineGo || check.Reason != pljit.CheckReasonCompileError || check.Detail == "" {
		t.Fatalf("Rust accepted constant non-string prefix: %+v", check)
	}

	before := requests.Load()
	runJSONOracle(t, []jsonOracleCase{{
		name: "9-invalid-prefix-variable", input: "[]", fails: true,
		source: fmt.Sprintf(`p=123; resp=http_request("POST",%s,{"extraHeader":"1"},{"a":"1"},p)`, url),
	}})
	if requests.Load() != before {
		t.Fatalf("invalid variable prefix issued %d requests", requests.Load()-before)
	}
}

func TestJITHTTPRequestBodyOracle(t *testing.T) {
	var requests atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		body, err := io.ReadAll(io.LimitReader(r.Body, 64*1024))
		if err != nil {
			t.Errorf("read body: %v", err)
			w.WriteHeader(500)
			return
		}
		_, _ = w.Write(body)
	}))
	defer server.Close()
	var cases []jsonOracleCase
	for _, tc := range []struct {
		name, setup, body string
		want              string
	}{
		{"text", ``, `"plain"`, "plain"},
		{"raw", `add_key(raw,"/wBh"); b64dec(raw);`, `raw`, string([]byte{0xff, 0, 'a'})},
		{"integer", ``, `123`, "123"},
		{"boolean", ``, `true`, "true"},
		{"boolean-false", ``, `false`, "false"},
		{"nil", ``, `nil`, ""},
		{"float-large", ``, `100000000000000000000.0`, "100000000000000000000"},
		{"float-small", ``, `0.00000001`, "0.00000001"},
		{"float-fraction", ``, `1.25`, "1.25"},
		{"float-negative-zero", `add_key(v,"-0"); cast(v,"float");`, `v`, "-0"},
		{"float-nan", `add_key(v,"NaN"); cast(v,"float");`, `v`, "NaN"},
		{"float-infinity", `add_key(v,"+Inf"); cast(v,"float");`, `v`, "+Inf"},
		{"list-nan-encoding-failure", `add_key(v,"NaN"); cast(v,"float");`, `[v]`, ""},
		{"list-escape", ``, `["<", "中"]`, `["\u003c","中"]`},
		{"map-key-order", `u="http://example.invalid/?%0A=line&A=letter"; m=url_parse(u);`, `m["params"]`, `{"A":"letter","\n":"line"}`},
		{"map-raw-key", `u="http://example.invalid/?%FF=one"; m=url_parse(u);`, `m["params"]`, `{"\ufffd":"one"}`},
		// Equal values make duplicate encoded-key ordering irrelevant while
		// still proving both original keys survive until wire serialization.
		{"map-distinct-raw-keys", `u="http://example.invalid/?%FF=one&%FE=one"; m=url_parse(u);`, `m["params"]`, `{"\ufffd":"one","\ufffd":"one"}`},
	} {
		cases = append(cases, jsonOracleCase{name: tc.name, input: `{}`,
			source: tc.setup + fmt.Sprintf(`r=http_request("POST", %q, body=%s); add_key(result,r["body"])`, server.URL, tc.body),
			key:    "result", want: tc.want})
	}
	runJSONOracle(t, cases)
	if !t.Failed() && requests.Load() != int64(len(cases)*50) {
		t.Fatalf("unexpected request count %d", requests.Load())
	}
}

func TestJITHTTPHeadersOracle(t *testing.T) {
	var requests atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		_, _ = w.Write([]byte(r.Header.Get("X-Test")))
	}))
	defer server.Close()
	var cases []jsonOracleCase
	for _, tc := range []struct {
		name, headers, want string
		fails               bool
	}{
		{"text", `{"x-test":"yes"}`, "yes", false},
		{"empty", `{"X-Test":""}`, "", false},
		{"integer-ignored", `{"X-Test":12}`, "", false},
		{"list-ignored", `{"X-Test":["a"]}`, "", false},
		{"non-map-string", `"text"`, "", true},
		{"non-map-list", `[]`, "", true},
		{"explicit-nil", `nil`, "", true},
	} {
		key := "result"
		if tc.fails {
			key = ""
		}
		cases = append(cases, jsonOracleCase{name: tc.name, input: `{}`, key: key, want: tc.want, fails: tc.fails,
			source: fmt.Sprintf(`h=%s; r=http_request("GET",%q,headers=h); add_key(result,r["body"])`, tc.headers, server.URL)})
	}
	for _, headers := range []string{`"bad"`, `headers="bad"`} {
		cases = append(cases, jsonOracleCase{name: "invalid-before-body/" + headers, input: `{}`, fails: true,
			source: fmt.Sprintf(`r=http_request("GET",%q,%s,body=add_key(unexpected,true))`, server.URL, headers)})
	}
	cases = append(cases, jsonOracleCase{name: "reordered-body-before-invalid-headers", input: `{}`, fails: true,
		source: fmt.Sprintf(`r=http_request("GET",%q,body=add_key(unexpected,true),headers="bad")`, server.URL)})
	cases = append(cases, jsonOracleCase{name: "invalid-prefix-before-headers", input: `{}`, fails: true,
		source: fmt.Sprintf(`p=123; r=http_request("GET",%q,headers=add_key(unexpected,true),prefix=p)`, server.URL)})
	runJSONOracle(t, cases)
	if !t.Failed() && requests.Load() != 200 {
		t.Fatalf("requests=%d want=200", requests.Load())
	}
}

func TestJITHTTPURLControlOracle(t *testing.T) {
	var requests atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.WriteHeader(204)
	}))
	defer server.Close()
	var cases []jsonOracleCase
	for name, url := range map[string]string{
		"leading-newline": "\n" + server.URL,
		"leading-space":   " " + server.URL,
		"embedded-tab":    strings.Replace(server.URL, "http:", "http:\t", 1),
		"trailing-cr":     server.URL + "/\r",
		"embedded-del":    server.URL + "/a\x7fb",
	} {
		cases = append(cases, jsonOracleCase{name: name, input: `{}`, key: "result", want: true,
			source: `json(_,url,trim_space=false); r=http_request("GET",url,body=add_key(unexpected,true)); add_key(result,r == nil)`})
		encoded, err := json.Marshal(map[string]string{"url": url})
		if err != nil {
			t.Fatal(err)
		}
		cases[len(cases)-1].input = string(encoded)
	}
	runJSONOracle(t, cases)
	if requests.Load() != 0 {
		t.Errorf("invalid URLs caused %d requests", requests.Load())
	}
}

func TestJITHTTPRedirectOracle(t *testing.T) {
	for _, tc := range []struct {
		name     string
		status   int
		location bool
		want     string
		follow   bool
	}{
		{"post-301", 301, true, "GET|", true},
		{"post-302", 302, true, "GET|", true},
		{"post-303", 303, true, "GET|", true},
		{"post-307", 307, true, "POST|payload", true},
		{"post-308", 308, true, "POST|payload", true},
		{"get-301", 301, true, "GET|", true},
		{"get-302", 302, true, "GET|", true},
		{"get-303", 303, true, "GET|", true},
		{"get-307", 307, true, "GET|payload", true},
		{"get-308", 308, true, "GET|payload", true},
		{"terminal-300", 300, true, "terminal", false},
		{"terminal-304", 304, true, "", false},
		{"terminal-305", 305, true, "terminal", false},
		{"missing-location", 302, false, "terminal", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var requests atomic.Int64
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				if r.URL.Path == "/start" {
					if tc.location {
						w.Header().Set("Location", "/finish")
					}
					w.WriteHeader(tc.status)
					if tc.status != 304 {
						_, _ = w.Write([]byte("terminal"))
					}
					return
				}
				body, err := io.ReadAll(io.LimitReader(r.Body, 4096))
				if err != nil {
					t.Errorf("read body: %v", err)
					return
				}
				_, _ = fmt.Fprintf(w, "%s|%s", r.Method, body)
			}))
			defer server.Close()
			method := "POST"
			if strings.HasPrefix(tc.name, "get-") {
				method = "GET"
			}
			runJSONOracle(t, []jsonOracleCase{{name: tc.name, input: `{}`,
				source: fmt.Sprintf(`r=http_request(%q, %q, body="payload"); add_key(result,r["body"]); add_key(code,r["status_code"])`, method, server.URL+"/start"),
				key:    "result", want: tc.want}})
			want := int64(50)
			if tc.follow {
				want = 100
			}
			if !t.Failed() && requests.Load() != want {
				t.Fatalf("requests %d want %d", requests.Load(), want)
			}
		})
	}
}

func TestJITHTTPRedirectLimitOracle(t *testing.T) {
	for _, hops := range []int{3, 4, 9, 10} {
		t.Run(fmt.Sprint(hops), func(t *testing.T) {
			var requests atomic.Int64
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				step, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/"))
				if err != nil {
					t.Errorf("step: %v", err)
					w.WriteHeader(400)
					return
				}
				if step < hops {
					w.Header().Set("Location", fmt.Sprintf("/%d", step+1))
					w.WriteHeader(http.StatusFound)
					return
				}
				_, _ = w.Write([]byte("done"))
			}))
			defer server.Close()
			runJSONOracle(t, []jsonOracleCase{{name: "limit", input: `{}`,
				source: fmt.Sprintf(`r=http_request("GET", %q); add_key(result,r == nil)`, server.URL+"/0"),
				key:    "result", want: hops >= 10}})
			if !t.Failed() {
				want := int64(50 * min(hops+1, 10))
				if got := requests.Load(); got != want {
					t.Fatalf("requests=%d want=%d", got, want)
				}
			}
		})
	}
}

// Keep each policy in a fresh process so the matrix also detects unrelated
// package-global HTTP state. ReplaceNetFilter itself now atomically replaces its
// immutable policy; the in-process reload test below pins that contract.
func TestJITHTTPPolicyMatrixOracle(t *testing.T) {
	const childKey = "DATAKIT_TEST_JIT_HTTP_POLICY"
	type scenario struct {
		Name    string
		Policy  pljit.HTTPServiceConfig
		Blocked bool
	}
	if raw := os.Getenv(childKey); raw != "" {
		var tc scenario
		if err := json.Unmarshal([]byte(raw), &tc); err != nil {
			t.Fatal(err)
		}
		funcs.ReplaceNetFilter(tc.Policy.DisableInternalNetwork, tc.Policy.CIDRWhitelist, tc.Policy.HostWhitelist)
		var requests atomic.Int64
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requests.Add(1)
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()
		cases := []jsonOracleCase{{
			name: tc.Name, input: `{}`,
			source: fmt.Sprintf(`r=http_request("GET", %q); add_key(result,r == nil)`, server.URL),
			key:    "result", want: tc.Blocked,
		}}
		if tc.Blocked {
			cases = append(cases,
				jsonOracleCase{name: "blocked-skips-headers", input: `{}`, key: "result", want: true,
					source: fmt.Sprintf(`r=http_request("GET",%q,headers=add_key(unexpected,true)); add_key(result,r == nil)`, server.URL)},
				jsonOracleCase{name: "blocked-skips-body", input: `{}`, key: "result", want: true,
					source: fmt.Sprintf(`r=http_request("GET",%q,body=add_key(unexpected,true)); add_key(result,r == nil)`, server.URL)},
			)
		}
		runJSONOracleWithServices(t, cases, pljit.ServiceConfig{Version: 1, HTTP: tc.Policy})
		wantRequests := int64(50)
		if tc.Blocked {
			wantRequests = 0
		}
		if got := requests.Load(); got != wantRequests {
			t.Errorf("requests=%d want=%d", got, wantRequests)
		}
		return
	}
	for _, tc := range []scenario{
		{Name: "default-allow"},
		{Name: "private-deny", Policy: pljit.HTTPServiceConfig{DisableInternalNetwork: true}, Blocked: true},
		{Name: "cidr-allow", Policy: pljit.HTTPServiceConfig{DisableInternalNetwork: true, CIDRWhitelist: []string{"127.0.0.0/8"}}},
		{Name: "cidr-deny", Policy: pljit.HTTPServiceConfig{CIDRWhitelist: []string{"192.0.2.0/24"}}, Blocked: true},
		{Name: "host-allow", Policy: pljit.HTTPServiceConfig{DisableInternalNetwork: true, HostWhitelist: []string{"127.0.0.1"}}},
		{Name: "host-deny", Policy: pljit.HTTPServiceConfig{HostWhitelist: []string{"example.invalid"}}, Blocked: true},
		{Name: "host-whitespace-deny", Policy: pljit.HTTPServiceConfig{DisableInternalNetwork: true, HostWhitelist: []string{" 127.0.0.1 "}}, Blocked: true},
		{Name: "whitelist-union", Policy: pljit.HTTPServiceConfig{DisableInternalNetwork: true, HostWhitelist: []string{"example.invalid"}, CIDRWhitelist: []string{"127.0.0.0/8"}}},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			payload, err := json.Marshal(tc)
			if err != nil {
				t.Fatal(err)
			}
			binary, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			command := exec.Command(binary, "-test.run=^TestJITHTTPPolicyMatrixOracle$", "-test.timeout=30s")
			command.Env = append(os.Environ(), childKey+"="+string(payload))
			if output, err := command.CombinedOutput(); err != nil {
				t.Fatalf("policy subprocess: %v\n%s", err, output)
			}
		})
	}
}

func TestJITHTTPPolicyReloadReplacesPreviousWhitelist(t *testing.T) {
	var requests atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	t.Cleanup(func() { funcs.ReplaceNetFilter(false, nil, nil) })

	run := func(name string, policy pljit.HTTPServiceConfig, blocked bool) {
		t.Helper()
		funcs.ReplaceNetFilter(policy.DisableInternalNetwork, policy.CIDRWhitelist, policy.HostWhitelist)
		before := requests.Load()
		runJSONOracleWithServices(t, []jsonOracleCase{{
			name: name, input: `{}`,
			source: fmt.Sprintf(`r=http_request("GET",%q); add_key(result,r == nil)`, server.URL),
			key:    "result", want: blocked,
		}}, pljit.ServiceConfig{Version: 1, HTTP: policy})
		wantDelta := int64(50)
		if blocked {
			wantDelta = 0
		}
		if got := requests.Load() - before; got != wantDelta {
			t.Fatalf("%s requests=%d want=%d", name, got, wantDelta)
		}
	}

	run("allow-loopback", pljit.HTTPServiceConfig{
		DisableInternalNetwork: true,
		HostWhitelist:          []string{"127.0.0.1"},
	}, false)
	run("replace-with-deny", pljit.HTTPServiceConfig{
		DisableInternalNetwork: true,
		HostWhitelist:          []string{"example.invalid"},
	}, true)
	run("clear-policy", pljit.HTTPServiceConfig{}, false)
}

// Source-derived oracle: pipeline-go v1.4.3 HTTPRequest returns string(body).
// Exercise the real network and release library, not a mocked Rust service.
// This also exposes differences in the policy passed to the two engines.
func TestJITHTTPResponseBytesOracle(t *testing.T) {
	var requests atomic.Int64
	body := []byte{0xff, 0, 'a'}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write(body)
	}))
	defer server.Close()
	t.Cleanup(func() { t.Logf("actual HTTP requests (both engines): %d", requests.Load()) })
	runJSONOracle(t, []jsonOracleCase{{
		name:   "raw-response",
		input:  `{}`,
		source: fmt.Sprintf(`r=http_request("GET", %q); add_key(result,r["body"]); add_key(code,r["status_code"]); add_key(length,strlen(result)); add_key(digest,hash(result,"sha256"))`, server.URL),
		key:    "result", want: string(body),
	}})
	if !t.Failed() && requests.Load() != 50 {
		t.Fatalf("expected 25 Go + 25 JIT requests, got %d", requests.Load())
	}
}

// Use DataKit's actual disable hook, restoring the registry after the serial test.
// Verify both no network requests and no argument evaluation when disabled.
func TestJITHTTPDisabledFunctionOracle(t *testing.T) {
	original := funcs.FuncsMap["http_request"]
	plval.DisableExternalRequestsFunc()
	t.Cleanup(func() {
		plval.EnableExternalRequestsFunc()
		funcs.FuncsMap["http_request"] = original
	})
	var requests atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	runJSONOracleWithServices(t, []jsonOracleCase{{
		name: "disabled-returns-nil", input: `{}`,
		source: fmt.Sprintf(`r=http_request("GET", %q); add_key(result,r == nil)`, server.URL),
		key:    "result", want: true,
	}, {
		name: "disabled-skips-error", input: `{}`,
		source: `r=http_request(1/divisor, "http://127.0.0.1/"); add_key(result,r == nil)`,
		key:    "result", want: true,
	}, {
		name: "disabled-skips-side-effect", input: `{}`,
		source: fmt.Sprintf(`r=http_request("GET", %q, body=add_key(unexpected,true)); add_key(result,r == nil)`, server.URL),
		key:    "result", want: true,
	}}, pljit.ServiceConfig{Version: 1, DisableHTTPRequestFunc: true})
	if requests.Load() != 0 {
		t.Errorf("disabled HTTP function issued %d requests", requests.Load())
	}
}

func TestPipelineHTTPDisableReloadIsReversible(t *testing.T) {
	var requests atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	t.Cleanup(plval.EnableExternalRequestsFunc)

	call := func(disabled bool, wantNil bool) {
		t.Helper()
		if disabled {
			plval.DisableExternalRequestsFunc()
		} else {
			plval.EnableExternalRequestsFunc()
		}
		runJSONOracleWithServices(t, []jsonOracleCase{{
			name: fmt.Sprintf("disabled-%t", disabled), input: `{}`,
			source: fmt.Sprintf(`r=http_request("GET",%q); add_key(result,r == nil)`, server.URL),
			key:    "result", want: wantNil,
		}}, pljit.ServiceConfig{Version: 1, DisableHTTPRequestFunc: disabled})
	}

	call(false, false)
	call(true, true)
	call(false, false)
	if got := requests.Load(); got != 100 {
		t.Fatalf("requests=%d want=100 across enable-disable-enable", got)
	}
}
