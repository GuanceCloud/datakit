package dql

import (
	"encoding/json"
	"testing"

	ifxModel "github.com/influxdata/influxdb1-client/models"

	"gitlab.jiagouyun.com/cloudcare-tools/kodo/dql/parser"
)

func TestParseQuery(t *testing.T) {
	var queryCases = []struct {
		in  *singleQuery
		out string // out 没用到
	}{
		{
			in: &singleQuery{
				Query:     "O::`host_processes`:(cmdline) {}",
				TimeRange: []int64{1621921819000},
			},
			out: `"{"aggs":{},"_source":["last_update_time","cmdline"],"query":{"bool":{"must":[{"bool":{"should":[{"term":{"class":{"value":"host_processes"}}}]}},{"range":{"last_update_time":{"gte":"1622009622371"}}},{"range":{"last_update_time":{"lte":"1622009922371"}}}]}},"size":1000,"sort":[{"last_update_time":{"missing":"_first","order":"desc","unmapped_type":"string"}}],"track_total_hits":true}`,
		},
	}

	for index, query := range queryCases {
		astResults, err := ParseQuery(query.in)
		if err != nil {
			t.Fatalf("[%d] ERROR: %s", index, query.in.Query)
		}

		// 因为各个 namespace 翻译行为不同，可能会以时间等非常量值进行加盐，所以无法对最终查询语句做相等判断
		// 此处简单打印即可，便于观测
		for idx, ast := range astResults {
			t.Logf("[%d][%d] TransQuery: %s", index, idx, ast.Q)
		}
	}
}

func TestFillResult(t *testing.T) {
	for idx, tt := range fillCases {
		// if err := tt.FillResult(); err != nil {
		// 	t.Log(err)
		// }
		tt.FillConstant()

		t.Logf("[%d] OK testing:\nlen(results):%d\n%+#v\n", idx, len(tt.res[0].Values), tt.res)
	}
}

