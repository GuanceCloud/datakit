// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package netstat

import (
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/testutils"
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
		netConnections: ConnectionStat4Test,
		netInfos:       []*NetInfos{},
		tagger:         testutils.NewTaggerHost(),
	}
	i.platform = "linux" // runtime.GOOS
	i.setup()
	if err := i.collect(); err != nil {
		t.Error(err)
	}

	assert.Equal(t, 2, len(i.collectCache))

	expected := map[string]map[string]int64{
		"4": {
			"tcp_close":       0,
			"tcp_close_wait":  6,
			"tcp_closing":     0,
			"tcp_established": 25,
			"tcp_fin_wait1":   0,
			"tcp_fin_wait2":   0,
			"tcp_last_ack":    0,
			"tcp_listen":      4,
			"tcp_none":        0,
			"tcp_syn_recv":    0,
			"tcp_syn_sent":    3,
			"tcp_time_wait":   0,
			"udp_socket":      0,
		},
		"unknown": {
			"tcp_close":       0,
			"tcp_close_wait":  0,
			"tcp_closing":     0,
			"tcp_established": 0,
			"tcp_fin_wait1":   0,
			"tcp_fin_wait2":   0,
			"tcp_last_ack":    0,
			"tcp_listen":      2,
			"tcp_none":        0,
			"tcp_syn_recv":    0,
			"tcp_syn_sent":    0,
			"tcp_time_wait":   0,
			"udp_socket":      16,
		},
	}

	for _, v := range i.collectCache {
		ipVersion := v.MapTags()["ip_version"]
		collect := v.InfluxFields()
		if expectedFields, ok := expected[ipVersion]; ok {
			for k, v := range expectedFields {
				assertEqualint(t, v, collect[k].(int64), k)
			}
		}
	}
}

func assertEqualint(t *testing.T, expected, actual int64, mName string) {
	t.Helper()
	if expected != actual {
		t.Errorf("error: "+mName+" expected: %d \t actual %d", expected, actual)
	}
}

// testing netstat by ports

type testStruct struct {
	name       string
	connStats  []net.ConnectionStat
	addrPorts  []*portConf
	wantFields []*netInfo
	wantTags   []map[string]string
	wantErr    bool
}

// return  the correct  source connection data
func ConnectionStat4TestPort(conn []net.ConnectionStat) func() ([]net.ConnectionStat, error) {
	return func() ([]net.ConnectionStat, error) {
		return conn, nil
	}
}

// sort data,the result is index
type testingCatche struct {
	index      int
	tagsString string
}

// sort by tags content
func sortData(collects []*point.Point) []testingCatche {
	testingCatches := make([]testingCatche, len(collects))
	for i := 0; i < len(collects); i++ {
		testingCatches[i].index = i
		strs := make([]string, 0)

		for k, v := range collects[i].InfluxTags() {
			vStr := v
			_ = vStr

			// strs = append(strs, k+v)
			strs = append(strs, fmt.Sprintf("%d%d", k, v))
		}
		sort.Strings(strs)

		// all tags k/v -> one string
		testingCatches[i].tagsString = func(ss []string) string {
			s := ""
			for j := 0; j < len(ss); j++ {
				s += ss[j]
			}
			return s
		}(strs)
	}

	// sort testingCatches by testingCatches[i].tagsString
	sort.SliceStable(testingCatches, func(i, j int) bool {
		return testingCatches[i].tagsString < testingCatches[j].tagsString
	})

	return testingCatches
}

