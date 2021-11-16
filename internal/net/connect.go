package net

import (
	"net"
	"time"
)

// RawConnect 验证host:port是否监听，类似telenet host port.
func RawConnect(host, port string, timeout time.Duration) error {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), timeout)
	if err != nil {
		return err
	}
	if conn != nil {
		defer conn.Close() //nolint:errcheck
		return nil
	}
	return nil
}
