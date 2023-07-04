// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	T "testing"
	"time"

	"github.com/stretchr/testify/assert"
	iprom "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/prom"
)

func TestNewPromRunnerWithTomlConfig(t *T.T) {
	testcase := []struct {
		in  string
		out []*promRunner
	}{
		{
			in: `
			    [[inputs.prom]]
			      urls = ["http://127.0.0.1:9527/v1/metric"]
			      source = "kodo-prom"
			      interval = "20s"
			      timeout = "10s"
			    
			      [[inputs.prom.measurements]]
			        prefix = "kodo_api"
			        name = "kodo_api"
			    
			      [inputs.prom.tags]
			       # some_tag = "some_value"
			       # more_tag = "some_other_value"
		`,
			out: []*promRunner{
				{
					conf: &promConfig{
						Source:   "kodo-prom",
						Interval: time.Second * 20,
						Timeout:  time.Second * 10,
						URLs:     []string{"http://127.0.0.1:9527/v1/metric"},
						Measurements: []iprom.Rule{
							{
								Prefix: "kodo_api",
								Name:   "kodo_api",
							},
						},
						Tags: map[string]string{},
					},
				},
			},
		},

		{
			in: `
			    [[inputs.prom]]
			      urls = ["http://127.0.0.1:9527/v1/metric"]
			      source = "kodo-prom"
			      interval = "20s"
			      timeout = "10s"

			    [[inputs.prom]]
			      urls = ["http://127.0.0.1:9527/v1/metric"]
			      source = "kodo-prom-2"
			      interval = "30s"
			      timeout = "20s"
		`,
			out: []*promRunner{
				{
					conf: &promConfig{
						Source:   "kodo-prom",
						Interval: time.Second * 20,
						Timeout:  time.Second * 10,
						URLs:     []string{"http://127.0.0.1:9527/v1/metric"},
					},
				},
				{
					conf: &promConfig{
						Source:   "kodo-prom-2",
						Interval: time.Second * 30,
						Timeout:  time.Second * 20,
						URLs:     []string{"http://127.0.0.1:9527/v1/metric"},
					},
				},
			},
		},
	}

	for _, tc := range testcase {
		rs, err := newPromRunnerWithTomlConfig(tc.in)
		assert.NoError(t, err)

		assert.Equal(t, len(tc.out), len(rs))

		for idx := range tc.out {
			assert.Equal(t, tc.out[idx].conf, rs[idx].conf)
		}
	}
}
