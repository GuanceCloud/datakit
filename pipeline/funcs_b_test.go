package pipeline

import (
	"strconv"
	"testing"
	"fmt"
	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/geo"
)

type funcCase struct {
	desc     string
	data     string
	script   string
	expected interface{}
	key      string
	err      error
	fail     bool
}

type EscapeError string

func (e EscapeError) Error() string {
	return "invalid URL escape " + strconv.Quote(string(e))
}

func TestJsonFunc(t *testing.T) {
	var testCase = []*funcCase{
		// {
		// 	data:     `{
		// 	  "name": {"first": "Tom", "last": "Anderson"},
		// 	  "age":37,
		// 	  "children": ["Sara","Alex","Jack"],
		// 	  "fav.movie": "Deer Hunter",
		// 	  "friends": [
		// 	    ["ig", "fb", "tw"],
		// 	    ["fb", "tw"],
		// 	    ["ig", "tw"]
		// 	  ]
		// 	}`,
		// 	script:   `json(_, friends[0][0])`,
		// 	expected: "ig",
		// 	key:      "friends[0][0]",
		// 	err:      nil,
		// },
		{
			data:     `{
			  "name": {"first": "Tom", "last": "Anderson"},
			  "age":37,
			  "children": ["Sara","Alex","Jack"],
			  "fav.movie": "Deer Hunter",
			  "friends": [
			    {"first": "Dale", "last": "Murphy", "age": 44, "nets": ["ig", "fb", "tw"]},
			    {"first": "Roger", "last": "Craig", "age": 68, "nets": ["fb", "tw"]},
			    {"first": "Jane", "last": "Murphy", "age": 47, "nets": ["ig", "tw"]}
			  ]
			}`,
			script:   `json(_, name) json(name, first)`,
			expected: "Tom",
			key:      "first",
			err:      nil,
		},
		{
			data:`[
				    {"first": "Dale", "last": "Murphy", "age": 44, "nets": ["ig", "fb", "tw"]},
				    {"first": "Roger", "last": "Craig", "age": 68, "nets": ["fb", "tw"]},
				    {"first": "Jane", "last": "Murphy", "age": 47, "nets": ["ig", "tw"]}
				]`,
			script:   `json(_, [0].nets[-1])`,
			expected: "tw",
			key:      "[0].nets[-1]",
			err:      nil,
		},
	}

	for _, tt := range testCase {
		p, err := NewPipeline(tt.script)

		assert.Equal(t, err, nil)

		p.Run(tt.data)

		r, err := p.getContentStr(tt.key)

		fmt.Println("======>", p.Output)

		assert.Equal(t, r, tt.expected)
	}
}

func TestDefaultTimeFunc(t *testing.T) {
	var testCase = []*funcCase{
		{
			data:     `{"a":{"time":"","second":2,"thrid":"abc","forth":true},"age":47}`,
			script:   `json(_, a.time) default_time(a.time)`,
			expected: nil,
			key:      "a.time",
			err:      nil,
		},
		{
			data:     `{"a":{"time":"06/Jan/2017:16:16:37 +0000","second":2,"thrid":"abc","forth":true},"age":47}`,
			script:   `json(_, a.time) default_time(a.time)`,
			expected: "1483719397000000000",
			key:      "a.time",
			err:      nil,
		},
		{
			data:     `{"a":{"time":"2014-12-16 06:20:00 UTC","second":2,"thrid":"abc","forth":true},"age":47}`,
			script:   `json(_, a.time) default_time(a.time)`,
			expected: "1418682000000000000",
			key:      "a.time",
			err:      nil,
		},
	}

	for _, tt := range testCase {
		p, err := NewPipeline(tt.script)

		assert.Equal(t, err, nil)

		p.Run(tt.data)

		r, err := p.getContentStr(tt.key)

		assert.Equal(t, r, tt.expected)
	}
}

