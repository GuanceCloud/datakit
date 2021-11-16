package pipeline

import (
	"testing"

	"github.com/ip2location/ip2location-go"
	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

type mockGEO struct{}

func (m *mockGEO) Get(ip string) (*ip2location.IP2Locationrecord, error) {
	return &ip2location.IP2Locationrecord{
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
		Country_short: func() string {
			if ip == "unknown-country-short" {
				return geoDefaultVal
			} else {
				return "CN"
			}
		}(),
	}, nil
}

func TestGeoIpFunc(t *testing.T) {
	defaultGEO = &mockGEO{}

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

		p, err := NewPipeline(tt.script)
		if err != nil {
			tu.Equals(t, err.Error(), p.lastErr)

			if tt.fail {
				continue
			} else {
				tu.Ok(t, err) // always fail
			}
		}

		assertEqual(t, err, p.lastErr)

		p.Run(tt.data)
		if p.lastErr != nil {
			t.Errorf("lastErr: %s", p.lastErr)
			continue
		}

		t.Logf("output %+#v", p.Output)

		for k, v := range tt.expected {
			r, err := p.getContentStr(k)
			tu.Ok(t, err)

			tu.Assert(t, r == v, "%s != %s, output: %+#v", r, v, p.Output)
		}
	}
}
