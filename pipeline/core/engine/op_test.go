// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package engine

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/engine/funcs"
)

func TestOp(t *testing.T) {
	pl := `
	b = 1 + 1
	a = (b + 2) == 4 || False
	c = a * 3 + +100 + -10 + 
		3/1.1
	d = 4 / 3
	e = "dwdw" + "3"
	add_key(e)


	map_a = {"a": 1, "b" :2 , "1.1": 2, "nil": [1,2.1,"3"], "1": nil}

	f = map_a["a"]

	
	aaaa = 1.0 == (b = 1)

	a = v = a
	x7 = [1, 2.1, "3"]
	if b == 2 {
		x = 2
		for i = 1; i < 4; i = i+1 {
			x1 = 1 + x
			e = e + "1"
			if i == 2 {
				break
			}
			continue
			e = e + "2"
		}
	}
	ddd = "" 
	
	# 无序遍历 key
	# for x in {'a': 1, "b":2, "c":3} {
	# 	ddd = ddd + x
	# }

	# add_key(ddd)

	abc = {
		"a": [1,2,3],
		"d": "a",
		"1": 2,
		"d": nil
	}
	add_key(abc)
	abc["a"][-1] = 5
	add_key(abc)
` + ";`aa dw.` = abc;" + "add_key(`aa dw.`)" + `
for x in [1,2,3 ] {
	for y in [1,3,4] {
		if y == 3 {
			break
		}
		continue
	}
	break
}

add_key(len1, len([12,2]))
add_key(len2, len("123"))

for a = 0; a < 12; a = a + 1 {
	if a > 5 {
	  add_key(ef, a)
	  break
	}
	continue
	a = a - 1
  }

`

	scripts, errs := ParseScript(map[string]string{
		"abc.p": pl,
	}, map[string]string{
		"abc.p": "",
	}, funcs.FuncsMap, funcs.FuncsCheckMap)
	if len(errs) != 0 {
		t.Fatal(errs["abc.p"])
	}

	script := scripts["abc.p"]

	m, tags, f, tn, drop, err := RunScript(script, "test", nil, nil, time.Now(), nil)
	t.Log(m, tags, tn, drop)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, map[string]any{
		"aa dw.": `{"1":2,"a":[1,2,5],"d":null}`,
		"abc":    `{"1":2,"a":[1,2,5],"d":null}`,
		"e":      "dwdw3",
		"ef":     int64(6),
		"len1":   int64(2),
		"len2":   int64(3),
	}, f)
}
