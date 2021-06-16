package net

import (
	"testing"
	"time"
)

func TestRawConnect(t *testing.T) {
	var cases = []struct {
		host string
		port string
		fail bool
	}{
		{
			"127.0.0.1",
			"22", // 假设本机开启 sshd
			false,
		},
		{
			"127.0.0.1",
			"65534",
			true,
		},
	}

	for idx, tc := range cases {
		err := RawConnect(tc.host, tc.port, time.Second)
		if err != nil && tc.fail {
			t.Logf("[ OK ] index: %d failed of connect: %s\n", idx, err)
		} else if err != nil {
			t.Logf("[ ERROR ] index: %d failed of connect: %s\n", idx, err)
		} else {
			t.Logf("[ OK ] index: %d connect success %s:%s\n", idx, tc.host, tc.port)
		}
	}
}
