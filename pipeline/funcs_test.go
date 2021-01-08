package pipeline

import (
	"testing"

	"github.com/tidwall/gjson"
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

	p := NewProcedure(script)

	p.ProcessText(js)

	assertEqual(t, p.lastErr, nil)

	r := gjson.GetBytes(p.Content, "clientip")
	assertEqual(t, r.String(), "127.0.0.1")
}

func TestRenameFunc(t *testing.T) {
	js := `{"a":{"first":2.3,"second":2,"thrid":"abc","forth":true},"age":47}`
	script := `rename(a.second, bb);`

	p := NewProcedure(script)
	p.ProcessText(js)

	assertEqual(t, p.lastErr, nil)

	r := gjson.GetBytes(p.Content, "bb")
	assertEqual(t, r.String(), "2")
}

func TestExprFunc(t *testing.T) {
	js := `{"a":{"first":2.3,"second":2,"thrid":"abc","forth":true},"age":47}`
	script := `expr(a.second*10+(2+3)*5, bb);`

	p := NewProcedure(script)
	p.ProcessText(js)

	r := gjson.GetBytes(p.Content, "bb")
	assertEqual(t, r.String(), "45")
}

func TestUrlencodeFunc(t *testing.T) {
	js := `{"a":{"url":"http://www.example.org/default.html?ct=32&op=92&item=98","second":2},"age":47}`
	script := `urldecode(a.url, bb);`

	p := NewProcedure(script)
	p.ProcessText(js)

	r := gjson.GetBytes(p.Content, "bb")
	assertEqual(t, r.String(), "45")
}

func TestDatetimeFunc(t *testing.T) {
	js := `{"a":{"date":"2021.01.07 12:12", "second":2},"age":47}`
	script := `datetime(a.date, 'yyyy-mm-dd hh:MM:ss', new_date);`

	p := NewProcedure(script)
	p.ProcessText(js)

	r := gjson.GetBytes(p.Content, "bb")
	assertEqual(t, r.String(), "45")
}

func TestGroupFunc(t *testing.T) {
	js := `{"a":{"status": 200,"age":47}`
	script := `group(a.status, [200-299], "ok", bb);`

	p := NewProcedure(script)
	p.ProcessText(js)

	r := gjson.GetBytes(p.Content, "bb")
	assertEqual(t, r.String(), "45")
}

func TestGroupInFunc(t *testing.T) {
	js := `{"a":{"status": 200,"age":47}`
	script := `group(a.status, [200, 201], "ok", bb);`

	p := NewProcedure(script)
	p.ProcessText(js)

	r := gjson.GetBytes(p.Content, "bb")
	assertEqual(t, r.String(), "45")
}

func TestCastFloat2IntFunc(t *testing.T) {
	js := `{"a":{"first":2.3,"second":2,"thrid":"abc","forth":true},"age":47}`
	script := `cast(bb, a.first, "int");`

	p := NewProcedure(script)
	p.ProcessText(js)

	r := gjson.GetBytes(p.Content, "bb")
	assertEqual(t, r.String(), "2")
}

func TestCastInt2FloatFunc(t *testing.T) {
	js := `{"a":{"first":2.3,"second":2,"thrid":"abc","forth":true},"age":47}`
	script := `cast(bb, a.second, "float");`

	p := NewProcedure(script)
	p.ProcessText(js)

	r := gjson.GetBytes(p.Content, "bb")
	assertEqual(t, r.String(), "2")
}

func TestStringfFunc(t *testing.T) {
	js := `{"a":{"first":2.3,"second":2,"thrid":"abc","forth":true},"age":47}`
	script := `stringf(bb, "%d %s %v", a.second, a.thrid, a.forth);`

	p := NewProcedure(script)
	p.ProcessText(js)

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
	p := NewProcedure(script)
	p.ProcessText(js)

	r := gjson.GetBytes(p.Content, "bb")
	assertEqual(t, r.String(), "2")

	r = gjson.GetBytes(p.Content, "cc")
	assertEqual(t, r.String(), "45")

	r = gjson.GetBytes(p.Content, "dd")
	assertEqual(t, r.String(), "2")

	r = gjson.GetBytes(p.Content, "ee")
	assertEqual(t, r.String(), "2 abc true")
}
