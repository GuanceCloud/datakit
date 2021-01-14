package pipeline

import (
	"testing"
)

func assertEqual(t *testing.T, a, b interface{}) {
	t.Helper()
	if a != b {
		t.Errorf("Not Equal. %d %d", a, b)
	}
}
func TestJsonFunc(t *testing.T) {
	script := `json(_, friends[-1])
json(_, name)
json(name, last, last_new)`
	js := `
{
  "name": {"first": "Tom", "last": "Anderson"},
  "age":37,
  "children": ["Sara","Alex","Jack"],
  "fav.movie": "Deer Hunter",
  "friends": [
    ["ig", "fb", "tw"],
    ["fb", "tw"],
    ["ig", "tw"]
  ]
}
`
	p, err := NewPipeline(script)
	assertEqual(t, err, nil)

	p.Run(js)
	t.Log(p.lastErr)
	//assertEqual(t, p.lastErr, nil)
	t.Log(p.lastErr)
	t.Log(p.Output)

}

func TestGrokFunc(t *testing.T) {
	script := `
add_pattern("_second", "(?:(?:[0-5]?[0-9]|60)(?:[:.,][0-9]+)?)")
add_pattern("_minute", "(?:[0-5][0-9])")
add_pattern("_hour", "(?:2[0123]|[01]?[0-9])")
add_pattern("time", "([^0-9]?)%{_hour:hour}:%{_minute:minute}(?::%{_second:second})([^0-9]?)")
grok(_, "%{time}")`

	p, err := NewPipeline(script)

	assertEqual(t, err, nil)

	p.Run("12:13:14")
	assertEqual(t, p.lastErr, nil)

	hour, _ := p.getContentStr("hour")
	assertEqual(t, hour, "12")

	minute, _ := p.getContentStr("minute")
	assertEqual(t, minute, "13")

	second, _ := p.getContentStr("second")
	assertEqual(t, second, "14")
}