func assertEquaFields(t *testing.T, collectFields map[string]interface{}, wantFields *netInfo, name string, j int) {
	t.Helper()
	for k, v := range collectFields {
		vWant := 0
		vGot, ok := v.(int64)
		if !ok {
			continue
		}
		switch k {
		case "tcp_established":
			vWant = wantFields.tcpEstablished
		case "tcp_syn_sent":
			vWant = wantFields.tcpSynSent
		case "tcp_syn_recv":
			vWant = wantFields.tcpSynRecv
		case "tcp_fin_wait1":
			vWant = wantFields.tcpFinWait1
		case "tcp_fin_wait2":
			vWant = wantFields.tcpFinWait2
		case "tcp_time_wait":
			vWant = wantFields.tcpTimeWait
		case "tcp_close":
			vWant = wantFields.tcpClose
		case "tcp_close_wait":
			vWant = wantFields.tcpCloseWait
		case "tcp_last_ack":
			vWant = wantFields.tcpLastAck
		case "tcp_listen":
			vWant = wantFields.tcpListen
		case "tcp_closing":
			vWant = wantFields.tcpClosing
		case "tcp_none":
			vWant = wantFields.tcpNone
		case "udp_socket":
			vWant = wantFields.udpSocket
		case "pid":
			vWant = wantFields.pid
		}
		if int64(vWant) != vGot {
			t.Errorf("error: testname:"+name+", index:%d, field:"+k+", want: %d  got %d", j, vWant, vGot)
		}
	}
}

func assertEquaTags(t *testing.T, collectTags map[string]string, wantTags map[string]string, name string, j int) {
	t.Helper()
	if !reflect.DeepEqual(collectTags, wantTags) {
		t.Errorf("error: testname:"+name+", index:%d, want: %d  got %d", j, wantTags, collectTags)
	}
}

func TestCollectByPort(t *testing.T) {
	tests := testData
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// init
			i := &Input{
				// source connection data
				netConnections: ConnectionStat4TestPort(tt.connStats),
				// .conf data
				AddrPorts: tt.addrPorts,
				// must init this map, otherwise will panic
				netInfos: []*NetInfos{},
				tagger:   testutils.NewTaggerHost(),
			}

			// run function
			i.platform = "linux" // runtime.GOOS
			i.setup()
			if err := i.collect(); err != nil {
				t.Error(err)
			}

			// compare length
			if len(tt.wantFields) != len(i.collectCachePort) {
				t.Errorf("error: testName:"+tt.name+", error length, want: %d  got %d", len(tt.wantFields), len(i.collectCachePort))
			}

			// sort data
			sortCatches := sortData(i.collectCachePort)

			for k, v := range i.collectCachePort {
				fmt.Printf("pt[%02d] = %v\n", k, v.LineProto())
			}

			// check
			for j := 0; j < len(i.collectCachePort); j++ {
				// i.collectCachePort's order after sort
				collectFields := i.collectCachePort[sortCatches[j].index].InfluxFields()
				assertEquaFields(t, collectFields, tt.wantFields[j], tt.name, j)

				collectTags := i.collectCachePort[sortCatches[j].index].MapTags()
				assertEquaTags(t, collectTags, tt.wantTags[j], tt.name, j)
			}
		})
	}
}

// data of testing netstat by ports

