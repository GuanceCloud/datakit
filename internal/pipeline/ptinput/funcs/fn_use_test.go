// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"testing"
	"time"

	"github.com/GuanceCloud/platypus/pkg/inimpl/guancecloud/input"
	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/ptinput"
)

func TestUse(t *testing.T) {
	case1 := map[string]string{
		"a.p": "if true {use(\"b.p\")}",
		"b.p": "add_key(b,1)",
		"d.p": "use(\"c.p\")",
		"c.p": "use(\"a.p\") use(\"d.p\") use(\"fcName.p\")",
	}

	ret := [2][]string{
		{"a.p", "b.p"},
		{"d.p", "c.p"},
	}

	timenow := time.Now()

	retCheck := input.Point{
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

	for _, name := range ret[0] {
		pt := ptinput.GetPoint()
		ptinput.InitPt(pt, "default", map[string]string{"ax": "1"}, nil, timenow)
		errR := runScript(ret1[name], pt)

		if errR != nil {
			ptinput.PutPoint(pt)
			t.Fatal(errR)
		}
		assert.Equal(t, retCheck.Tags, pt.Tags)
		assert.Equal(t, retCheck.Fields, pt.Fields)
		assert.Equal(t, retCheck.Drop, pt.Drop)
		assert.Equal(t, retCheck.Measurement, pt.Name)

		ptinput.PutPoint(pt)
	}
}
