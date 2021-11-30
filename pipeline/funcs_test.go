package pipeline

import (
	"testing"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
)

func assertEqual(t *testing.T, a, b interface{}) {
	t.Helper()
	if a != b {
		t.Errorf("Not Equal. %v %v", a, b)
	}
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
		rename(newhour, hour)
	`
	p, err := NewPipeline(script)

	assertEqual(t, err, nil)

	p.Run("12:13:14")

	assertEqual(t, p.lastErr, nil)

	r, _ := p.getContentStr("newhour")

	assertEqual(t, r, "12")
}

func TestExprFunc(t *testing.T) {
	js := `{"a":{"first":2.3,"second":2,"third":"abc","forth":true},"age":47}`
	script := `json(_, a.second)
		cast(a.second, "int")
		expr(a.second*10+(2+3)*5, bb)
	`
	p, err := NewPipeline(script)
	assertEqual(t, err, nil)

	p.Run(js)
	assertEqual(t, p.lastErr, nil)

	if v, err := p.getContentStr("bb"); err != nil {
		t.Error(err)
	} else {
		assertEqual(t, v, "45")
	}
}

func TestCastFloat2IntFunc(t *testing.T) {
	js := `{"a":{"first":2.3,"second":2,"third":"abc","forth":true},"age":47}`
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
	js := `{"a":{"first":2.3,"second":2,"third":"abc","forth":true},"age":47}`
	script := `json(_, a.second)
cast(a.second, "float")
`
	p, err := NewPipeline(script)
	assertEqual(t, err, nil)

	p.Run(js)

	v, _ := p.getContentStr("a.second")
	assertEqual(t, v, "2")
}

// a.second 为 float 类型
func TestStringfFunc(t *testing.T) {
	// js := `{"a":{"first":2.3,"second":2,"third":"abc","forth":true},"age":47}`
	// script := `stringf(bb, "%d %s %v", a.second, a.thrid, a.forth);`
	js := `{"a":{"first":2.3,"second":2,"third":"abc","forth":true},"age":47}`
	script := `json(_, a.second)
		json(_, a.thrid)
		json(_, a.forth)
		cast(a.second, "int")
		strfmt(bb, "%d %s %v", a.second, a.thrid, a.forth)
	`
	p, err := NewPipeline(script)
	assertEqual(t, err, nil)

	p.Run(js)
	v, _ := p.getContent("bb")
	assertEqual(t, v, "2 abc true")
}

func TestUppercaseFunc(t *testing.T) {
	js := `{"a":{"first":2.3,"second":2,"third":"abc","forth":true},"age":47}`
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
	js := `{"a":{"first":2.3,"second":2,"third":"aBC","forth":true},"age":47}`
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
	js := `{"a":{"first":2.3,"second":2,"third":"aBC","forth":true},"age":47}`
	script := `add_key(aa, 3)
`
	p, err := NewPipeline(script)
	assertEqual(t, err, nil)

	p.Run(js)
	v, _ := p.getContentStr("aa")
	assertEqual(t, v, "3")
}

func TestDropkeyFunc(t *testing.T) {
	js := `{"a":{"first":2.3,"second":2,"third":"aBC","forth":true},"age":47}`
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

func TestTimeParse(t *testing.T) {
	cases := []*struct {
		fmt    string
		input  string
		output int64
		pass   bool
	}{
		{
			fmt:    "2006-01-02 15:04:05",
			input:  "2021-07-20 18:00:00",
			output: 1626804000000000000,
		},
		{
			fmt:    "01/02/06",
			input:  "07/20/21",
			output: 1626739200000000000,
		},
	}

	tz, _ := time.LoadLocation("UTC")
	for _, c := range cases {
		if pt, err := time.ParseInLocation(c.fmt, c.input, tz); err == nil {
			if pt.UnixNano() == c.output {
				c.pass = true
			} else {
				t.Error("act:", pt.UnixNano(),
					"exp:", c.output, "delta:", pt.UnixNano()-c.output)
			}
		} else {
			t.Error(err)
		}
	}
}

func TestDefaultTime(t *testing.T) {
	cases := []struct {
		name              string
		time              string
		timezone          string
		expectedTimestamp int64
	}{
		// test different timezones
		{
			name:              "normal",
			time:              "2017-12-29T12:33:33.095243Z",
			timezone:          "Asia/Shanghai",
			expectedTimestamp: 1514522013095243000,
		},
		{
			name:              "normal",
			time:              "2017-12-29T12:33:33.095243Z",
			timezone:          "+8",
			expectedTimestamp: 1514522013095243000,
		},
		{
			name:              "normal",
			time:              "2017-12-29T12:33:33.095243Z",
			timezone:          "MST",
			expectedTimestamp: 1514576013095243000,
		},
		{
			name:              "normal",
			time:              "2017-12-29T12:33:33.095243Z",
			timezone:          "America/Los_Angeles",
			expectedTimestamp: 1514579613095243000,
		},
		{
			name:              "normal",
			time:              "2021-11-29 09:42:42.927",
			timezone:          "Asia/Shanghai",
			expectedTimestamp: 1638150162927000000,
		},
		{
			name:              "normal",
			time:              "2021-11-29 09:42:42.927",
			timezone:          "America/Chicago",
			expectedTimestamp: 1638200562927000000,
		},
		{
			name:              "normal",
			time:              "2021-11-29 09:42:42.927",
			timezone:          "UTC",
			expectedTimestamp: 1638178962927000000,
		},
		{
			name:              "normal",
			time:              "2021/11/29 09:42:42.927",
			timezone:          "Europe/London",
			expectedTimestamp: 1638178962927000000,
		},
		{
			name:              "normal",
			time:              "2021/11/29 09:42:42.927",
			timezone:          "+0",
			expectedTimestamp: 1638178962927000000,
		},
		{
			name:              "normal",
			time:              "2021-11-29 09:42:42.927 +0000 UTC",
			timezone:          "",
			expectedTimestamp: 1638178962927000000,
		},
		// test different formats
		{
			name:              "normal",
			time:              "Mon 20 Sep 2021 09:42:42",
			timezone:          "MST",
			expectedTimestamp: 1632156162000000000,
		},
		{
			name:              "normal",
			time:              "2015-09-30 18:48:56.35272715 +0000 UTC",
			timezone:          "",
			expectedTimestamp: 1443638936352727150,
		},
		{
			name:              "normal",
			time:              "2015-09-30 18:48:56.35272715 +0900 UTC",
			timezone:          "",
			expectedTimestamp: 1443606536352727150,
		},
		{
			name:              "normal",
			time:              "2021-11-29 13:36:56.352",
			timezone:          "UTC",
			expectedTimestamp: 1638193016352000000,
		},
		{
			name:              "normal",
			time:              "2021年11月29日 13:36:56.352",
			timezone:          "UTC",
			expectedTimestamp: 1638193016352000000,
		},
		{
			name:              "normal",
			time:              "2021-11-29T12:00:11.986+0800",
			timezone:          "UTC",
			expectedTimestamp: 1638158411986000000,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pl := &Pipeline{
				Output: map[string]interface{}{
					"time": tc.time,
				},
				ast: &parser.Ast{
					Functions: []*parser.FuncExpr{
						{
							Name: "default_time",
							Param: []parser.Node{
								&parser.Identifier{Name: "time"},
								&parser.StringLiteral{Val: tc.timezone},
							},
						},
					},
				},
			}
			p, err := DefaultTime(pl, pl.ast.Functions[0])
			if err != nil {
				t.Error(err)
			}
			assertEqual(t, p.Output["time"], tc.expectedTimestamp)
		})
	}
}
