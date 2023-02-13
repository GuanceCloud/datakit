// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"testing"
	"time"

	tu "github.com/GuanceCloud/cliutils/testutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/ptinput"
)

func TestExit(t *testing.T) {
	cases := []struct {
		name, pl, in string
		outkey       string
		expect       interface{}
		fail         bool
	}{
		{
			name: "cast int",
			in:   `162.62.81.1 - - [29/Nov/2021:07:30:50 +0000] "123 /?signature=b8d8ea&timestamp=1638171049 HTTP/1.1" 200 413 "-" "Mozilla/4.0"`,
			pl: `
	grok(_, "%{IPORHOST:client_ip} %{NOTSPACE} %{NOTSPACE} \\[%{HTTPDATE:time}\\] \"%{DATA:data} %{GREEDYDATA} HTTP/%{NUMBER}\" %{INT:status_code} %{INT:bytes}")
	exit()
	cast(data, "int")
	`,
			outkey: "data",
			expect: "123",
			fail:   false,
		},
	}

	for idx, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runner, err := NewTestingRunner(tc.pl)
			if err != nil {
				if tc.fail {
					t.Logf("[%d]expect error: %s", idx, err)
				} else {
					t.Errorf("[%d] failed: %s", idx, err)
				}
				return
			}
			pt := ptinput.GetPoint()
			ptinput.InitPt(pt, "test", nil, map[string]any{"message": tc.in}, time.Now())
			errR := runScript(runner, pt)

			if errR != nil {
				ptinput.PutPoint(pt)
				t.Fatal(errR)
			}
			if pt.Drop {
				return
			}
			v := pt.Fields[tc.outkey]
			tu.Equals(t, tc.expect, v)
			t.Logf("[%d] PASS", idx)
			ptinput.PutPoint(pt)
		})
	}
}
