// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package jit

import (
	"os"
	"testing"

	"github.com/DataDog/datadog-agent/pkg/obfuscate"
)

func TestSQLCoverPipelineGoUpstreamCorpus(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := NewRunner(path, "pipeline-go-1.4.3-datakit", 16)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()

	cases := []struct {
		name     string
		source   string
		message  string
		key      string
		expected any
	}{
		{"comparison", `json(_,str_msg); sql_cover(str_msg); add_key(after,true)`, `{"str_msg":"select abc from def where x > 3 and y < 5"}`, "str_msg", "select abc from def where x > ? and y < ?"},
		{"dollar-quote", `json(_,str_msg); sql_cover(str_msg); add_key(after,true)`, `{"str_msg":"SELECT $func$INSERT INTO table VALUES ('a', 1, 2)$func$ FROM users"}`, "str_msg", "SELECT ? FROM users"},
		{"aliases", `json(_,str_msg); sql_cover(str_msg); add_key(after,true)`, `{"str_msg":"SELECT Codi , Nom_CA AS Nom, Descripció_CAT AS Descripció FROM ProtValAptitud WHERE Vigent=1 ORDER BY Ordre, Codi"}`, "str_msg", "SELECT Codi, Nom_CA, Descripció_CAT FROM ProtValAptitud WHERE Vigent = ? ORDER BY Ordre, Codi"},
		{"literal-is-missing-key", `sql_cover("SELECT Codi FROM table WHERE id=1"); add_key(after,true)`, "unchanged", "SELECT Codi FROM table WHERE id=1", nil},
		{"quoted-string", `json(_,str_msg); sql_cover(str_msg); add_key(after,true)`, `{"str_msg":"SELECT ('/uffd')"}`, "str_msg", "SELECT ( ? )"},
		{"multiline", `sql_cover(_); add_key(after,true)`, "select abc from def where x > 3 and y < 5\n\t\t\t\t\t\tSELECT ( ? )", "message", "select abc from def where x > ? and y < ? SELECT ( ? )"},
		{"hash-comment", `sql_cover(_); add_key(after,true)`, "#test\nselect abc from def where x > 3 and y < 5", "message", "select abc from def where x > ? and y < ?"},
	}
	dataDog := obfuscate.NewObfuscator(obfuscate.Config{})
	for _, tc := range cases {
		if tc.expected == nil || tc.key == "message" || tc.name == "literal-is-missing-key" {
			continue
		}
		var input string
		switch tc.name {
		case "comparison":
			input = "select abc from def where x > 3 and y < 5"
		case "dollar-quote":
			input = "SELECT $func$INSERT INTO table VALUES ('a', 1, 2)$func$ FROM users"
		case "aliases":
			input = "SELECT Codi , Nom_CA AS Nom, Descripció_CAT AS Descripció FROM ProtValAptitud WHERE Vigent=1 ORDER BY Ordre, Codi"
		case "quoted-string":
			input = "SELECT ('/uffd')"
		}
		oracle, err := dataDog.ObfuscateSQLString(input)
		if err != nil || oracle.Query != tc.expected {
			t.Fatalf("pipeline-go DataDog oracle %s=%q error=%v want %#v", tc.name, oracle.Query, err, tc.expected)
		}
	}

	for _, size := range []int{1, 2, 4, 8, 10, 128} {
		t.Run("batch-"+itoa(size), func(t *testing.T) {
			for _, tc := range cases {
				check := runner.Check(tc.source)
				if check.Route != RouteJITNative || check.Capabilities.ExecutionMode != "native" ||
					check.Capabilities.RequiredHostFlags != 0 || len(check.Capabilities.HostCalls) != 0 {
					t.Fatalf("%s route=%+v", tc.name, check)
				}
				points := make([]Point, size)
				for i := range points {
					points[i] = Point{Version: 1, Category: "logging", Measurement: tc.name,
						Fields: map[string]any{"message": tc.message, "sequence": int64(i)}}
				}
				input, err := EncodeFlatPoints(points)
				if err != nil {
					t.Fatal(err)
				}
				batch, err := runner.Process(tc.source, input)
				if err != nil || len(batch.Records) != size {
					t.Fatalf("%s records=%d error=%v", tc.name, len(batch.Records), err)
				}
				for i, record := range batch.Records {
					if record.Status != TerminalOK {
						t.Fatalf("%s record %d terminal=%d error=%s", tc.name, i, record.Status, record.Error)
					}
					applyHostCompatRecord(t, batch, i, &points[i])
					if tc.expected == nil {
						if _, ok := points[i].Fields[tc.key]; ok {
							t.Fatalf("%s record %d unexpectedly set %q: %#v", tc.name, i, tc.key, points[i])
						}
					} else if got := points[i].Fields[tc.key]; got != tc.expected {
						t.Fatalf("%s record %d %s=%#v want %#v", tc.name, i, tc.key, got, tc.expected)
					}
					if points[i].Fields["after"] != true || points[i].Fields["sequence"] != int64(i) {
						t.Fatalf("%s record %d lost continuation/input: %#v", tc.name, i, points[i])
					}
				}
			}
		})
	}
}
