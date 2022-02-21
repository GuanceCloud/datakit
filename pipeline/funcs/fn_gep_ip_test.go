package funcs

import (
	"testing"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/ipdb"
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
		data   string
		script string

		expected map[string]string

		fail bool
	}{
		{
			data: `{"ip":"1.2.3.4-something", "second":2,"third":"abc","forth":true}`,
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
			data: `{"ip":"unknown-city", "second":2,"third":"abc","forth":true}`,
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
			data: `{"aa": {"ip":"116.228.89.xxx"}, "second":2,"third":"abc","forth":true}`,
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
			data: `{"aa": {"ip":"unknown-region"}, "second":2,"third":"abc","forth":true}`,
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
			data: `{"aa": {"ip":"unknown-country-short"}, "second":2,"third":"abc","forth":true}`,
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

	for idx, tt := range cases {
		t.Logf("case %d...", idx)

		runner, err := NewTestingRunner(tt.script)
		if err != nil {
			t.Errorf("[%d] failed: %s", idx, err)
			return
		}

		if err := runner.Run(tt.data); err != nil {
			t.Error(err)
			return
		}
		t.Log(runner.Result())

		for k, v := range tt.expected {
			r, err := runner.GetContentStr(k)
			tu.Ok(t, err)

			tu.Assert(t, r == v, "%s != %s, output: %+#v", r, v)
		}
	}
}
