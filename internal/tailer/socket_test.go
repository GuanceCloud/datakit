// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package tailer

import (
	"reflect"
	"testing"
)

func Test_spiltBuffer(t *testing.T) {
	type args struct {
		fromCache string
		date      string
		full      bool
	}
	tests := []struct {
		name          string
		args          args
		wantPipdata   []string
		wantCacheDate string
	}{
		// 0055-rc.local-exist update to monitor\n0055-rc.local-exist update to monitor
		{
			name: "case01", args: args{
				fromCache: "",
				date: `0055-rc.local-exist update to monitor
0055-rc.local-exist update to`, full: true,
			},
			wantCacheDate: "0055-rc.local-exist update to",
			wantPipdata:   []string{"0055-rc.local-exist update to monitor"},
		},

		{
			name: "case02", args: args{
				fromCache: "0055-rc",
				date: `.local-exist update to monitor
0055-rc.local-exist update to`, full: true,
			},
			wantCacheDate: "0055-rc.local-exist update to",
			wantPipdata:   []string{"0055-rc.local-exist update to monitor"},
		},

		{
			name: "case03", args: args{
				fromCache: "",
				date: `2021-12-22T14:12:42 ERROR internal.lua luafuncs/monitor.go:297  0055update to mon
0055-rc.local-exist update to
`,
				full: false,
			},
			wantCacheDate: "",
			wantPipdata:   []string{"2021-12-22T14:12:42 ERROR internal.lua luafuncs/monitor.go:297  0055update to mon", "0055-rc.local-exist update to", ""},
		},
		{
			name: "case04", args: args{
				fromCache: "",
				date:      `2021-12-22T14:12:42 ERROR internal.lua luafuncs/monitor.go:297  0055update to mon`,
				full:      false,
			},
			wantCacheDate: "",
			wantPipdata:   []string{"2021-12-22T14:12:42 ERROR internal.lua luafuncs/monitor.go:297  0055update to mon"},
		},
	}
	sl := &socketLogger{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPipdata, gotCacheDate := sl.spiltBuffer(tt.args.fromCache, tt.args.date, tt.args.full)
			if !reflect.DeepEqual(gotPipdata, tt.wantPipdata) {
				t.Errorf("gotPipdata len=%d want len=%d", len(gotPipdata), len(tt.wantPipdata))
				t.Errorf("spiltBuffer() gotPipdata = %v, want %v", gotPipdata, tt.wantPipdata)
			}
			if gotCacheDate != tt.wantCacheDate {
				t.Errorf("spiltBuffer() gotCacheDate = %v, want %v", gotCacheDate, tt.wantCacheDate)
			}
		})
	}
}

func Test_mkServer(t *testing.T) {
	type args struct {
		socket string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// we use random port(:0) here, see https://stackoverflow.com/a/43425461/342348
		{
			name:    "case1",
			args:    args{socket: "tcp://127.0.0.1:0"},
			wantErr: false,
		},
		{
			name:    "case2",
			args:    args{socket: "udp://127.0.0.1:0"}, // tcp 和 udp 可以使用同一端口
			wantErr: false,
		},
		{
			name:    "case4",
			args:    args{socket: "udp1://127.0.0.1:0"}, // err socket
			wantErr: true,
		},
		{
			name:    "case5",
			args:    args{socket: "udp127.0.0.1:0"}, // err socket
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotS, err := mkServer(tt.args.socket)
			if (err != nil) != tt.wantErr {
				t.Errorf("case:%s mkServer() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}

			if gotS != nil {
				if gotS.lis != nil {
					t.Logf("TCP addr: %s", gotS.lis.Addr().String())
				}

				if gotS.conn != nil {
					t.Logf("UDP addr: %+#v", gotS.conn.LocalAddr().String())
				}
			}
		})
	}
}
