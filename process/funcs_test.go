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

func TestGrokFunc(t *testing.T) {
	js := `127.0.0.1 - - [23/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 207`
	script := `grok("%{COMMONAPACHELOG}");`

	nodes, err := parser.ParseFuncExpr(script)
	assertEqual(t, err, nil)

	p := NewProcedure(nil)
	p = p.ProcessLog(js, nodes)
	assertEqual(t, p.lastErr, nil)

	r := gjson.GetBytes(p.Content, "clientip")
	assertEqual(t, r.String(), "127.0.0.1")
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


func TestExprFunc(t *testing.T) {
	js := `{"a":{"first":2.3,"second":2,"thrid":"abc","forth":true},"age":47}`

	script := `expr(a.second*10+(2+3)*5, bb);`
	nodes, err := parser.ParseFuncExpr(script)
	assertEqual(t, err, nil)

	p := NewProcedure(nil)
	p = p.ProcessLog(js, nodes)
	assertEqual(t, p.lastErr, nil)

	r := gjson.GetBytes(p.Content, "bb")
	assertEqual(t, r.String(), "45")
}

func TestExprFunc(t *testing.T) {
	js := `{"a":{"first":2.3,"second":2,"url":"abc","forth":true},"age":47}`

	script := `urlencode(a.url, bb);`
	nodes, err := parser.ParseFuncExpr(script)
	assertEqual(t, err, nil)

	p := NewProcedure(nil)
	p = p.ProcessLog(js, nodes)
	assertEqual(t, p.lastErr, nil)

	r := gjson.GetBytes(p.Content, "bb")
	assertEqual(t, r.String(), "45")
}

func TestCastFloat2IntFunc(t *testing.T) {
	js := `{"a":{"first":2.3,"second":2,"thrid":"abc","forth":true},"age":47}`

	script := `cast(bb, a.first, "int");`
	nodes, err := parser.ParseFuncExpr(script)
	assertEqual(t, err, nil)

	p := NewProcedure(nil)
	p = p.ProcessLog(js, nodes)
	assertEqual(t, p.lastErr, nil)

	r := gjson.GetBytes(p.Content, "bb")
	assertEqual(t, r.String(), "2")
}

func TestCastInt2FloatFunc(t *testing.T) {
	js := `{"a":{"first":2.3,"second":2,"thrid":"abc","forth":true},"age":47}`

	script := `cast(bb, a.second, "float");`
	nodes, err := parser.ParseFuncExpr(script)
	assertEqual(t, err, nil)

	p := NewProcedure(nil)
	p = p.ProcessLog(js, nodes)
	assertEqual(t, p.lastErr, nil)

	r := gjson.GetBytes(p.Content, "bb")
	assertEqual(t, r.String(), "2")
}

func TestStringfFunc(t *testing.T) {
	js := `{"a":{"first":2.3,"second":2,"thrid":"abc","forth":true},"age":47}`

	script := `stringf(bb, "%d %s %v", a.second, a.thrid, a.forth);`
	nodes, err := parser.ParseFuncExpr(script)
	assertEqual(t, err, nil)

	p := NewProcedure(nil)
	p = p.ProcessLog(js, nodes)
	assertEqual(t, p.lastErr, nil)

	r := gjson.GetBytes(p.Content, "bb")
	assertEqual(t, r.String(), "2 abc true")
}

func TestScriptFunc(t *testing.T) {
	js := `{"a":{"first":2.3,"second":2,"thrid":"abc","forth":true},"age":47}`

	script := `rename(a.second, bb);
expr(a.second*10+(2+3)*5, cc);
cast(dd, a.first, "int");

stringf(ee, "%d %s %v", a.second, a.thrid, a.forth)
`
	nodes, err := parser.ParseFuncExpr(script)
	assertEqual(t, err, nil)

	p := NewProcedure(nil)
	p = p.ProcessLog(js, nodes)
	assertEqual(t, p.lastErr, nil)

	r := gjson.GetBytes(p.Content, "bb")
	assertEqual(t, r.String(), "2")

	r = gjson.GetBytes(p.Content, "cc")
	assertEqual(t, r.String(), "45")

	r = gjson.GetBytes(p.Content, "dd")
	assertEqual(t, r.String(), "2")

	r = gjson.GetBytes(p.Content, "ee")
	assertEqual(t, r.String(), "2 abc true")
}