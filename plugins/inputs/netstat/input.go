// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2022-present Guance, Inc.

// Package netstat collects host netstat metrics.
package netstat

import (
	"encoding/json"
	"fmt"
	"net/netip"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/v3/net"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	_ inputs.ReadEnv   = (*Input)(nil)
	_ inputs.Singleton = (*Input)(nil)
)

const (
	minInterval = time.Second
	maxInterval = time.Minute
)

const (
	inputName      = "netstat"
	metricName     = inputName
	inputNamePort  = "netstat_port"
	metricNamePort = inputNamePort
	// conf File samples, reflected in the document.
	sampleCfg = `
[[inputs.netstat]]
  ##(optional) collect interval, default is 10 seconds
  interval = '10s'

  ##(optional) addr_ports which counts separately, default is []
  ## addr_ports measurements will be "netstat_port" 
  ## server may have multiple network cards, 
  ## only display this addr+port, example "1.1.1.1:80"
  ## display by server if only have port number, example "443"
  ## should add tags for some port, example ["80","443","0"] add tags
  #[[inputs.netstat.addr_ports]]
  #   ports = ["1.1.1.1:80","443","0"]
  #   [inputs.netstat.addr_ports.tags]
  #		project = "datakit"
  #		yyy = "xxx"
  #		service = "http"

  ## display this port by ip, example "*:9529"
  #[[inputs.netstat.addr_ports]]
  #	ports = ["*:9529"]

[inputs.netstat.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"`
)

type portConf struct {
	Ports []string
	Tags  map[string]string
}

var l = logger.DefaultSLogger(inputName)

type netInfo struct {
	tcpEstablished int
	tcpSynSent     int
	tcpSynRecv     int
	tcpFinWait1    int
	tcpFinWait2    int
	tcpTimeWait    int
	tcpClose       int
	tcpCloseWait   int
	tcpLastAck     int
	tcpListen      int
	tcpClosing     int
	tcpNone        int
	udpSocket      int
	pid            int
}

func newNetInfo() *netInfo {
	n := &netInfo{}
	return n
}

func (n *netInfo) toMap() map[string]interface{} {
	return map[string]interface{}{
		"tcp_established": n.tcpEstablished,
		"tcp_syn_sent":    n.tcpSynSent,
		"tcp_syn_recv":    n.tcpSynRecv,
		"tcp_fin_wait1":   n.tcpFinWait1,
		"tcp_fin_wait2":   n.tcpFinWait2,
		"tcp_time_wait":   n.tcpTimeWait,
		"tcp_close":       n.tcpClose,
		"tcp_close_wait":  n.tcpCloseWait,
		"tcp_last_ack":    n.tcpLastAck,
		"tcp_listen":      n.tcpListen,
		"tcp_closing":     n.tcpClosing,
		"tcp_none":        n.tcpNone,
		"udp_socket":      n.udpSocket,
		"pid":             n.pid,
	}
}

type Input struct {
	Interval         datakit.Duration
	Tags             map[string]string // Indicator name
	collectCache     []inputs.Measurement
	collectCachePort []inputs.Measurement
	platform         string
	netConnections   NetConnections // A function Type, the instance of Input calls the function
	semStop          *cliutils.Sem  // start stop signal
	AddrPorts        []*portConf    `toml:"addr_ports"` // the ip and port that must show
	netInfos         map[string]*netInfo
}

func (ipt *Input) Singleton() {
}

// netStat Measurement structure.
type netStatMeasurement struct {
	name   string                 // Indicator set name ，here is "netstat"
	tags   map[string]string      // Indicator name
	fields map[string]interface{} // Indicator measurement results
}

// LineProto data formatting, submit through FeedMeasurement.
func (n *netStatMeasurement) LineProto() (*point.Point, error) {
	return point.NewPoint(n.name, n.tags, n.fields, point.MOpt())
}

