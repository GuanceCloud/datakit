package pipeline

import (
	"testing"
	"time"

	"github.com/araddon/dateparse"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
)

//  时间解析补充函数
func BenchmarkParseDatePattern(b *testing.B) {
	for n := 0; n < b.N; n++ {
		if _, err := parseDatePattern("Tue May 18 06:25:05.176170 2021"); err != nil {
			b.Error(err)
		}
	}
}

// dataparse库
func BenchmarkDateparseParseIn(b *testing.B) {
	for n := 0; n < b.N; n++ {
		if _, err := dateparse.ParseIn("2017-12-29T12:33:33.095243Z", nil); err != nil {
			b.Error(err)
		}
	}
}

// dataparse库 指定时区
func BenchmarkDateparseParseInTZ(b *testing.B) {
	tz, _ := time.LoadLocation("Asia/Shanghai")
	for n := 0; n < b.N; n++ {
		if _, err := dateparse.ParseIn("2017-12-29T12:33:33.095243Z", tz); err != nil {
			b.Error(err)
		}
	}
}

// default_time， pipeline 时间解析函数
func BenchmarkTimeDefault(b *testing.B) {
	for n := 0; n < b.N; n++ {
		p := &Pipeline{
			Output: map[string]interface{}{
				"time": "2017-12-29T12:33:33.095243Z",
			},
			ast: &parser.Ast{
				Functions: []*parser.FuncExpr{
					{
						Name: "default_time",
						Param: []parser.Node{
							&parser.Identifier{Name: "time"},
							&parser.StringLiteral{Val: ""},
						},
					},
				},
			},
		}
		if _, err := DefaultTime(p, p.ast.Functions[0]); err != nil {
			b.Error(err)
		}
	}
}

// // default_time， pipeline 时间解析函数, 设置时区
func BenchmarkTimeDefaultTZ(b *testing.B) {
	for n := 0; n < b.N; n++ {
		p := &Pipeline{
			Output: map[string]interface{}{
				"time": "2017-12-29T12:33:33.095243Z",
			},
			ast: &parser.Ast{
				Functions: []*parser.FuncExpr{
					{
						Name: "default_time",
						Param: []parser.Node{
							&parser.Identifier{Name: "time"},
							&parser.StringLiteral{Val: "Asia/Shanghai"},
						},
					},
				},
			},
		}
		if _, err := DefaultTime(p, p.ast.Functions[0]); err != nil {
			b.Error(err)
		}
	}
}

// add_pattern 函数， pipeline 添加模式
func BenchmarkAddPattern(b *testing.B) {
	for n := 0; n < b.N; n++ {
		p := &Pipeline{
			Output: map[string]interface{}{
				"time": "2017-12-29T12:33:33.095243Z",
			},
			ast: &parser.Ast{
				Functions: []*parser.FuncExpr{
					{
						Name: "add_pattern",
						Param: []parser.Node{
							&parser.StringLiteral{Val: "time1"},
							&parser.StringLiteral{Val: "[\\w:\\.\\+-]+?"},
						},
					},
				},
			},
		}
		if _, err := AddPattern(p, p.ast.Functions[0]); err != nil {
			b.Error(err)
		}
	}
}

// default_time_with_fmt， pipeline 根据指定的时间 fmt 解析时间
func BenchmarkTimeDefaultWithTfmt(b *testing.B) {
	for n := 0; n < b.N; n++ {
		p := &Pipeline{
			Output: map[string]interface{}{
				"time": "2017-12-29T12:33:33.095243Z", // "2017-12-29T12:33:33.095243Z+0800"
			},
			ast: &parser.Ast{
				Functions: []*parser.FuncExpr{
					{
						Name: "default_time_with_fmt",
						Param: []parser.Node{
							&parser.Identifier{Name: "time"},
							&parser.StringLiteral{Val: "2006-01-02T15:04:05.000000Z"}, // 2006-01-02T15:04:05.000000Z-0700
							&parser.StringLiteral{Val: ""},                            // "Asia/Shanghai"
						},
					},
				},
			},
		}
		if _, err := DefaultTimeWithFmt(p, p.ast.Functions[0]); err != nil {
			b.Error(err)
		}
	}
}

