package pipeline

import (
	"testing"
	"fmt"
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
	script := `grok(_, "%{COMMONAPACHELOG}");`

	p := NewPipeline(script)
	p.Run(js)
	assertEqual(t, p.lastErr, nil)

	r := p.getContentStr("clientip")
	assertEqual(t, r, "127.0.0.1")
}

func TestRenameFunc(t *testing.T) {
	js := `127.0.0.1 - - [23/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 207`
	script := `grok(_, "%{COMMONAPACHELOG}");
rename(newkey, clientip)`

	p := NewPipeline(script)
	p.Run(js)
	assertEqual(t, p.lastErr, nil)

	r := p.getContentStr("newkey")
	assertEqual(t, r, "127.0.0.1")
}

func TestExprFunc(t *testing.T) {

	js := `{"a":{"first":2.3,"second":2,"thrid":"abc","forth":true},"age":47}`
	script := `json(_, a.second);
cast(a.second, "int");
expr(a.second*10+(2+3)*5, bb);
`
	p := NewPipeline(script)
	p.Run(js)
	assertEqual(t, p.lastErr, nil)

	assertEqual(t, p.getContentStr("bb"), "45")
}

func TestUrlencodeFunc(t *testing.T) {
	js := `{"a":{"url":"http%3A%2F%2Fwww.baidu.com%2Fs%3Fwd%3D%E8%87%AA%E7%94%B1%E5%BA%A6","second":2},"age":47}`
	script := `url_decode(a.url);`

	p := NewProcedure(script)
	p.Run(js)

	r := p.getContentStr("a.url")

	fmt.Println("======>", r)

	assertEqual(t, r, "http://www.baidu.com/s?wd=自由度")
}

func TestUserAgentFunc(t *testing.T) {
	js := `{"a":{"userAgent":"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/36.0.1985.125 Safari/537.36","second":2},"age":47}`
	script := `useragent(a.userAgent, bb);`

	p := NewProcedure(script)
	p.ProcessText(js)

	r := gjson.GetBytes(p.Content, "bb")

	assertEqual(t, r.Get("Mobile").Bool(), false)
}

func TestDatetimeFunc(t *testing.T) {
	js := `{"a":{"date":"23/Apr/2014:22:58:32 +0200", "second":2},"age":47}`
	script := `datetime(a.date, 'yyyy-mm-dd hh:MM:ss', new_date);`

	p := NewProcedure(script)
	p.ProcessText(js)

	r := gjson.GetBytes(p.Content, "new_date")

	fmt.Println("result ==>", r.String())
	assertEqual(t, r.String(), "45")
	//js := `{"a":{"url":"http://www.example.org/default.html?ct=32&op=92&item=98","second":2},"age":47}`
	//script := `urldecode(a.url, bb);`
	//
	//p := NewProcedure(script)
	//p.ProcessText(js)
	//
	//r := gjson.GetBytes(p.Content, "bb")
	//assertEqual(t, r.String(), "45")
}

func TestDatetimeFunc(t *testing.T) {
	//js := `{"a":{"date":"2021.01.07 12:12", "second":2},"age":47}`
	//script := `datetime(a.date, 'yyyy-mm-dd hh:MM:ss', new_date);`
	//
	//p := NewProcedure(script)
	//p.ProcessText(js)
	//
	//r := gjson.GetBytes(p.Content, "bb")
	//assertEqual(t, r.String(), "45")
}

func TestGroupFunc(t *testing.T) {
	//js := `{"a":{"status": 200,"age":47}`
	//script := `group(a.status, [200-299], "ok", bb);`
	//
	//p := NewProcedure(script)
	//p.ProcessText(js)
	//
	//r := gjson.GetBytes(p.Content, "bb")
	//assertEqual(t, r.String(), "45")
}

func TestGroupInFunc(t *testing.T) {
	//js := `{"a":{"status": 200,"age":47}`
	//script := `group(a.status, [200, 201], "ok", bb);`
	//
	//p := NewProcedure(script)
	//p.ProcessText(js)
	//
	//r := gjson.GetBytes(p.Content, "bb")
	//assertEqual(t, r.String(), "45")
}

func TestCastFloat2IntFunc(t *testing.T) {
	js := `{"a":{"first":2.3,"second":2,"thrid":"abc","forth":true},"age":47}`
	script := `json(_, a.first);
cast(a.first, "int");
`
	p := NewPipeline(script)
	p.Run(js)

	assertEqual(t, p.getContentStr("a.first"), "2")
}

func TestCastInt2FloatFunc(t *testing.T) {
	js := `{"a":{"first":2.3,"second":2,"thrid":"abc","forth":true},"age":47}`
	script := `json(_, a.second);
cast(a.second, "float");
`
	p := NewPipeline(script)
	p.Run(js)
	assertEqual(t, p.getContentStr("a.second"), "2")
}

func TestStringfFunc(t *testing.T) {
	//js := `{"a":{"first":2.3,"second":2,"thrid":"abc","forth":true},"age":47}`
	//script := `stringf(bb, "%d %s %v", a.second, a.thrid, a.forth);`
	js := `{"a":{"first":2.3,"second":2,"thrid":"abc","forth":true},"age":47}`
	script := `json(_, a.second);
json(_, a.thrid);
json(_, a.forth);
strfmt(bb, "%d %s %v", a.second, a.thrid, a.forth);
`
	p := NewPipeline(script)
	p.Run(js)
	assertEqual(t, p.getContent("bb"), "2 abc true")
}

func TestUppercaseFunc(t *testing.T) {
	js := `{"a":{"first":2.3,"second":2,"thrid":"abc","forth":true},"age":47}`
	script := `json(_, a.thrid);
uppercase(a.thrid);
`
	p := NewPipeline(script)
	p.Run(js)
	assertEqual(t, p.getContentStr("a.thrid"), "ABC")
}

func TestLowercaseFunc(t *testing.T) {
	js := `{"a":{"first":2.3,"second":2,"thrid":"aBC","forth":true},"age":47}`
	script := `json(_, a.thrid);
lowercase(a.thrid);
`
	p := NewPipeline(script)
	p.Run(js)
	assertEqual(t, p.getContentStr("a.thrid"), "abc")
}