// Info , reflected in the document
//nolint:lll
func (n *netStatMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricName,
		Fields: map[string]interface{}{
			"tcp_established": NewFieldInfoC("ESTABLISHED : The number of TCP state be open connection, data received to be delivered to the user. "),
			"tcp_syn_sent":    NewFieldInfoC("SYN_SENT : The number of TCP state be waiting for a machine connection request after sending a connecting request."),
			"tcp_syn_recv":    NewFieldInfoC("SYN_RECV : The number of TCP state be waiting for confirmation of connection acknowledgement after both sender and receiver has sent / received connection request."),
			"tcp_fin_wait1":   NewFieldInfoC("FIN_WAIT1 : The number of TCP state be waiting for a connection termination request from remote TCP host or acknowledgment of connection termination request sent previously."),
			"tcp_fin_wait2":   NewFieldInfoC("FIN_WAIT2 : The number of TCP state be waiting for connection termination request from remote TCP host."),
			"tcp_time_wait":   NewFieldInfoC("TIME_WAIT : The number of TCP state be waiting sufficient time to pass to ensure remote TCP host received acknowledgement of its request for connection termination."),
			"tcp_close":       NewFieldInfoC("CLOSE : The number of TCP state be waiting for a connection termination request acknowledgement from remote TCP host."),
			"tcp_close_wait":  NewFieldInfoC("CLOSE_WAIT : The number of TCP state be waiting for a connection termination request from local user."),
			"tcp_last_ack":    NewFieldInfoC("LAST_ACK : The number of TCP state be waiting for connection termination request acknowledgement previously sent to remote TCP host including its acknowledgement of connection termination request."),
			"tcp_listen":      NewFieldInfoC("LISTEN : The number of TCP state be waiting for a connection request from any remote TCP host."),
			"tcp_closing":     NewFieldInfoC("CLOSING : The number of TCP state be waiting for a connection termination request acknowledgement from remote TCP host."),
			"tcp_none":        NewFieldInfoC("NONE"),
			"udp_socket":      NewFieldInfoC("UDP : The number of UDP connection."),
			"pid":             NewFieldInfoC("pid."),
		},

		Tags: map[string]interface{}{
			"host":      &inputs.TagInfo{Desc: "Host name"},
			"addr_port": &inputs.TagInfo{Desc: "addr and port"},
		},
	}
}

// Collect Get, Aggregate, Calculate Data.
func (ipt *Input) Collect() error {
	ipt.collectCache = make([]inputs.Measurement, 0)
	ipt.collectCachePort = make([]inputs.Measurement, 0)

	// get data
	netConns, err := ipt.netConnections()
	if err != nil {
		return fmt.Errorf("error getting virtual memory info: %w", err)
	}

	// count every indicator
	for _, netConn := range netConns {
		ipt.handleNetConn(netConn)
	}

	// collectCache tags
	tags := map[string]string{}
	for k, v := range ipt.Tags {
		tags[k] = v
	}

	// handle  fields
	fields := ipt.netInfos["all"].toMap()
	delete(fields, "pid")

	// Append to the collectCache, the Run() function will handle it
	ipt.collectCache = append(ipt.collectCache, &netStatMeasurement{
		name: inputName,
		tags: tags,
		// "all" mean count by server
		fields: fields,
	})

	// Append port data to the collectCachePort, the Run() function will handle it
	for addrPort, value := range ipt.netInfos {
		if addrPort == "all" {
			continue
		}

		// tag of addr+port
		portTags := map[string]string{"addr_port": addrPort}
		for k, v := range ipt.Tags {
			portTags[k] = v
		}

		// add tag follow ports
	loppOut:
		for _, addrInfo := range ipt.AddrPorts {
			for _, port := range addrInfo.Ports {
				if strings.Index(port, "*:") == 0 {
					// handle like "*:8080"

					// must compare the real port
					portLeft, _ := getPort(port)
					portRight, _ := getPort(port)
					if portLeft == portRight {
						for k, v := range addrInfo.Tags {
							portTags[k] = v
						}
						break loppOut
					}
				} else if port == addrPort {
					// handle like "8080", "1.1.1.1:90"
					for k, v := range addrInfo.Tags {
						portTags[k] = v
					}
					break loppOut
				}
			}
		}

		// Append to the collectCachePort, the Run() function will handle it
		ipt.collectCachePort = append(ipt.collectCachePort, &netStatMeasurement{
			name:   inputNamePort,
			tags:   portTags,
			fields: value.toMap(),
		})
	}

	return err
}

