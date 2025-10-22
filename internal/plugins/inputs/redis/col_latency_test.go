// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package redis

import (
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/testutils"
)

func Test_parseLatencyData(t *testing.T) {
	type fields struct {
		host            string
		tag             map[string]string
		tagger          datakit.GlobalTagger
		latencyLastTime map[string]time.Time
	}

	type args struct {
		list string
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "basic",
			fields: fields{
				host:            "localhost",
				tag:             map[string]string{"foo": "bar"},
				tagger:          testutils.DefaultMockTagger(),
				latencyLastTime: map[string]time.Time{},
			},
			args: args{
				list: "latency latest: [[command 1699346177 250 1000]]",
			},
			want: []string{
				"redis_latency,foo=bar,host=localhost cost_time=250i,event_name=\"command\",max_cost_time=1000i,message=\"command cost time 250ms,max_cost_time 1000ms\",occur_time=1699346177i,status=\"info\"",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipt := defaultInput()
			ipt.Tags = tt.fields.tag
			inst := newInstance()
			inst.ipt = ipt
			inst.host = "localhost"
			inst.setup()

			got := inst.parseLatencyData(tt.args.list)

			gotStr := []string{}
			for _, v := range got {
				s := v.LineProto()
				s = s[:strings.LastIndex(s, " ")]
				gotStr = append(gotStr, s)
			}
			sort.Strings(gotStr)

			assert.Equal(t, tt.want, gotStr)
		})
	}
}