func TestUrlencodeFunc(t *testing.T) {
	var testCase = []*funcCase{
		{
			data:     `{"url[0]":"+%3F%26%3D%23%2B%25%21%3C%3E%23%22%7B%7D%7C%5C%5E%5B%5D%60%E2%98%BA%09%3A%2F%40%24%27%28%29%2A%2C%3B","second":2}`,
			script:   "json(_, `url[0]`) url_decode(`url[0]`)",
			expected: " ?&=#+%!<>#\"{}|\\^[]`☺\t:/@$'()*,;",
			key:      "url[0]",
			err:      nil,
		},
		{
			data:     `{"url":"http%3a%2f%2fwww.baidu.com%2fs%3fwd%3d%e6%b5%8b%e8%af%95","second":2}`,
			script:   `json(_, url) url_decode(url)`,
			expected: `http://www.baidu.com/s?wd=测试`,
			key:      "url",
			err:      nil,
		},
		{
			data:     `{"url":"","second":2}`,
			script:   `json(_, url) url_decode(url)`,
			expected: ``,
			key:      "url",
			err:      nil,
		},
		{
			data:     `{"url":"+%3F%26%3D%23%2B%25%21%3C%3E%23%22%7B%7D%7C%5C%5E%5B%5D%60%E2%98%BA%09%3A%2F%40%24%27%28%29%2A%2C%3B","second":2}`,
			script:   `json(_, url) url_decode(url)`,
			expected: " ?&=#+%!<>#\"{}|\\^[]`☺\t:/@$'()*,;",
			key:      "url",
			err:      nil,
		},
		{
			data:     `{"url":"+%3F%26%3D%23%2B%25%21%3C%3E%23%22%7B%7D%7C%5C%5E%5B%5D%60%E2%98%BA%09%3A%2F%40%24%27%28%29%2A%2C%3B","second":2}`,
			script:   `json(_, url) url_decode("url", "aaa")`,
			expected: "+%3F%26%3D%23%2B%25%21%3C%3E%23%22%7B%7D%7C%5C%5E%5B%5D%60%E2%98%BA%09%3A%2F%40%24%27%28%29%2A%2C%3B",
			key:      "url",
			err:      nil,
		},
		{
			data:     `{"aa":{"url":"http%3a%2f%2fwww.baidu.com%2fs%3fwd%3d%e6%b5%8b%e8%af%95"},"second":2}`,
			script:   `json(_, aa.url) url_decode(aa.url)`,
			expected: `http://www.baidu.com/s?wd=测试`,
			key:      "aa.url",
			err:      nil,
		},
		{
			data:     `{"aa":{"aa.url":"http%3a%2f%2fwww.baidu.com%2fs%3fwd%3d%e6%b5%8b%e8%af%95"},"second":2}`,
			script:   "json(_, aa.`aa.url`) url_decode(aa.`aa.url`)",
			expected: `http://www.baidu.com/s?wd=测试`,
			key:      "aa.aa.url",
			err:      nil,
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
		{
			data:     `{"ip":"116.228.89.206", "second":2,"thrid":"abc","forth":true}`,
			script:   `json(_, ip) geoip(ip)`,
			expected: "Shanghai",
			key:      "city",
			err:      nil,
		},
		{
			data:     `{"ip":"192.168.0.1", "second":2,"thrid":"abc","forth":true}`,
			script:   `json(_, ip) geoip(ip)`,
			expected: "-",
			key:      "city",
			err:      nil,
		},
		{
			data:     `{"ip":"", "second":2,"thrid":"abc","forth":true}`,
			script:   `json(_, "ip") geoip(ip)`,
			expected: "unknown",
			key:      "city",
			err:      nil,
		},
		{
			data:     `{"aa": {"ip":"116.228.89.206"}, "second":2,"thrid":"abc","forth":true}`,
			script:   `json(_, aa.ip) geoip("aa.ip")`,
			expected: "Shanghai",
			key:      "city",
			err:      nil,
		},
	}

	geo.Init()

	for _, tt := range testCase {
		p, err := NewPipeline(tt.script)
		assertEqual(t, err, p.lastErr)

		p.Run(tt.data)

		r, err := p.getContentStr(tt.key)

		assertEqual(t, r, tt.expected)
	}
}

func TestUserAgentFunc(t *testing.T) {
	var testCase = []*funcCase{
		{
			data:     `{"userAgent":"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/36.0.1985.125 Safari/537.36", "second":2,"thrid":"abc","forth":true}`,
			script:   `json(_, userAgent) user_agent(userAgent)`,
			expected: "Windows 7",
			key:      "os",
			err:      nil,
		},
		{
			data:     `{"userAgent":"Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"}`,
			script:   `json(_, userAgent) user_agent(userAgent)`,
			expected: "Googlebot",
			key:      "browser",
			err:      nil,
		},
		{
			data:     `{"userAgent":"Mozilla/5.0 (iPhone; CPU iPhone OS 6_0 like Mac OS X) AppleWebKit/536.26 (KHTML, like Gecko) Version/6.0 Mobile/10A5376e Safari/8536.25 (compatible; Googlebot/2.1; +http://www.google.com/bot.html"}`,
			script:   `json(_, userAgent) user_agent(userAgent)`,
			expected: "",
			key:      "engine",
			err:      nil,
		},
		{
			data:     `{"userAgent":"Mozilla/5.0 (compatible; bingbot/2.0; +http://www.bing.com/bingbot.htm)"}`,
			script:   `json(_, userAgent) user_agent(userAgent)`,
			expected: "bingbot",
			key:      "browser",
			err:      nil,
		},
	}

	for _, tt := range testCase {
		p, err := NewPipeline(tt.script)
		assertEqual(t, err, p.lastErr)

		p.Run(tt.data)

		r, err := p.getContentStr(tt.key)

		assertEqual(t, r, tt.expected)
	}
}

func TestDatetimeFunc(t *testing.T) {
	var testCase = []*funcCase{
		{
			data:     `{"a":{"timestamp": "1610960605", "second":2},"age":47}`,
			script:   `json(_, a.timestamp) datetime(a.timestamp, 's', 'RFC3339')`,
			expected: "2021-01-18T17:03:25+08:00",
			key:      "a.timestamp",
			err:      nil,
		},
		{
			data:     `{"a":{"timestamp": "1610960605000", "second":2},"age":47}`,
			script:   `json(_, a.timestamp) datetime(a.timestamp, 'ms', 'RFC3339')`,
			expected: "2021-01-18T17:03:25+08:00",
			key:      "a.timestamp",
			err:      nil,
		},
	}

	for _, tt := range testCase {
		p, err := NewPipeline(tt.script)
		assertEqual(t, err, p.lastErr)

		p.Run(tt.data)

		r, err := p.getContentStr(tt.key)

		assertEqual(t, r, tt.expected)
	}
}

func TestGroupFunc(t *testing.T) {
	var testCase = []*funcCase{
		{
			data:     `{"status": 200,"age":47}`,
			script:   `json(_, status) group_between(status, [200, 400], false, newkey)`,
			expected: false,
			key:      "newkey",
			err:      nil,
		},
		{
			data:     `{"status": 200,"age":47}`,
			script:   `json(_, status) group_between(status, [200, 400], 10, newkey)`,
			expected: int64(10),
			key:      "newkey",
			err:      nil,
		},
		{
			data:     `{"status": 200,"age":47}`,
			script:   `json(_, status) group_between(status, [, 400], "ok", newkey)`,
			expected: nil,
			key:      "newkey",
			err:      nil,
		},
		{
			data:     `{"status": 200,"age":47}`,
			script:   `json(_, status) group_between(status, [200, 400], "ok", newkey)`,
			expected: "ok",
			key:      "newkey",
			err:      nil,
		},
		{
			data:     `{"status": 200,"age":47}`,
			script:   `json(_, status) group_between(status, [200, 299], "ok")`,
			expected: "ok",
			key:      "status",
			err:      nil,
		},
		{
			data:     `{"status": 200,"age":47}`,
			script:   `json(_, status) group_between(status, [299, 200], "ok")`,
			expected: float64(200),
			key:      "status",
			err:      nil,
		},
		{
			data:     `{"status": 200,"age":47}`,
			script:   `json(_, status) group_between(status, [299, 200], "ok", newkey)`,
			expected: float64(200),
			key:      "status",
			err:      nil,
		},
		{
			data:     `{"status": 200,"age":47}`,
			script:   `json(_, status) group_between(status, [200, 299], "ok", newkey)`,
			expected: "ok",
			key:      "newkey",
			err:      nil,
		},
		{
			data:     `{"status": 200,"age":47}`,
			script:   `json(_, status) group_between(status, [300, 400], "ok", newkey)`,
			expected: nil,
			key:      "newkey",
			err:      nil,
		},
	}

	for _, tt := range testCase {
		p, err := NewPipeline(tt.script)
		assertEqual(t, err, p.lastErr)

		p.Run(tt.data)

		r, err := p.getContent(tt.key)

		assertEqual(t, r, tt.expected)
	}
}

func TestGroupInFunc(t *testing.T) {
	var testCase = []*funcCase{
		{
			data:     `{"status": true,"age":"47"}`,
			script:   `json(_, status) group_in(status, [true], "ok", "newkey")`,
			expected: "ok",
			key:      "newkey",
			err:      nil,
		},
		{
			data:     `{"status": true,"age":"47"}`,
			script:   `json(_, status) group_in(status, [true], "ok", "newkey")`,
			expected: "ok",
			key:      "newkey",
			err:      nil,
		},
		{
			data:     `{"status": "aa","age":"47"}`,
			script:   `json(_, status) group_in(status, [], "ok")`,
			expected: "aa",
			key:      "status",
			err:      nil,
		},
		{
			data:     `{"status": "aa","age":"47"}`,
			script:   `json(_, status) group_in(status, ["aa"], "ok", "newkey")`,
			expected: "ok",
			key:      "newkey",
			err:      nil,
		},
		{
			data:     `{"status": "test","age":"47"}`,
			script:   `json(_, status) group_in(status, [200, 47, "test"], "ok", newkey)`,
			expected: "ok",
			key:      "newkey",
			err:      nil,
		},
		{
			data:     `{"status": "test","age":"47"}`,
			script:   `json(_, status) group_in(status, [200, "test"], "ok")`,
			expected: "ok",
			key:      "status",
			err:      nil,
		},
	}

	for _, tt := range testCase {
		p, err := NewPipeline(tt.script)
		assertEqual(t, err, p.lastErr)

		p.Run(tt.data)

		r, err := p.getContentStr(tt.key)

		assertEqual(t, r, tt.expected)
	}
}

func TestNullIfFunc(t *testing.T) {
	var testCase = []*funcCase{
		{
			data:     `{"a":{"first": 1,"second":2,"thrid":"aBC","forth":true},"age":47}`,
			script:   `json(_, a.first) nullif(a.first, "1")`,
			expected: float64(1),
			key:      "a.first",
			err:      nil,
		},
		{
			data:     `{"a":{"first": "1","second":2,"thrid":"aBC","forth":true},"age":47}`,
			script:   `json(_, a.first) nullif(a.first, 1)`,
			expected: "1",
			key:      "a.first",
			err:      nil,
		},
		{
			data:     `{"a":{"first": "","second":2,"thrid":"aBC","forth":true},"age":47}`,
			script:   `json(_, a.first) nullif(a.first, "")`,
			expected: nil,
			key:      "a.first",
			err:      nil,
		},
		{
			data:     `{"a":{"first": null,"second":2,"thrid":"aBC","forth":true},"age":47}`,
			script:   `json(_, a.first) nullif(a.first, nil)`,
			expected: nil,
			key:      "a.first",
			err:      nil,
		},
		{
			data:     `{"a":{"first": true,"second":2,"thrid":"aBC","forth":true},"age":47}`,
			script:   `json(_, a.first) nullif(a.first, true)`,
			expected: nil,
			key:      "a.first",
			err:      nil,
		},
		{
			data:     `{"a":{"first": 2.3, "second":2,"thrid":"aBC","forth":true},"age":47}`,
			script:   `json(_, a.first) nullif(a.first, 2.3)`,
			expected: nil,
			key:      "a.first",
			err:      nil,
		},
		{
			data:     `{"a":{"first": 2,"second":2,"thrid":"aBC","forth":true},"age":47}`,
			script:   `json(_, a.first) nullif(a.first, 2, "newkey")`,
			expected: nil,
			key:      "newkey",
			err:      nil,
		},
		{
			data:     `{"a":{"first":"2.3","second":2,"thrid":"aBC","forth":true},"age":47}`,
			script:   `json(_, a.first) nullif(a.first, "2.3", "newkey")`,
			expected: nil,
			key:      "newkey",
			err:      nil,
		},
	}

	for _, tt := range testCase {
		p, err := NewPipeline(tt.script)
		assertEqual(t, err, p.lastErr)

		p.Run(tt.data)

		r, err := p.getContent(tt.key)

		assertEqual(t, r, tt.expected)
	}
}

func TestJsonAllFunc(t *testing.T) {
	var testCase = []*funcCase{
		{
			data:     `{
			  "name": {"first": "Tom", "last": "Anderson"},
			  "age":37,
			  "children": ["Sara","Alex","Jack"],
			  "fav.movie": "Deer Hunter",
			  "friends": [
			    {"first": "Dale", "last": "Murphy", "age": 44, "nets": ["ig", "fb", "tw"]},
			    {"first": "Roger", "last": "Craig", "age": 68, "nets": ["fb", "tw"]},
			    {"first": "Jane", "last": "Murphy", "age": 47, "nets": ["ig", "tw"]}
			  ]
			}`,
			script:   `json_all()`,
			expected: "Sara",
			key:      "children[0]",
			err:      nil,
		},
		{
			data: `[
			    {"first": "Dale", "last": "Murphy", "age": 44, "nets": ["ig", "fb", "tw"]},
			    {"first": "Roger", "last": "Craig", "age": 68, "nets": ["fb", "tw"]},
			    {"first": "Jane", "last": "Murphy", "age": 47, "nets": ["ig", "tw"]}
			  ]`,
			script:   `json_all()`,
			expected: "Dale",
			key:      "[0].first",
			err:      nil,
		},
	}

	for _, tt := range testCase {
		p, err := NewPipeline(tt.script)
		assertEqual(t, err, p.lastErr)

		p.Run(tt.data)

		r, err := p.getContentStr(tt.key)

		fmt.Println("======>", p.Output)

		assertEqual(t, r, tt.expected)
	}

	js := `
{
  "name": {"first": "Tom", "last": "Anderson"},
  "age":37,
  "children": ["Sara","Alex","Jack"],
  "fav.movie": "Deer Hunter",
  "friends": [
    {"first": "Dale", "last": "Murphy", "age": 44, "nets": ["ig", "fb", "tw"]},
    {"first": "Roger", "last": "Craig", "age": 68, "nets": ["fb", "tw"]},
    {"first": "Jane", "last": "Murphy", "age": 47, "nets": ["ig", "tw"]}
  ]
}
`
	script := `json_all()`

	p, err := NewPipeline(script)
	assertEqual(t, err, nil)

	p.Run(js)

	r, err := p.getContent("children[0]")

	assertEqual(t, r, "Sara")
}
