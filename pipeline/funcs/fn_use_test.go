package funcs

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
)

func TestUse(t *testing.T) {
	case1 := map[string]string{
		"a.p": "if true {use(\"b.p\")}",
		"b.p": "add_key(b,1)",
		"d.p": "use(\"c.p\")",
		"c.p": "use(\"a.p\") use(\"d.p\") use(\"fcName.p\")",
		"f.p": "s",
	}

	ret := [2][]string{
		{"a.p", "b.p"},
		{"d.p", "c.p", "f.p"},
	}

	timenow := time.Now()

	retCheck := parser.Output{
		Measurement: "default",
		Tags: map[string]string{
			"ax": "1",
		},
		Fields: map[string]interface{}{
			"b": int64(1),
		},
		Time: timenow,
		Drop: false,
	}

	ret1, ret2 := NewTestingRunner2(case1)
	assert.Equal(t, len(ret[0]), len(ret1))
	assert.Equal(t, len(ret[1]), len(ret2))

	for _, v := range ret[0] {
		if _, ok := ret1[v]; !ok {
			t.Error(v)
		}
	}

	for _, v := range ret[1] {
		if _, ok := ret2[v]; !ok {
			t.Error(v)
		}
	}

	fc := func(ng *parser.Engine) (*parser.Output, error) {
		return ng.Run("default", map[string]string{
			"ax": "1",
		}, nil, "message", timenow)
	}

	for _, name := range ret[0] {
		o, err := fc(ret1[name])
		if err != nil {
			t.Error(err)
			continue
		}
		assert.Equal(t, retCheck.Tags, o.Tags)
		assert.Equal(t, retCheck.Fields, o.Fields)
		assert.Equal(t, retCheck.Drop, o.Drop)
		assert.Equal(t, retCheck.Measurement, o.Measurement)
	}
}
