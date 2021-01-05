package process

import (
	"testing"

	"github.com/tidwall/gjson"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/process/parser"
)


func assertEqual(t *testing.T, a, b interface{}) {
	t.Helper()
	if a != b {
		t.Errorf("Not Equal. %d %d", a, b)
	}
}

func TestRenameFunc(t *testing.T) {
	js := `{"a":{"first":2.3,"second":2,"thrid":"abc","forth":true},"age":47}`

	script := `rename(a.second, bb);`
	nodes, err := parser.ParseFuncExpr(script)
	assertEqual(t, err, nil)

	p := NewProcedure(nil)
	p = p.ProcessLog(js, nodes)
	assertEqual(t, p.lastErr, nil)

	r := gjson.GetBytes(p.Content, "bb")
	assertEqual(t, r.String(), "2")
}