package dql

import (
	"encoding/json"
	"testing"
)

var (
	jsonCases = []struct {
		in   *InnerParse
		fail bool
	}{
		{
			in: &InnerParse{DQLs: []string{`M::mem [1615883598:1615883608:auto(5)]`}},
		},
		{
			in: &InnerParse{DQLs: []string{`M::mem:(avg(int(usage)))`}},
		},
		{
			in: &InnerParse{DQLs: []string{`M::cpu {x > 0, f in [1,2,3,4]}`}},
		},
		{
			in: &InnerParse{DQLs: []string{`cpu:(f1, f2) {host="abc"} LINK O::ecs:(a1,a2) {a1 = "xxx"} WITH {f1=a1}`}},
		},
		{
			in: &InnerParse{DQLs: []string{`show_tag_value(cpu)`}},
		},
		{
			in: &InnerParse{DQLs: []string{`show_tag_value(cpu, keyin = ["region", "host"]) {service="redis"} [10m] LIMIT 3 OFFSET 5`}},
		},
		{
			in: &InnerParse{DQLs: []string{"difference(\"m::`datakit`:(last(`heap_alloc`)) by `host`\").derivative()"}},
		},
	}
)

func TestParseDQLToJSON(t *testing.T) {
	for idx, c := range jsonCases {
		res := ParseDQLToJSONV2(c.in)
		j, err := json.MarshalIndent(res, "", "  ")
		if err != nil {
			t.Errorf("[%d] ERROR -> %s\n", idx, err)
			continue
		}
		t.Logf("[%d] OK -> ResultJSON:\n%s\n", idx, j)
	}
}
