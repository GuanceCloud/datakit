// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/ptinput"
)

func TestConvW3CTraceID(t *testing.T) {
	cases := []struct {
		name, pl, in string
		key          string
		fail         bool
		expect       any
	}{
		{
			name: "value type: string",
			in:   `18962fdd9eea517f2ae0771ea69d6e16`,
			pl: `
			grok(_, "%{NOTSPACE:trace_id}")

			conv_traceid_w3c_to_dd(trace_id)
`,
			key:    "trace_id",
			expect: "3089600317904219670",
			fail:   false,
		},

		{
			name: "value type: string",
			in:   `2f7cdb2b45447bcb43820ddf56d9a654`,
			pl: `
			grok(_, "%{NOTSPACE:trace_id}")

			conv_traceid_w3c_to_dd(trace_id)
`,
			key:    "trace_id",
			expect: "4864465800399529556",
			fail:   false,
		},

		{
			name: "value type: string",
			in:   `43820ddf56d9a654`,
			pl: `
			grok(_, "%{NOTSPACE:trace_id}")

			conv_traceid_w3c_to_dd(trace_id)
`,
			key:    "trace_id",
			expect: "4864465800399529556",
			fail:   false,
		},

		{
			name: "value type: string",
			in:   `03820ddf56d9a654`,
			pl: `
			grok(_, "%{NOTSPACE:trace_id}")

			conv_traceid_w3c_to_dd(trace_id)
`,
			key:    "trace_id",
			expect: "252779781972141652",
			fail:   false,
		},

		{
			name: "value type: string",
			in:   `0f7cdb2b45447bcb43820ddf56d9a654`,
			pl: `
			grok(_, "%{NOTSPACE:trace_id}")

			conv_traceid_w3c_to_dd(trace_id)
`,
			key:    "trace_id",
			expect: "4864465800399529556",
			fail:   false,
		},

		{
			name: "value type: string",
			in:   `10f7cdb2b45447bcb43820ddf56d9a654`,
			pl: `
			grok(_, "%{NOTSPACE:trace_id}")

			conv_traceid_w3c_to_dd(trace_id)
`,
			key: "trace_id",
			// 原样返回
			expect: "10f7cdb2b45447bcb43820ddf56d9a654",
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
			pt := ptinput.NewPlPoint(point.Logging, "test", nil, map[string]any{"message": tc.in}, time.Now())

			errR := runScript(runner, pt)

			if errR == nil {
				v, _, ok := pt.Get(tc.key)
				assert.Equal(t, nil, ok)
				assert.Equal(t, tc.expect, v)
				t.Logf("[%d] PASS", idx)
			} else {
				t.Error(errR)
			}
		})
	}
}