func TestRenameFunc(t *testing.T) {
	script := `
add_pattern("_second", "(?:(?:[0-5]?[0-9]|60)(?:[:.,][0-9]+)?)")
add_pattern("_minute", "(?:[0-5][0-9])")
add_pattern("_hour", "(?:2[0123]|[01]?[0-9])")
add_pattern("time", "([^0-9]?)%{_hour:hour}:%{_minute:minute}(?::%{_second:second})([^0-9]?)")
grok(_, "%{time}")
rename(newhour, hour)`

	p, err := NewPipeline(script)
	assertEqual(t, err, nil)

	p.Run("12:13:14")
	assertEqual(t, p.lastErr, nil)

	r, _ := p.getContentStr("newhour")
	assertEqual(t, r, "12")
}
//
//func TestExprFunc(t *testing.T) {
//
//	js := `{"a":{"first":2.3,"second":2,"thrid":"abc","forth":true},"age":47}`
//	script := `json(_, a.second);
//cast(a.second, "int");
//expr(a.second*10+(2+3)*5, bb);
//`
//	p, err := NewPipeline(script)
//	assertEqual(t, err, nil)
//
//	p.Run(js)
//	assertEqual(t, p.lastErr, nil)
//
//	assertEqual(t, p.getContentStr("bb"), "45")
//}
//
//func TestDefaultTimeFunc(t *testing.T) {
//
//	js := `{"a":{"time":"2014/04/08 22:05","second":2,"thrid":"abc","forth":true},"age":47}`
//	script := `json(_, a.time);
//default_time(a.second);
//`
//	p, err := NewPipeline(script)
//	assertEqual(t, err, nil)
//
//	p.Run(js)
//	assertEqual(t, p.lastErr, nil)
//
//	r := p.getContent("a.time")
//
//	assertEqual(t, r, 1396965900)
//}
//
//func TestUrlencodeFunc(t *testing.T) {
//	js := `{"url":"http%3A%2F%2Fwww.baidu.com%2Fs%3Fwd%3D%E8%87%AA%E7%94%B1%E5%BA%A6","second":2}`
//	script := `json(_, url); url_decode(url);`
//
//	p, err := NewPipeline(script)
//	assertEqual(t, err, nil)
//
//	p.Run(js)
//
//	r := p.getContentStr("url")
//
//	assertEqual(t, r, "http://www.baidu.com/s?wd=自由度")
//}
//
//func TestGeoIpFunc(t *testing.T) {
//	js := `{"a":{"ip":"116.228.89.206", "second":2,"thrid":"abc","forth":true},"age":47}`
//	script := `json(_, a.ip); geoip(a.ip);`
//
//	p, err := NewPipeline(script)
//	assertEqual(t, err, nil)
//
//	p.Run(js)
//
//	r := p.getContentStr("city")
//	assertEqual(t, r, "Shanghai")
//}
//
//func TestUserAgentFunc(t *testing.T) {
//	js := `{"a":{"userAgent":"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/36.0.1985.125 Safari/537.36","second":2},"age":47}`
//	script := `json(_, a.userAgent); user_agent(a.userAgent)`
//
//	p, err := NewPipeline(script)
//	assertEqual(t, err, nil)
//
//	p.Run(js)
//
//	r := p.getContentStr("os")
//
//	assertEqual(t, r, "Windows 7")
//}
//
//func TestDatetimeFunc(t *testing.T) {
//	js := `{"a":{"timestamp": "1610103765000", "second":2},"age":47}`
//	script := `json(_, a.timestamp); datetime(a.timestamp, 'ms', 'YYYY-MM-dd hh:mm:ss');`
//
//	p, err := NewPipeline(script)
//	assertEqual(t, err, nil)
//
//	p.Run(js)
//
//	r := p.getContent("a.timestamp")
//
//	fmt.Println("====>", r)
//
//	assertEqual(t, r, "2021-01-08 07:02:45")
//}
//
//func TestGroupFunc(t *testing.T) {
//	js := `{"a":{"status": 200,"age":47}`
//	script := `json(_, a.status); group_between(a.status, [200, 299], "ok", newkey);`
//
//	p, err := NewPipeline(script)
//	assertEqual(t, err, nil)
//
//	p.Run(js)
//
//	r := p.getContent("newkey")
//
//	assertEqual(t, r, "ok")
//}
//
//func TestGroupInFunc(t *testing.T) {
//	js := `{"a":{"status": "test","age":"47"}`
//	script := `json(_, a.status); group_in(a.status, [200, 47, "test"], "ok", newkey);`
//
//	p, err := NewPipeline(script)
//	assertEqual(t, err, nil)
//
//	p.Run(js)
//
//	r := p.getContent("newkey")
//	assertEqual(t, r, "ok")
//}
//
//func TestCastFloat2IntFunc(t *testing.T) {
//	js := `{"a":{"first":2.3,"second":2,"thrid":"abc","forth":true},"age":47}`
//	script := `json(_, a.first);
//cast(a.first, "int");
//`
//	p, err := NewPipeline(script)
//	assertEqual(t, err, nil)
//
//	p.Run(js)
//
//	assertEqual(t, p.getContentStr("a.first"), "2")
//}
//
//func TestCastInt2FloatFunc(t *testing.T) {
//	js := `{"a":{"first":2.3,"second":2,"thrid":"abc","forth":true},"age":47}`
//	script := `json(_, a.second);
//cast(a.second, "float");
//`
//	p, err := NewPipeline(script)
//	assertEqual(t, err, nil)
//
//	p.Run(js)
//	assertEqual(t, p.getContentStr("a.second"), "2")
//}
//
//func TestStringfFunc(t *testing.T) {
//	//js := `{"a":{"first":2.3,"second":2,"thrid":"abc","forth":true},"age":47}`
//	//script := `stringf(bb, "%d %s %v", a.second, a.thrid, a.forth);`
//	js := `{"a":{"first":2.3,"second":2,"thrid":"abc","forth":true},"age":47}`
//	script := `json(_, a.second);
//json(_, a.thrid);
//json(_, a.forth);
//strfmt(bb, "%d %s %v", a.second, a.thrid, a.forth);
//`
//	p, err := NewPipeline(script)
//	assertEqual(t, err, nil)
//
//	p.Run(js)
//	assertEqual(t, p.getContent("bb"), "2 abc true")
//}
//
func TestUppercaseFunc(t *testing.T) {
	js := `{"a":{"first":2.3,"second":2,"thrid":"abc","forth":true},"age":47}`
	script := `json(_, a.thrid)
uppercase(a.thrid)
`
	p, err := NewPipeline(script)
	assertEqual(t, err, nil)

	p.Run(js)

	v, _ := p.getContent("a.thrid")
	t.Log(v, "ABC")
	t.Log(p.Output)
	assertEqual(t, v, "ABC")
}

func TestLowercaseFunc(t *testing.T) {
	js := `{"a":{"first":2.3,"second":2,"thrid":"aBC","forth":true},"age":47}`
	script := `json(_, a.thrid)
lowercase(a.thrid)
`
	p, err := NewPipeline(script)
	t.Log(err)
	assertEqual(t, err, nil)

	p.Run(js)
	v, _ := p.getContentStr("a.thrid")
	assertEqual(t, v, "abc")
}

func TestAddkeyFunc(t *testing.T) {
	js := `{"a":{"first":2.3,"second":2,"thrid":"aBC","forth":true},"age":47}`
	script := `add_key(aa, 3)
`
	p, err := NewPipeline(script)
	assertEqual(t, err, nil)

	p.Run(js)
	v, _ := p.getContentStr("aa")
	assertEqual(t, v, "3")
}

func TestDropkeyFunc(t *testing.T) {
	js := `{"a":{"first":2.3,"second":2,"thrid":"aBC","forth":true},"age":47}`
	script := `json(_, a.thrid)
json(_, a.first)
drop_key(a.thrid)
`
	p, err := NewPipeline(script)
	assertEqual(t, err, nil)

	p.Run(js)
	v, _ := p.getContentStr("a.first")
	assertEqual(t, v, "2.3")
	t.Log(p.Output)
}

//func TestNullIfFunc(t *testing.T) {
//	js := `{"a":{"first":2.3,"second":2,"thrid":"aBC","forth":true},"age":47}`
//	script := `json(_, a.first);
//nullif(a.first, 2.3)
//`
//	p, err := NewPipeline(script)
//	assertEqual(t, err, nil)
//
//	p.Run(js)
//
//	r := p.getContent("a.first")
//	assertEqual(t, r, nil)
//}
