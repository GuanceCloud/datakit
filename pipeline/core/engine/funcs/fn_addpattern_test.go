// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAddPattern(t *testing.T) {
	cases := []struct {
		name, pl, in string
		outkey       string
		expect       interface{}
		fail         bool
	}{
		{
			name: "pattern: http version (-s)",
			in:   `162.62.81.1 - - [29/Nov/2021:07:30:50 +0000] "POST /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413 "-" "Mozilla/4.0"`,
			pl: `
		add_pattern("http_version", "[\\d\\.]+")
		grok(_, "%{IPORHOST:client_ip} %{NOTSPACE} %{NOTSPACE} \\[%{HTTPDATE:time}\\] \"%{DATA} %{GREEDYDATA} HTTP/%{http_version:http_version}\" %{INT:status_code} %{INT:bytes}")
		cast(http_version, "float")
		`,
			outkey: "http_version",
			expect: 1.1,
			fail:   false,
		},
		{
			name: "pattern: http version (-d)",
			in:   `162.62.81.1 - - [29/Nov/2021:07:30:50 +0000] "POST /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413 "-" "Mozilla/4.0"`,
			pl: `
		add_pattern("num", "\\d")
		add_pattern("http_version", "[%{num}\\.]+")
		grok(_, "%{IPORHOST:client_ip} %{NOTSPACE} %{NOTSPACE} \\[%{HTTPDATE:time}\\] \"%{DATA} %{GREEDYDATA} HTTP/%{http_version:http_version}\" %{INT:status_code} %{INT:bytes}")
		cast(http_version, "float")
		`,
			outkey: "http_version",
			expect: 1.1,
			fail:   false,
		},
		{
			name: "pattern: http version (-o)",
			in:   `162.62.81.1 - - [29/Nov/2021:07:30:50 +0000] "POST /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413 "-" "Mozilla/4.0"`,
			pl: `
		add_pattern("num", "\\d")
		add_pattern("http_version", "\\d")
		add_pattern("http_version", "[%{num}\\.]+")
		grok(_, "%{IPORHOST:client_ip} %{NOTSPACE} %{NOTSPACE} \\[%{HTTPDATE:time}\\] \"%{DATA} %{GREEDYDATA} HTTP/%{http_version:http_version}[\\.\\d]+\" %{INT:status_code} %{INT:bytes}")
		cast(http_version, "float")
		`,
			outkey: "http_version",
			expect: float64(1),
			fail:   false,
		},
		{
			name: "pattern: http version (-fd)",
			in:   `162.62.81.1 - - [29/Nov/2021:07:30:50 +0000] "POST /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413 "-" "Mozilla/4.0"`,
			pl: `
		add_pattern("_num", "\\d")
		add_pattern("http_version", "[%{_num}\\.]+")
		grok(_, "%{IPORHOST:client_ip} %{NOTSPACE} %{NOTSPACE} \\[%{HTTPDATE:time}\\] \"%{DATA} %{GREEDYDATA} HTTP/%{http_version:http_version}\" %{INT:status_code} %{INT:bytes}")
		cast(http_version, "float")
		`,
			outkey: "http_version",
			expect: float64(1.1),
			fail:   false,
		},
		{
			name: "pattern: http version (-fu)",
			in:   `162.62.81.1 - - [29/Nov/2021:07:30:50 +0000] "POST /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413 "-" "Mozilla/4.0"`,
			pl: `
		add_pattern("Num", "\\d")
		add_pattern("http_version", "[%{Num}\\.]+")
		grok(_, "%{IPORHOST:client_ip} %{NOTSPACE} %{NOTSPACE} \\[%{HTTPDATE:time}\\] \"%{DATA} %{GREEDYDATA} HTTP/%{http_version:http_version}\" %{INT:status_code} %{INT:bytes}")
		cast(http_version, "float")
		`,
			outkey: "http_version",
			expect: float64(1.1),
			fail:   false,
		},
		{
			name: "pattern: http version (-fn)",
			in:   `162.62.81.1 - - [29/Nov/2021:07:30:50 +0000] "POST /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413 "-" "Mozilla/4.0"`,
			pl: `
		add_pattern("1Num", "\\d")
		add_pattern("http_version", "[%{1Num}\\.]+")
		grok(_, "%{IPORHOST:client_ip} %{NOTSPACE} %{NOTSPACE} \\[%{HTTPDATE:time}\\] \"%{DATA} %{GREEDYDATA} HTTP/%{http_version:http_version}\" %{INT:status_code} %{INT:bytes}")
		cast(http_version, "float")
		`,
			outkey: "http_version",
			expect: float64(1.1),
			fail:   false,
		},
		{
			name: "pattern: http version (-opm)", // 测试替换部分 global pattern 可被替换
			in:   `162.62.81.1 - - [29/Nov/2021:07:30:50 +0000] "POST /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413 "-" "Mozilla/4.0"`,
			pl: `
		add_pattern("NUMBER", "\\d")
		grok(_, "%{IPORHOST:client_ip} %{NOTSPACE} %{NOTSPACE} \\[%{HTTPDATE:time}\\] \"%{DATA} %{GREEDYDATA} HTTP/%{NUMBER:http_version}.*\" %{INT:status_code} %{INT:bytes}")
		cast(http_version, "float")
		`,
			outkey: "http_version",
			expect: float64(1),
			fail:   false,
		},
		{
			name: "pattern: http version (-opm2)",
			in:   `162.62.81.1 - - [29/Nov/2021:07:30:50 +0000] "POST /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413 "-" "Mozilla/4.0"`,
			pl: `
		add_pattern("numb2", "%{NUMBER}")
		grok(_, "%{IPORHOST:client_ip} %{NOTSPACE} %{NOTSPACE} \\[%{HTTPDATE:time}\\] \"%{DATA} %{GREEDYDATA} HTTP/%{numb2:http_version}.*\" %{INT:status_code} %{INT:bytes}")
		cast(http_version, "float")
		`,
			outkey: "http_version",
			expect: float64(1.1),
			fail:   false,
		},
		{
			name: "pattern: http version (-opm3)",
			in:   `162.62.81.1 - - [29/Nov/2021:07:30:50 +0000] "POST /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413 "-" "Mozilla/4.0"`,
			pl: `
		add_pattern("numb2", "%{NUMBER}")
		grok(_, "%{IPORHOST:client_ip} %{NOTSPACE} %{NOTSPACE} \\[%{HTTPDATE:time}\\] \"%{DATA} %{GREEDYDATA} HTTP/%{numb2:http_version}.*\" %{INT:status_code} %{INT:bytes}")
		add_pattern("numb3", "\\d")
		grok(_, "%{IPORHOST:client_ip} %{NOTSPACE} %{NOTSPACE} \\[%{HTTPDATE:time}\\] \"%{DATA} %{GREEDYDATA} HTTP/%{numb3:http_version_int}.*\" %{INT:status_code} %{INT:bytes}")
		cast(http_version_int, "float")
		`,
			outkey: "http_version_int",
			expect: float64(1.),
			fail:   false,
		},
		{
			name: "pattern: http version (-opm4)",
			in:   `162.62.81.1 - - [29/Nov/2021:07:30:50 +0000] "POST /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413 "-" "Mozilla/4.0"`,
			pl: `
	add_pattern("numb2", "%{NUMBER}")
	grok(_, "%{IPORHOST:client_ip} %{NOTSPACE} %{NOTSPACE} \\[%{HTTPDATE:time}\\] \"%{DATA} %{GREEDYDATA} HTTP/%{numb2:http_version}.*\" %{INT:status_code} %{INT:bytes}")
	add_pattern("numb2", "\\d")
	grok(_, "%{IPORHOST:client_ip} %{NOTSPACE} %{NOTSPACE} \\[%{HTTPDATE:time}\\] \"%{DATA} %{GREEDYDATA} HTTP/%{numb2:http_version_int}.*\" %{INT:status_code} %{INT:bytes}")
	cast(http_version_int, "float")
	`,
			outkey: "http_version_int",
			expect: float64(1),
			fail:   false,
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
			_, _, f, _, _, err := runScript(runner, "test", nil, map[string]any{"message": tc.in}, time.Now())
			if err != nil {
				t.Fatal(err)
			}
			v := f[tc.outkey]
			if assert.Equal(t, tc.expect, v) {
				t.Logf("[%d] PASS", idx)
			}
		})
	}
}
