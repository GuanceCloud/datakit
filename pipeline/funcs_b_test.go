package pipeline

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type funcCase struct {
	data     string
	script   string
	expected interface{}
	key      string
	fail     bool
}

type EscapeError string

func (e EscapeError) Error() string {
	return "invalid URL escape " + strconv.Quote(string(e))
}

func TestJsonFunc(t *testing.T) {
	testCase := []*funcCase{
		{
			data: `{
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
		},
		{
			data: `[
				    {"first": "Dale", "last": "Murphy", "age": 44, "nets": ["ig", "fb", "tw"]},
				    {"first": "Roger", "last": "Craig", "age": 68, "nets": ["fb", "tw"]},
				    {"first": "Jane", "last": "Murphy", "age": 47, "nets": ["ig", "tw"]}
				]`,
			script:   `json(_, [0].nets[-1])`,
			expected: "tw",
			key:      "[0].nets[-1]",
		},
	}

	for _, tt := range testCase {
		p, err := NewPipeline(tt.script)

		assert.Equal(t, err, nil)

		p.Run(tt.data)

		r, err := p.getContentStr(tt.key)

		assertEqual(t, err, nil)

		assert.Equal(t, r, tt.expected)
	}
}

func TestDefaultTimeFunc(t *testing.T) {
	testCase := []*funcCase{
		{
			data:     `{"a":{"time":"14 May 2019 19:11:40.164","second":2,"third":"abc","forth":true},"age":47}`,
			script:   `json(_, a.time) default_time(a.time)`,
			expected: int64(1557832300164000000),
			key:      "a.time",
		},
		{
			data:     `{"a":{"time":"06/Jan/2017:16:16:37 +0000","second":2,"third":"abc","forth":true},"age":47}`,
			script:   `json(_, a.time) default_time(a.time)`,
			expected: int64(1483719397000000000),
			key:      "a.time",
		},
		{
			data:     `{"a":{"time":"2014-12-16 06:20:00 UTC","second":2,"third":"abc","forth":true},"age":47}`,
			script:   `json(_, a.time) default_time(a.time)`,
			expected: int64(1418682000000000000),
			key:      "a.time",
		},
		{
			data:     `{"a":{"time":"171113 14:14:20","second":2,"third":"abc","forth":true},"age":47}`,
			script:   `json(_, a.time) default_time(a.time)`,
			expected: int64(1510582460000000000),
			key:      "a.time",
		},

		{
			data:     `{"str":"2021/02/27 - 08:11:46"}`,
			script:   `json(_, str) default_time(str)`,
			expected: int64(1614413506000000000),
			key:      "str",
		},

		{
			data:     `{"a":{"time":"2021-03-15 13:50:47,000"}}`,
			script:   `json(_, a.time) default_time(a.time)`,
			expected: int64(1615787447000000000),
			key:      "a.time",
		},

		{
			// data:     `{"a":{"time":"2021-03-15 13:50:47,000 UTC"}}`,
			data:     `{"a":{"time":"2021-03-15 13:50:47,000"}}`,
			script:   `json(_, a.time) default_time(a.time)`,
			expected: int64(1615787447000000000),
			key:      "a.time",
		},

		{
			data:     `{"a":{"time":"15 Mar 06:20:12.000"}}`,
			script:   `json(_, a.time) default_time(a.time)`,
			expected: int64(time.Second * 1615789212),
			key:      "a.time",
		},

		{
			data:     `{"a":{"time":"Wed Mar 17 15:40:49 CST 2021"}}`,
			script:   `json(_, a.time) default_time(a.time)`,
			expected: int64(time.Second * 1615966849),
			key:      "a.time",
		},
		{
			data:     `{"a":{"time":"Wed Mar 17 15:40:49 CST 2021"}}`,
			script:   `json(_, a.time) default_time(a.time, "Asia/Shanghai")`,
			expected: int64(time.Second * 1615966849),
			key:      "a.time",
		},
	}

	for idx, tt := range testCase {
		t.Logf("-=-=-=-=-=-=-=-=[ %d ]-=-=-=-=-=-=-=-=-=-=", idx+1)
		p, err := NewPipeline(tt.script)

		assert.Equal(t, err, nil)

		p.Run(tt.data)

		r, err := p.getContent(tt.key)
		assertEqual(t, err, nil)

		ok := assert.Equal(t, r, tt.expected)

		t.Logf("[passed? %v]out: %s <> %v", ok, tt.data, p.Output)
	}
}

func TestUrlencodeFunc(t *testing.T) {
	testCase := []*funcCase{
		{
			data:     `{"url[0]":"+%3F%26%3D%23%2B%25%21%3C%3E%23%22%7B%7D%7C%5C%5E%5B%5D%60%E2%98%BA%09%3A%2F%40%24%27%28%29%2A%2C%3B","second":2}`,
			script:   "json(_, `url[0]`) url_decode(`url[0]`)",
			expected: " ?&=#+%!<>#\"{}|\\^[]`☺\t:/@$'()*,;",
			key:      "url[0]",
		},
		{
			data:     `{"url":"http%3a%2f%2fwww.baidu.com%2fs%3fwd%3d%e6%b5%8b%e8%af%95","second":2}`,
			script:   `json(_, url) url_decode(url)`,
			expected: `http://www.baidu.com/s?wd=测试`,
			key:      "url",
		},
		{
			data:     `{"url":"","second":2}`,
			script:   `json(_, url) url_decode(url)`,
			expected: ``,
			key:      "url",
		},
		{
			data:     `{"url":"+%3F%26%3D%23%2B%25%21%3C%3E%23%22%7B%7D%7C%5C%5E%5B%5D%60%E2%98%BA%09%3A%2F%40%24%27%28%29%2A%2C%3B","second":2}`,
			script:   `json(_, url) url_decode(url)`,
			expected: " ?&=#+%!<>#\"{}|\\^[]`☺\t:/@$'()*,;",
			key:      "url",
		},
		{
			data:     `{"url":"+%3F%26%3D%23%2B%25%21%3C%3E%23%22%7B%7D%7C%5C%5E%5B%5D%60%E2%98%BA%09%3A%2F%40%24%27%28%29%2A%2C%3B","second":2}`,
			script:   `json(_, url) url_decode("url", "aaa")`,
			expected: "+%3F%26%3D%23%2B%25%21%3C%3E%23%22%7B%7D%7C%5C%5E%5B%5D%60%E2%98%BA%09%3A%2F%40%24%27%28%29%2A%2C%3B",
			key:      "url",
		},
		{
			data:     `{"aa":{"url":"http%3a%2f%2fwww.baidu.com%2fs%3fwd%3d%e6%b5%8b%e8%af%95"},"second":2}`,
			script:   `json(_, aa.url) url_decode(aa.url)`,
			expected: `http://www.baidu.com/s?wd=测试`,
			key:      "aa.url",
		},
		{
			data:     `{"aa":{"aa.url":"http%3a%2f%2fwww.baidu.com%2fs%3fwd%3d%e6%b5%8b%e8%af%95"},"second":2}`,
			script:   "json(_, aa.`aa.url`) url_decode(aa.`aa.url`)",
			expected: `http://www.baidu.com/s?wd=测试`,
			key:      "aa.aa.url",
		},
	}

	for _, tt := range testCase {
		p, err := NewPipeline(tt.script)

		assertEqual(t, err, nil)

		p.Run(tt.data)

		r, err := p.getContentStr(tt.key)
		assertEqual(t, err, nil)

		assertEqual(t, r, tt.expected)
	}
}

