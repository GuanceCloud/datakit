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

func TestInput_parseReplicData(t *testing.T) {
	type fields struct {
		Host     string
		Tags     map[string]string
		Election bool
		tagger   datakit.GlobalTagger
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
			name: "ok have slave no election",
			fields: fields{
				Host:     "localhost",
				Tags:     map[string]string{"foo": "bar"},
				Election: false,
				tagger:   testutils.DefaultMockTagger(),
			},
			args: args{
				list: mockReplicaData01,
			},
			want: []string{
				"redis_replica,foo=bar,host=HOST,slave_addr=127.0.0.1:6379,slave_id=1 repl_delay=10",
				"redis_replica,foo=bar,host=HOST,slave_addr=127.0.0.1:6380,slave_id=0 repl_delay=10",
			},
			wantErr: false,
		},
		{
			name: "ok have slave election",
			fields: fields{
				Host:     "localhost",
				Tags:     map[string]string{"foo": "bar"},
				Election: true,
				tagger:   testutils.DefaultMockTagger(),
			},
			args: args{
				list: mockReplicaData01,
			},
			want: []string{
				"redis_replica,election=TRUE,foo=bar,slave_addr=127.0.0.1:6379,slave_id=1 repl_delay=10",
				"redis_replica,election=TRUE,foo=bar,slave_addr=127.0.0.1:6380,slave_id=0 repl_delay=10",
			},
			wantErr: false,
		},
		{
			name: "ok no slave",
			fields: fields{
				Host:     "localhost",
				Tags:     map[string]string{"foo": "bar"},
				Election: false,
				tagger:   testutils.DefaultMockTagger(),
			},
			args: args{
				list: mockReplicaData02,
			},
			want:    []string{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipt := &Input{
				Host:     tt.fields.Host,
				Tags:     tt.fields.Tags,
				Election: tt.fields.Election,
				tagger:   tt.fields.tagger,
			}

			ipt.setup()

			got, err := ipt.parseReplicaData(tt.args.list)
			if (err != nil) != tt.wantErr {
				t.Errorf("Input.parseReplicData() error = %v, wantErr %v", err, tt.wantErr)
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

var mockReplicaData01 = `# Replication
role:master
connected_slaves:2
slave0:ip=127.0.0.1,port=6380,state=online,offset=4046,lag=0
slave1:ip=127.0.0.1,port=6379,state=online,offset=4046,lag=0
master_failover_state:no-failover
master_replid:458e0f8b3f40adb50a3dca8cc90e76b96937107b
master_replid2:943718468a346457e58ad233607f464998e6159c
master_repl_offset:4056
second_repl_offset:2591
repl_backlog_active:1
repl_backlog_size:1048576
repl_backlog_first_byte_offset:15
repl_backlog_histlen:4032
`

var mockReplicaData02 = `# Replication
role:master
connected_slaves:0
master_failover_state:no-failover
master_replid:bc5312e5e73f9a533a13a48659f2b6dfb1a28082
master_replid2:0000000000000000000000000000000000000000
master_repl_offset:0
second_repl_offset:-1
repl_backlog_active:0
repl_backlog_size:1048576
repl_backlog_first_byte_offset:0
repl_backlog_histlen:0
`
