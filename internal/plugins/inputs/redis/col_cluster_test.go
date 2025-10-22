// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package redis

import (
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/testutils"
)

func Test_parseClusterData(t *testing.T) {
	type fields struct {
		host   string
		tags   map[string]string
		tagger datakit.GlobalTagger
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
				host:   "localhost",
				tags:   map[string]string{"foo": "bar"},
				tagger: testutils.DefaultMockTagger(),
			},
			args: args{
				list: mockClusterData01,
			},
			want: []string{
				"redis_cluster,foo=bar,host=localhost cluster_current_epoch=6,cluster_known_nodes=6,cluster_my_epoch=2,cluster_size=3,cluster_slots_assigned=16384,cluster_slots_fail=0,cluster_slots_ok=16384,cluster_slots_pfail=0,cluster_state=1,cluster_stats_messages_auth_ack_sent=0,cluster_stats_messages_received=1483968,cluster_stats_messages_sent=1483972,total_cluster_links_buffer_limit_exceeded=0",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipt := defaultInput()
			ipt.Tags = tt.fields.tags

			inst := newInstance()
			inst.ipt = ipt
			inst.host = tt.fields.host
			inst.setup()

			got, err := inst.parseClusterData(tt.args.list)
			if (err != nil) != tt.wantErr {
				t.Errorf("Input.ParseClusterData() error = %v, wantErr %v", err, tt.wantErr)
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

			assert.Equal(t, tt.want, gotStr)
		})
	}
}

var mockClusterData01 = `cluster_state:ok
cluster_slots_assigned:16384
cluster_slots_ok:16384
cluster_slots_pfail:0
cluster_slots_fail:0
cluster_known_nodes:6
cluster_size:3
cluster_current_epoch:6
cluster_my_epoch:2
cluster_stats_messages_sent:1483972
cluster_stats_messages_received:1483968
total_cluster_links_buffer_limit_exceeded:0
cluster_stats_messages_auth-ack_sent:0
`
