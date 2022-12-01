// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package netstat

import (
	"testing"

	"github.com/shirou/gopsutil/v3/net"
	"github.com/stretchr/testify/assert"
)

var connectionStats = []net.ConnectionStat{
	{Fd: 18, Family: 2, Type: 2, Laddr: net.Addr{IP: "*", Port: 137}, Raddr: net.Addr{IP: "", Port: 0}, Status: "", Uids: nil, Pid: 1},
	{Fd: 31, Family: 2, Type: 2, Laddr: net.Addr{IP: "*", Port: 138}, Raddr: net.Addr{IP: "", Port: 0}, Status: "", Uids: nil, Pid: 1},
	{Fd: 6, Family: 2, Type: 2, Laddr: net.Addr{IP: "*", Port: 5353}, Raddr: net.Addr{IP: "", Port: 0}, Status: "", Uids: nil, Pid: 186},
	{Fd: 7, Family: 30, Type: 2, Laddr: net.Addr{IP: "*", Port: 5353}, Raddr: net.Addr{IP: "", Port: 0}, Status: "", Uids: nil, Pid: 186},
	{Fd: 21, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 53182}, Raddr: net.Addr{IP: "10.100.65.71", Port: 8009}, Status: "ESTABLISHED", Uids: nil, Pid: 3059},
	{Fd: 22, Family: 2, Type: 1, Laddr: net.Addr{IP: "127.0.0.1", Port: 54064}, Raddr: net.Addr{IP: "127.0.0.1", Port: 8000}, Status: "ESTABLISHED", Uids: nil, Pid: 3059},
	{Fd: 24, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 53262}, Raddr: net.Addr{IP: "120.92.145.239", Port: 7826}, Status: "ESTABLISHED", Uids: nil, Pid: 3059},
	{Fd: 25, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 54067}, Raddr: net.Addr{IP: "172.217.163.42", Port: 443}, Status: "SYN_SENT", Uids: nil, Pid: 3059},
	{Fd: 26, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 53307}, Raddr: net.Addr{IP: "122.9.67.165", Port: 443}, Status: "ESTABLISHED", Uids: nil, Pid: 3059},
	{Fd: 27, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 53303}, Raddr: net.Addr{IP: "223.109.175.205", Port: 443}, Status: "ESTABLISHED", Uids: nil, Pid: 3059},
	{Fd: 28, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 53260}, Raddr: net.Addr{IP: "120.92.145.239", Port: 7826}, Status: "ESTABLISHED", Uids: nil, Pid: 3059},
	{Fd: 29, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 54069}, Raddr: net.Addr{IP: "172.217.163.42", Port: 443}, Status: "SYN_SENT", Uids: nil, Pid: 3059},
	{Fd: 30, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 54050}, Raddr: net.Addr{IP: "172.217.163.42", Port: 443}, Status: "SYN_SENT", Uids: nil, Pid: 3059},
	{Fd: 32, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 53982}, Raddr: net.Addr{IP: "121.40.139.49", Port: 443}, Status: "ESTABLISHED", Uids: nil, Pid: 3059},
	{Fd: 33, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 53981}, Raddr: net.Addr{IP: "121.40.139.49", Port: 443}, Status: "ESTABLISHED", Uids: nil, Pid: 3059},
	{Fd: 37, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 53986}, Raddr: net.Addr{IP: "52.83.230.200", Port: 443}, Status: "ESTABLISHED", Uids: nil, Pid: 3059},
	{Fd: 39, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 54045}, Raddr: net.Addr{IP: "36.156.125.230", Port: 443}, Status: "ESTABLISHED", Uids: nil, Pid: 3059},
	{Fd: 43, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 54046}, Raddr: net.Addr{IP: "36.156.125.230", Port: 443}, Status: "ESTABLISHED", Uids: nil, Pid: 3059},
	{Fd: 44, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 54044}, Raddr: net.Addr{IP: "36.156.208.222", Port: 443}, Status: "ESTABLISHED", Uids: nil, Pid: 3059},
	{Fd: 47, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 54061}, Raddr: net.Addr{IP: "223.111.250.56", Port: 443}, Status: "ESTABLISHED", Uids: nil, Pid: 3059},
	{Fd: 49, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 54060}, Raddr: net.Addr{IP: "112.13.92.248", Port: 443}, Status: "ESTABLISHED", Uids: nil, Pid: 3059},
	{Fd: 63, Family: 30, Type: 2, Laddr: net.Addr{IP: "*", Port: 5353}, Raddr: net.Addr{IP: "", Port: 0}, Status: "", Uids: nil, Pid: 3059},
	{Fd: 78, Family: 30, Type: 2, Laddr: net.Addr{IP: "*", Port: 5353}, Raddr: net.Addr{IP: "", Port: 0}, Status: "", Uids: nil, Pid: 3059},
	{Fd: 82, Family: 2, Type: 2, Laddr: net.Addr{IP: "*", Port: 5353}, Raddr: net.Addr{IP: "", Port: 0}, Status: "", Uids: nil, Pid: 3059},
	{Fd: 83, Family: 30, Type: 2, Laddr: net.Addr{IP: "*", Port: 5353}, Raddr: net.Addr{IP: "", Port: 0}, Status: "", Uids: nil, Pid: 3059},
	{Fd: 85, Family: 30, Type: 2, Laddr: net.Addr{IP: "*", Port: 5353}, Raddr: net.Addr{IP: "", Port: 0}, Status: "", Uids: nil, Pid: 3059},
	{Fd: 86, Family: 30, Type: 2, Laddr: net.Addr{IP: "*", Port: 5353}, Raddr: net.Addr{IP: "", Port: 0}, Status: "", Uids: nil, Pid: 3059},
	{Fd: 87, Family: 30, Type: 2, Laddr: net.Addr{IP: "*", Port: 5353}, Raddr: net.Addr{IP: "", Port: 0}, Status: "", Uids: nil, Pid: 3059},
	{Fd: 88, Family: 30, Type: 2, Laddr: net.Addr{IP: "*", Port: 5353}, Raddr: net.Addr{IP: "", Port: 0}, Status: "", Uids: nil, Pid: 3059},
	{Fd: 89, Family: 30, Type: 2, Laddr: net.Addr{IP: "*", Port: 5353}, Raddr: net.Addr{IP: "", Port: 0}, Status: "", Uids: nil, Pid: 3059},
	{Fd: 35, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 53388}, Raddr: net.Addr{IP: "59.82.44.105", Port: 443}, Status: "ESTABLISHED", Uids: nil, Pid: 4285},
	{Fd: 147, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 53775}, Raddr: net.Addr{IP: "59.82.31.244", Port: 443}, Status: "ESTABLISHED", Uids: nil, Pid: 4285},
	{Fd: 217, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 54055}, Raddr: net.Addr{IP: "203.119.244.127", Port: 443}, Status: "ESTABLISHED", Uids: nil, Pid: 4285},
	{Fd: 244, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 53180}, Raddr: net.Addr{IP: "203.119.205.54", Port: 443}, Status: "ESTABLISHED", Uids: nil, Pid: 4285},
	{Fd: 33, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 53448}, Raddr: net.Addr{IP: "121.36.83.100", Port: 80}, Status: "CLOSE_WAIT", Uids: nil, Pid: 26791},
	{Fd: 42, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 59863}, Raddr: net.Addr{IP: "114.116.235.116", Port: 443}, Status: "CLOSED", Uids: nil, Pid: 26791},
	{Fd: 88, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 59924}, Raddr: net.Addr{IP: "120.92.43.165", Port: 443}, Status: "CLOSE_WAIT", Uids: nil, Pid: 26791},
	{Fd: 95, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 59928}, Raddr: net.Addr{IP: "124.70.24.185", Port: 443}, Status: "CLOSE_WAIT", Uids: nil, Pid: 26791},
	{Fd: 100, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 59944}, Raddr: net.Addr{IP: "110.43.67.232", Port: 443}, Status: "CLOSE_WAIT", Uids: nil, Pid: 26791},
	{Fd: 3, Family: 2, Type: 2, Laddr: net.Addr{IP: "*", Port: 137}, Raddr: net.Addr{IP: "", Port: 0}, Status: "", Uids: nil, Pid: 38050},
	{Fd: 4, Family: 2, Type: 2, Laddr: net.Addr{IP: "*", Port: 138}, Raddr: net.Addr{IP: "", Port: 0}, Status: "", Uids: nil, Pid: 38050},
	{Fd: 39, Family: 30, Type: 1, Laddr: net.Addr{IP: "*", Port: 54003}, Raddr: net.Addr{IP: "", Port: 0}, Status: "LISTEN", Uids: nil, Pid: 38332},
	{Fd: 40, Family: 30, Type: 1, Laddr: net.Addr{IP: "127.0.0.1", Port: 54003}, Raddr: net.Addr{IP: "127.0.0.1", Port: 54005}, Status: "ESTABLISHED", Uids: nil, Pid: 38332},
	{Fd: 15, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 59864}, Raddr: net.Addr{IP: "121.36.83.100", Port: 80}, Status: "ESTABLISHED", Uids: nil, Pid: 41852},
	{Fd: 22, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 65523}, Raddr: net.Addr{IP: "111.31.6.3", Port: 443}, Status: "CLOSE_WAIT", Uids: nil, Pid: 41852},
	{Fd: 38, Family: 2, Type: 2, Laddr: net.Addr{IP: "*", Port: 55890}, Raddr: net.Addr{IP: "", Port: 0}, Status: "", Uids: nil, Pid: 41852},
	{Fd: 193, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 59923}, Raddr: net.Addr{IP: "36.156.49.191", Port: 443}, Status: "ESTABLISHED", Uids: nil, Pid: 41852},
	{Fd: 212, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 59922}, Raddr: net.Addr{IP: "110.43.67.230", Port: 443}, Status: "CLOSE_WAIT", Uids: nil, Pid: 41852},
	{Fd: 17, Family: 30, Type: 1, Laddr: net.Addr{IP: "127.0.0.1", Port: 6942}, Raddr: net.Addr{IP: "", Port: 0}, Status: "LISTEN", Uids: nil, Pid: 44342},
	{Fd: 22, Family: 30, Type: 1, Laddr: net.Addr{IP: "127.0.0.1", Port: 63342}, Raddr: net.Addr{IP: "", Port: 0}, Status: "LISTEN", Uids: nil, Pid: 44342},
	{Fd: 287, Family: 30, Type: 1, Laddr: net.Addr{IP: "127.0.0.1", Port: 54005}, Raddr: net.Addr{IP: "127.0.0.1", Port: 54003}, Status: "ESTABLISHED", Uids: nil, Pid: 44342},
	{Fd: 12, Family: 2, Type: 1, Laddr: net.Addr{IP: "127.0.0.1", Port: 9529}, Raddr: net.Addr{IP: "", Port: 0}, Status: "LISTEN", Uids: nil, Pid: 47143},
	{Fd: 14, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 54062}, Raddr: net.Addr{IP: "47.110.144.10", Port: 443}, Status: "ESTABLISHED", Uids: nil, Pid: 47143},
	{Fd: 3, Family: 2, Type: 1, Laddr: net.Addr{IP: "127.0.0.1", Port: 8000}, Raddr: net.Addr{IP: "", Port: 0}, Status: "LISTEN", Uids: nil, Pid: 62951},
	{Fd: 4, Family: 2, Type: 1, Laddr: net.Addr{IP: "127.0.0.1", Port: 8000}, Raddr: net.Addr{IP: "127.0.0.1", Port: 54064}, Status: "ESTABLISHED", Uids: nil, Pid: 62951},
	{Fd: 149, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 53225}, Raddr: net.Addr{IP: "117.184.242.106", Port: 80}, Status: "ESTABLISHED", Uids: nil, Pid: 69248},
	{Fd: 17, Family: 30, Type: 1, Laddr: net.Addr{IP: "*", Port: 7123}, Raddr: net.Addr{IP: "", Port: 0}, Status: "LISTEN", Uids: nil, Pid: 80181},
}

