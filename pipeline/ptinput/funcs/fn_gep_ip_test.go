// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"testing"
	"time"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/ipdb"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/ptinput"
)

type mockGEO struct{}

func (m *mockGEO) Init(dataDir string, config map[string]string) {}
func (m *mockGEO) SearchIsp(ip string) string                    { return "" }

func (m *mockGEO) Geo(ip string) (*ipdb.IPdbRecord, error) {
	return &ipdb.IPdbRecord{
		City: func() string {
			if ip == "unknown-city" {
				return geoDefaultVal
			} else {
				return "Shanghai"
			}
		}(),
		Region: func() string {
			if ip == "unknown-region" {
				return geoDefaultVal
			} else {
				return "Shanghai"
			}
		}(),
		Country: func() string {
			if ip == "unknown-country-short" {
				return geoDefaultVal
			} else {
				return "CN"
			}
		}(),
	}, nil
}

func TestGeoIpFunc(t *testing.T) {
	ipdbInstance = &mockGEO{}
	cases := []struct {
		in     string
		script string

		expected map[string]string

		fail bool
	}{
		{
			in: `{"ip":"1.2.3.4-something", "second":2,"third":"abc","forth":true}`,
			script: `
				json(_, ip)
				geoip(ip)`,
			expected: map[string]string{
				"city":     "Shanghai",
				"country":  "CN",
				"province": "Shanghai",
				"isp":      geoDefaultVal,
			},
		},

		{
			in: `{"ip":"unknown-city", "second":2,"third":"abc","forth":true}`,
			script: `
				json(_, ip)
				geoip(ip)`,
			expected: map[string]string{
				"city":     geoDefaultVal,
				"country":  "CN",
				"province": "Shanghai",
				"isp":      geoDefaultVal,
			},
		},

		{
			in: `{"aa": {"ip":"116.228.89.xxx"}, "second":2,"third":"abc","forth":true}`,
			script: `
				json(_, aa.ip)
				geoip(aa.ip)`,
			expected: map[string]string{
				"city":     "Shanghai",
				"country":  "CN",
				"province": "Shanghai",
				"isp":      geoDefaultVal,
			},
		},

		{
			in: `{"aa": {"ip":"unknown-region"}, "second":2,"third":"abc","forth":true}`,
			script: `
				json(_, aa.ip)
				geoip(aa.ip)`,
			expected: map[string]string{
				"city":     "Shanghai",
				"country":  "CN",
				"province": geoDefaultVal,
				"isp":      geoDefaultVal,
			},
		},

		{
			in: `{"aa": {"ip":"unknown-country-short"}, "second":2,"third":"abc","forth":true}`,
			script: `
				json(_, aa.ip)
				geoip(aa.ip)`,
			expected: map[string]string{
				"city":     "Shanghai",
				"country":  geoDefaultVal,
				"province": "Shanghai",
				"isp":      geoDefaultVal,
			},
		},
	}

	for idx, tc := range cases {
		t.Logf("case %d...", idx)

		runner, err := NewTestingRunner(tc.script)
		if err != nil {
			t.Errorf("[%d] failed: %s", idx, err)
			return
		}

		pt := ptinput.GetPoint()
		ptinput.InitPt(pt, "test", nil, map[string]any{"message": tc.in}, time.Now())
		errR := runScript(runner, pt)

		if errR != nil {
			ptinput.PutPoint(pt)
			t.Fatal(errR)
		}

		for k, v := range tc.expected {
			r, ok := pt.Fields[k]
			tu.Assert(t, ok == true, "!ok")
			tu.Assert(t, r == v, "%s != %s, output: %+#v", r, v)
		}
		ptinput.PutPoint(pt)
	}
}