func (ipt *Input) handleNetConn(netConn net.ConnectionStat) {
	var key string // key word of field Status
	var gotAddrPort string

	// make key
	if netConn.Type == syscall.SOCK_DGRAM {
		key = "UDP"
		// continue // UDP has no status
	} else {
		key = netConn.Status
	}

	// make mertric netstat
	ipt.addCounts("all", key, netConn.Laddr.IP, int(netConn.Pid), int(netConn.Laddr.Port))

	// make mertric netstat_port
	for _, addrInfo := range ipt.AddrPorts {
		for _, port := range addrInfo.Ports {
			if strings.Index(port, "*:") == 0 {
				// handle like "*:8080"
				gotPort := strconv.Itoa(int(netConn.Laddr.Port))
				if port == gotPort {
					// like "*:8080", nust show add ip prefix
					gotAddrPort = netConn.Laddr.IP + ":" + gotPort
					if _, err := netip.ParseAddrPort(gotAddrPort); err == nil {
						// use typ make form netConn just now
						ipt.addCounts(gotAddrPort, key, netConn.Laddr.IP, int(netConn.Pid), int(netConn.Laddr.Port))
						return
					}
				}
			} else {
				// handle like "8080", "1.1.1.1:90"
				gotPort := strconv.Itoa(int(netConn.Laddr.Port))
				gotAddrPort = netConn.Laddr.IP + ":" + gotPort
				if port == gotPort || port == gotAddrPort {
					// use typ from ipt.addrPorts
					ipt.addCounts(port, key, netConn.Laddr.IP, int(netConn.Pid), int(netConn.Laddr.Port))
					return
				}
			}
		}
	}
}

// addCounts add counts.
func (ipt *Input) addCounts(typ, key, addr string, pid int, port int) {
	// check map key, creat if non-existent
	if _, ok := ipt.netInfos[typ]; !ok {
		ipt.netInfos[typ] = newNetInfo()
	}

	// add pid
	ipt.netInfos[typ].pid = pid

	// add netInfos data
	switch key {
	case "ESTABLISHED":
		ipt.netInfos[typ].tcpEstablished++
	case "SYN_SENT":
		ipt.netInfos[typ].tcpSynSent++
	case "SYN_RECV":
		ipt.netInfos[typ].tcpSynRecv++
	case "FIN_WAIT1":
		ipt.netInfos[typ].tcpFinWait1++
	case "FIN_WAIT2":
		ipt.netInfos[typ].tcpFinWait2++
	case "TIME_WAIT":
		ipt.netInfos[typ].tcpTimeWait++
	case "CLOSE":
		ipt.netInfos[typ].tcpClose++
	case "CLOSE_WAIT":
		ipt.netInfos[typ].tcpCloseWait++
	case "LAST_ACK":
		ipt.netInfos[typ].tcpLastAck++
	case "LISTEN":
		ipt.netInfos[typ].tcpListen++
	case "CLOSING":
		ipt.netInfos[typ].tcpClosing++
	case "NONE":
		ipt.netInfos[typ].tcpNone++
	case "UDP":
		ipt.netInfos[typ].udpSocket++
	}
}

