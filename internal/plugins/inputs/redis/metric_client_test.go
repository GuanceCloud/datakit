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

func TestInput_parseClientData(t *testing.T) {
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
			name: "no election",
			fields: fields{
				Host:     "localhost",
				Tags:     map[string]string{"foo": "bar"},
				Election: false,
				tagger:   testutils.DefaultMockTagger(),
			},
			args: args{
				list: mockClient01(),
			},
			want: []string{
				"redis_client,addr=127.0.0.1,foo=bar,host=HOST,name=unknown age=26441,argv_mem=0,db=0,fd=12,idle=24342,multi=-1,multi_mem=0,obl=0,oll=0,omem=0,psub=0,qbuf=0,qbuf_free=0,redir=-1,resp=2,ssub=0,sub=0,tot_mem=1800",
				"redis_client,addr=172.17.0.1,foo=bar,host=HOST,name=unknown age=253,argv_mem=10,db=0,fd=13,idle=0,multi=-1,multi_mem=0,obl=0,oll=0,omem=0,psub=0,qbuf=26,qbuf_free=20448,redir=-1,resp=2,ssub=0,sub=0,tot_mem=22298",
			},
			wantErr: false,
		},
		{
			name: "election",
			fields: fields{
				Host:     "localhost",
				Tags:     map[string]string{"foo": "bar"},
				Election: true,
				tagger:   testutils.DefaultMockTagger(),
			},
			args: args{
				list: mockClient01(),
			},
			want: []string{
				"redis_client,addr=127.0.0.1,election=TRUE,foo=bar,name=unknown age=26441,argv_mem=0,db=0,fd=12,idle=24342,multi=-1,multi_mem=0,obl=0,oll=0,omem=0,psub=0,qbuf=0,qbuf_free=0,redir=-1,resp=2,ssub=0,sub=0,tot_mem=1800",
				"redis_client,addr=172.17.0.1,election=TRUE,foo=bar,name=unknown age=253,argv_mem=10,db=0,fd=13,idle=0,multi=-1,multi_mem=0,obl=0,oll=0,omem=0,psub=0,qbuf=26,qbuf_free=20448,redir=-1,resp=2,ssub=0,sub=0,tot_mem=22298",
			},
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

			got, err := ipt.parseClientData(tt.args.list)
			if (err != nil) != tt.wantErr {
				t.Errorf("Input.parseClientData() error = %v, wantErr %v", err, tt.wantErr)
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

func mockClient01() string {
	return `id=4 addr=127.0.0.1:33294 laddr=127.0.0.1:6379 fd=12 name= age=26441 idle=24342 flags=N db=0 sub=0 psub=0 ssub=0 multi=-1 qbuf=0 qbuf-free=0 argv-mem=0 multi-mem=0 rbs=1024 rbp=0 obl=0 oll=0 omem=0 tot-mem=1800 events=r cmd=slowlog|get user=default redir=-1 resp=2
	id=16 addr=172.17.0.1:41942 laddr=172.17.0.2:6379 fd=13 name= age=253 idle=0 flags=N db=0 sub=0 psub=0 ssub=0 multi=-1 qbuf=26 qbuf-free=20448 argv-mem=10 multi-mem=0 rbs=1024 rbp=1024 obl=0 oll=0 omem=0 tot-mem=22298 events=r cmd=client|list user=default redir=-1 resp=2`
}