var testData = []testStruct{
	{
		name:      "01 ports, no tag",
		connStats: connectionStats01,
		addrPorts: []*portConf{
			{
				Ports: []string{"8000", "5353", "80"},
			},
		},
		wantFields: []*netInfo{
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 9, 3059},
			{1, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 62951},
			{12, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3059},
		},
		wantTags: []map[string]string{
			{"addr_port": "5353", "ip_version": "unknown", "host": "HOST"},
			{"addr_port": "8000", "ip_version": "4", "host": "HOST"},
			{"addr_port": "80", "ip_version": "4", "host": "HOST"},
		},
		wantErr: false,
	},
	{
		name:      "02 with *:port && tag",
		connStats: connectionStats02,
		addrPorts: []*portConf{
			{
				PortsMatch: []string{"*:80"},
				Tags:       map[string]string{"service": "http"},
			},
			{
				PortsMatch: []string{"*:443"},
				Tags:       map[string]string{"service": "https"},
			},
		},
		wantFields: []*netInfo{
			{4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3059},
			{1, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3059},
			{4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3059},
			{4, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3059},
		},
		wantTags: []map[string]string{
			{"service": "https", "addr_port": "10.100.64.115:443", "ip_version": "4", "host": "HOST"},
			{"service": "http", "addr_port": "10.100.64.115:80", "ip_version": "4", "host": "HOST"},
			{"service": "https", "addr_port": "10.100.64.119:443", "ip_version": "4", "host": "HOST"},
			{"service": "http", "addr_port": "10.100.64.119:80", "ip_version": "4", "host": "HOST"},
		},
		wantErr: false,
	},
	{
		name:      "03 with ip",
		connStats: connectionStats03,
		addrPorts: []*portConf{
			{
				Ports: []string{"10.100.64.115:443", "80", "443"},
			},
		},
		wantFields: []*netInfo{
			{3, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4285},
			{4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4285},
			{5, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3059},
		},
		wantTags: []map[string]string{
			{"addr_port": "10.100.64.115:443", "ip_version": "4", "host": "HOST"},
			{"addr_port": "443", "ip_version": "4", "host": "HOST"},
			{"addr_port": "80", "ip_version": "4", "host": "HOST"},
		},
		wantErr: false,
	},
}