func TestUserAgentFunc(t *testing.T) {
	testCase := []*funcCase{
		{
			data:     `{"userAgent":"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/36.0.1985.125 Safari/537.36", "second":2,"third":"abc","forth":true}`,
			script:   `json(_, userAgent) user_agent(userAgent)`,
			expected: "Windows 7",
			key:      "os",
		},
		{
			data:     `{"userAgent":"Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"}`,
			script:   `json(_, userAgent) user_agent(userAgent)`,
			expected: "Googlebot",
			key:      "browser",
		},
		{
			data:     `{"userAgent":"Mozilla/5.0 (iPhone; CPU iPhone OS 6_0 like Mac OS X) AppleWebKit/536.26 (KHTML, like Gecko) Version/6.0 Mobile/10A5376e Safari/8536.25 (compatible; Googlebot/2.1; +http://www.google.com/bot.html"}`,
			script:   `json(_, userAgent) user_agent(userAgent)`,
			expected: "",
			key:      "engine",
		},
		{
			data:     `{"userAgent":"Mozilla/5.0 (compatible; bingbot/2.0; +http://www.bing.com/bingbot.htm)"}`,
			script:   `json(_, userAgent) user_agent(userAgent)`,
			expected: "bingbot",
			key:      "browser",
		},
	}

	for _, tt := range testCase {
		p, err := NewPipeline(tt.script)
		assertEqual(t, err, p.lastErr)

		p.Run(tt.data)

		r, err := p.getContentStr(tt.key)
		assertEqual(t, err, nil)

		assertEqual(t, r, tt.expected)
	}
}

