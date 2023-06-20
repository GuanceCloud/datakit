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

func TestCast(t *testing.T) {
	cases := []struct {
		name, pl, in string
		outkey       string
		expect       interface{}
		fail         bool
	}{
		{
			name: "cast int",
			in:   `162.62.81.1 - - [29/Nov/2021:07:30:50 +0000] "123 /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413 "-" "Mozilla/4.0"`,
			pl: `
	grok(_, "%{IPORHOST:client_ip} %{NOTSPACE} %{NOTSPACE} \\[%{HTTPDATE:time}\\] \"%{DATA:data} %{GREEDYDATA} HTTP/%{NUMBER}\" %{INT:status_code} %{INT:bytes}")
	cast(data, "int")
	`,
			outkey: "data",
			expect: int64(123),
			fail:   false,
		},
		{
			name: "cast string",
			in:   `162.62.81.1 - - [29/Nov/2021:07:30:50 +0000] "123 /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413 "-" "Mozilla/4.0"`,
			pl: `
	grok(_, "%{IPORHOST:client_ip} %{NOTSPACE} %{NOTSPACE} \\[%{HTTPDATE:time}\\] \"%{DATA:data} %{GREEDYDATA} HTTP/%{NUMBER}\" %{INT:status_code} %{INT:bytes}")
	cast(data, "str")
	`,
			outkey: "data",
			expect: "123",
			fail:   false,
		},
		{
			name: "cast bool",
			in:   `162.62.81.1 - - [29/Nov/2021:07:30:50 +0000] "true /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413 "-" "Mozilla/4.0"`,
			pl: `
	grok(_, "%{IPORHOST:client_ip} %{NOTSPACE} %{NOTSPACE} \\[%{HTTPDATE:time}\\] \"%{DATA:data} %{GREEDYDATA} HTTP/%{NUMBER}\" %{INT:status_code} %{INT:bytes}")
	cast(data, "bool")
	`,
			outkey: "data",
			expect: true,
			fail:   false,
		},
		{
			name: "cast float",
			in:   `162.62.81.1 - - [29/Nov/2021:07:30:50 +0000] "123 /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413 "-" "Mozilla/4.0"`,
			pl: `
	grok(_, "%{IPORHOST:client_ip} %{NOTSPACE} %{NOTSPACE} \\[%{HTTPDATE:time}\\] \"%{DATA:data} %{GREEDYDATA} HTTP/%{NUMBER}\" %{INT:status_code} %{INT:bytes}")
	cast(data, "float")
	`,
			outkey: "data",
			expect: float64(123),
			fail:   false,
		},
		{
			name: "cast float",
			in:   `162.62.81.1 - - [29/Nov/2021:07:30:50 +0000] "12.3 /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413 "-" "Mozilla/4.0"`,
			pl: `
	grok(_, "%{IPORHOST:client_ip} %{NOTSPACE} %{NOTSPACE} \\[%{HTTPDATE:time}\\] \"%{DATA:data} %{GREEDYDATA} HTTP/%{NUMBER}\" %{INT:status_code} %{INT:bytes}")
	cast(data, "float")
	`,
			outkey: "data",
			expect: float64(12.3),
			fail:   false,
		},
		{
			name: "cast float ",
			in:   `162.62.81.1 - - [29/Nov/2021:07:30:50 +0000] "-123. /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413 "-" "Mozilla/4.0"`,
			pl: `
	grok(_, "%{IPORHOST:client_ip} %{NOTSPACE} %{NOTSPACE} \\[%{HTTPDATE:time}\\] \"%{DATA:data} %{GREEDYDATA} HTTP/%{NUMBER}\" %{INT:status_code} %{INT:bytes}")
	cast(data, "float")
	`,
			outkey: "data",
			expect: float64(-123),
			fail:   false,
		},
		{
			name: "cast float ",
			in:   `162.62.81.1 - - [29/Nov/2021:07:30:50 +0000] "-.12 /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413 "-" "Mozilla/4.0"`,
			pl: `
	grok(_, "%{IPORHOST:client_ip} %{NOTSPACE} %{NOTSPACE} \\[%{HTTPDATE:time}\\] \"%{DATA:data} %{GREEDYDATA} HTTP/%{NUMBER}\" %{INT:status_code} %{INT:bytes}")
	cast(data, "float")
	`,
			outkey: "data",
			expect: float64(-.12),
			fail:   false,
		},
		{
			name: "cast int ",
			in:   `162.62.81.1 - - [29/Nov/2021:07:30:50 +0000] "+12.6 /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413 "-" "Mozilla/4.0"`,
			pl: `
	grok(_, "%{IPORHOST:client_ip} %{NOTSPACE} %{NOTSPACE} \\[%{HTTPDATE:time}\\] \"%{DATA:data} %{GREEDYDATA} HTTP/%{NUMBER}\" %{INT:status_code} %{INT:bytes}")
	cast(data, "int")
	`,
			outkey: "data",
			expect: int64(+12),
			fail:   false,
		},
		{
			name: "cast intx ",
			in:   `162.62.81.1 - - [29/Nov/2021:07:30:50 +0000] "+12.6 /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413 "-" "Mozilla/4.0"`,
			pl: `
	grok(_, "%{IPORHOST:client_ip} %{NOTSPACE} %{NOTSPACE} \\[%{HTTPDATE:time}\\] \"%{DATA:data} %{GREEDYDATA} HTTP/%{NUMBER}\" %{INT:status_code} %{INT:bytes}")
	cast(data, "intx")
	`,
			outkey: "data",
			expect: "+12.6",
			fail:   true,
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

			pt := ptinput.NewPlPoint(
				point.Logging, "test", nil, map[string]any{"message": tc.in}, time.Now())
			errR := runScript(runner, pt)

			if errR != nil {
				t.Fatal(errR)
			} else {
				v, _, _ := pt.Get(tc.outkey)
				tu.Equals(t, tc.expect, v)
				t.Logf("[%d] PASS", idx)
			}
		})
	}
}
