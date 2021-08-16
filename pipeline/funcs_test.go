package pipeline

import (
	"testing"
	"time"
)

func assertEqual(t *testing.T, a, b interface{}) {
	t.Helper()
	if a != b {
		t.Errorf("Not Equal. %d %d", a, b)
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

	js := `{"a":{"first":2.3,"second":2,"thrid":"abc","forth":true},"age":47}`
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

// a.second 为 float 类型
func TestStringfFunc(t *testing.T) {
	//js := `{"a":{"first":2.3,"second":2,"thrid":"abc","forth":true},"age":47}`
	//script := `stringf(bb, "%d %s %v", a.second, a.thrid, a.forth);`
	js := `{"a":{"first":2.3,"second":2,"thrid":"abc","forth":true},"age":47}`
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