func TestDatetimeFunc(t *testing.T) {
	testCase := []*funcCase{
		{
			data:     `{"a":{"timestamp": "1610960605", "second":2},"age":47}`,
			script:   `json(_, a.timestamp) datetime(a.timestamp, 's', 'RFC3339')`,
			expected: "2021-01-18T17:03:25+08:00",
			key:      "a.timestamp",
		},
		{
			data:     `{"a":{"timestamp": "1610960605000", "second":2},"age":47}`,
			script:   `json(_, a.timestamp) datetime(a.timestamp, 'ms', 'RFC3339')`,
			expected: "2021-01-18T17:03:25+08:00",
			key:      "a.timestamp",
		},
	}

	for _, tt := range testCase {
		p, err := NewPipeline(tt.script)
		assertEqual(t, err, p.lastErr)

		p.Run(tt.data)

		r, err := p.getContentStr(tt.key)
		assertEqual(t, err, nil)

		assertEqual(t, r, tt.expected)
	}
}

func TestGroupFunc(t *testing.T) {
	testCase := []*funcCase{
		{
			data:     `{"status": 200,"age":47}`,
			script:   `json(_, status) group_between(status, [200, 400], false, newkey)`,
			expected: false,
			key:      "newkey",
		},
		{
			data:     `{"status": 200,"age":47}`,
			script:   `json(_, status) group_between(status, [200, 400], 10, newkey)`,
			expected: int64(10),
			key:      "newkey",
		},
		{
			data:     `{"status": 200,"age":47}`,
			script:   `json(_, status) group_between(status, [, 400], "ok", newkey)`,
			expected: nil,
			key:      "newkey",
		},
		{
			data:     `{"status": 200,"age":47}`,
			script:   `json(_, status) group_between(status, [200, 400], "ok", newkey)`,
			expected: "ok",
			key:      "newkey",
		},
		{
			data:     `{"status": 200,"age":47}`,
			script:   `json(_, status) group_between(status, [200, 299], "ok")`,
			expected: "ok",
			key:      "status",
		},
		{
			data:     `{"status": 200,"age":47}`,
			script:   `json(_, status) group_between(status, [299, 200], "ok")`,
			expected: float64(200),
			key:      "status",
		},
		{
			data:     `{"status": 200,"age":47}`,
			script:   `json(_, status) group_between(status, [299, 200], "ok", newkey)`,
			expected: float64(200),
			key:      "status",
		},
		{
			data:     `{"status": 200,"age":47}`,
			script:   `json(_, status) group_between(status, [200, 299], "ok", newkey)`,
			expected: "ok",
			key:      "newkey",
		},
		{
			data:     `{"status": 200,"age":47}`,
			script:   `json(_, status) group_between(status, [300, 400], "ok", newkey)`,
			expected: nil,
			key:      "newkey",
		},
	}

	for _, tt := range testCase {
		p, err := NewPipeline(tt.script)
		assertEqual(t, err, p.lastErr)

		p.Run(tt.data)

		r, err := p.getContent(tt.key)
		assertEqual(t, err, nil)

		assertEqual(t, r, tt.expected)
	}
}

func TestGroupInFunc(t *testing.T) {
	testCase := []*funcCase{
		{
			data:     `{"status": true,"age":"47"}`,
			script:   `json(_, status) group_in(status, [true], "ok", "newkey")`,
			expected: "ok",
			key:      "newkey",
		},
		{
			data:     `{"status": true,"age":"47"}`,
			script:   `json(_, status) group_in(status, [true], "ok", "newkey")`,
			expected: "ok",
			key:      "newkey",
		},
		{
			data:     `{"status": "aa","age":"47"}`,
			script:   `json(_, status) group_in(status, [], "ok")`,
			expected: "aa",
			key:      "status",
		},
		{
			data:     `{"status": "aa","age":"47"}`,
			script:   `json(_, status) group_in(status, ["aa"], "ok", "newkey")`,
			expected: "ok",
			key:      "newkey",
		},
		{
			data:     `{"status": "test","age":"47"}`,
			script:   `json(_, status) group_in(status, [200, 47, "test"], "ok", newkey)`,
			expected: "ok",
			key:      "newkey",
		},
		{
			data:     `{"status": "test","age":"47"}`,
			script:   `json(_, status) group_in(status, [200, "test"], "ok")`,
			expected: "ok",
			key:      "status",
		},
	}

	for _, tt := range testCase {
		p, err := NewPipeline(tt.script)
		assertEqual(t, err, p.lastErr)

		p.Run(tt.data)

		r, err := p.getContentStr(tt.key)
		assertEqual(t, err, nil)

		assertEqual(t, r, tt.expected)
	}
}

