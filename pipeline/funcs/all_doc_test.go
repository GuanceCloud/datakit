package funcs

import (
	"strings"
	"testing"
)

func TestAllDoc(t *testing.T) {
	funcNameMap := map[string]bool{}
	for name := range PipelineFunctionDocs {
		funcNameMap[strings.TrimSuffix(name, "()")] = true
	}
	for fn := range FuncsMap {
		if fn == "json_all" || fn == "default_time_with_fmt" {
			continue
		}
		if _, has := funcNameMap[fn]; !has {
			t.Errorf("func %s exits in FuncsMap but not in PipelineFunctionDocs", fn)
		}
	}
}