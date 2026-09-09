// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package pipeline

import (
	"fmt"
	"os"
	"testing"

	"github.com/GuanceCloud/cliutils/point"
	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
)

// TestJITPointKeyAdmissionMatrix compares the common getKeyName grammar used
// by pipeline-go with the JIT checker across Point-reading and Point-writing
// builtins. A matching function name alone is not sufficient coverage: every
// legal key syntax must route before a script is admitted to production.
func TestJITPointKeyAdmissionMatrix(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := pljit.NewRunner(path, "pipeline-go-1.4.3-datakit", 256)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()

	calls := []string{
		`add_key(%s, 1)`,
		`adjust_timezone(%s)`,
		`cast(%s, "str")`,
		`cover(%s, [1, 2])`,
		`datetime(%s, "s", "RFC3339", "UTC")`,
		`decode(%s, "utf-8")`,
		`default_time(%s)`,
		`drop_key(%s)`,
		`get_key(%s)`,
		`gjson(%s, "x", output)`,
		`grok(%s, "x")`,
		`group_in(%s, ["x"], "matched", output)`,
		`json(%s, "x", output)`,
		`json_all(%s)`,
		`kv_split(%s)`,
		`lowercase(%s)`,
		`nullif(%s, "")`,
		`parse_duration(%s)`,
		`rename(%s, output)`,
		`replace(%s, "x", "y")`,
		`set_measurement(%s, false)`,
		`set_tag(%s, "x")`,
		`sql_cover(%s)`,
		`strfmt(%s, "%%s", "x")`,
		`trim(%s)`,
		`uppercase(%s)`,
		`url_decode(%s)`,
		`user_agent(%s)`,
		`xml(%s, "/root", output)`,
	}
	keys := []string{`key`, `"key"`, `obj.name`, `obj[0]`, `(key)`, `$data.key`}
	for _, call := range calls {
		for _, key := range keys {
			source := fmt.Sprintf(call, key)
			name := source
			if len(name) > 80 {
				name = name[:80]
			}
			t.Run(name, func(t *testing.T) {
				_, goErr := NewPlScriptSimple(point.Logging, "key-matrix.p", source)
				check := runner.Check(source)
				goAccepts := goErr == nil
				jitAccepts := check.Route == pljit.RouteJITNative
				if goAccepts != jitAccepts {
					t.Fatalf("source=%q Go error=%v JIT=%+v", source, goErr, check)
				}
			})
		}
	}
}
