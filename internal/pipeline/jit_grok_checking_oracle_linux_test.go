// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package pipeline

import (
	"os"
	"testing"

	"github.com/GuanceCloud/cliutils/point"
	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
)

func TestJITGrokCheckingOracle(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real runtime")
	}
	runner, err := pljit.NewRunner(path, "pipeline-go-1.4.3-datakit", 64)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	for _, args := range []string{
		``, `message`, `message,"%{WORD:value}"`,
		`message,"%{WORD:value}",true`, `message,"%{WORD:value}",false`,
		`message,"%{WORD:value}",1`,
		`message,"%{WORD:value}",true,false`, `123,"x"`, `"message","x"`,
		`obj.name,"x"`, `obj[0],"x"`, `(message),"x"`,
		`message,"["`, `message,"%{UNKNOWN}"`,
	} {
		for _, nested := range []bool{false, true} {
			source := "grok(" + args + ")"
			if nested {
				source = "add_key(result," + source + ")"
			}
			t.Run(source, func(t *testing.T) {
				_, goErr := NewPlScriptSimple(point.Logging, "grok-check.p", source)
				check := runner.Check(source)
				if (goErr == nil) != (check.Route == pljit.RouteJITNative) {
					t.Fatalf("Go error=%v; JIT=%+v", goErr, check)
				}
			})
		}
	}
	for _, source := range []string{
		`pattern="%{WORD:value}"; grok(message,pattern)`,
		`trim=true; grok(message,"%{WORD:value}",trim)`,
		`grok(message,("%{WORD:value}"))`,
	} {
		_, goErr := NewPlScriptSimple(point.Logging, "grok-check.p", source)
		check := runner.Check(source)
		if (goErr == nil) != (check.Route == pljit.RouteJITNative) {
			t.Fatalf("Go error=%v; source=%q JIT=%+v", goErr, source, check)
		}
	}
}
