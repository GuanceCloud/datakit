// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package sinkdataway

import (
	"testing"

	"github.com/stretchr/testify/assert"
	lp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/lineproto"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/filter"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
)

// go test -v -timeout 30s -run ^TestGetURLFromMapConfig$ gitlab.jiagouyun.com/cloudcare-tools/datakit/io/sink/sinkdataway
func TestGetURLFromMapConfig(t *testing.T) {
	cases := []struct {
		name string
		in   map[string]interface{}
		out  string
	}{
		{
			name: "normal",
			in: map[string]interface{}{
				"target": "dataway",
				"url":    "https://openway.guance.com?token=tkn_xxxxx",
				"proxy":  "127.0.0.1:1080",
				"host":   "",
			},
			out: "https://openway.guance.com?token=tkn_xxxxx",
		},
		{
			name: "env",
			in: map[string]interface{}{
				"target": "dataway",
				"url":    "https://openway.guance.com",
				"token":  "tkn_xxxxx",
				"proxy":  "127.0.0.1:1080",
				"host":   "",
			},
			out: "https://openway.guance.com?token=tkn_xxxxx",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := getURLFromMapConfig(tc.in)
			assert.NoError(t, err)
			assert.Equal(t, tc.out, out)
		})
	}
}

// go test -v -timeout 30s -run ^TestFiltered$ gitlab.jiagouyun.com/cloudcare-tools/datakit/io/sink/sinkdataway
func TestFiltered(t *testing.T) {
	cases := []struct {
		name           string
		inCategory     string
		inLineProtocol string
		inFilters      []string
		out            bool
	}{
		{
			name:           "cpu",
			inCategory:     datakit.Metric,
			inLineProtocol: `cpu,cpu=cpu-total,host=izbp15lfa9l8950ivlj2gjz load5s=1i,usage_guest=0,usage_guest_nice=0,usage_idle=80.47606989203526,usage_iowait=0.15193719929354482,usage_irq=0,usage_nice=0,usage_softirq=0.050645733097848276,usage_steal=0,usage_system=4.76069891114016,usage_total=19.523930107964745,usage_user=14.560648265644621 1661232665480124338`,
			inFilters:      []string{`{cpu='cpu-total'}`},
			out:            true,
		},
		{
			name:       "need fixed tag source",
			inCategory: datakit.Logging,
			inLineProtocol: `test111,filename=log,filepath=/var/log/scheck/log,host=izbp15lfa9l8950ivlj2gjz,service=test111 log_read_lines=74i,log_read_offset=27082037i,message="2022-08-23T13:48:38.118+0800	INFO	output	output/datakit.go:85	init output for datakit ok,path=http://127.0.0.1:9529/v1/write/security?version=1.0.8&input=scheck",status="unknown" 1661233718379737907`,
			inFilters: []string{`{source='test111'}`},
			out:       true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pts, err := lp.ParsePoints([]byte(tc.inLineProtocol), nil)
			assert.NoError(t, err)
			conds, err := filter.GetConds(tc.inFilters)
			assert.NoError(t, err)
			for _, pt := range pts {
				out, err := filter.CheckPointFiltered(conds, tc.inCategory, &point.Point{Point: pt})
				assert.NoError(t, err)
				assert.Equal(t, tc.out, out)
			}
		})
	}
}
