package funcs

import (
	"strings"
	"testing"
)

func TestAllDoc(t *testing.T) {
	protoPrefix, descPrefix := "函数原型：", "函数说明："
	funcNameMap := map[string]bool{}
	for name := range PipelineFunctionDocs {
		funcNameMap[strings.TrimSuffix(name, "()")] = true
	}
	for fn := range FuncsMap {
		if fn == "json_all" || fn == "default_time_with_fmt" || fn == "expr" {
			continue
		}
		if _, has := funcNameMap[fn]; !has {
			t.Errorf("func %s exists in FuncsMap but not in PipelineFunctionDocs", fn)
		}
	}
	for fn, plDoc := range PipelineFunctionDocs {
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
