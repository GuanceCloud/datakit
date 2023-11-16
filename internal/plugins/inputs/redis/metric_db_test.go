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

func TestInput_parseDBData(t *testing.T) {
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
				list: mockDBData01,
			},
			want: []string{
				"redis_db,db_name=db0,foo=bar,host=HOST avg_ttl=30904274304765,expires=117,keys=43706",
			},
			wantErr: false,
		},
		{
			name: "election localhost 01 ",
			fields: fields{
				Host:     "localhost",
				Tags:     map[string]string{"foo": "bar"},
				Election: true,
				tagger:   testutils.DefaultMockTagger(),
			},
			args: args{
				list: mockDBData01,
			},
			want: []string{
				"redis_db,db_name=db0,election=TRUE,foo=bar avg_ttl=30904274304765,expires=117,keys=43706",
			},
			wantErr: false,
		},

		{
			name: "election localhost 02",
			fields: fields{
				Host:     "127.0.0.1",
				Tags:     map[string]string{"foo": "bar"},
				Election: true,
				tagger:   testutils.DefaultMockTagger(),
			},
			args: args{
				list: mockDBData01,
			},
			want: []string{
				"redis_db,db_name=db0,election=TRUE,foo=bar avg_ttl=30904274304765,expires=117,keys=43706",
			},
			wantErr: false,
		},
		{
			name: "election not localhost",
			fields: fields{
				Host:     "172.2.0.1",
				Tags:     map[string]string{"foo": "bar"},
				Election: true,
				tagger:   testutils.DefaultMockTagger(),
			},
			args: args{
				list: mockDBData01,
			},
			want: []string{
				"redis_db,db_name=db0,election=TRUE,foo=bar,host=172.2.0.1 avg_ttl=30904274304765,expires=117,keys=43706",
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

			got, err := ipt.parseDBData(tt.args.list)
			if (err != nil) != tt.wantErr {
				t.Errorf("Input.parseDBData() error = %v, wantErr %v", err, tt.wantErr)
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

var mockDBData01 = "db0:keys=43706,expires=117,avg_ttl=30904274304765"
