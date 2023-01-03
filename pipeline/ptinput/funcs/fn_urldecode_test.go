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

func TestURLDecode(t *testing.T) {
	cases := []struct {
		name     string
		pl, in   string
		outKey   string
		expected interface{}
		fail     bool
	}{
		{
			name: "normal",
			pl: `json(_, url) 
			url_decode(url)`,
			in:       `{"url":"http%3a%2f%2fwww.baidu.com%2fs%3fwd%3d%e6%b5%8b%e8%af%95"}`,
			outKey:   "url",
			expected: "http://www.baidu.com/s?wd=测试",
			fail:     false,
		},

		{
			name: "normal",
			pl: `json(_, url) 
			url_decode(url)`,
			in:       `{"url":"https:%2F%2Fkubernetes.io%2Fdocs%2Freference%2Faccess-authn-authz%2Fbootstrap-tokens%2F"}`,
			outKey:   "url",
			expected: "https://kubernetes.io/docs/reference/access-authn-authz/bootstrap-tokens/",
			fail:     false,
		},

		{
			name: "non-existed key",
			pl: `json(_, url) 
			url_decode(link)`,
			in:       `{"url":"https:%2F%2Fkubernetes.io%2Fdocs%2Freference%2Faccess-authn-authz%2Fbootstrap-tokens%2F"}`,
			outKey:   "link",
			expected: nil,
			fail:     true,
		},

		{
			name: "invalid arg",
			pl:   `url_decode("https:%2F%2Fkubernetes.io%2Fdocs%2Freference%2Faccess-authn-authz%2Fbootstrap-tokens%2F")`,
			in:   `{"url":"https:%2F%2Fkubernetes.io%2Fdocs%2Freference%2Faccess-authn-authz%2Fbootstrap-tokens%2F"}`,
			fail: true,
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

			if v, ok := pt.Fields[tc.outKey]; !ok {
				if !tc.fail {
					t.Errorf("[%d]expect error", idx)
				}
			} else {
				tu.Equals(t, tc.expected, v)
				t.Logf("[%d] PASS", idx)
			}
			ptinput.PutPoint(pt)
		})
	}
}
