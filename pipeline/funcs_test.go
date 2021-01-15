package pipeline

import (
	"testing"
	"strconv"
	"fmt"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/geo"
)

type funcCase struct {
	data     string
	script   string
	expected string
	key      string
	err      error
	fail     bool
}

type EscapeError string

func (e EscapeError) Error() string {
	return "invalid URL escape " + strconv.Quote(string(e))
}

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

func TestExprFunc(t *testing.T) {

	js := `{"a":{"first":2.3,"second":2,"thrid":"abc","forth":true},"age":47}`
	script := `json(_, a.second)
cast(a.second, "int")
expr(a.second*10+(2+3)*5, bb)
`
	p, err := NewPipeline(script)
	assertEqual(t, err, nil)

	p.Run(js)
	assertEqual(t, p.lastErr, nil)

	v, _ := p.getContentStr("bb")
	assertEqual(t, v, "45")
}

func TestDefaultTimeFunc(t *testing.T) {

	js := `{"a":{"time":"2014/04/08 22:05","second":2,"thrid":"abc","forth":true},"age":47}`
	script := `json(_, a.time)
default_time(a.second);
`
	p, err := NewPipeline(script)
	assertEqual(t, err, nil)

	p.Run(js)
	assertEqual(t, p.lastErr, nil)

	r, _ := p.getContent("a.time")

	assertEqual(t, r, 1396965900)
}

func TestUrlencodeFunc(t *testing.T) {
	var testCase = []*funcCase{
		{
			data: `{"url":"http%3a%2f%2fwww.baidu.com%2fs%3fwd%3d%e6%b5%8b%e8%af%95","second":2}`,
			script: `json(_, url) url_decode(url)`,
			expected: `http://www.baidu.com/s?wd=测试`,
			key: "url",
			err: nil,
		},
		{
			data: `{"url":"","second":2}`,
			script: `json(_, url) url_decode(url)`,
			expected: ``,
			key: "url",
			err: nil,
		},
		{
			data: `{"url":"+%3F%26%3D%23%2B%25%21%3C%3E%23%22%7B%7D%7C%5C%5E%5B%5D%60%E2%98%BA%09%3A%2F%40%24%27%28%29%2A%2C%3B","second":2}`,
			script: `json(_, url) url_decode(url)`,
			expected: " ?&=#+%!<>#\"{}|\\^[]`☺\t:/@$'()*,;",
			key: "url",
			err: nil,
		},
		{
			data: `{"url[0]":"+%3F%26%3D%23%2B%25%21%3C%3E%23%22%7B%7D%7C%5C%5E%5B%5D%60%E2%98%BA%09%3A%2F%40%24%27%28%29%2A%2C%3B","second":2}`,
			script: `json(_, "url[0]") url_decode("url[0]")`,
			expected: " ?&=#+%!<>#\"{}|\\^[]`☺\t:/@$'()*,;",
			key: "url[0]",
			err: nil,
		},
		{
			data: `{"url":"+%3F%26%3D%23%2B%25%21%3C%3E%23%22%7B%7D%7C%5C%5E%5B%5D%60%E2%98%BA%09%3A%2F%40%24%27%28%29%2A%2C%3B","second":2}`,
			script: `json(_, "url") url_decode("url", "aaa")`,
			expected: " ?&=#+%!<>#\"{}|\\^[]`☺\t:/@$'()*,;",
			key: "url",
			err: nil,
		},
		{
			data: `{"aa":{"url":"http%3a%2f%2fwww.baidu.com%2fs%3fwd%3d%e6%b5%8b%e8%af%95"},"second":2}`,
			script: `json(_, aa.url) url_decode(aa.url)`,
			expected: `http://www.baidu.com/s?wd=测试`,
			key: "aa.url",
			err: nil,
		},
		{
			data: `{"aa":{"url":"http%3a%2f%2fwww.baidu.com%2fs%3fwd%3d%e6%b5%8b%e8%af%95"},"second":2}`,
			script: `json(_, "aa.url") url_decode("aa.url")`,
			expected: `http://www.baidu.com/s?wd=测试`,
			key: "aa.url",
			err: nil,
		},
		{
			data: `{"aa":{"aa.url":"http%3a%2f%2fwww.baidu.com%2fs%3fwd%3d%e6%b5%8b%e8%af%95"},"second":2}`,
			script: `json(_, aa."aa.url") url_decode("aa.aa.url")`,
			expected: `http://www.baidu.com/s?wd=测试`,
			key: "aa.aa.url",
			err: nil,
		},
	}

	for _, tt := range testCase {
		p, err := NewPipeline(tt.script)

		assertEqual(t, err, nil)

		p.Run(tt.data)

		r, err := p.getContentStr(tt.key)

		assertEqual(t, r, tt.expected)
	}
}

