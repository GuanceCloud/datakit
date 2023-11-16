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

func TestInput_parseCommandData(t *testing.T) {
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
				list: mockCommandData01,
			},
			want: []string{
				"redis_command_stat,foo=bar,host=HOST,method=cmdstat_client|list calls=1,failed_calls=0,rejected_calls=0,usec=25,usec_per_call=25",
				"redis_command_stat,foo=bar,host=HOST,method=cmdstat_cluster|info calls=2,failed_calls=0,rejected_calls=0,usec=93,usec_per_call=46.5",
				"redis_command_stat,foo=bar,host=HOST,method=cmdstat_command|docs calls=2,failed_calls=0,rejected_calls=0,usec=4112,usec_per_call=2056",
				"redis_command_stat,foo=bar,host=HOST,method=cmdstat_info calls=5,failed_calls=0,rejected_calls=0,usec=378,usec_per_call=75.6",
				"redis_command_stat,foo=bar,host=HOST,method=cmdstat_ping calls=1,failed_calls=0,rejected_calls=0,usec=6,usec_per_call=6",
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
				list: mockCommandData01,
			},
			want: []string{
				"redis_command_stat,election=TRUE,foo=bar,method=cmdstat_client|list calls=1,failed_calls=0,rejected_calls=0,usec=25,usec_per_call=25",
				"redis_command_stat,election=TRUE,foo=bar,method=cmdstat_cluster|info calls=2,failed_calls=0,rejected_calls=0,usec=93,usec_per_call=46.5",
				"redis_command_stat,election=TRUE,foo=bar,method=cmdstat_command|docs calls=2,failed_calls=0,rejected_calls=0,usec=4112,usec_per_call=2056",
				"redis_command_stat,election=TRUE,foo=bar,method=cmdstat_info calls=5,failed_calls=0,rejected_calls=0,usec=378,usec_per_call=75.6",
				"redis_command_stat,election=TRUE,foo=bar,method=cmdstat_ping calls=1,failed_calls=0,rejected_calls=0,usec=6,usec_per_call=6",
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

			got, err := ipt.parseCommandData(tt.args.list)
			if (err != nil) != tt.wantErr {
				t.Errorf("Input.parseCommandData() error = %v, wantErr %v", err, tt.wantErr)
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

var mockCommandData01 = `cmdstat_client|list:calls=1,usec=25,usec_per_call=25.00,rejected_calls=0,failed_calls=0
cmdstat_cluster|info:calls=2,usec=93,usec_per_call=46.50,rejected_calls=0,failed_calls=0
cmdstat_info:calls=5,usec=378,usec_per_call=75.60,rejected_calls=0,failed_calls=0
cmdstat_ping:calls=1,usec=6,usec_per_call=6.00,rejected_calls=0,failed_calls=0
cmdstat_command|docs:calls=2,usec=4112,usec_per_call=2056.00,rejected_calls=0,failed_calls=0
`