func TestNullIfFunc(t *testing.T) {
	testCase := []*funcCase{
		{
			data:     `{"a":{"first": 1,"second":2,"third":"aBC","forth":true},"age":47}`,
			script:   `json(_, a.first) nullif(a.first, "1")`,
			expected: float64(1),
			key:      "a.first",
		},
		{
			data:     `{"a":{"first": "1","second":2,"third":"aBC","forth":true},"age":47}`,
			script:   `json(_, a.first) nullif(a.first, 1)`,
			expected: "1",
			key:      "a.first",
		},
		{
			data:     `{"a":{"first": "","second":2,"third":"aBC","forth":true},"age":47}`,
			script:   `json(_, a.first) nullif(a.first, "")`,
			expected: nil,
			key:      "a.first",
		},
		{
			data:     `{"a":{"first": null,"second":2,"third":"aBC","forth":true},"age":47}`,
			script:   `json(_, a.first) nullif(a.first, nil)`,
			expected: nil,
			key:      "a.first",
		},
		{
			data:     `{"a":{"first": true,"second":2,"third":"aBC","forth":true},"age":47}`,
			script:   `json(_, a.first) nullif(a.first, true)`,
			expected: nil,
			key:      "a.first",
		},
		{
			data:     `{"a":{"first": 2.3, "second":2,"third":"aBC","forth":true},"age":47}`,
			script:   `json(_, a.first) nullif(a.first, 2.3)`,
			expected: nil,
			key:      "a.first",
		},
		{
			data:     `{"a":{"first": 2,"second":2,"third":"aBC","forth":true},"age":47}`,
			script:   `json(_, a.first) nullif(a.first, 2, "newkey")`,
			expected: nil,
			key:      "newkey",
		},
		{
			data:     `{"a":{"first":"2.3","second":2,"third":"aBC","forth":true},"age":47}`,
			script:   `json(_, a.first) nullif(a.first, "2.3", "newkey")`,
			expected: nil,
			key:      "newkey",
		},
	}

	for _, tt := range testCase {
		p, err := NewPipeline(tt.script)
		assertEqual(t, err, p.lastErr)

		p.Run(tt.data)

		r, err := p.getContent(tt.key)
		assertEqual(t, err, nil)

		assertEqual(t, r, tt.expected)
	}
}

func TestParseDuration(t *testing.T) {
	cases := []*funcCase{
		{
			data:     `{"str": "1s"}`,
			script:   `json(_, str) parse_duration(str)`,
			expected: int64(time.Second),
			key:      "str",
		},

		{
			data:     `{"str": "1ms"}`,
			script:   `json(_, str) parse_duration(str)`,
			expected: int64(time.Millisecond),
			key:      "str",
		},

		{
			data:     `{"str": "1us"}`,
			script:   `json(_, str) parse_duration(str)`,
			expected: int64(time.Microsecond),
			key:      "str",
		},

		{
			data:     `{"str": "1µs"}`,
			script:   `json(_, str) parse_duration(str)`,
			expected: int64(time.Microsecond),
			key:      "str",
		},

		{
			data:     `{"str": "1m"}`,
			script:   `json(_, str) parse_duration(str)`,
			expected: int64(time.Minute),
			key:      "str",
		},

		{
			data:     `{"str": "1h"}`,
			script:   `json(_, str) parse_duration(str)`,
			expected: int64(time.Hour),
			key:      "str",
		},

		{
			data:     `{"str": "-23h"}`,
			script:   `json(_, str) parse_duration(str)`,
			expected: -23 * int64(time.Hour),
			key:      "str",
		},

		{
			data:     `{"str": "-23ns"}`,
			script:   `json(_, str) parse_duration(str)`,
			expected: int64(-23),
			key:      "str",
		},

		{
			data:     `{"str": "-2.3s"}`,
			script:   `json(_, str) parse_duration(str)`,
			expected: int64(time.Second*-2 - 300*time.Millisecond),
			key:      "str",
		},

		{
			data:   `{"str": "1uuus"}`,
			script: `json(_, str) parse_duration(str)`,
			key:    "str",
			fail:   true,
		},
	}

	for _, tt := range cases {
		p, err := NewPipeline(tt.script)
		assertEqual(t, err, p.lastErr)

		p.Run(tt.data)

		r, err := p.getContent(tt.key)
		assertEqual(t, err, nil)

		if !tt.fail {
			assertEqual(t, r, tt.expected)
		}
	}
}

