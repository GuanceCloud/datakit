package parser

import (
	"fmt"
	"strconv"
	"testing"
)

var MaxLimitStr = strconv.Itoa(MaxLimit)

// 类似influxdb的查询结果
var testCases = []map[string]string{

	{
		// 1, 测试逻辑not
		"index":    "1",
		"input":    "L::c1:(age){ age != 20 }",
		"expected": `{"aggs":{},"_source":["date","age"],"query":{"bool":{"must":[{"bool":{"should":[{"term":{"source":{"value":"c1"}}}]}},{"bool":{"must_not":[{"term":{"age":{"value":"20"}}}]}}]}},"size":1000,"sort":[{"date":{"missing":"_first","order":"desc","unmapped_type":"string"}}],"track_total_hits":true}`,
	},
	{
		// 2, 测试正则匹配
		"index":    "2",
		"input":    "L::c1:(age, name){ age > 19 and name=re(`go.*`)  }",
		"expected": `{"aggs":{},"_source":["date","age","name"],"query":{"bool":{"must":[{"bool":{"should":[{"term":{"source":{"value":"c1"}}}]}},{"bool":{"must":[{"range":{"age":{"gt":"19"}}},{"regexp":{"name":{"value":"go.*"}}}]}}]}},"size":1000,"sort":[{"date":{"missing":"_first","order":"desc","unmapped_type":"string"}}],"track_total_hits":true}`,
	},
	{
		// 3, 测试逻辑组合
		"index":    "3",
		"input":    "L::c1:(age, height){ age != 20 or height > 150 }",
		"expected": `{"aggs":{},"_source":["date","age","height"],"query":{"bool":{"must":[{"bool":{"should":[{"term":{"source":{"value":"c1"}}}]}},{"bool":{"should":[{"bool":{"must_not":[{"term":{"age":{"value":"20"}}}]}},{"range":{"height":{"gt":"150"}}}]}}]}},"size":1000,"sort":[{"date":{"missing":"_first","order":"desc","unmapped_type":"string"}}],"track_total_hits":true}`,
	},
	{
		// 4, 测试基本聚合
		"index":    "4",
		"input":    "L::c1:(age, height) by age, height limit 3",
		"expected": `{"aggs":{"age":{"aggs":{"height":{"aggs":{"top_hits":{"top_hits":{"_source":["date","age","height"],"size":1,"sort":[{"date":{"missing":"_first","order":"desc","unmapped_type":"string"}}]}}},"terms":{"field":"height","size":10}}},"terms":{"field":"age","size":3}}},"_source":["date","age","height"],"query":{"bool":{"must":[{"bool":{"should":[{"term":{"source":{"value":"c1"}}}]}}]}},"size":0}`,
	},
	{
		// 5, 测试时间聚合
		"index":    "5",
		"input":    "L::c1 [2020-11-11 12:10:30:2020-11-11 12:25:30:1m]",
		"expected": `{"aggs":{"time":{"date_histogram":{"field":"date","interval":"60000ms"}}},"query":{"bool":{"must":[{"bool":{"should":[{"term":{"` + LCLASS + `":{"value":"c1"}}}]}},{"range":{"date":{"gte":"1605096630000"}}},{"range":{"date":{"lte":"1605097530000"}}}]}},"size":0}`,
	},
	{
		// 6, 指标聚合
		"index":    "6",
		"input":    "L::c1:(max(age))",
		"expected": `{"aggs":{"max_age":{"max":{"field":"age"}}},"query":{"bool":{"must":[{"bool":{"should":[{"term":{"source":{"value":"c1"}}}]}}]}},"size":0}`,
	},
	{
		// 7, match查询
		"index":    "7",
		"input":    "L::c1:(age, height) {name=match(`golang`)}",
		"expected": `{"aggs":{},"_source":["date","age","height"],"query":{"bool":{"must":[{"bool":{"should":[{"term":{"source":{"value":"c1"}}}]}},{"match":{"name":{"query":"golang"}}}]}},"size":1000,"sort":[{"date":{"missing":"_first","order":"desc","unmapped_type":"string"}}],"track_total_hits":true}`,
	},
	{
		// 8, class分类正则匹配
		"index":    "8",
		"input":    "L::re(`c.*`):(age){ age > 20 }",
		"expected": `{"aggs":{},"_source":["date","age"],"query":{"bool":{"must":[{"bool":{"should":[{"regexp":{"source":{"value":"c.*"}}}]}},{"range":{"age":{"gt":"20"}}}]}},"size":1000,"sort":[{"date":{"missing":"_first","order":"desc","unmapped_type":"string"}}],"track_total_hits":true}`,
	},
	{
		// 9, class多个分类
		"index":    "9",
		"input":    "L::re(`c.*`), t1:(age){ age > 20 }",
		"expected": `{"aggs":{},"_source":["date","age"],"query":{"bool":{"must":[{"bool":{"should":[{"term":{"source":{"value":"t1"}}},{"regexp":{"source":{"value":"c.*"}}}]}},{"range":{"age":{"gt":"20"}}}]}},"size":1000,"sort":[{"date":{"missing":"_first","order":"desc","unmapped_type":"string"}}],"track_total_hits":true}`,
	},
	{
		// 10, 测试逻辑组合, 有（）情况
		"index":    "10",
		"input":    "L::c1:(age, height){ (age < 20 or age > 10) and (height > 150 or height < 180) }",
		"expected": `{"aggs":{},"_source":["date","age","height"],"query":{"bool":{"must":[{"bool":{"should":[{"term":{"source":{"value":"c1"}}}]}},{"bool":{"must":[{"bool":{"should":[{"range":{"age":{"lt":"20"}}},{"range":{"age":{"gt":"10"}}}]}},{"bool":{"should":[{"range":{"height":{"gt":"150"}}},{"range":{"height":{"lt":"180"}}}]}}]}}]}},"size":1000,"sort":[{"date":{"missing":"_first","order":"desc","unmapped_type":"string"}}],"track_total_hits":true}`,
	},
	// 11, tophits，删除tophits函数
	{
		// 12, query order by
		"index":    "12",
		"input":    "L::c1:(age, height) order by age, height desc",
		"expected": `{"aggs":{},"_source":["date","age","height"],"query":{"bool":{"must":[{"bool":{"should":[{"term":{"source":{"value":"c1"}}}]}}]}},"size":1000,"sort":[{"age":{"missing":"_first","order":"asc","unmapped_type":"string"}},{"height":{"missing":"_last","order":"desc","unmapped_type":"string"}}],"track_total_hits":true}`,
	},
	{
		// 13, aggs order by
		"index":    "13",
		"input":    "L::c1:(age, height) by age order by age desc limit 3",
		"expected": `{"aggs":{"age":{"aggs":{"top_hits":{"top_hits":{"_source":["date","age","height"],"size":1,"sort":[{"date":{"missing":"_first","order":"desc","unmapped_type":"string"}}]}}},"terms":{"field":"age","order":{"_key":"desc"},"size":3}}},"_source":["date","age","height"],"query":{"bool":{"must":[{"bool":{"should":[{"term":{"source":{"value":"c1"}}}]}}]}},"size":0}`,
	},
	{
		// 14, 日志类型查询
		"index":    "14",
		"input":    "L::c1:(age, height) {age > 20}",
		"expected": `{"aggs":{},"_source":["date","age","height"],"query":{"bool":{"must":[{"bool":{"should":[{"term":{"` + LCLASS + `":{"value":"c1"}}}]}},{"range":{"age":{"gt":"20"}}}]}},"size":1000,"sort":[{"date":{"missing":"_first","order":"desc","unmapped_type":"string"}}],"track_total_hits":true}`,
	},
	{
		// 15, 日志时间聚合
		"index":    "15",
		"input":    "L::c1 [2020-11-11 12:10:30:2020-11-11 12:25:30:1m]",
		"expected": `{"aggs":{"time":{"date_histogram":{"field":"date","interval":"60000ms"}}},"query":{"bool":{"must":[{"bool":{"should":[{"term":{"` + LCLASS + `":{"value":"c1"}}}]}},{"range":{"date":{"gte":"1605096630000"}}},{"range":{"date":{"lte":"1605097530000"}}}]}},"size":0}`,
	},
	{
		// 16, 指标聚合
		"index":    "16",
		"input":    "L::c1:(max(age) as maxAge)",
		"expected": `{"aggs":{"maxAge":{"max":{"field":"age"}}},"query":{"bool":{"must":[{"bool":{"should":[{"term":{"source":{"value":"c1"}}}]}}]}},"size":0}`,
	},
	{
		// 17, show object
		"index":    "17",
		"input":    "show_object_class()",
		"expected": `{"aggs":{"aggs1":{"terms":{"field":"class","size":"` + string(MaxLimitStr) + `"}}},"size":0}`,
	},
	{
		// 18, show logging
		"index":    "18",
		"input":    "show_logging_source()",
		"expected": `{"aggs":{"aggs1":{"terms":{"field":"` + LCLASS + `","size":"` + string(MaxLimitStr) + `"}}},"size":0}`,
	},
	{
		// 19, show event
		"index":    "19",
		"input":    "show_event_source()",
		"expected": `{"aggs":{"aggs1":{"terms":{"field":"` + ECLASS + `","size":"` + string(MaxLimitStr) + `"}}},"size":0}`,
	},
	{
		// 20, show tracing
		"index":    "20",
		"input":    "show_tracing_service()",
		"expected": `{"aggs":{"aggs1":{"terms":{"field":"` + TCLASS + `","size":"` + string(MaxLimitStr) + `"}}},"size":0}`,
	},
	{
		// 21, event类型查询,带有汉字
		"index":    "21",
		"input":    "event::`测试`:(age, height) {age > 20}",
		"expected": `{"aggs":{},"_source":["date","age","height"],"query":{"bool":{"must":[{"bool":{"should":[{"term":{"source":{"value":"测试"}}}]}},{"range":{"age":{"gt":"20"}}}]}},"size":1000,"sort":[{"date":{"missing":"_first","order":"desc","unmapped_type":"string"}}],"track_total_hits":true}`,
	},
	{
		// 22, 测试limit
		"index":    "22",
		"input":    "L::c1:(age){ age > 20 } limit 2",
		"expected": `{"aggs":{},"_source":["date","age"],"query":{"bool":{"must":[{"bool":{"should":[{"term":{"source":{"value":"c1"}}}]}},{"range":{"age":{"gt":"20"}}}]}},"size":2,"sort":[{"date":{"missing":"_first","order":"desc","unmapped_type":"string"}}],"track_total_hits":true}`,
	},
	{
		// 23, 查询返回所有字段
		"index":    "23",
		"input":    "L::c1:(`*`){ age > 20 }",
		"expected": `{"aggs":{},"query":{"bool":{"must":[{"bool":{"should":[{"term":{"source":{"value":"c1"}}}]}},{"range":{"age":{"gt":"20"}}}]}},"size":1000,"sort":[{"date":{"missing":"_first","order":"desc","unmapped_type":"string"}}],"track_total_hits":true}`,
	},
	{
		// 24, 别名测试
		"index":    "24",
		"input":    "L::c1:(age as a){ a > 20 }",
		"expected": `{"aggs":{},"_source":["date","age"],"query":{"bool":{"must":[{"bool":{"should":[{"term":{"source":{"value":"c1"}}}]}},{"range":{"age":{"gt":"20"}}}]}},"size":1000,"sort":[{"date":{"missing":"_first","order":"desc","unmapped_type":"string"}}],"track_total_hits":true}`,
	},
	{
		// 25, 聚合别名测试
		"index":    "25",
		"input":    "L::c1:(age as a) by a limit 3",
		"expected": `{"aggs":{"a":{"aggs":{"top_hits":{"top_hits":{"_source":["date","age"],"size":1,"sort":[{"date":{"missing":"_first","order":"desc","unmapped_type":"string"}}]}}},"terms":{"field":"age","size":3}}},"_source":["date","age"],"query":{"bool":{"must":[{"bool":{"should":[{"term":{"source":{"value":"c1"}}}]}}]}},"size":0}`,
	},
	{
		// 26, 测试count函数
		"index":    "25",
		"input":    "L::c1:(count(age)) by age",
		"expected": `{"aggs":{"age":{"aggs":{"count_age":{"value_count":{"field":"age"}}},"terms":{"field":"age","size":10}}},"query":{"bool":{"must":[{"bool":{"should":[{"term":{"source":{"value":"c1"}}}]}}]}},"size":0}`,
	},

	{
		// 27, rum类型查询
		"index":    "27",
		"input":    "R::c1:(age, height) {age > 20}",
		"expected": `{"aggs":{},"_source":["date","age","height"],"query":{"bool":{"must":[{"bool":{"should":[{"term":{"` + RCLASS + `":{"value":"c1"}}}]}},{"range":{"age":{"gt":"20"}}}]}},"size":1000,"sort":[{"date":{"missing":"_first","order":"desc","unmapped_type":"string"}}],"track_total_hits":true}`,
	},
	{
		// 28, rum show函数
		"index":    "28",
		"input":    "show_rum_type()",
		"expected": `{"aggs":{"aggs1":{"terms":{"field":"` + RCLASS + `","size":"` + string(MaxLimitStr) + `"}}},"size":0}`,
	},
	{
		// 29, top函数
		"index":    "29",
		"input":    "L::c1:(top(age, 3))",
		"expected": `{"aggs":{"top_age":{"top_hits":{"_source":["date","age"],"size":"3","sort":[{"age":{"missing":"_first","order":"desc","unmapped_type":"string"}}]}}},"query":{"bool":{"must":[{"bool":{"should":[{"term":{"source":{"value":"c1"}}}]}},{"exists":{"field":"age"}}]}},"size":0}`,
	},
	{
		// 30, bottom函数
		"index":    "30",
		"input":    "L::c1:(bottom(age, 10))",
		"expected": `{"aggs":{"bottom_age":{"top_hits":{"_source":["date","age"],"size":"10","sort":[{"age":{"missing":"_first","order":"asc","unmapped_type":"string"}}]}}},"query":{"bool":{"must":[{"bool":{"should":[{"term":{"source":{"value":"c1"}}}]}},{"exists":{"field":"age"}}]}},"size":0}`,
	},
	{
		// 31, first函数
		"index":    "31",
		"input":    "L::c1:(first(age))",
		"expected": `{"aggs":{"first_age":{"top_hits":{"_source":["date","age"],"size":"1","sort":[{"date":{"missing":"_first","order":"asc","unmapped_type":"string"}}]}}},"query":{"bool":{"must":[{"bool":{"should":[{"term":{"source":{"value":"c1"}}}]}},{"exists":{"field":"age"}}]}},"size":0}`,
	},
	{
		// 32, last函数
		"index":    "32",
		"input":    "L::c1:(last(age))",
		"expected": `{"aggs":{"last_age":{"top_hits":{"_source":["date","age"],"size":"1","sort":[{"date":{"missing":"_last","order":"desc","unmapped_type":"string"}}]}}},"query":{"bool":{"must":[{"bool":{"should":[{"term":{"source":{"value":"c1"}}}]}},{"exists":{"field":"age"}}]}},"size":0}`,
	},
	{
		// 33, distinct函数
		"index":    "33",
		"input":    "L::c1:(distinct(age))",
		"expected": `{"aggs":{"distinct_age":{"terms":{"field":"age","size":1000}}},"query":{"bool":{"must":[{"bool":{"should":[{"term":{"source":{"value":"c1"}}}]}}]}},"size":0}`,
	},
	{
		// 34, countdistinct函数
		"index":    "34",
		"input":    "L::c1:(count_distinct(age))",
		"expected": `{"aggs":{"count_distinct_age":{"cardinality":{"field":"age"}}},"query":{"bool":{"must":[{"bool":{"should":[{"term":{"source":{"value":"c1"}}}]}}]}},"size":0}`,
	},
	{
		// 35, text类型字段聚合
		"index":    "35",
		"input":    "L::c1:() by `message`",
		"expected": `{"aggs":{"message":{"aggs":{"top_hits":{"top_hits":{"_source":[],"size":1,"sort":[{"date":{"missing":"_first","order":"desc","unmapped_type":"string"}}]}}},"terms":{"field":"message.keyword","size":10}}},"query":{"bool":{"must":[{"bool":{"should":[{"term":{"source":{"value":"c1"}}}]}}]}},"size":0}`,
	},
	{
		// 36, text类型字段term查询
		"index":    "36",
		"input":    "L::c1:() {`message`=`testValue`}",
		"expected": `{"aggs":{},"query":{"bool":{"must":[{"bool":{"should":[{"term":{"source":{"value":"c1"}}}]}},{"term":{"message.keyword":{"value":"testValue"}}}]}},"size":1000,"sort":[{"date":{"missing":"_first","order":"desc","unmapped_type":"string"}}],"track_total_hits":true}`,
	},
	{
		// 37, show fields
		"index":    "37",
		"input":    "show_rum_field(`resource`)",
		"expected": `{"term":{"source":{"value":"resource"}}},`,
	},

	{
		// 38, count text field
		"index":    "38",
		"input":    "L::`nginx`:(COUNT(`message`))",
		"expected": `{"aggs":{"count_message":{"value_count":{"field":"message.keyword"}}},"query":{"bool":{"must":[{"bool":{"should":[{"term":{"source":{"value":"nginx"}}}]}}]}},"size":0}`,
	},

	{
		// 39, 参数字段为"", 非``
		"index":    "39",
		"input":    `show_logging_field("nginx")`,
		"expected": `{"term":{"source":{"value":"nginx"}}},`,
	},

	{
		// 40, 数值型指标聚合函数，添加int转换
		"index":    "40",
		"input":    `L::nginx:(avg(int(__errorCode)))`,
		"expected": `{"aggs":{"avg___errorCode":{"avg":{"script":{"source":"if (doc['__errorCode'].size() \u003e 0) {Integer.parseInt(doc['__errorCode'].value)}"}}}},"query":{"bool":{"must":[{"bool":{"should":[{"term":{"source":{"value":"nginx"}}}]}}]}},"size":0}`,
	},
	{
		// 41, in查询
		"index":    "41",
		"input":    "L::c1:(age){ age in [20, `40`, '50']}",
		"expected": `{"aggs":{},"_source":["date","age"],"query":{"bool":{"must":[{"bool":{"should":[{"term":{"source":{"value":"c1"}}}]}},{"bool":{"should":[{"term":{"age":{"value":20}}},{"term":{"age":{"value":"40"}}},{"term":{"age":{"value":"50"}}}]}}]}},"size":1000,"sort":[{"date":{"missing":"_first","order":"desc","unmapped_type":"string"}}],"track_total_hits":true}`,
	},

	{
		// 42, offset query
		"index":    "42",
		"input":    "L::c1:(age){ } offset 9",
		"expected": `{"aggs":{},"_source":["date","age"],"query":{"bool":{"must":[{"bool":{"should":[{"term":{"source":{"value":"c1"}}}]}}]}},"size":1000,"from":9,"sort":[{"date":{"missing":"_first","order":"desc","unmapped_type":"string"}}],"track_total_hits":true}`,
	},

	{
		// 43, script查询
		"index":    "43",
		"input":    `L::c1:(age){ __script=script("if (doc.containsKey('f1') && doc.containsKey('f2') && !doc['f1'].empty && !doc['f2'].empty) { return (doc['f1'].value + doc['f2'].value) > 40 } else { return false}") }`,
		"expected": `{"aggs":{},"_source":["date","age"],"query":{"bool":{"must":[{"bool":{"should":[{"term":{"source":{"value":"c1"}}}]}},{"script":{"script":"if (doc.containsKey('f1') \u0026\u0026 doc.containsKey('f2') \u0026\u0026 !doc['f1'].empty \u0026\u0026 !doc['f2'].empty) { return (doc['f1'].value + doc['f2'].value) \u003e 40 } else { return false}"}}]}},"size":1000,"sort":[{"date":{"missing":"_first","order":"desc","unmapped_type":"string"}}],"track_total_hits":true}`,
	},
	{
		// 44, 聚合内置函数script
		"index":    "44",
		"input":    "L::c1:(avg(script('doc.f1.value + doc.f2.value'))){ age != 20 }",
		"expected": `{"aggs":{"avg_script":{"avg":{"script":{"source":"doc.f1.value + doc.f2.value"}}}},"query":{"bool":{"must":[{"bool":{"should":[{"term":{"source":{"value":"c1"}}}]}},{"bool":{"must_not":[{"term":{"age":{"value":"20"}}}]}}]}},"size":0}`,
	},

	{
		// 45, percent函数
		"index":    "44",
		"input":    "L::c1:(percentile(duration, 75)){ age != 20 }",
		"expected": `{"aggs":{"percentile_duration_75.0":{"percentiles":{"field":"duration","percents":[75]}}},"query":{"bool":{"must":[{"bool":{"should":[{"term":{"source":{"value":"c1"}}}]}},{"bool":{"must_not":[{"term":{"age":{"value":"20"}}}]}}]}},"size":0}`,
	},

	{
		// 46, wildcard函数
		"index":    "45",
		"input":    "L::c1:(){ f1=wildcard(`*{*`) }",
		"expected": `{"aggs":{},"query":{"bool":{"must":[{"bool":{"should":[{"term":{"source":{"value":"c1"}}}]}},{"wildcard":{"f1":{"value":"*{*"}}}]}},"size":1000,"sort":[{"date":{"missing":"_first","order":"desc","unmapped_type":"string"}}],"track_total_hits":true}`,
	},

	{
		// 47, 多个order by
		"index":    "47",
		"input":    "L::c1:(age){ age != 20 } order by f1 asc, f2 desc",
		"expected": `{"aggs":{},"_source":["date","age"],"query":{"bool":{"must":[{"bool":{"should":[{"term":{"source":{"value":"c1"}}}]}},{"bool":{"must_not":[{"term":{"age":{"value":"20"}}}]}}]}},"size":1000,"sort":[{"f1":{"missing":"_first","order":"asc","unmapped_type":"string"}},{"f2":{"missing":"_last","order":"desc","unmapped_type":"string"}}],"track_total_hits":true}`,
	},

	{
		// 48, histogram函数
		"index":    "48",
		"input":    "E::`monitor`:(histogram(date_range, 300, 6060, 100))",
		"expected": `{"aggs":{"date":{"aggs":{"date_range":{"value_count":{"field":"date_range"}}},"histogram":{"extended_bounds":{"max":6060,"min":300},"field":"date_range","interval":100}}},"query":{"bool":{"must":[{"bool":{"should":[{"term":{"source":{"value":"monitor"}}}]}},{"range":{"date_range":{"gte":300,"lte":6060}}}]}},"size":0}`,
	},

	{
		// 49, 聚合分页
		"index":    "49",
		"input":    "E::`monitor`:(count(host)) by host, status limit 3 offset 1",
		"expected": `{"aggs":{"host":{"aggs":{"status":{"aggs":{"count_host":{"value_count":{"field":"host"}}},"terms":{"field":"status","size":10}}},"terms":{"field":"host","size":4}}},"query":{"bool":{"must":[{"bool":{"should":[{"term":{"source":{"value":"monitor"}}}]}}]}},"size":0}`,
	},

	{
		// 50, 聚合分页
		"index":    "50",
		"input":    "E::`monitor`:(count(host)) by host limit 3 offset 1",
		"expected": `{"aggs":{"host":{"aggs":{"count_host":{"value_count":{"field":"host"}}},"terms":{"field":"host","size":4}}},"query":{"bool":{"must":[{"bool":{"should":[{"term":{"source":{"value":"monitor"}}}]}}]}},"size":0}`,
	},

	{
		// 51, query string
		"index":    "51",
		"input":    "L::re(`.*`) {`message`=querystring('GIN')} limit 1",
		"expected": `{"aggs":{},"query":{"bool":{"must":[{"query_string":{"analyze_wildcard":true,"default_field":"message","default_operator":"AND","query":"gin"}}]}},"size":1,"sort":[{"date":{"missing":"_first","order":"desc","unmapped_type":"string"}}],"track_total_hits":true}`,
	},

	{
		// 52, query string
		"index":    "52",
		"input":    "T::`ddtrace-web-dd`:(percentile(duration, 75) as percent) by service order by `percent.75` desc",
		"expected": `{"aggs":{"service":{"aggs":{"percent":{"percentiles":{"field":"duration","percents":[75]}}},"terms":{"field":"service","order":{"percent.75":"desc"},"size":10}}},"query":{"bool":{"must":[{"bool":{"should":[{"term":{"service":{"value":"ddtrace-web-dd"}}}]}}]}},"size":0}`,
	},
	{
		// 53, show security source
		"index":    "53",
		"input":    "show_security_category()",
		"expected": `{"aggs":{"aggs1":{"terms":{"field":"category","size":"` + string(MaxLimitStr) + `"}}},"size":0}`,
	},
	{
		// 54, show security source
		"index":    "54",
		"input":    "show_security_source()",
		"expected": `{"aggs":{"aggs1":{"terms":{"field":"category","size":"` + string(MaxLimitStr) + `"}}},"size":0}`,
	},
}

