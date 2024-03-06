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

	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/getdatassh"
)

func TestInput_checkConf(t *testing.T) {
	tests := []struct {
		name    string
		ipt     *Input
		wantErr bool
	}{
		{
			name: "all-nil",
			ipt: &Input{
				BinPath:    "",
				tagger:     &mockTagger{},
				mergedTags: make(map[string]urlTags),
			},
			wantErr: true,
		},
		{
			name: "have-bin-path",
			ipt: &Input{
				BinPath:    "chronyc",
				tagger:     &mockTagger{},
				mergedTags: make(map[string]urlTags),
			},
			wantErr: false,
		},
		{
			name: "error-url",
			ipt: &Input{
				BinPath: "",
				SSHServers: getdatassh.SSHServers{
					RemoteAddrs:     []string{"abcd"},
					RemoteUsers:     []string{"abcd"},
					RemotePasswords: []string{"abcd"},
					RemoteCommand:   "chronyc",
				},
				tagger:     &mockTagger{},
				mergedTags: make(map[string]urlTags),
			},
			wantErr: false,
		},
		{
			name: "ip-url",
			ipt: &Input{
				BinPath: "",
				SSHServers: getdatassh.SSHServers{
					RemoteAddrs:     []string{"127.0.0.1"},
					RemoteUsers:     []string{"abcd"},
					RemotePasswords: []string{"abcd"},
					RemoteCommand:   "chronyc",
				},
				tagger:     &mockTagger{},
				mergedTags: make(map[string]urlTags),
			},
			wantErr: false,
		},
		{
			name: "with-port",
			ipt: &Input{
				BinPath: "",
				SSHServers: getdatassh.SSHServers{
					RemoteAddrs:     []string{"127.0.0.1:22"},
					RemoteUsers:     []string{"abcd"},
					RemotePasswords: []string{"abcd"},
					RemoteCommand:   "chronyc",
				},
				tagger:     &mockTagger{},
				mergedTags: make(map[string]urlTags),
			},
			wantErr: false,
		},
		{
			name: "http-ip-url",
			ipt: &Input{
				BinPath: "",
				SSHServers: getdatassh.SSHServers{
					RemoteAddrs:     []string{"http://127.0.0.1"},
					RemoteUsers:     []string{"abcd"},
					RemotePasswords: []string{"abcd"},
					RemoteCommand:   "chronyc",
				},
				tagger:     &mockTagger{},
				mergedTags: make(map[string]urlTags),
			},
			wantErr: false,
		},
		{
			name: "domain",
			ipt: &Input{
				BinPath: "",
				SSHServers: getdatassh.SSHServers{
					RemoteAddrs:     []string{"www.baidu.com"},
					RemoteUsers:     []string{"abcd"},
					RemotePasswords: []string{"abcd"},
					RemoteCommand:   "chronyc",
				},
				tagger:     &mockTagger{},
				mergedTags: make(map[string]urlTags),
			},
			wantErr: false,
		},
		{
			name: "rsa-path",
			ipt: &Input{
				BinPath: "",
				SSHServers: getdatassh.SSHServers{
					RemoteAddrs:    []string{"www.baidu.com"},
					RemoteRsaPaths: []string{"/home/your_name/.ssh/id_rsa"},
					RemoteCommand:  "chronyc",
				},
				tagger:     &mockTagger{},
				mergedTags: make(map[string]urlTags),
			},
			wantErr: false,
		},
		{
			name: "no-remote-command",
			ipt: &Input{
				BinPath: "",
				SSHServers: getdatassh.SSHServers{
					RemoteAddrs:    []string{"www.baidu.com"},
					RemoteRsaPaths: []string{"/home/your_name/.ssh/id_rsa"},
				},
				tagger:     &mockTagger{},
				mergedTags: make(map[string]urlTags),
			},
			wantErr: true,
		},
		{
			name: "both-no-rsa-users",
			ipt: &Input{
				BinPath: "",
				SSHServers: getdatassh.SSHServers{
					RemoteAddrs:   []string{"www.baidu.com"},
					RemoteCommand: "chronyc",
				},
				tagger:     &mockTagger{},
				mergedTags: make(map[string]urlTags),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipt := tt.ipt

			if err := ipt.checkConf(); (err != nil) != tt.wantErr {
				t.Errorf("Input.checkConf() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInput_getPts(t *testing.T) {
	type args struct {
		data []*getdatassh.SSHData
	}
	tests := []struct {
		name    string
		ipt     *Input
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "election-and-remote",
			ipt: &Input{
				SSHServers: getdatassh.SSHServers{
					RemoteAddrs:     []string{"192.168.1.1:22"},
					RemoteUsers:     []string{"remote_login_name"},
					RemotePasswords: []string{"remote_login_password"},
					RemoteCommand:   "chronyc -n tracking",
				},
				Election:   true,
				tagger:     &mockTagger{},
				mergedTags: make(map[string]urlTags),
			},
			args: args{mockDataRemote()},
			want: []string{
				`chrony,host=192.168.1.1,leap_status=normal,reference_id=CA760182,stratum=2 frequency=-1.452,last_offset=-0.00029172,residual_freq=-0.094,rms_offset=0.00476266,root_delay=0.04132754,root_dispersion=0.003143095,skew=4.524,system_time=-0,update_interval=65.3`,
			},
			wantErr: false,
		},
		{
			name: "un-election-and-remote",
			ipt: &Input{
				SSHServers: getdatassh.SSHServers{
					RemoteAddrs:     []string{"192.168.1.1:22"},
					RemoteUsers:     []string{"remote_login_name"},
					RemotePasswords: []string{"remote_login_password"},
					RemoteCommand:   "chronyc -n tracking",
				},
				Election:   false,
				tagger:     &mockTagger{},
				mergedTags: make(map[string]urlTags),
			},
			args: args{mockDataRemote()},
			want: []string{
				`chrony,host=192.168.1.1,leap_status=normal,reference_id=CA760182,stratum=2 frequency=-1.452,last_offset=-0.00029172,residual_freq=-0.094,rms_offset=0.00476266,root_delay=0.04132754,root_dispersion=0.003143095,skew=4.524,system_time=-0,update_interval=65.3`,
			},
			wantErr: false,
		},
		{
			name: "election-and-local",
			ipt: &Input{
				SSHServers: getdatassh.SSHServers{
					RemoteAddrs:     []string{},
					RemoteUsers:     []string{},
					RemotePasswords: []string{},
					RemoteCommand:   "",
				},
				BinPath:    "chronyc",
				Election:   true,
				tagger:     &mockTagger{},
				mergedTags: make(map[string]urlTags),
			},
			args: args{mockDataLocal()},
			want: []string{
				`chrony,leap_status=normal,reference_id=CA760182,stratum=2 frequency=-1.452,last_offset=-0.00029172,residual_freq=-0.094,rms_offset=0.00476266,root_delay=0.04132754,root_dispersion=0.003143095,skew=4.524,system_time=-0,update_interval=65.3`,
			},
			wantErr: false,
		},
		{
			name: "election-and-local-extra-tag",
			ipt: &Input{
				SSHServers: getdatassh.SSHServers{
					RemoteAddrs:     []string{},
					RemoteUsers:     []string{},
					RemotePasswords: []string{},
					RemoteCommand:   "",
				},
				BinPath:  "chronyc",
				Election: true,
				Tags: map[string]string{
					"some_tag":  "some_value",
					"some_tag2": "some_value2",
				},
				tagger:     &mockTagger{},
				mergedTags: make(map[string]urlTags),
			},
			args: args{mockDataLocal()},
			want: []string{
				`chrony,leap_status=normal,reference_id=CA760182,some_tag=some_value,some_tag2=some_value2,stratum=2 frequency=-1.452,last_offset=-0.00029172,residual_freq=-0.094,rms_offset=0.00476266,root_delay=0.04132754,root_dispersion=0.003143095,skew=4.524,system_time=-0,update_interval=65.3`,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipt := tt.ipt
			err := ipt.setup()
			assert.NoError(t, err)

			ipt.collectCache = make([]*point.Point, 0)
			err = ipt.getPts(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Input.getPts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(ipt.collectCache) != len(tt.want) {
				t.Errorf("got %d points, want %d points,", len(ipt.collectCache), len(tt.want))
			}

			var got []string
			for _, p := range ipt.collectCache {
				s := p.LineProto()
				// remove timestamp
				s = s[:strings.LastIndex(s, " ")]
				got = append(got, s)
			}
			sort.Strings(got)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Input.getPts() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInput_getPts_with_host(t *testing.T) {
	type args struct {
		data []*getdatassh.SSHData
	}
	tests := []struct {
		name    string
		ipt     *Input
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "no-election-and-local",
			ipt: &Input{
				SSHServers: getdatassh.SSHServers{
					RemoteAddrs:     []string{},
					RemoteUsers:     []string{},
					RemotePasswords: []string{},
					RemoteCommand:   "",
				},
				BinPath:    "chronyc",
				Election:   false,
				tagger:     &mockTagger{},
				mergedTags: make(map[string]urlTags),
			},
			args: args{mockDataLocal()},
			want: []string{
				`chrony,host=me,leap_status=normal,reference_id=CA760182,stratum=2 frequency=-1.452,last_offset=-0.00029172,residual_freq=-0.094,rms_offset=0.00476266,root_delay=0.04132754,root_dispersion=0.003143095,skew=4.524,system_time=-0,update_interval=65.3`,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipt := tt.ipt
			err := ipt.setup()
			assert.NoError(t, err)

			ipt.collectCache = make([]*point.Point, 0)
			err = ipt.getPts(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Input.getPts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(ipt.collectCache) != len(tt.want) {
				t.Errorf("got %d points, want %d points,", len(ipt.collectCache), len(tt.want))
			}

			var got []string
			for _, p := range ipt.collectCache {
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

func mockDataRemote() []*getdatassh.SSHData {
	mockData := make([]*getdatassh.SSHData, 0)
	mockData = append(mockData, &getdatassh.SSHData{
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

func mockDataLocal() []*getdatassh.SSHData {
	mockData := make([]*getdatassh.SSHData, 0)
	mockData = append(mockData, &getdatassh.SSHData{
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

type mockTagger struct{}

func (t *mockTagger) HostTags() map[string]string {
	return map[string]string{
		"host": "me",
	}
}

func (t *mockTagger) ElectionTags() map[string]string {
	return nil
}

func (g *mockTagger) UpdateVersion() {}

func (g *mockTagger) Updated() bool { return false }