// default_time
func BenchmarkParseLog(b *testing.B) {
	script := `
	add_pattern("date1", "[\\w:\\.\\+-]+?")
	add_pattern("date2", "[\\w:\\.\\+-]+?")
	add_pattern("date3", "[\\w:\\.\\+-]+?")
	add_pattern("date4", "[\\w:\\.\\+-]+?")
	grok(_, "%{TIMESTAMP_ISO8601:time}\\s+%{INT:thread_id}\\s+%{WORD:operation}\\s+%{GREEDYDATA:raw_query}")
	cast(thread_id, "int")
	default_time(time)
		`
	pl, err := NewPipeline(script)
	if err != nil {
		b.Error(err)
	}
	data := `2017-12-29T12:33:33.095243Z         2 Query     SELECT TABLE_SCHEMA, TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE CREATE_OPTIONS LIKE '%partitioned%'`

	for n := 0; n < b.N; n++ {
		pl.Run(data)
	}
}

func BenchmarkParseLog_tz(b *testing.B) {
	script := `
	add_pattern("date1", "(\\d+/\\w+/[\\d:]+ [\\d+-]+)")
	add_pattern("date2", "[\\w:\\.\\+-]+?")
	add_pattern("date3", "[\\w:\\.\\+-]+?")
	add_pattern("date4", "[\\w:\\.\\+-]+?")
	grok(_, "%{TIMESTAMP_ISO8601:time}\\s+%{IP:ip}\\s+%{INT:thread_id}\\s+")

	grok(_, "%{TIMESTAMP_ISO8601:time}\\s+%{IP:ip}\\s+%{INT:thread_id}\\s+")
	grok(_, "%{TIMESTAMP_ISO8601:time}\\s+%{IP:ip}\\s+%{INT:thread_id}\\s+")
	grok(_, "%{TIMESTAMP_ISO8601:time}\\s+%{IP:ip}\\s+%{INT:thread_id}\\s+")


	cast(thread_id, "int")
	default_time(time, "Asia/Shanghai")
		`
	pl, err := NewPipeline(script)
	if err != nil {
		b.Error(err)
	}
	data := `2017-12-29T12:33:33.095243Z     1.1.1.1    2 `

	for n := 0; n < b.N; n++ {
		pl.Run(data)
	}
}

func BenchmarkGrok(b *testing.B) {
	script := `
#grok(_, "%{IPV6:client_ip}")
#grok(_, "%{IPV6:client_ip} %{NOTSPACE:http_ident} %{NOTSPACE:http_auth} \\[%{HTTPDATE:time}\\] \"%{DATA:http_method} %{URIPATHPARAM:http_url} HTTP/%{NUMBER:http_version}\" %{INT:status_code} %{INT:bytes}")
grok(_, "%{IPORHOST:client_ip} %{NOTSPACE:http_ident} %{NOTSPACE:http_auth} \\[%{HTTPDATE:time}\\] \"%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\" %{INT:status_code} %{INT:bytes}")
`
	pl, err := NewPipeline(script)
	if err != nil {
		b.Error(err)
	}
	data := `127.0.0.1 - - [21/Jul/2021:14:14:38 +0800] "GET /?1 HTTP/1.1" 200 2178 "-" "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.164 Safari/537.36"`
	// data := `fe80:d::127.0.0.1 - - [21/Jul/2021:14:14:38 +0800] "GET /?1 HTTP/1.1" 200 2178 "-" "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.164 Safari/537.36"`

	for n := 0; n < b.N; n++ {
		pl.Run(data)
		// b.Error(pl.Output)
	}
}

