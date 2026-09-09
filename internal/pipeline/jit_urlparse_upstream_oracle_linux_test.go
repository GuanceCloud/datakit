// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

// Adapted from pipeline-go v1.4.3 ptinput/funcs/fn_url_parse_test.go.
// Copyright 2021-present Guance, Inc. MIT License; see testdata/PIPELINE_GO_TEST_LICENSE.txt.
package pipeline

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/GuanceCloud/cliutils/point"
	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
)

func TestJITUpstreamURLParseOracle(t *testing.T) {
	var cases []jsonOracleCase
	for i, tc := range []struct{ input, call, key, access, want, tail string }{
		{"https://www.baidu.com", `url_parse(url)`, "scheme", `["scheme"]`, "https", ""},
		{"http://127.0.0.1:9529", `url_parse(url)`, "host", `["host"]`, "127.0.0.1:9529", ""},
		{"http://127.0.0.1:9529", `url_parse(url)`, "port", `["port"]`, "9529", ""},
		{"http://127.0.0.1:9529/v1/metrics", `url_parse(url)`, "path", `["path"]`, "/v1/metrics", ""},
		{"http://127.0.0.1:9529/v1/metrics?arg1=v1&arg2=v2", `url_parse(url)`, "a", `["params"]["arg1"]`, "v1", ""},
		{"http://127.0.0.1:9529/v1/metrics?arg1=v1&arg2=v2&arg2=v3", `url_parse(url)`, "a", `["params"]["arg2"]`, "v2,v3", ""},
		{"https://www.baidu.com", `url_parse(url, "up_")`, "scheme", `["up_scheme"]`, "https", `if m["scheme"] != nil { add_key(unexpected_unprefixed_key, true) }`},
		{"http://127.0.0.1:9529/v1/metrics?arg1=v1&arg2=v2", `url_parse(url, "up_")`, "a", `["up_params"]["arg1"]`, "v1", ""},
		{"https://www.baidu.com", `url_parse(url, p)`, "scheme", `["up_scheme"]`, "https", ""},
		{"https://www.baidu.com", `url_parse(url, prefix="up_")`, "scheme", `["up_scheme"]`, "https", ""},
		{"https://www.baidu.com", `url_parse(url, "")`, "scheme", `["scheme"]`, "https", ""},
		{"/var/log/datakit/log", `url_parse(url)`, "p", `["path"]`, "/var/log/datakit/log", ""},
	} {
		prefix := "json(_, url)\n"
		if i == 8 {
			prefix += "p = \"up_\"\n"
		}
		cases = append(cases, jsonOracleCase{name: fmt.Sprint(i), input: fmt.Sprintf(`{"url":%q}`, tc.input),
			source: prefix + fmt.Sprintf("m = %s\nadd_key(%s, m%s)\n%s", tc.call, tc.key, tc.access, tc.tail), key: tc.key, want: tc.want})
	}
	runJSONOracle(t, cases)
}

func TestJITURLParseBoundaryOracle(t *testing.T) {
	runURLComponentOracle(t, []string{
		"https://EXAMPLE.com:443/a/../b", "https://example.com", "//EXAMPLE.com:80/path",
		"relative/../path?a=1", "mailto:user@example.com", "http:opaque", "",
		"https://example.com/a%2Fb?q=a+b&q=c%2Bd", "https://example.com/?x=1;x=2&good=ok",
		"https://example.com/?bad=%zz&good=ok", "http://[::1]:8080/x", "http://user:pass@example.com/",
		"http://例子.测试/路径", "http://example.com/%zz", "http://example.com:bad/",
		"http://example.com/a\nb", "1abc:foo", "https://example.com/a#%zz",
	})
}

func TestJITURLAuthorityOracle(t *testing.T) {
	runURLComponentOracle(t, []string{
		"http://%41.example/", "http://example%2ecom/", "http://%E4%BE%8B.example/",
		"http://[fe80::1%25eth0]:8080/", "http://[fe80::1%25eth%30]/",
		"http://[::1", "http://[::1]suffix/", "http://[not-ip]/",
		"http://example.com:/", "http://example.com:999999/", "http://example.com:00080/",
		"http://user name@example.com/", "http://us%20er@example.com/", "http://user%zz@example.com/",
		"http://user|name@example.com/", "http://user:pass@EXAMPLE.com:80/",
	})
}