var connectionStats01 = []net.ConnectionStat{
	{Fd: 18, Family: 2, Type: 2, Laddr: net.Addr{IP: "*", Port: 137}, Raddr: net.Addr{IP: "", Port: 0}, Status: "", Uids: nil, Pid: 1},
	{Fd: 31, Family: 2, Type: 2, Laddr: net.Addr{IP: "*", Port: 138}, Raddr: net.Addr{IP: "", Port: 0}, Status: "", Uids: nil, Pid: 1},
	{Fd: 21, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 53182}, Raddr: net.Addr{IP: "10.100.65.71", Port: 8009}, Status: "ESTABLISHED", Uids: nil, Pid: 3059},
	{Fd: 22, Family: 2, Type: 1, Laddr: net.Addr{IP: "127.0.0.1", Port: 54064}, Raddr: net.Addr{IP: "127.0.0.1", Port: 8000}, Status: "ESTABLISHED", Uids: nil, Pid: 3059},
	{Fd: 24, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 80}, Raddr: net.Addr{IP: "120.92.145.239", Port: 7826}, Status: "ESTABLISHED", Uids: nil, Pid: 3059},
	{Fd: 25, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 80}, Raddr: net.Addr{IP: "172.217.163.42", Port: 443}, Status: "SYN_SENT", Uids: nil, Pid: 3059},
	{Fd: 26, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 80}, Raddr: net.Addr{IP: "122.9.67.165", Port: 443}, Status: "ESTABLISHED", Uids: nil, Pid: 3059},
	{Fd: 27, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 80}, Raddr: net.Addr{IP: "223.109.175.205", Port: 443}, Status: "ESTABLISHED", Uids: nil, Pid: 3059},
	{Fd: 28, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.114", Port: 80}, Raddr: net.Addr{IP: "120.92.145.239", Port: 7826}, Status: "ESTABLISHED", Uids: nil, Pid: 3059},
	{Fd: 29, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.114", Port: 80}, Raddr: net.Addr{IP: "172.217.163.42", Port: 443}, Status: "SYN_SENT", Uids: nil, Pid: 3059},
	{Fd: 30, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.114", Port: 80}, Raddr: net.Addr{IP: "172.217.163.42", Port: 443}, Status: "SYN_SENT", Uids: nil, Pid: 3059},
	{Fd: 32, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.114", Port: 80}, Raddr: net.Addr{IP: "121.40.139.49", Port: 443}, Status: "ESTABLISHED", Uids: nil, Pid: 3059},
	{Fd: 33, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.114", Port: 80}, Raddr: net.Addr{IP: "121.40.139.49", Port: 443}, Status: "ESTABLISHED", Uids: nil, Pid: 3059},
	{Fd: 37, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.114", Port: 80}, Raddr: net.Addr{IP: "52.83.230.200", Port: 443}, Status: "ESTABLISHED", Uids: nil, Pid: 3059},
	{Fd: 39, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.114", Port: 80}, Raddr: net.Addr{IP: "36.156.125.230", Port: 443}, Status: "ESTABLISHED", Uids: nil, Pid: 3059},
	{Fd: 43, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.114", Port: 80}, Raddr: net.Addr{IP: "36.156.125.230", Port: 443}, Status: "ESTABLISHED", Uids: nil, Pid: 3059},
	{Fd: 44, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.114", Port: 80}, Raddr: net.Addr{IP: "36.156.208.222", Port: 443}, Status: "ESTABLISHED", Uids: nil, Pid: 3059},
	{Fd: 47, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.114", Port: 80}, Raddr: net.Addr{IP: "223.111.250.56", Port: 443}, Status: "ESTABLISHED", Uids: nil, Pid: 3059},
	{Fd: 49, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.114", Port: 80}, Raddr: net.Addr{IP: "112.13.92.248", Port: 443}, Status: "ESTABLISHED", Uids: nil, Pid: 3059},
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

var connectionStats02 = []net.ConnectionStat{
	{Fd: 24, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 80}, Raddr: net.Addr{IP: "120.92.145.239", Port: 7826}, Status: "ESTABLISHED", Uids: nil, Pid: 3059},
	{Fd: 25, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 80}, Raddr: net.Addr{IP: "172.217.163.42", Port: 443}, Status: "SYN_SENT", Uids: nil, Pid: 3059},
	{Fd: 26, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 80}, Raddr: net.Addr{IP: "122.9.67.165", Port: 443}, Status: "ESTABLISHED", Uids: nil, Pid: 3059},
	{Fd: 27, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 80}, Raddr: net.Addr{IP: "223.109.175.205", Port: 443}, Status: "ESTABLISHED", Uids: nil, Pid: 3059},
	{Fd: 28, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 80}, Raddr: net.Addr{IP: "120.92.145.239", Port: 7826}, Status: "ESTABLISHED", Uids: nil, Pid: 3059},
	{Fd: 29, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.115", Port: 80}, Raddr: net.Addr{IP: "172.217.163.42", Port: 443}, Status: "SYN_SENT", Uids: nil, Pid: 3059},
	{Fd: 30, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.115", Port: 80}, Raddr: net.Addr{IP: "172.217.163.42", Port: 443}, Status: "SYN_SENT", Uids: nil, Pid: 3059},
	{Fd: 32, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.115", Port: 80}, Raddr: net.Addr{IP: "121.40.139.49", Port: 443}, Status: "ESTABLISHED", Uids: nil, Pid: 3059},
	{Fd: 33, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.115", Port: 443}, Raddr: net.Addr{IP: "121.40.139.49", Port: 443}, Status: "ESTABLISHED", Uids: nil, Pid: 3059},
	{Fd: 37, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.115", Port: 443}, Raddr: net.Addr{IP: "52.83.230.200", Port: 443}, Status: "ESTABLISHED", Uids: nil, Pid: 3059},
	{Fd: 39, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.115", Port: 443}, Raddr: net.Addr{IP: "36.156.125.230", Port: 443}, Status: "ESTABLISHED", Uids: nil, Pid: 3059},
	{Fd: 43, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.115", Port: 443}, Raddr: net.Addr{IP: "36.156.125.230", Port: 443}, Status: "ESTABLISHED", Uids: nil, Pid: 3059},
	{Fd: 44, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 443}, Raddr: net.Addr{IP: "36.156.208.222", Port: 443}, Status: "ESTABLISHED", Uids: nil, Pid: 3059},
	{Fd: 47, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 443}, Raddr: net.Addr{IP: "223.111.250.56", Port: 443}, Status: "ESTABLISHED", Uids: nil, Pid: 3059},
	{Fd: 49, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 443}, Raddr: net.Addr{IP: "112.13.92.248", Port: 443}, Status: "ESTABLISHED", Uids: nil, Pid: 3059},
	{Fd: 35, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 443}, Raddr: net.Addr{IP: "59.82.44.105", Port: 443}, Status: "ESTABLISHED", Uids: nil, Pid: 3059},
	{Fd: 21, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 53182}, Raddr: net.Addr{IP: "10.100.65.71", Port: 8009}, Status: "ESTABLISHED", Uids: nil, Pid: 3059},
	{Fd: 22, Family: 2, Type: 1, Laddr: net.Addr{IP: "127.0.0.1", Port: 8080}, Raddr: net.Addr{IP: "127.0.0.1", Port: 8000}, Status: "ESTABLISHED", Uids: nil, Pid: 3059},
	{Fd: 147, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 53775}, Raddr: net.Addr{IP: "59.82.31.244", Port: 443}, Status: "ESTABLISHED", Uids: nil, Pid: 4285},
	{Fd: 217, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 54055}, Raddr: net.Addr{IP: "203.119.244.127", Port: 443}, Status: "ESTABLISHED", Uids: nil, Pid: 4285},
	{Fd: 244, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 53180}, Raddr: net.Addr{IP: "203.119.205.54", Port: 443}, Status: "ESTABLISHED", Uids: nil, Pid: 4285},
	{Fd: 33, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 53448}, Raddr: net.Addr{IP: "121.36.83.100", Port: 80}, Status: "CLOSE_WAIT", Uids: nil, Pid: 26791},
	{Fd: 42, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 59863}, Raddr: net.Addr{IP: "114.116.235.116", Port: 443}, Status: "CLOSED", Uids: nil, Pid: 26791},
	{Fd: 88, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 59924}, Raddr: net.Addr{IP: "120.92.43.165", Port: 443}, Status: "CLOSE_WAIT", Uids: nil, Pid: 26791},
	{Fd: 95, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 59928}, Raddr: net.Addr{IP: "124.70.24.185", Port: 443}, Status: "CLOSE_WAIT", Uids: nil, Pid: 26791},
	{Fd: 100, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 59944}, Raddr: net.Addr{IP: "110.43.67.232", Port: 443}, Status: "CLOSE_WAIT", Uids: nil, Pid: 26791},
	{Fd: 149, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 53225}, Raddr: net.Addr{IP: "117.184.242.106", Port: 80}, Status: "ESTABLISHED", Uids: nil, Pid: 69248},
}