func BenchmarkParseLogNginx(b *testing.B) {
	script := `
add_pattern("date2", "%{YEAR}[./]%{MONTHNUM}[./]%{MONTHDAY} %{TIME}")

# access log
grok(_, "%{IPORHOST:client_ip} %{NOTSPACE:http_ident} %{NOTSPACE:http_auth} \\[%{HTTPDATE:time}\\] \"%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\" %{INT:status_code} %{INT:bytes}")


add_pattern("access_common", "%{IPORHOST:client_ip} %{NOTSPACE:http_ident} %{NOTSPACE:http_auth} \\[%{HTTPDATE:time}\\] \"%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\" %{INT:status_code} %{INT:bytes}")
grok(_, '%{access_common} "%{NOTSPACE:referrer}" "%{GREEDYDATA:agent}')
user_agent(agent)

grok(_, "%{date2:time} \\[%{LOGLEVEL:status}\\] %{GREEDYDATA:msg}, client: %{IPORHOST:client_ip}, server: %{IPORHOST:server}, request: \"%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\", (upstream: \"%{GREEDYDATA:upstream}\", )?host: \"%{IPORHOST:ip_or_host}\"")
grok(_, "%{date2:time} \\[%{LOGLEVEL:status}\\] %{GREEDYDATA:msg}, client: %{IPORHOST:client_ip}, server: %{IPORHOST:server}, request: \"%{GREEDYDATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\", host: \"%{IPORHOST:ip_or_host}\"")
grok(_,"%{date2:time} \\[%{LOGLEVEL:status}\\] %{GREEDYDATA:msg}")

group_in(status, ["warn", "notice"], "warning")
group_in(status, ["error", "crit", "alert", "emerg"], "error")

cast(status_code, "int")
cast(bytes, "int")

group_between(status_code, [200,299], "OK", status)
group_between(status_code, [300,399], "notice", status)
group_between(status_code, [400,499], "warning", status)
group_between(status_code, [500,599], "error", status)


nullif(http_ident, "-")
nullif(http_auth, "-")
nullif(upstream, "")
default_time(time)
`
	pl, err := NewPipeline(script)
	if err != nil {
		b.Error(err)
	}
	data := `127.0.0.1 - - [21/Jul/2021:14:14:38 +0800] "GET /?1 HTTP/1.1" 200 2178 "-" "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.164 Safari/537.36"`
	for n := 0; n < b.N; n++ {
		pl.Run(data)
	}
}

// default_time_with_fmt
func BenchmarkParseLogWithTfmt(b *testing.B) {
	script := `
	add_pattern("date1", "[\\w:\\.\\+-]+?")
	add_pattern("date2", "[\\w:\\.\\+-]+?")
	add_pattern("date3", "[\\w:\\.\\+-]+?")
	add_pattern("date4", "[\\w:\\.\\+-]+?")	grok(_, "%{TIMESTAMP_ISO8601:time}\\s+%{INT:thread_id}\\s+%{WORD:operation}\\s+%{GREEDYDATA:raw_query}")
	cast(thread_id, "int")
	default_time_with_fmt(time, "2006-01-02T15:04:05.000000Z")
	`
	// "2006-01-02T15:04:05.000000Z"
	pl, err := NewPipeline(script)
	if err != nil {
		b.Error(err)
	}
	data := `2021-07-20T12:33:33.095243Z         2 Query     SELECT TABLE_SCHEMA, TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE CREATE_OPTIONS LIKE '%partitioned%'`

	for n := 0; n < b.N; n++ {
		pl.Run(data)
	}
}

// default_time_with_fmt， timezone
func BenchmarkParseLogWithTfmt_tz(b *testing.B) {
	script := `
	add_pattern("date1", "[\\w:\\.\\+-]+?")
	add_pattern("date2", "[\\w:\\.\\+-]+?")
	add_pattern("date3", "[\\w:\\.\\+-]+?")
	add_pattern("date4", "[\\w:\\.\\+-]+?")
	grok(_, "%{TIMESTAMP_ISO8601:time}\\s+%{INT:thread_id}\\s+%{WORD:operation}\\s+%{GREEDYDATA:raw_query}")
	cast(thread_id, "int")
	default_time_with_fmt(time, "2006-01-02T15:04:05.000000Z", "Asia/Shanghai")
	`
	// "2006-01-02T15:04:05.000000Z"
	pl, err := NewPipeline(script)
	if err != nil {
		b.Error(err)
	}
	data := `2021-07-20T12:33:33.095243Z         2 Query     SELECT TABLE_SCHEMA, TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE CREATE_OPTIONS LIKE '%partitioned%'`

	for n := 0; n < b.N; n++ {
		pl.Run(data)
	}
}

func BenchmarkParseLogWithTfmt_NoAddPattern(b *testing.B) {
	script := `
	grok(_, "%{TIMESTAMP_ISO8601:time}\\s+%{INT:thread_id}\\s+%{WORD:operation}\\s+%{GREEDYDATA:raw_query}")
	cast(thread_id, "int")
	default_time_with_fmt(time, "2006-01-02T15:04:05.000000Z", "Asia/Shanghai")
	`
	// "2006-01-02T15:04:05.000000Z"
	pl, err := NewPipeline(script)
	if err != nil {
		b.Error(err)
	}
	data := `2021-07-20T12:33:33.095243Z         2 Query     SELECT TABLE_SCHEMA, TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE CREATE_OPTIONS LIKE '%partitioned%'`

	for n := 0; n < b.N; n++ {
		pl.Run(data)
	}
}
