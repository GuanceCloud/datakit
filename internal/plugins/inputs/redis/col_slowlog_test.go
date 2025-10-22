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

func Test_parseSlowData(t *testing.T) {
	type fields struct {
		host          string
		tags          map[string]string
		tagger        datakit.GlobalTagger
		slowlogMaxLen int
	}
	type args struct {
		slowlogs any
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "ok-no-selection",
			fields: fields{
				host:          "localhost",
				slowlogMaxLen: 128,
				tags:          map[string]string{"foo": "bar"},
				tagger:        testutils.DefaultMockTagger(),
			},
			args: args{
				slowlogs: mockSlowLogs01(),
			},
			want: []string{
				// not host=HOST, because have host=localhost
				`redis_slowlog,foo=bar,host=localhost,server=localhost:6379 client_addr="127.0.0.1:33294",client_name="",command="debug sleep .25",message="debug sleep .25 cost time 250092us",slowlog_95percentile=250092,slowlog_avg=131128.5,slowlog_id=3i,slowlog_max=250092i,slowlog_median=131128i,slowlog_micros=250092i,status="WARNING"`,
				`redis_slowlog,foo=bar,host=localhost,server=localhost:6379 client_addr="127.0.0.1:33294",client_name="",command="debug sleep .25",message="debug sleep .25 cost time 250122us",slowlog_95percentile=1000091,slowlog_avg=378117.5,slowlog_id=1i,slowlog_max=1000091i,slowlog_median=250107i,slowlog_micros=250122i,status="WARNING"`,
				`redis_slowlog,foo=bar,host=localhost,server=localhost:6379 client_addr="127.0.0.1:33294",client_name="",command="debug sleep 1",message="debug sleep 1 cost time 1000087us",slowlog_95percentile=1000091,slowlog_avg=502511.4,slowlog_id=0i,slowlog_max=1000091i,slowlog_median=250122i,slowlog_micros=1000087i,status="WARNING"`,
				`redis_slowlog,foo=bar,host=localhost,server=localhost:6379 client_addr="127.0.0.1:33294",client_name="",command="debug sleep 1",message="debug sleep 1 cost time 1000091us",slowlog_95percentile=1000091,slowlog_avg=420782.6666666667,slowlog_id=2i,slowlog_max=1000091i,slowlog_median=250092i,slowlog_micros=1000091i,status="WARNING"`,
				`redis_slowlog,foo=bar,host=localhost,server=localhost:6379 client_addr="172.17.0.1:60188",client_name="",command="info ALL",message="info ALL cost time 12165us",slowlog_95percentile=12165,slowlog_avg=12165,slowlog_id=4i,slowlog_max=12165i,slowlog_median=12165i,slowlog_micros=12165i,status="WARNING"`,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipt := defaultInput()
			ipt.Tags = tt.fields.tags

			ipt.setup()
			ipt.startUpUnix = 0

			inst := newInstance()
			inst.ipt = ipt
			inst.host = tt.fields.host
			inst.addr = tt.fields.host + ":6379"
			inst.setup()

			got, err := inst.parseSlowData(tt.args.slowlogs)
			if (err != nil) != tt.wantErr {
				t.Errorf("Input.parseSlowData() error = %v, wantErr %v", err, tt.wantErr)
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

type slowLog struct {
	id        int64
	startTime int64
	duration  int64
	command   []string
	addr      string
	other     string
}

func mockSlowLogs01() any {
	mockData := []slowLog{
		{
			id:        4,
			startTime: 1699238829,
			duration:  12165,
			command:   []string{"info", "ALL"},
			addr:      "172.17.0.1:60188",
			other:     "",
		},
		{
			id:        3,
			startTime: 1699236425,
			duration:  250092,
			command:   []string{"debug", "sleep", ".25"},
			addr:      "127.0.0.1:33294",
			other:     "",
		},
		{
			id:        2,
			startTime: 1699236422,
			duration:  1000091,
			command:   []string{"debug", "sleep", "1"},
			addr:      "127.0.0.1:33294",
			other:     "",
		},
		{
			id:        1,
			startTime: 1699236401,
			duration:  250122,
			command:   []string{"debug", "sleep", ".25"},
			addr:      "127.0.0.1:33294",
			other:     "",
		},
		{
			id:        0,
			startTime: 1699236394,
			duration:  1000087,
			command:   []string{"debug", "sleep", "1"},
			addr:      "127.0.0.1:33294",
			other:     "",
		},
	}

	slowlogs := []any{}
	for _, data := range mockData {
		temp := []any{}
		temp = append(temp, data.id)
		temp = append(temp, data.startTime)
		temp = append(temp, data.duration)

		subTemp := []any{}
		for _, v := range data.command {
			subTemp = append(subTemp, v)
		}
		temp = append(temp, subTemp)

		temp = append(temp, data.addr)
		temp = append(temp, data.other)

		slowlogs = append(slowlogs, temp)
	}

	return slowlogs
}