var connectionStats03 = []net.ConnectionStat{
	{Fd: 21, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 53182}, Raddr: net.Addr{IP: "10.100.65.71", Port: 8009}, Status: "ESTABLISHED", Uids: nil, Pid: 3059},
	{Fd: 22, Family: 2, Type: 1, Laddr: net.Addr{IP: "127.0.0.1", Port: 8080}, Raddr: net.Addr{IP: "127.0.0.1", Port: 8000}, Status: "ESTABLISHED", Uids: nil, Pid: 3059},
	{Fd: 24, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 80}, Raddr: net.Addr{IP: "120.92.145.239", Port: 7826}, Status: "ESTABLISHED", Uids: nil, Pid: 3059},
	{Fd: 25, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 80}, Raddr: net.Addr{IP: "172.217.163.42", Port: 443}, Status: "SYN_SENT", Uids: nil, Pid: 3059},
	{Fd: 26, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 80}, Raddr: net.Addr{IP: "122.9.67.165", Port: 443}, Status: "ESTABLISHED", Uids: nil, Pid: 3059},
	{Fd: 27, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 80}, Raddr: net.Addr{IP: "223.109.175.205", Port: 443}, Status: "ESTABLISHED", Uids: nil, Pid: 3059},
	{Fd: 28, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 80}, Raddr: net.Addr{IP: "120.92.145.239", Port: 7826}, Status: "ESTABLISHED", Uids: nil, Pid: 3059},
	{Fd: 29, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.115", Port: 80}, Raddr: net.Addr{IP: "172.217.163.42", Port: 443}, Status: "SYN_SENT", Uids: nil, Pid: 3059},
	{Fd: 30, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.115", Port: 80}, Raddr: net.Addr{IP: "172.217.163.42", Port: 443}, Status: "SYN_SENT", Uids: nil, Pid: 3059},
	{Fd: 32, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.115", Port: 80}, Raddr: net.Addr{IP: "121.40.139.49", Port: 443}, Status: "ESTABLISHED", Uids: nil, Pid: 3059},
	{Fd: 33, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.115", Port: 443}, Raddr: net.Addr{IP: "121.40.139.49", Port: 443}, Status: "SYN_SENT", Uids: nil, Pid: 4285},
	{Fd: 37, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.115", Port: 443}, Raddr: net.Addr{IP: "52.83.230.200", Port: 443}, Status: "ESTABLISHED", Uids: nil, Pid: 4285},
	{Fd: 39, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.115", Port: 443}, Raddr: net.Addr{IP: "36.156.125.230", Port: 443}, Status: "ESTABLISHED", Uids: nil, Pid: 4285},
	{Fd: 43, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.115", Port: 443}, Raddr: net.Addr{IP: "36.156.125.230", Port: 443}, Status: "ESTABLISHED", Uids: nil, Pid: 4285},
	{Fd: 44, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 443}, Raddr: net.Addr{IP: "36.156.208.222", Port: 443}, Status: "ESTABLISHED", Uids: nil, Pid: 4285},
	{Fd: 47, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 443}, Raddr: net.Addr{IP: "223.111.250.56", Port: 443}, Status: "ESTABLISHED", Uids: nil, Pid: 4285},
	{Fd: 49, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 443}, Raddr: net.Addr{IP: "112.13.92.248", Port: 443}, Status: "ESTABLISHED", Uids: nil, Pid: 4285},
	{Fd: 35, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 443}, Raddr: net.Addr{IP: "59.82.44.105", Port: 443}, Status: "ESTABLISHED", Uids: nil, Pid: 4285},
	{Fd: 147, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 53775}, Raddr: net.Addr{IP: "59.82.31.244", Port: 443}, Status: "ESTABLISHED", Uids: nil, Pid: 4285},
	{Fd: 217, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 54055}, Raddr: net.Addr{IP: "203.119.244.127", Port: 443}, Status: "ESTABLISHED", Uids: nil, Pid: 4285},
	{Fd: 244, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 53180}, Raddr: net.Addr{IP: "203.119.205.54", Port: 443}, Status: "ESTABLISHED", Uids: nil, Pid: 4285},
	{Fd: 33, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 53448}, Raddr: net.Addr{IP: "121.36.83.100", Port: 80}, Status: "CLOSE_WAIT", Uids: nil, Pid: 26791},
	{Fd: 42, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 59863}, Raddr: net.Addr{IP: "114.116.235.116", Port: 443}, Status: "CLOSED", Uids: nil, Pid: 26791},
	{Fd: 88, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 59924}, Raddr: net.Addr{IP: "120.92.43.165", Port: 443}, Status: "CLOSE_WAIT", Uids: nil, Pid: 26791},
	{Fd: 95, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 59928}, Raddr: net.Addr{IP: "124.70.24.185", Port: 443}, Status: "CLOSE_WAIT", Uids: nil, Pid: 26791},
	{Fd: 100, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 59944}, Raddr: net.Addr{IP: "110.43.67.232", Port: 443}, Status: "CLOSE_WAIT", Uids: nil, Pid: 26791},
	{Fd: 149, Family: 2, Type: 1, Laddr: net.Addr{IP: "10.100.64.119", Port: 53225}, Raddr: net.Addr{IP: "117.184.242.106", Port: 80}, Status: "ESTABLISHED", Uids: nil, Pid: 69248},
}