// Run Start the process of timing acquisition.
// If this indicator is included in the list to be collected, it will only be called once.
// The for{} loops every tick.
func (ipt *Input) Run() {
	l = logger.SLogger(inputName)
	l.Infof("netStat input started")
	ipt.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval.Duration)
	tick := time.NewTicker(ipt.Interval.Duration)
	defer tick.Stop()

	for {
		start := time.Now()

		// Collect() to get data
		if err := ipt.Collect(); err != nil {
			l.Errorf("Collect: %s", err)
			io.FeedLastError(inputName, err.Error())
		}

		// If there is data in the collectCache, submit it
		if len(ipt.collectCache) > 0 {
			if err := inputs.FeedMeasurement(metricName, datakit.Metric, ipt.collectCache,
				&io.Option{CollectCost: time.Since(start)}); err != nil {
				fmt.Println(err)
				l.Errorf("FeedMeasurement: %s", err)
			}
		}

		// If there is data in the collectCachePort, submit it
		if len(ipt.collectCachePort) > 0 {
			if err := inputs.FeedMeasurement(metricNamePort, datakit.Metric, ipt.collectCachePort,
				&io.Option{CollectCost: time.Since(start)}); err != nil {
				fmt.Println(err)
				l.Errorf("FeedMeasurement: %s", err)
			}
		}

		select {
		case <-tick.C:

		case <-datakit.Exit.Wait():
			l.Infof("memory input exit")
			return

		case <-ipt.semStop.Wait():
			l.Infof("memory input return")
			return
		}
	}
}

// get the port example "*:8080" "1.1.1.1:8080" "8080".
func getPort(s string) (string, error) {
	if _, err := strconv.Atoi(s); err == nil {
		// only port
		return s, nil
	} else {
		// get last part of string
		strs := strings.Split(s, ":")
		str := strs[len(strs)-1]

		if _, err := strconv.Atoi(str); err == nil {
			// right port
			return str, nil
		} else {
			return "", fmt.Errorf("get port form string error")
		}
	}
}

// Terminate Stop.
func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

// Catalog catalog.
func (*Input) Catalog() string {
	return "host"
}

// SampleConfig : conf File samples, reflected in the document.
func (*Input) SampleConfig() string {
	return sampleCfg
}

// AvailableArchs : OS support, reflected in the document.
func (*Input) AvailableArchs() []string {
	return datakit.AllOS
}

// SampleMeasurement Sample measurement results, reflected in the document.
func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&netStatMeasurement{},
	}
}

// ReadEnv support envs：only for K8S.
func (ipt *Input) ReadEnv(envs map[string]string) {
	// ENV_INPUT_NETSTAT_TAGS : "a=b,c=d"
	if tagsStr, ok := envs["ENV_INPUT_NETSTAT_TAGS"]; ok {
		tags := config.ParseGlobalTags(tagsStr)
		for k, v := range tags {
			ipt.Tags[k] = v
		}
	}

	// ENV_INPUT_NETSTAT_INTERVAL : datakit.Duration
	if str, ok := envs["ENV_INPUT_NETSTAT_INTERVAL"]; ok {
		da, err := time.ParseDuration(str)
		if err != nil {
			l.Warnf("parse ENV_INPUT_NETSTAT_INTERVAL to time.Duration: %s, ignore", err)
		} else {
			ipt.Interval.Duration = config.ProtectedInterval(minInterval,
				maxInterval,
				da)
		}
	}

	// ENV_INPUT_NETSTAT_ADDR_PORTS
	if str, ok := envs["ENV_INPUT_NETSTAT_ADDR_PORTS"]; ok {
		var strs []string
		err := json.Unmarshal([]byte(str), &strs)
		if err != nil {
			l.Warnf("parse ENV_INPUT_NETSTAT_ADDR_PORTS: %s, ignore", err)
		} else {
			ipt.AddrPorts = []*portConf{
				// only pots, no tags
				{Ports: strs},
			}
		}
	}
}

func newDefaultInput() *Input {
	ipt := &Input{
		netConnections: GetNetConnections,
		platform:       runtime.GOOS,
		Interval:       datakit.Duration{Duration: time.Second * 10},
		semStop:        cliutils.NewSem(),
		Tags:           make(map[string]string),
		netInfos:       make(map[string]*netInfo),
	}
	return ipt
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return newDefaultInput()
	})
}