func ConnectionStat4Test() ([]net.ConnectionStat, error) {
	return connectionStats, nil
}

func TestNetStatCollect(t *testing.T) {
	i := &Input{
		netConnections: ConnectionStat4Test, AddrPorts: []*portConf{
			{
				Ports: []string{"127.0.0.1:8000", "5353", "*:53303"},
				Tags:  map[string]string{"service": "http"},
			},
		},
		netInfos: make(map[string]*netInfo),
	}
	i.platform = "linux" // runtime.GOOS
	if err := i.Collect(); err != nil {
		t.Error(err)
	}
	collect := i.collectCache[0].(*netStatMeasurement).fields

	assertEqualint(t, 0, collect["tcp_close"].(int), "tcp_close")
	assertEqualint(t, 6, collect["tcp_close_wait"].(int), "tcp_close_wait")
	assertEqualint(t, 0, collect["tcp_closing"].(int), "tcp_closing")
	assertEqualint(t, 25, collect["tcp_established"].(int), "tcp_established")
	assertEqualint(t, 0, collect["tcp_fin_wait1"].(int), "tcp_fin_wait1")
	assertEqualint(t, 0, collect["tcp_fin_wait2"].(int), "tcp_fin_wait2")
	assertEqualint(t, 0, collect["tcp_last_ack"].(int), "tcp_last_ack")
	assertEqualint(t, 6, collect["tcp_listen"].(int), "tcp_listen")
	assertEqualint(t, 0, collect["tcp_none"].(int), "tcp_none")
	assertEqualint(t, 0, collect["tcp_syn_recv"].(int), "tcp_syn_recv")
	assertEqualint(t, 3, collect["tcp_syn_sent"].(int), "tcp_syn_sent")
	assertEqualint(t, 0, collect["tcp_time_wait"].(int), "tcp_time_wait")
	assertEqualint(t, 16, collect["udp_socket"].(int), "udp_socket")

	// assert collectCachePort
	collect = i.collectCachePort[0].(*netStatMeasurement).fields
	assertEqualint(t, 0, collect["tcp_close"].(int), "tcp_close")
	assertEqualint(t, 0, collect["tcp_close_wait"].(int), "tcp_close_wait")
	assertEqualint(t, 0, collect["tcp_closing"].(int), "tcp_closing")
	assertEqualint(t, 0, collect["tcp_established"].(int), "tcp_established")
	assertEqualint(t, 0, collect["tcp_fin_wait1"].(int), "tcp_fin_wait1")
	assertEqualint(t, 0, collect["tcp_fin_wait2"].(int), "tcp_fin_wait2")
	assertEqualint(t, 0, collect["tcp_last_ack"].(int), "tcp_last_ack")
	assertEqualint(t, 0, collect["tcp_listen"].(int), "tcp_listen")
	assertEqualint(t, 0, collect["tcp_none"].(int), "tcp_none")
	assertEqualint(t, 0, collect["tcp_syn_recv"].(int), "tcp_syn_recv")
	assertEqualint(t, 0, collect["tcp_syn_sent"].(int), "tcp_syn_sent")
	assertEqualint(t, 0, collect["tcp_time_wait"].(int), "tcp_time_wait")
	assertEqualint(t, 11, collect["udp_socket"].(int), "udp_socket")

	collect = i.collectCachePort[1].(*netStatMeasurement).fields
	assertEqualint(t, 0, collect["tcp_close"].(int), "tcp_close")
	assertEqualint(t, 0, collect["tcp_close_wait"].(int), "tcp_close_wait")
	assertEqualint(t, 0, collect["tcp_closing"].(int), "tcp_closing")
	assertEqualint(t, 1, collect["tcp_established"].(int), "tcp_established")
	assertEqualint(t, 0, collect["tcp_fin_wait1"].(int), "tcp_fin_wait1")
	assertEqualint(t, 0, collect["tcp_fin_wait2"].(int), "tcp_fin_wait2")
	assertEqualint(t, 0, collect["tcp_last_ack"].(int), "tcp_last_ack")
	assertEqualint(t, 1, collect["tcp_listen"].(int), "tcp_listen")
	assertEqualint(t, 0, collect["tcp_none"].(int), "tcp_none")
	assertEqualint(t, 0, collect["tcp_syn_recv"].(int), "tcp_syn_recv")
	assertEqualint(t, 0, collect["tcp_syn_sent"].(int), "tcp_syn_sent")
	assertEqualint(t, 0, collect["tcp_time_wait"].(int), "tcp_time_wait")
	assertEqualint(t, 0, collect["udp_socket"].(int), "udp_socket")

	tags := i.collectCachePort[0].(*netStatMeasurement).tags
	assert.EqualValues(t, "5353", tags["addr_port"])

	tags = i.collectCachePort[1].(*netStatMeasurement).tags
	assert.EqualValues(t, "127.0.0.1:8000", tags["addr_port"])
}

func assertEqualint(t *testing.T, expected, actual int, mName string) {
	t.Helper()
	if expected != actual {
		t.Errorf("error: "+mName+" expected: %d \t actual %d", expected, actual)
	}
}
