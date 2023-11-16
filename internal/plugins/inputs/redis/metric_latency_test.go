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

func TestInput_parseLatencyData(t *testing.T) {
	type fields struct {
		Host            string
		Tags            map[string]string
		Election        bool
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
			name: "ok 1 line data no election",
			fields: fields{
				Host:            "localhost",
				Tags:            map[string]string{"foo": "bar"},
				Election:        false,
				tagger:          testutils.DefaultMockTagger(),
				latencyLastTime: map[string]time.Time{},
			},
			args: args{
				list: "latency latest: [[command 1699346177 250 1000]]",
			},
			want: []string{
				"redis_latency,foo=bar,host=HOST cost_time=250i,event_name=\"command\",max_cost_time=1000i,message=\"command cost time 250ms,max_cost_time 1000ms\",occur_time=1699346177i,status=\"unknown\"",
			},
			wantErr: false,
		},
		{
			name: "ok 1 line data election",
			fields: fields{
				Host:            "localhost",
				Tags:            map[string]string{"foo": "bar"},
				Election:        true,
				tagger:          testutils.DefaultMockTagger(),
				latencyLastTime: map[string]time.Time{},
			},
			args: args{
				list: "latency latest: [[command 1699346177 250 1000]]",
			},
			want: []string{
				"redis_latency,election=TRUE,foo=bar cost_time=250i,event_name=\"command\",max_cost_time=1000i,message=\"command cost time 250ms,max_cost_time 1000ms\",occur_time=1699346177i,status=\"unknown\"",
			},
			wantErr: false,
		},
		{
			name: "ok 2 line data no election",
			fields: fields{
				Host:            "localhost",
				Tags:            map[string]string{"foo": "bar"},
				Election:        false,
				tagger:          testutils.DefaultMockTagger(),
				latencyLastTime: map[string]time.Time{},
			},
			args: args{
				list: "latency latest: [[command 1699346177 250 1000] [xxxxx 1699346178 251 1001]]",
			},
			want: []string{
				"redis_latency,foo=bar,host=HOST cost_time=250i,event_name=\"command\",max_cost_time=1000i,message=\"command cost time 250ms,max_cost_time 1000ms\",occur_time=1699346177i,status=\"unknown\"",
				"redis_latency,foo=bar,host=HOST cost_time=251i,event_name=\"xxxxx\",max_cost_time=1001i,message=\"xxxxx cost time 251ms,max_cost_time 1001ms\",occur_time=1699346178i,status=\"unknown\"",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipt := &Input{
				Host:            tt.fields.Host,
				Tags:            tt.fields.Tags,
				Election:        tt.fields.Election,
				tagger:          tt.fields.tagger,
				latencyLastTime: tt.fields.latencyLastTime,
			}

			ipt.setup()

			got, err := ipt.parseLatencyData(tt.args.list)
			if (err != nil) != tt.wantErr {
				t.Errorf("Input.parseLatencyData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			gotStr := []string{}
			for _, v := range got {
				s := v.LineProto()
				s = s[:strings.LastIndex(s, " ")]
				gotStr = append(gotStr, s)
			}
			sort.Strings(gotStr)

			assert.Equal(t, gotStr, tt.want)
		})
	}
}
