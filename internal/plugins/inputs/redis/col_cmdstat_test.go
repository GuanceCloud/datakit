// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package redis

import (
	"sort"
	"strings"
	T "testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/testutils"
)

func TestInput_parseCommandData(t *T.T) {
	mockCommandData01 := `cmdstat_client|list:calls=1,usec=25,usec_per_call=25.00,rejected_calls=0,failed_calls=0
cmdstat_cluster|info:calls=2,usec=93,usec_per_call=46.50,rejected_calls=0,failed_calls=0
cmdstat_info:calls=5,usec=378,usec_per_call=75.60,rejected_calls=0,failed_calls=0
cmdstat_ping:calls=1,usec=6,usec_per_call=6.00,rejected_calls=0,failed_calls=0
cmdstat_command|docs:calls=2,usec=4112,usec_per_call=2056.00,rejected_calls=0,failed_calls=0
`

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
		name   string
		fields fields
		args   args
		want   []string
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
				"redis_command_stat,foo=bar,host=localhost,method=client|list calls=1,failed_calls=0,rejected_calls=0,usec=25,usec_per_call=25",
				"redis_command_stat,foo=bar,host=localhost,method=cluster|info calls=2,failed_calls=0,rejected_calls=0,usec=93,usec_per_call=46.5",
				"redis_command_stat,foo=bar,host=localhost,method=command|docs calls=2,failed_calls=0,rejected_calls=0,usec=4112,usec_per_call=2056",
				"redis_command_stat,foo=bar,host=localhost,method=info calls=5,failed_calls=0,rejected_calls=0,usec=378,usec_per_call=75.6",
				"redis_command_stat,foo=bar,host=localhost,method=ping calls=1,failed_calls=0,rejected_calls=0,usec=6,usec_per_call=6",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *T.T) {
			ipt := defaultInput()
			ipt.Tags = tt.fields.Tags

			inst := newInstance()
			inst.ipt = ipt
			inst.host = tt.fields.Host

			inst.setup()

			got, err := inst.parseCommandData(tt.args.list)
			require.NoError(t, err)

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

	t.Run("ignore-bad-command", func(t *T.T) {
		badCommand := `# invalid value
cmdstat_client|list:calls=1,usec=25,usec_per_call=25.00,rejected_calls=0,failed_calls=,invalid=1=2

# no command details: ignored
cmdstat_client|list

# point no fields: ignored
cmdstat_client|list:calls=

# key got no value
cmdstat_client|list:calls=1,usec=25,usec_per_call=25.00,rejected_calls,failed_calls=0`

		ipt := defaultInput()
		inst := newInstance()
		inst.ipt = ipt
		inst.host = "HOST"

		inst.setup()

		pts, err := inst.parseCommandData(badCommand)
		for _, pt := range pts {
			pt.SetTime(time.Unix(0, 123))
		}

		require.NoError(t, err)
		assert.Len(t, pts, 2)

		for _, pt := range pts {
			t.Logf("%s", pt.Pretty())
		}

		assert.Equal(
			t,
			"redis_command_stat,host=HOST,method=client|list calls=1,rejected_calls=0,usec=25,usec_per_call=25 123",
			pts[0].LineProto(),
		)

		assert.Equal(
			t,
			"redis_command_stat,host=HOST,method=client|list calls=1,failed_calls=0,usec=25,usec_per_call=25 123",
			pts[1].LineProto(),
		)
	})
}
