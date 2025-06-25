//go:build !windows
// +build !windows

package driver

import (
	"fmt"
	"net"
	"syscall"
	"time"
)

func CreateDialer(timeout timeoutParams) net.Dialer {
	dialer := net.Dialer{
		Timeout:   time.Duration(timeout.connect_timeout) * time.Second,
		KeepAlive: time.Duration(timeout.keepalive_interval) * time.Second, //此处参数同时作用于keepalive_idle和keepalive_interval
		Control: func(network, address string, c syscall.RawConn) error {
			var controlErr error
			err := c.Control(func(fd uintptr) {
				//通过系统调用依次设置keepalive_count、tcp_user_timeout
				//keepalive_idle和keepalive_interval必须在上述KeepAlive设置，否则会被KeepAlive的默认值15s覆盖
				controlErr = syscall.SetsockoptInt(int(fd), syscall.IPPROTO_TCP, syscall.TCP_KEEPCNT, timeout.keepalive_count)
				if controlErr != nil {
					return
				}
				controlErr = syscall.SetsockoptInt(int(fd), syscall.IPPROTO_TCP, 0x12, timeout.tcp_user_timeout)
				if controlErr != nil {
					return
				}
			})
			if err != nil {
				return fmt.Errorf("raw control error: %w", err)
			}
			return controlErr
		},
	}
	return dialer
}
