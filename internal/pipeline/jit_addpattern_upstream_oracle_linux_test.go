// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

// Adapted from pipeline-go v1.4.3 ptinput/funcs/fn_addpattern_test.go.
// Copyright 2021-present Guance, Inc. MIT License; see
// testdata/PIPELINE_GO_TEST_LICENSE.txt.
package pipeline

import (
	"fmt"
	"testing"
)

func TestJITUpstreamAddPatternOracle(t *testing.T) {
	const input = `162.62.81.1 - - [29/Nov/2021:07:30:50 +0000] "POST /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413 "-" "Mozilla/4.0"`
	// Only factor out the common upstream expression; retain capture suffixes
	// and instruction order, especially pattern redefinitions between groks.
	grok := func(version string) string {
		return `grok(_, "%{IPORHOST:client_ip} %{NOTSPACE} %{NOTSPACE} \\[%{HTTPDATE:time}\\] \"%{DATA} %{GREEDYDATA} HTTP/` + version + `\" %{INT:status_code} %{INT:bytes}")` + "\n"
	}
	pattern := func(name, value string) string { return fmt.Sprintf("add_pattern(%q, %q)\n", name, value) }
	var cases []jsonOracleCase
	add := func(name, source, key string, want float64) {
		cases = append(cases, jsonOracleCase{name: name, source: source + fmt.Sprintf("cast(%s, \"float\")", key), input: input, key: key, want: want})
	}
	add("00-s", pattern("http_version", `[\d\.]+`)+grok(`%{http_version:http_version}`), "http_version", 1.1)
	add("01-d", pattern("num", `\d`)+pattern("http_version", `[%{num}\.]+`)+grok(`%{http_version:http_version}`), "http_version", 1.1)
	add("02-o", pattern("num", `\d`)+pattern("http_version", `\d`)+pattern("http_version", `[%{num}\.]+`)+grok(`%{http_version:http_version}[\\.\\d]+`), "http_version", 1)
	for i, name := range []string{"_num", "Num", "1Num"} {
		add(fmt.Sprintf("%02d-name-%s", i+3, name), pattern(name, `\d`)+pattern("http_version", `[%{`+name+`}\.]+`)+grok(`%{http_version:http_version}`), "http_version", 1.1)
	}
	add("06-opm", pattern("NUMBER", `\d`)+grok(`%{NUMBER:http_version}.*`), "http_version", 1)
	add("07-opm2", pattern("numb2", `%{NUMBER}`)+grok(`%{numb2:http_version}.*`), "http_version", 1.1)
	add("08-opm3", pattern("numb2", `%{NUMBER}`)+grok(`%{numb2:http_version}.*`)+pattern("numb3", `\d`)+grok(`%{numb3:http_version_int}.*`), "http_version_int", 1)
	add("09-opm4", pattern("numb2", `%{NUMBER}`)+grok(`%{numb2:http_version}.*`)+pattern("numb2", `\d`)+grok(`%{numb2:http_version_int}.*`), "http_version_int", 1)
	runJSONOracle(t, cases)
}
