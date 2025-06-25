//go:build !linux
// +build !linux

package driver

import (
	"net"
	"time"
)

func CreateDialer(timeout timeoutParams) net.Dialer {
	dialer := net.Dialer{
		Timeout:   time.Duration(timeout.connect_timeout) * time.Second,
		KeepAlive: time.Duration(timeout.keepalive_interval) * time.Second, //此处参数同时作用于keepalive_idle和keepalive_interval
	}
	return dialer
}
