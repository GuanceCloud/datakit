// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2022-present Guance, Inc.

// Package chrony collects chrony metrics.
package chrony

import (
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
)

func TestInput_checkConf(t *testing.T) {
	type fields struct {
		BinPath      string
		Interval     time.Duration
		Timeout      time.Duration
		SSHServers   datakit.SSHServers
		Tags         map[string]string
		Election     bool
		collectCache []*point.Point
		platform     string
		feeder       io.Feeder
		semStop      *cliutils.Sem
		pause        bool
		pauseCh      chan bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "all-nil",
			fields: fields{
				BinPath: "",
			},
			wantErr: true,
		},
		{
			name: "have-bin-path",
			fields: fields{
				BinPath: "chronyc",
			},
			wantErr: false,
		},
		{
			name: "error-url",
			fields: fields{
				BinPath: "",
				SSHServers: datakit.SSHServers{
					RemoteAddrs:     []string{"abcd"},
					RemoteUsers:     []string{"abcd"},
					RemotePasswords: []string{"abcd"},
					RemoteCommand:   "chronyc",
				},
			},
			wantErr: false,
		},
		{
			name: "ip-url",
			fields: fields{
				BinPath: "",
				SSHServers: datakit.SSHServers{
					RemoteAddrs:     []string{"127.0.0.1"},
					RemoteUsers:     []string{"abcd"},
					RemotePasswords: []string{"abcd"},
					RemoteCommand:   "chronyc",
				},
			},
			wantErr: false,
		},
		{
			name: "with-port",
			fields: fields{
				BinPath: "",
				SSHServers: datakit.SSHServers{
					RemoteAddrs:     []string{"127.0.0.1:22"},
					RemoteUsers:     []string{"abcd"},
					RemotePasswords: []string{"abcd"},
					RemoteCommand:   "chronyc",
				},
			},
			wantErr: false,
		},
		{
			name: "http-ip-url",
			fields: fields{
				BinPath: "",
				SSHServers: datakit.SSHServers{
					RemoteAddrs:     []string{"http://127.0.0.1"},
					RemoteUsers:     []string{"abcd"},
					RemotePasswords: []string{"abcd"},
					RemoteCommand:   "chronyc",
				},
			},
			wantErr: false,
		},
		{
			name: "domain",
			fields: fields{
				BinPath: "",
				SSHServers: datakit.SSHServers{
					RemoteAddrs:     []string{"www.baidu.com"},
					RemoteUsers:     []string{"abcd"},
					RemotePasswords: []string{"abcd"},
					RemoteCommand:   "chronyc",
				},
			},
			wantErr: false,
		},
		{
			name: "rsa-path",
			fields: fields{
				BinPath: "",
				SSHServers: datakit.SSHServers{
					RemoteAddrs:    []string{"www.baidu.com"},
					RemoteRsaPaths: []string{"/home/your_name/.ssh/id_rsa"},
					RemoteCommand:  "chronyc",
				},
			},
			wantErr: false,
		},
		{
			name: "no-remote-command",
			fields: fields{
				BinPath: "",
				SSHServers: datakit.SSHServers{
					RemoteAddrs:    []string{"www.baidu.com"},
					RemoteRsaPaths: []string{"/home/your_name/.ssh/id_rsa"},
				},
			},
			wantErr: true,
		},
		{
			name: "both-no-rsa-users",
			fields: fields{
				BinPath: "",
				SSHServers: datakit.SSHServers{
					RemoteAddrs:   []string{"www.baidu.com"},
					RemoteCommand: "chronyc",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipt := &Input{
				BinPath:      tt.fields.BinPath,
				Interval:     tt.fields.Interval,
				Timeout:      tt.fields.Timeout,
				SSHServers:   tt.fields.SSHServers,
				Tags:         tt.fields.Tags,
				Election:     tt.fields.Election,
				collectCache: tt.fields.collectCache,
				platform:     tt.fields.platform,
				feeder:       tt.fields.feeder,
				semStop:      tt.fields.semStop,
				pause:        tt.fields.pause,
				pauseCh:      tt.fields.pauseCh,
			}
			if err := ipt.checkConf(); (err != nil) != tt.wantErr {
				t.Errorf("Input.checkConf() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInput_getPts(t *testing.T) {
	type fields struct {
		Interval     time.Duration
		Timeout      time.Duration
		BinPath      string
		SSHServers   datakit.SSHServers
		Tags         map[string]string
		Election     bool
		collectCache []*point.Point
		platform     string
		feeder       io.Feeder
		semStop      *cliutils.Sem
		pause        bool
		pauseCh      chan bool
	}
	type args struct {
		data []datakit.SSHData
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "election-and-remote",
			fields: fields{
				Election: true,
			},
			args: args{mockDataRemote()},
			want: []string{
				`chrony,host=192.168.1.1:22,leap_status=normal,reference_id=CA760182,stratum=2 frequency=-1.452,last_offset=-0.00029172,residual_freq=-0.094,rms_offset=0.00476266,root_delay=0.04132754,root_dispersion=0.003143095,skew=4.524,system_time=-0,update_interval=65.3`,
			},
			wantErr: false,
		},
		{
			name: "un-election-and-remote",
			fields: fields{
				Election: false,
			},
			args: args{mockDataRemote()},
			want: []string{
				`chrony,host=192.168.1.1:22,leap_status=normal,reference_id=CA760182,stratum=2 frequency=-1.452,last_offset=-0.00029172,residual_freq=-0.094,rms_offset=0.00476266,root_delay=0.04132754,root_dispersion=0.003143095,skew=4.524,system_time=-0,update_interval=65.3`,
			},
			wantErr: false,
		},
		{
			name: "election-and-local",
			fields: fields{
				Election: true,
			},
			args: args{mockDataLocal()},
			want: []string{
				`chrony,leap_status=normal,reference_id=CA760182,stratum=2 frequency=-1.452,last_offset=-0.00029172,residual_freq=-0.094,rms_offset=0.00476266,root_delay=0.04132754,root_dispersion=0.003143095,skew=4.524,system_time=-0,update_interval=65.3`,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipt := &Input{
				Interval:     tt.fields.Interval,
				Timeout:      tt.fields.Timeout,
				BinPath:      tt.fields.BinPath,
				SSHServers:   tt.fields.SSHServers,
				Tags:         tt.fields.Tags,
				Election:     tt.fields.Election,
				collectCache: tt.fields.collectCache,
				platform:     tt.fields.platform,
				feeder:       tt.fields.feeder,
				semStop:      tt.fields.semStop,
				pause:        tt.fields.pause,
				pauseCh:      tt.fields.pauseCh,
			}
			points, err := ipt.getPts(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Input.getPts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(points) != len(tt.want) {
				t.Errorf("got %d points, want %d points,", len(points), len(tt.want))
			}

			var got []string
			for _, p := range points {
				s := p.LineProto()
				// remove timestamp
				s = s[:strings.LastIndex(s, " ")]
				got = append(got, s)
			}
			sort.Strings(got)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Input.getPts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func mockDataRemote() []datakit.SSHData {
	mockData := make([]datakit.SSHData, 0)
	mockData = append(mockData, datakit.SSHData{
		Server: "192.168.1.1:22",
		Data: []byte(`Reference ID    : CA760182 (202.118.1.130)
Stratum         : 2
Ref time (UTC)  : Wed Jun 07 06:22:16 2023
System time     : 0.000000000 seconds slow of NTP time
Last offset     : -0.000291720 seconds
RMS offset      : 0.004762660 seconds
Frequency       : 1.452 ppm slow
Residual freq   : -0.094 ppm
Skew            : 4.524 ppm
Root delay      : 0.041327540 seconds
Root dispersion : 0.003143095 seconds
Update interval : 65.3 seconds
Leap status     : Normal
`),
	})

	return mockData
}

func mockDataLocal() []datakit.SSHData {
	mockData := make([]datakit.SSHData, 0)
	mockData = append(mockData, datakit.SSHData{
		Server: "localhost",
		Data: []byte(`Reference ID    : CA760182 (202.118.1.130)
Stratum         : 2
Ref time (UTC)  : Wed Jun 07 06:22:16 2023
System time     : 0.000000000 seconds slow of NTP time
Last offset     : -0.000291720 seconds
RMS offset      : 0.004762660 seconds
Frequency       : 1.452 ppm slow
Residual freq   : -0.094 ppm
Skew            : 4.524 ppm
Root delay      : 0.041327540 seconds
Root dispersion : 0.003143095 seconds
Update interval : 65.3 seconds
Leap status     : Normal
`),
	})

	return mockData
}