func TestESParse(t *testing.T) {

	for _, item := range testCases {
		t.Logf("--start %s case--", item["index"])
		asts, perr := ParseDQL(item["input"])
		if perr != nil {
			t.Errorf(
				"parse error: the input is:\n\n %s \n\n err is:\n\n %s \n",
				item["input"],
				perr,
			)
		}
		fmt.Println(asts)
		switch asts.(type) {
		case Stmts:
			for _, ast := range asts.(Stmts) {
				switch ast.(type) {
				case *DFQuery:
					m := ast.(*DFQuery)
					// m.Rewrite(w)
					switch m.Namespace {
					case "object", "O", "logging", "L", "event", "E", "tracing", "T", "rum", "R", "security", "S":
						res, err := m.ESQL()
						logESQl(t, item, res.(string), err)
					}

				case *Show:
					show := ast.(*Show)

					switch show.Namespace {
					case "object", "O", "logging", "L", "event", "E", "tracing", "T", "rum", "R", "security", "S":
						res, err := show.ESQL()
						logESQl(t, item, res.(string), err)
					}
				}
			}
		}
	}
}

func logESQl(t *testing.T, item map[string]string, res string, err error) {
	// var esDql = map[string]string{}
	var dql string
	// jerr := json.Unmarshal([]byte(res), &esDql)

	// if jerr != nil {
	// 	t.Errorf(`es dql error`)
	// }

	dql = res

	if err != nil {
		t.Errorf(
			"esql error: the input is:\n\n %s \n\n err is:\n\n %s \n",
			item["input"],
			err,
		)
	}
	if dql != item["expected"] {
		t.Errorf(
			"esql unexpected: the input is :\n\n%s \n\n excepted is:\n\n  %s \n\n but output is:\n\n %s\n",
			item["input"],
			item["expected"],
			dql,
		)
	} else {
		t.Logf("--end %s case--", item["index"])
	}
}
