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

func TestInput_parseSlowData(t *testing.T) {
	type fields struct {
		Host          string
		Tags          map[string]string
		Election      bool
		tagger        datakit.GlobalTagger
		SlowlogMaxLen int
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
			name: "ok no selection",
			fields: fields{
				Host:          "localhost",
				SlowlogMaxLen: 128,
				Tags:          map[string]string{"foo": "bar"},
				Election:      false,
				tagger:        testutils.DefaultMockTagger(),
			},
			args: args{
				slowlogs: mockSlowLogs01(),
			},
			want: []string{
				// not host=HOST, because have host=localhost
				`redis_slowlog,foo=bar,host=localhost,service=redis command="debug sleep .25",message="debug sleep .25 cost time 250092us",slowlog_id=3i,slowlog_micros=250092i,status="WARNING"`,
				`redis_slowlog,foo=bar,host=localhost,service=redis command="debug sleep .25",message="debug sleep .25 cost time 250122us",slowlog_id=1i,slowlog_micros=250122i,status="WARNING"`,
				`redis_slowlog,foo=bar,host=localhost,service=redis command="debug sleep 1",message="debug sleep 1 cost time 1000087us",slowlog_id=0i,slowlog_micros=1000087i,status="WARNING"`,
				`redis_slowlog,foo=bar,host=localhost,service=redis command="debug sleep 1",message="debug sleep 1 cost time 1000091us",slowlog_id=2i,slowlog_micros=1000091i,status="WARNING"`,
				`redis_slowlog,foo=bar,host=localhost,service=redis command="info ALL",message="info ALL cost time 12165us",slowlog_id=4i,slowlog_micros=12165i,status="WARNING"`,
			},
			wantErr: false,
		},
		{
			name: "ok selection",
			fields: fields{
				Host:          "localhost",
				SlowlogMaxLen: 128,
				Tags:          map[string]string{"foo": "bar"},
				Election:      true,
				tagger:        testutils.DefaultMockTagger(),
			},
			args: args{
				slowlogs: mockSlowLogs01(),
			},
			want: []string{
				`redis_slowlog,election=TRUE,foo=bar,host=localhost,service=redis command="debug sleep .25",message="debug sleep .25 cost time 250092us",slowlog_id=3i,slowlog_micros=250092i,status="WARNING"`,
				`redis_slowlog,election=TRUE,foo=bar,host=localhost,service=redis command="debug sleep .25",message="debug sleep .25 cost time 250122us",slowlog_id=1i,slowlog_micros=250122i,status="WARNING"`,
				`redis_slowlog,election=TRUE,foo=bar,host=localhost,service=redis command="debug sleep 1",message="debug sleep 1 cost time 1000087us",slowlog_id=0i,slowlog_micros=1000087i,status="WARNING"`,
				`redis_slowlog,election=TRUE,foo=bar,host=localhost,service=redis command="debug sleep 1",message="debug sleep 1 cost time 1000091us",slowlog_id=2i,slowlog_micros=1000091i,status="WARNING"`,
				`redis_slowlog,election=TRUE,foo=bar,host=localhost,service=redis command="info ALL",message="info ALL cost time 12165us",slowlog_id=4i,slowlog_micros=12165i,status="WARNING"`,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipt := &Input{
				Host:          tt.fields.Host,
				Tags:          tt.fields.Tags,
				Election:      tt.fields.Election,
				tagger:        tt.fields.tagger,
				SlowlogMaxLen: tt.fields.SlowlogMaxLen,
				hashMap:       make([][16]byte, tt.fields.SlowlogMaxLen),
			}

			ipt.setup()
			ipt.startUpUnix = 0

			got, err := ipt.parseSlowData(tt.args.slowlogs)
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

			assert.Equal(t, gotStr, tt.want)
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
		{id: 4, startTime: 1699238829, duration: 12165, command: []string{"info", "ALL"}, addr: "172.17.0.1:60188", other: ""},
		{id: 3, startTime: 1699236425, duration: 250092, command: []string{"debug", "sleep", ".25"}, addr: "127.0.0.1:33294", other: ""},
		{id: 2, startTime: 1699236422, duration: 1000091, command: []string{"debug", "sleep", "1"}, addr: "127.0.0.1:33294", other: ""},
		{id: 1, startTime: 1699236401, duration: 250122, command: []string{"debug", "sleep", ".25"}, addr: "127.0.0.1:33294", other: ""},
		{id: 0, startTime: 1699236394, duration: 1000087, command: []string{"debug", "sleep", "1"}, addr: "127.0.0.1:33294", other: ""},
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
