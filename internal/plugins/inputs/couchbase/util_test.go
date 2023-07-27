// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package couchbase

import "testing"

func Test_parseURL(t *testing.T) {
	tests := []struct {
		name    string
		ipt     *Input
		wantErr bool
	}{
		{
			name: "ok case",
			ipt: &Input{
				Scheme:         "http",
				Host:           "127.0.0.1",
				Port:           8091,
				AdditionalPort: 9102,
			},
			wantErr: false,
		},
		{
			name: "ok https & localhost case",
			ipt: &Input{
				Scheme:         "https",
				Host:           "localhost",
				Port:           18091,
				AdditionalPort: 19102,
			},
			wantErr: false,
		},
		{
			name: "ok domain",
			ipt: &Input{
				Scheme:         "http",
				Host:           "www.baidu.com",
				Port:           8091,
				AdditionalPort: 9102,
			},
			wantErr: false,
		},
		{
			name: "error scheme",
			ipt: &Input{
				Scheme:         "errorHTTP",
				Host:           "127.0.0.1",
				Port:           8091,
				AdditionalPort: 9102,
			},
			wantErr: true,
		},
		{
			name: "error port",
			ipt: &Input{
				Scheme:         "http",
				Host:           "127.0.0.1",
				Port:           88888,
				AdditionalPort: 9102,
			},
			wantErr: true,
		},
		{
			name: "error AdditionalPort",
			ipt: &Input{
				Scheme:         "http",
				Host:           "127.0.0.1",
				Port:           8091,
				AdditionalPort: -9102,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.ipt.parseURL()
			if (err != nil) != tt.wantErr {
				t.Errorf("parseURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