var (
	toResult = func(str string) []ifxModel.Row {
		var res []ifxModel.Row
		if err := json.Unmarshal([]byte(str), &res); err != nil {
			panic(err)
		}
		return res
	}

	toAST = func(dql string) *parser.DFQuery {
		asts, err := parser.ParseDQL(dql)
		if err != nil {
			panic(err)
		}
		m, ok := asts.(parser.Stmts)[0].(*parser.DFQuery)
		if !ok {
			panic("Not DFQuery")
		}
		return m
	}

	fillCases = []RewriteData{
		{
			m:   toAST(`cpu:(fill(last(usage_user), linear)) [::3s] limit 50`), // influxdb: SELECT LAST(\"usage_user\") FROM \"cpu\" GROUP BY time(3000ms) LIMIT 50"
			res: toResult(`[{"name":"cpu","columns":["time","last"],"values":[[1607385600000,null],[1607385603000,null],[1607385606000,null],[1607385609000,2.347417840375587],[1607385612000,null],[1607385615000,null],[1607385618000,0.15625],[1607385621000,null],[1607385624000,null],[1607385627000,null],[1607385630000,4.212168486739469],[1607385633000,null],[1607385636000,null],[1607385639000,0.3125],[1607385642000,null],[1607385645000,null],[1607385648000,2.8125],[1607385651000,null],[1607385654000,null],[1607385657000,null],[1607385660000,3.43213728549142],[1607385663000,null],[1607385666000,null],[1607385669000,1.25],[1607385672000,null],[1607385675000,null],[1607385678000,0.15625],[1607385681000,null],[1607385684000,null],[1607385687000,null],[1607385690000,1.09375],[1607385693000,null],[1607385696000,null],[1607385699000,3.9001560062402496],[1607385702000,null],[1607385705000,null],[1607385708000,0.46875],[1607385711000,null],[1607385714000,null],[1607385717000,null],[1607385720000,7.165109034267913],[1607385723000,null],[1607385726000,null],[1607385729000,1.40625],[1607385732000,null],[1607385735000,null],[1607385738000,0.625],[1607385741000,null],[1607385744000,null],[1607385747000,null]]}]`),
		},
		{
			m:   toAST(`cpu:(fill(last(usage_system), linear)) [::3s] limit 100`), // influxdb: SELECT LAST(\"usage_system\") FROM \"cpu\" GROUP BY time(3000ms) LIMIT 100"
			res: toResult(`[{"name":"cpu","columns":["time","last"],"values":[[1607385600000,1.24804992199688],[1607385603000,null],[1607385606000,null],[1607385609000,0.7824726134585289],[1607385612000,null],[1607385615000,null],[1607385618000,0.3125],[1607385621000,null],[1607385624000,null],[1607385627000,null],[1607385630000,4.368174726989079],[1607385633000,null],[1607385636000,null],[1607385639000,0.46875],[1607385642000,null],[1607385645000,null],[1607385648000,2.1875],[1607385651000,null],[1607385654000,null],[1607385657000,null],[1607385660000,3.1201248049921997],[1607385663000,null],[1607385666000,null],[1607385669000,0.78125],[1607385672000,null],[1607385675000,null],[1607385678000,0.78125],[1607385681000,null],[1607385684000,null],[1607385687000,null],[1607385690000,1.25],[1607385693000,null],[1607385696000,null],[1607385699000,4.5241809672386895],[1607385702000,null],[1607385705000,null],[1607385708000,0.46875],[1607385711000,null],[1607385714000,null],[1607385717000,null],[1607385720000,4.984423676012461],[1607385723000,null],[1607385726000,null],[1607385729000,0.78125],[1607385732000,null],[1607385735000,null],[1607385738000,0.3125],[1607385741000,null],[1607385744000,null],[1607385747000,null],[1607385750000,0.46875],[1607385753000,null],[1607385756000,null],[1607385759000,3.7441497659906395],[1607385762000,null],[1607385765000,null],[1607385768000,1.40625],[1607385771000,null],[1607385774000,null],[1607385777000,null],[1607385780000,1.0920436817472698],[1607385783000,null],[1607385786000,null],[1607385789000,4.381846635367762],[1607385792000,null],[1607385795000,null],[1607385798000,0.625],[1607385801000,null],[1607385804000,null],[1607385807000,null],[1607385810000,0.3125],[1607385813000,null],[1607385816000,null],[1607385819000,0.78125],[1607385822000,null],[1607385825000,null],[1607385828000,5],[1607385831000,null],[1607385834000,null],[1607385837000,null],[1607385840000,0.7800312012480499],[1607385843000,null],[1607385846000,null],[1607385849000,5],[1607385852000,null],[1607385855000,null],[1607385858000,0.9389671361502347],[1607385861000,null],[1607385864000,null],[1607385867000,null],[1607385870000,0.15625],[1607385873000,null],[1607385876000,null],[1607385879000,0.46875],[1607385882000,null],[1607385885000,null],[1607385888000,3.75],[1607385891000,null],[1607385894000,null],[1607385897000,null]]}]`),
		},
		// FillConstant
		{
			// 正常数据集
			m:   toAST(`O::cpu:(fill(last(usage_user), 999)) [2021-02-01 14:05:00:2021-02-01 14:10:00:20s]`),
			res: toResult(`[{"name":"cpu","columns":["time","last"],"values":[[1612188400000,12],[1612188420000,12],[1612188440000,12],[1612188460000,12],[1612188480000,12]]}]`),
		},
		{
			// 非时间对齐
			m:   toAST(`O::cpu:(fill(last(usage_user), 999)) [2021-02-01 14:05:00:2021-02-01 14:10:00:20s]`),
			res: toResult(`[{"name":"cpu","columns":["time","last"],"values":[[1612188410000,12],[1612188430000,12],[1612188450000,12],[1612188470000,12],[1612188490000,12]]}]`),
		},
		{
			// 空集
			m:   toAST(`O::cpu:(fill(last(usage_user), 999)) [2021-02-01 14:05:00:2021-02-01 14:10:00:20s]`),
			res: toResult(`[{"name":"cpu","columns":["time","last"],"values":[]}]`),
		},
		{
			// 只有一个点
			m:   toAST(`O::cpu:(fill(last(usage_user), 999)) [2021-02-01 14:05:00:2021-02-01 14:10:00:20s]`),
			res: toResult(`[{"name":"cpu","columns":["time","last"],"values":[[1612188410000,12]]}]`),
		},
		{
			// 只有起始时间点
			m:   toAST(`O::cpu:(fill(last(usage_user), 999)) [2021-02-01 14:05:00:2021-02-01 14:10:00:20s]`),
			res: toResult(`[{"name":"cpu","columns":["time","last"],"values":[[1612188300000,12]]}]`),
		},
		{
			// 只有结束时间点
			m:   toAST(`O::cpu:(fill(last(usage_user), 999)) [2021-02-01 14:05:00:2021-02-01 14:10:00:20s]`),
			res: toResult(`[{"name":"cpu","columns":["time","last"],"values":[[1612188600000,12]]}]`),
		},
	}
)
