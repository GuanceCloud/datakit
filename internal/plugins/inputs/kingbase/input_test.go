// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kingbase

import (
	"net"
	"strings"
	"testing"

	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/kingbase/driver"
)

// func initIpt() *Input {
// 	i := defaultInput()
// 	i.Host = "junlin623.cn"
// 	i.Port = 54321
// 	i.User = "datakit"
// 	i.Password = "datakit"
// 	i.Database = "test"

// 	if err := i.setup(); err != nil {
// 		return nil
// 	}
// 	i.Init()

// 	return i
// }

// func TestCollect(t *testing.T) {
// 	i := initIpt()
// 	if i == nil {
// 		t.Fatal("init ipt error")
// 	}

// 	i.ptsTime = ntp.Now()
// 	if err := i.Collect(); err != nil {
// 		t.Error(err)
// 	}
// }

func TestGetHost(t *testing.T) {
	tests := []struct {
		name         string
		inputHost    string
		mockHostname string
		expectedHost string
		expectErr    bool
	}{
		{
			name:         "Empty Host",
			inputHost:    "",
			mockHostname: "test-machine",
			expectedHost: "test-machine",
		},
		{
			name:         "Localhost Host",
			inputHost:    "LocalHost",
			mockHostname: "test-machine",
			expectedHost: "test-machine",
		},
		{
			name:         "Loopback IP",
			inputHost:    "127.0.0.1",
			mockHostname: "test-machine",
			expectedHost: "test-machine",
		},
		{
			name:         "Non-Loopback IP",
			inputHost:    "192.168.1.1",
			mockHostname: "test-machine",
			expectedHost: "192.168.1.1",
		},
		{
			name:         "Non-IP Host",
			inputHost:    "example.com",
			mockHostname: "test-machine",
			expectedHost: "example.com",
		},
		{
			name:         "Hostname Error",
			inputHost:    "",
			mockHostname: "",
			expectedHost: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setHost := false
			host := strings.ToLower(tt.inputHost)
			switch host {
			case "", "localhost":
				setHost = true
			default:
				if ip := net.ParseIP(host); ip != nil && ip.IsLoopback() {
					setHost = true
				}
			}
			if setHost {
				host = tt.mockHostname
			}

			if host != tt.expectedHost {
				t.Errorf("Test %s: expected host %q, got %q", tt.name, tt.expectedHost, host)
			}
		})
	}
}
