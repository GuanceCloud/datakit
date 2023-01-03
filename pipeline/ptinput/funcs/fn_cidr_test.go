// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"testing"
	"time"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/ptinput"
)

func TestCIDR(t *testing.T) {
	cases := []struct {
		name     string
		outKey   string
		pl, in   string
		expected interface{}
		fail     bool
	}{
		{
			name: "ipv4_contains",
			pl: `ip = "192.0.2.233"
			if cidr(ip, "192.0.2.1/24") {
				add_key(ip_prefix, "192.0.2.1/24")
			}`,
			in:       ``,
			outKey:   "ip_prefix",
			expected: "192.0.2.1/24",
			fail:     false,
		},
		{
			name: "ipv4_not_contains",
			pl: `ip = "192.0.2.233"
			if cidr(ip, "192.0.1.1/24") {
				add_key(ip_prefix, "192.0.1.1/24")
			}`,
			in:     ``,
			outKey: "ip_prefix",
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
				t.Fatal(errR.Error())
			}

			if tc.fail {
				t.Logf("[%d]expect error: %s", idx, errR.Error())
			}
			v := pt.Fields[tc.outKey]
			tu.Equals(t, tc.expected, v)
			t.Logf("[%d] PASS", idx)
			ptinput.PutPoint(pt)
		})
	}
}
