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
	mockDBData01 := "db0:keys=43706,expires=117,avg_ttl=30904274304765"
	type fields struct {
		host     string
		tags     map[string]string
		election bool
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
			name: "basic",
			fields: fields{
				host:     "localhost",
				tags:     map[string]string{"foo": "bar"},
				election: false,
				tagger:   testutils.DefaultMockTagger(),
			},
			args: args{
				list: mockDBData01,
			},
			want: []string{
				"redis_db,db_name=db0,foo=bar,host=localhost avg_ttl=30904274304765,expires=117,keys=43706",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipt := defaultInput()
			ipt.Tags = tt.fields.tags

			inst := newInstance()
			inst.host = tt.fields.host
			inst.ipt = ipt
			inst.setup()

			got, err := inst.parseDBData(tt.args.list)
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

			assert.Equal(t, tt.want, gotStr)
		})
	}
}