func TestGeoIpFunc(t *testing.T) {
	var testCase = []*funcCase{
		// {
		// 	data: `{"ip":"116.228.89.206", "second":2,"thrid":"abc","forth":true}`,
		// 	script: `json(_, ip) geoip(ip)`,
		// 	expected: "Shanghai",
		// 	key: "city",
		// 	err: nil,
		// },
		// {
		// 	data: `{"ip":"192.168.0.1", "second":2,"thrid":"abc","forth":true}`,
		// 	script: `json(_, ip) geoip(ip)`,
		// 	expected: "-",
		// 	key: "city",
		// 	err: nil,
		// },
		{
			data: `{"ip":"192.168.0.1", "second":2,"thrid":"abc","forth":true}`,
			script: `json(_, "ip") geoip("ip")`,
			expected: "-",
			key: "city",
			err: nil,
		},
		// {
		// 	data: `{"ip":"192.168.0.1", "second":2,"thrid":"abc","forth":true}`,
		// 	script: `json(_, "ip") geoip(ip)`,
		// 	expected: "-",
		// 	key: "city",
		// 	err: nil,
		// },
	}

	geo.Init()

	for _, tt := range testCase {
		p, err := NewPipeline(tt.script)
		assertEqual(t, err, p.lastErr)

		p.Run(tt.data)

		r, err := p.getContentStr(tt.key)

		fmt.Println("=====>", p.Output)
		assertEqual(t, r, tt.expected)
	}
}

func TestUserAgentFunc(t *testing.T) {
	js := `{"a":{"userAgent":"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/36.0.1985.125 Safari/537.36","second":2},"age":47}`
	script := `json(_, a.userAgent)
user_agent(a.userAgent)
`
	p, err := NewPipeline(script)
	assertEqual(t, err, nil)

	p.Run(js)

	r, _ := p.getContentStr("os")

	assertEqual(t, r, "Windows 7")
}

func TestDatetimeFunc(t *testing.T) {
	js := `{"a":{"timestamp": "1610103765000", "second":2},"age":47}`
	script := `json(_, a.timestamp)
datetime(a.timestamp, 'ms', 'YYYY-MM-dd hh:mm:ss')`

	p, err := NewPipeline(script)
	assertEqual(t, err, nil)

	p.Run(js)

	r, _ := p.getContent("a.timestamp")

	assertEqual(t, r, "2021-01-08 07:02:45")
}

func TestGroupFunc(t *testing.T) {
	js := `{"status": 200,"age":47}`
	script := `json(_, status)
group_between(status, [200, 299], "ok", newkey)`

	p, err := NewPipeline(script)
	assertEqual(t, err, nil)

	p.Run(js)

	r, _ := p.getContent("newkey")

	assertEqual(t, r, "ok")
}

func TestGroupInFunc(t *testing.T) {
	js := `{"status": "test","age":"47"}`
	script := `json(_, status)
group_in(status, [200, 47, "test"], "ok", newkey)`

	p, err := NewPipeline(script)
	assertEqual(t, err, nil)

	p.Run(js)

	r, _ := p.getContent("newkey")
	assertEqual(t, r, "ok")
}

func TestCastFloat2IntFunc(t *testing.T) {
	js := `{"a":{"first":2.3,"second":2,"thrid":"abc","forth":true},"age":47}`
	script := `json(_, a.first)
cast(a.first, "int")
`
	p, err := NewPipeline(script)
	assertEqual(t, err, nil)

	p.Run(js)
	v, _ := p.getContentStr("a.first")

	assertEqual(t, v, "2")
}

func TestCastInt2FloatFunc(t *testing.T) {
	js := `{"a":{"first":2.3,"second":2,"thrid":"abc","forth":true},"age":47}`
	script := `json(_, a.second)
cast(a.second, "float")
`
	p, err := NewPipeline(script)
	assertEqual(t, err, nil)

	p.Run(js)

	v, _ := p.getContentStr("a.second")
	assertEqual(t, v, "2")
}

func TestStringfFunc(t *testing.T) {
	//js := `{"a":{"first":2.3,"second":2,"thrid":"abc","forth":true},"age":47}`
	//script := `stringf(bb, "%d %s %v", a.second, a.thrid, a.forth);`
	js := `{"a":{"first":2.3,"second":2,"thrid":"abc","forth":true},"age":47}`
	script := `json(_, a.second)
json(_, a.thrid)
cast(a.second, "int")
json(_, a.forth)
strfmt(bb, "%d %s %v", a.second, a.thrid, a.forth)
`
	p, err := NewPipeline(script)
	assertEqual(t, err, nil)

	p.Run(js)
	v, _ := p.getContent("bb")
	assertEqual(t, v, "2 abc true")
}

func TestUppercaseFunc(t *testing.T) {
	js := `{"a":{"first":2.3,"second":2,"thrid":"abc","forth":true},"age":47}`
	script := `json(_, a.thrid)
uppercase(a.thrid)
`
	p, err := NewPipeline(script)
	assertEqual(t, err, nil)

	p.Run(js)

	v, _ := p.getContent("a.thrid")
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
}

func TestNullIfFunc(t *testing.T) {
	js := `{"a":{"first":2.3,"second":2,"thrid":"aBC","forth":true},"age":47}`
	script := `json(_, a.first) nullif(a.first, 2.3)`
	p, err := NewPipeline(script)
	assertEqual(t, err, nil)

	p.Run(js)

	r, _ := p.getContent("a.first")
	assertEqual(t, r, nil)
}
