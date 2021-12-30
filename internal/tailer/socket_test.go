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
`, full: false,
			},
			wantCacheDate: "",
			wantPipdata:   []string{"2021-12-22T14:12:42 ERROR internal.lua luafuncs/monitor.go:297  0055update to mon", "0055-rc.local-exist update to"},
		},
	}
	sl := &socketLogger{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPipdata, gotCacheDate := sl.spiltBuffer(tt.args.fromCache, tt.args.date, tt.args.full)
			if !reflect.DeepEqual(gotPipdata, tt.wantPipdata) {
				t.Errorf("gotPipdata len=%d want len=%d", len(gotPipdata), len(tt.wantCacheDate))
				t.Errorf("spiltBuffer() gotPipdata = %v, want %v", gotPipdata, tt.wantPipdata)
			}
			if gotCacheDate != tt.wantCacheDate {
				t.Errorf("spiltBuffer() gotCacheDate = %v, want %v", gotCacheDate, tt.wantCacheDate)
			}
		})
	}
}