func TestJITURLRawQueryValueOracle(t *testing.T) {
	var cases []jsonOracleCase
	for i, tc := range []struct{ query, want string }{
		{"x=%FF", string([]byte{0xff})},
		{"x=%C0%AF", string([]byte{0xc0, 0xaf})},
		{"x=%00a", "\x00a"},
		{"x=%FF&x=%FE", string([]byte{0xff, ',', 0xfe})},
		{"x=%E4%B8%AD", "中"},
	} {
		input, err := json.Marshal(map[string]string{"url": "http://example.com/?" + tc.query})
		if err != nil {
			t.Fatal(err)
		}
		cases = append(cases, jsonOracleCase{name: fmt.Sprint(i), input: string(input), source: `json(_,url); m=url_parse(url); add_key(result,m["params"]["x"]); add_key(length,strlen(result)); add_key(digest,hash(result,"sha256"))`, key: "result", want: tc.want})
	}
	runJSONOracle(t, cases)
}

func TestJITURLRawQueryKeyOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		// A single invalid key isolates byte escaping from duplicate rendered-key ordering.
		// Distinct invalid-key identity is tested above; duplicate JSON ordering remains open.
		{name: "raw-key-json", input: `{"url":"http://example.com/?%FF=one"}`, source: `json(_,url); m=url_parse(url); add_key(result,m["params"])`, key: "result", want: `{"\ufffd":"one"}`},
		{name: "raw-update-text-delete", input: `{"url":"http://example.com/?%FF=one&%FE=two&remove=yes","key":"/w=="}`, source: `json(_,url); json(_,key); b64dec(key); m=url_parse(url); p=m["params"]; p[key]="updated"; add_key(result,p[key]); delete(p,"remove"); add_key(remaining,len(p))`, key: "result", want: "updated"},
		{name: "distinct-invalid-keys", input: `{"url":"http://example.com/?%FF=one&%FE=two"}`, source: `json(_,url); m=url_parse(url); add_key(result,len(m["params"]))`, key: "result", want: int64(2)},
		{name: "invalid-vs-replacement", input: `{"url":"http://example.com/?%FF=one&%EF%BF%BD=two"}`, source: `json(_,url); m=url_parse(url); add_key(result,m["params"]["�"])`, key: "result", want: "two"},
		{name: "raw-index", input: `{"url":"http://example.com/?%FF=one&%FE=two","key":"/w=="}`, source: `json(_,url); json(_,key); b64dec(key); m=url_parse(url); add_key(result,m["params"][key])`, key: "result", want: "one"},
	})
}

func TestJITURLRawInputBytesOracle(t *testing.T) {
	for _, tc := range []struct {
		name, input, access, want string
	}{
		{"absolute-path", "http://example.com/\xff", "path", "/\xff"},
		{"relative-path", "relative/\xfe", "path", "relative/\xfe"},
		{"authority", "http://\xff.example/", "host", "\xff.example"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			runJSONOracle(t, []jsonOracleCase{{
				name:   tc.name,
				input:  "test",
				source: fmt.Sprintf(`m=url_parse(raw_url); add_key(result,m[%q])`, tc.access),
				key:    "result",
				want:   tc.want,
			}}, func(pt *point.Point) {
				pt.AddKVs(point.NewKV("raw_url", tc.input))
			})
		})
	}
}

func TestJITJSONEncodedKeyOrderOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "escaped-control", input: `{"url":"http://example.com/?%0A=line&A=letter"}`, source: `json(_,url); m=url_parse(url); add_key(result,m["params"])`, key: "result", want: `{"A":"letter","\n":"line"}`},
		{name: "html-escape", input: `{"url":"http://example.com/?%3C=html&A=letter"}`, source: `json(_,url); m=url_parse(url); add_key(result,m["params"])`, key: "result", want: `{"A":"letter","\u003c":"html"}`},
		{name: "invalid-before-unicode", input: `{"url":"http://example.com/?%FF=raw&%E4%B8%AD=text"}`, source: `json(_,url); m=url_parse(url); add_key(result,m["params"])`, key: "result", want: `{"\ufffd":"raw","中":"text"}`},
	})
}