func TestParseDate(t *testing.T) {
	cases := []*funcCase{
		{
			data:     `{}`,
			script:   `parse_date(aa, "2021", "May", "12", "10", "10", "34", "", "Asia/Shanghai")`,
			expected: int64(1620785434000000000),
			key:      "aa",
		},
		{
			data:     `{}`,
			script:   `parse_date(aa, "2021", "12", "12", "10", "10", "34", "", "Asia/Shanghai")`,
			expected: int64(1639275034000000000),
			key:      "aa",
		},
		{
			data:     `{}`,
			script:   `parse_date(aa, "2021", "12", "12", "10", "10", "34", "100", "Asia/Shanghai")`,
			expected: int64(1639275034000000100),
			key:      "aa",
		},
		{
			data:     `{}`,
			script:   `parse_date(aa, "20", "February", "12", "10", "10", "34", "", "+8")`,
			expected: int64(1581473434000000000),
			key:      "aa",
		},
	}

	for _, tt := range cases {
		p, err := NewPipeline(tt.script)
		assertEqual(t, err, p.lastErr)

		p.Run(tt.data)
		r, err := p.getContent(tt.key)
		assertEqual(t, err, nil)

		if !tt.fail {
			assertEqual(t, r, tt.expected)
		}
	}
}

func TestJsonAllFunc(t *testing.T) {
	testCase := []*funcCase{
		{
			data: `{
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
		},
	}

	for _, tt := range testCase {
		p, err := NewPipeline(tt.script)
		assertEqual(t, err, p.lastErr)

		p.Run(tt.data)

		r, err := p.getContentStr(tt.key)
		assertEqual(t, err, nil)

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
	assertEqual(t, err, nil)

	assertEqual(t, r, "Sara")
}

func TestDz(t *testing.T) {
	cases := []*funcCase{
		{
			data:     `{"str": "13838130517"}`,
			script:   `json(_, str) cover(str, [8, 13])`,
			expected: "1383813****",
			key:      "str",
		},
		{
			data:     `{"str": "13838130517"}`,
			script:   `json(_, str) cover(str, [8, 11])`,
			expected: "1383813****",
			key:      "str",
		},
		{
			data:     `{"str": "13838130517"}`,
			script:   `json(_, str) cover(str, [2, 4])`,
			expected: "1***8130517",
			key:      "str",
		},
		{
			data:     `{"str": "13838130517"}`,
			script:   `json(_, str) cover(str, [0, 3])`,
			expected: "***38130517",
			key:      "str",
		},
		{
			data:     `{"str": "13838130517"}`,
			script:   `json(_, str) cover(str, [1, 1])`,
			expected: "*3838130517",
			key:      "str",
		},
		{
			data:     `{"str": "刘少波"}`,
			script:   `json(_, str) cover(str, [2, 2])`,
			expected: "刘＊波",
			key:      "str",
		},
	}

	for _, tt := range cases {
		p, err := NewPipeline(tt.script)
		assertEqual(t, err, p.lastErr)

		p.Run(tt.data)
		r, err := p.getContentStr(tt.key)
		assertEqual(t, err, nil)

		if !tt.fail {
			assertEqual(t, r, tt.expected)
		}
	}
}

func TestReplace(t *testing.T) {
	cases := []*funcCase{
		{
			data:     `{"str": "13789123014"}`,
			script:   `json(_, str) replace(str, "(1[0-9]{2})[0-9]{4}([0-9]{4})", "$1****$2")`,
			expected: "137****3014",
			key:      "str",
		},
		{
			data:     `{"str": "zhang san"}`,
			script:   `json(_, str) replace(str, "([a-z]*) \\w*", "$1 ***")`,
			expected: "zhang ***",
			key:      "str",
		},
		{
			data:     `{"str": "362201200005302565"}`,
			script:   `json(_, str) replace(str, "([1-9]{4})[0-9]{10}([0-9]{4})", "$1**********$2")`,
			expected: "3622**********2565",
			key:      "str",
		},
		{
			data:     `{"str": "小阿卡"}`,
			script:   `json(_, str) replace(str, '([\u4e00-\u9fa5])[\u4e00-\u9fa5]([\u4e00-\u9fa5])', "$1＊$2")`,
			expected: "小＊卡",
			key:      "str",
		},
	}

	for _, tt := range cases {
		p, err := NewPipeline(tt.script)
		assertEqual(t, err, p.lastErr)

		p.Run(tt.data)

		r, err := p.getContentStr(tt.key)
		assertEqual(t, err, nil)

		if !tt.fail {
			assertEqual(t, r, tt.expected)
		}
	}
}
