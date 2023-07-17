// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"strings"
	"testing"
)

func TestAllDocEn(t *testing.T) {
	if len(PipelineFunctionDocs) != len(PipelineFunctionDocsEN) {
		t.Fatal("len(PipelineFunctionDocs) != len(PipelineFunctionDocsEN)")
	}
	protoPrefix, descPrefix := "Function prototype: ", "Function description: "
	funcNameMap := map[string]bool{}
	for name := range PipelineFunctionDocsEN {
		funcNameMap[strings.TrimSuffix(name, "()")] = true
	}
	for fn := range FuncsMap {
		if fn == "json_all" ||
			fn == "default_time_with_fmt" ||
			fn == "expr" ||
			fn == "vaild_json" {
			continue
		}
		if _, has := funcNameMap[fn]; !has {
			t.Errorf("func %s exists in FuncsMap but not in PipelineFunctionDocs", fn)
		}
	}
	for fn, plDoc := range PipelineFunctionDocsEN {
		lines := strings.Split(plDoc.Doc, "\n")
		var hasProto, hasDesc bool
		for _, line := range lines {
			if strings.HasPrefix(line, protoPrefix) {
				hasProto = true
			}
			if strings.HasPrefix(line, descPrefix) {
				hasDesc = true
			}
		}
		// These fields are needed by front-end.
		if !hasDesc {
			t.Errorf("%s does not contain '%s'", fn, protoPrefix)
		}
		if !hasProto {
			t.Errorf("%s does not contain '%s'", fn, descPrefix)
		}
	}
}