func runURLComponentOracle(t *testing.T, inputs []string) {
	t.Helper()
	var cases []jsonOracleCase
	for i, input := range inputs {
		encoded, err := json.Marshal(map[string]string{"url": input})
		if err != nil {
			t.Fatal(err)
		}
		parsed, parseErr := url.Parse(input)
		source := `json(_, url); m = url_parse(url)`
		if parseErr != nil {
			cases = append(cases, jsonOracleCase{name: fmt.Sprintf("%d-invalid", i), input: string(encoded), source: source, fails: true})
			continue
		}
		params := map[string]string{}
		for key, values := range parsed.Query() {
			params[key] = strings.Join(values, ",")
		}
		for _, field := range []struct {
			key   string
			value any
		}{
			{"scheme", parsed.Scheme}, {"host", parsed.Host}, {"port", parsed.Port()}, {"path", parsed.Path},
		} {
			cases = append(cases, jsonOracleCase{name: fmt.Sprintf("%d-%s", i, field.key), input: string(encoded), source: source + fmt.Sprintf(`; add_key(result,m[%q])`, field.key), key: "result", want: field.value})
		}
		paramsJSON, err := json.Marshal(params)
		if err != nil {
			t.Fatal(err)
		}
		cases = append(cases, jsonOracleCase{name: fmt.Sprintf("%d-params", i), input: string(encoded), source: source + `; add_key(result,m["params"])`, key: "result", want: string(paramsJSON)})
	}
	runJSONOracle(t, cases)
}

func TestJITURLQueryFilteringOracle(t *testing.T) {
	var cases []jsonOracleCase
	for i, tc := range []struct{ query, expected string }{
		{`x=1;x=2&good=ok`, `{"good":"ok"}`},
		{`bad=%zz&good=ok`, `{"good":"ok"}`},
		{`bad%=v&good=ok`, `{"good":"ok"}`},
		{`a=%&b=%1&c=ok`, `{"c":"ok"}`},
		{`x=1&x=%zz&x=2`, `{"x":"1,2"}`},
		{`x=a%3Bb&x=c%2Bd`, `{"x":"a;b,c+d"}`},
		{`a+b=c+d&empty=&bare&=value`, `{"":"value","a b":"c d","bare":"","empty":""}`},
		{`&&x=1&&`, `{"x":"1"}`},
		{`x=%2520`, `{"x":"%20"}`},
		{`x=1#fragment?bad=%25zz`, `{"x":"1"}`},
	} {
		input, err := json.Marshal(map[string]string{"url": "https://example.com/?" + tc.query})
		if err != nil {
			t.Fatal(err)
		}
		cases = append(cases, jsonOracleCase{name: fmt.Sprint(i), input: string(input), source: `json(_,url); m = url_parse(url); add_key(result,m["params"])`, key: "result", want: tc.expected})
	}
	runJSONOracle(t, cases)
}

func TestJITUpstreamURLParseErrors(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real runtime")
	}
	runner, err := pljit.NewRunner(path, "pipeline-go-1.4.3-datakit", 2)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	for i, call := range []string{`url_parse(url, "up_", 2)`, `url_parse(url, 2)`} {
		t.Run(fmt.Sprint(i+12), func(t *testing.T) {
			source := "json(_, url)\nm = " + call
			if _, err := NewPlScriptSimple(point.Logging, "url.p", source); err == nil {
				t.Fatal("expected Go checking rejection")
			}
			check := runner.Check(source)
			if check.Route != pljit.RoutePipelineGo || check.Reason != pljit.CheckReasonCompileError {
				t.Fatalf("expected JIT checking rejection: %+v", check)
			}
		})
	}
	runJSONOracle(t, []jsonOracleCase{{name: "14-variable-prefix", source: `json(_, url); p = 123; m = url_parse(url, p)`, input: `{"url":"http://127.0.0.1:9529/v1/metrics?arg1=v1&arg2=v2"}`, fails: true}})
}
