// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	tu "github.com/GuanceCloud/cliutils/testutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/ptinput"
)

func TestURLParse(t *testing.T) {
	cases := []struct {
		name     string
		pl, in   string
		outKey   string
		expected interface{}
		fail     bool
	}{
		{
			name: "port",
			pl: `
json(_, url)
m = url_parse(url)
add_key(scheme, m["scheme"])
`,
			in:       `{"url": "https://www.baidu.com"}`,
			outKey:   "scheme",
			expected: "https",
			fail:     false,
		},
		{
			name: "host",
			pl: `
json(_, url)
m = url_parse(url)
add_key(host, m["host"])
`,
			in:       `{"url": "http://127.0.0.1:9529"}`,
			outKey:   "host",
			expected: "127.0.0.1:9529",
			fail:     false,
		},
		{
			name: "port",
			pl: `
json(_, url)
m = url_parse(url)
add_key(port, m["port"])
`,
			in:       `{"url": "http://127.0.0.1:9529"}`,
			outKey:   "port",
			expected: "9529",
			fail:     false,
		},
		{
			name: "path",
			pl: `
json(_, url)
m = url_parse(url)
add_key(path, m["path"])
`,
			in:       `{"url": "http://127.0.0.1:9529/v1/metrics"}`,
			outKey:   "path",
			expected: "/v1/metrics",
			fail:     false,
		},
		{
			name: "arg1",
			pl: `
json(_, url)
m = url_parse(url)
add_key(a, m["params"]["arg1"])
`,
			in:       `{"url": "http://127.0.0.1:9529/v1/metrics?arg1=v1&arg2=v2"}`,
			outKey:   "a",
			expected: "v1",
			fail:     false,
		},
		{
			name: "arg2",
			pl: `
json(_, url)
m = url_parse(url)
add_key(a, m["params"]["arg2"])
`,
			in:       `{"url": "http://127.0.0.1:9529/v1/metrics?arg1=v1&arg2=v2&arg2=v3"}`,
			outKey:   "a",
			expected: "v2,v3",
			fail:     false,
		},
		{
			name: "invalid url",
			pl: `
json(_, url)
m = url_parse(url)
add_key(p, m["path"])
`,
			in:   `{"url": "/var/log/datakit/log"}`,
			fail: true,
		},
		{
			name: "two many args",
			pl: `
json(_, url)
m = url_parse(url, 2)
`,
			in:   `{"url": "http://127.0.0.1:9529/v1/metrics?arg1=v1&arg2=v2"}`,
			fail: true,
		},
	}

	for idx, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runner, err := NewTestingRunner(tc.pl)
			if err != nil {
				if tc.fail {
					t.Logf("[%d]expect error: %s", idx, err)
				} else {
					t.Errorf("[%d] failed: %s", idx, err)
				}
				return
			}

			var pt ptinput.PlInputPt = ptinput.NewPlPoint(point.Logging, "test", nil, map[string]any{"message": tc.in}, time.Now())
			errR := runScript(runner, pt)
			if errR != nil {
				t.Fatal(errR)
			}

			if v, istag, ok := pt.GetWithIsTag(tc.outKey); !ok {
				if !tc.fail {
					t.Errorf("[%d]key %s, error: %s", idx, tc.outKey, err)
				}
			} else {
				if istag {
					t.Errorf("key %s should be a field", tc.outKey)
				} else {
					tu.Equals(t, tc.expected, v)
					t.Logf("[%d] PASS", idx)
				}
			}
		})
	}
}
