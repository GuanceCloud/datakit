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

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/shirou/gopsutil/v3/net"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
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
  ##(Optional) Collect interval, default is 10 seconds
  interval = '10s'

  ## The ports you want display
  ## Can add tags too
  # [[inputs.netstat.addr_ports]]
  #   ports = ["80","443"]

  ## Groups of ports and add different tags to facilitate statistics
  # [[inputs.netstat.addr_ports]]
  #   ports = ["80","443"]
  #   [inputs.netstat.addr_ports.tags]
  #     service = "http"
  # [[inputs.netstat.addr_ports]]
  #   ports = ["9529"]
  #   [inputs.netstat.addr_ports.tags]
  #     service = "datakit"
  #     foo = "bar"

  ## Server may have multiple network cards
  ## Display only some network cards
  ## Can add tags too
  # [[inputs.netstat.addr_ports]]
  #   ports = ["1.1.1.1:80","2.2.2.2:80"]
  #   ports_match is preferred if both ports and ports_match configured
  #   ports_match = ["*:80","*:443"]

[inputs.netstat.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"`
)

type portConf struct {
	Ports      []string          `toml:"ports"`       // monitor addr:port or port
	PortsMatch []string          `toml:"ports_match"` // monitor *:port, will shield Ports
	Tags       map[string]string `toml:"tags"`
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
	return &netInfo{}
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

type NetInfos struct {
	typ       string
	tags      map[string]string
	ipVersion string // ip version
	netInfo   *netInfo
}

type Input struct {
	Interval         datakit.Duration
	Tags             map[string]string // Indicator name
	collectCache     []*point.Point
	collectCachePort []*point.Point
	platform         string
	netConnections   NetConnections // A function Type, the instance of Input calls the function
	semStop          *cliutils.Sem  // start stop signal
	AddrPorts        []*portConf    `toml:"addr_ports"` // the ip and port that must show
	netInfos         []*NetInfos    // cache metric,
	feeder           dkio.Feeder
	opt              point.Option
}

func (ipt *Input) Singleton() {}

// netStat Measurement structure.
type netStatMeasurement struct {
	name   string                 // Indicator set name ，here is "netstat"
	tags   map[string]string      // Indicator name
	fields map[string]interface{} // Indicator measurement results
	ts     time.Time
	ipt    *Input
}

// Point implement MeasurementV2.
func (m *netStatMeasurement) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(m.ts), m.ipt.opt)

	return point.NewPointV2([]byte(m.name),
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

// LineProto data formatting, submit through FeedMeasurement.
func (*netStatMeasurement) LineProto() (*dkpt.Point, error) {
	// return point.NewPoint(n.name, n.tags, n.fields, point.MOpt())
	return nil, fmt.Errorf("not implement")
}

// Info , reflected in the document
//nolint:lll
func (*netStatMeasurement) Info() *inputs.MeasurementInfo {
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
			"pid":             NewFieldInfoC("PID. Optional."),
		},

		Tags: map[string]interface{}{
			"host":       &inputs.TagInfo{Desc: "Host name"},
			"addr_port":  &inputs.TagInfo{Desc: "Addr and port. Optional."},
			"ip_version": &inputs.TagInfo{Desc: "IP version, 4 for IPV4, 6 for IPV6, unknown for others"},
		},
	}
}

// Collect Get, Aggregate, Calculate Data.
func (ipt *Input) Collect() error {
	ipt.netInfos = make([]*NetInfos, 0)
	ipt.collectCache = make([]*point.Point, 0)
	ipt.collectCachePort = make([]*point.Point, 0)

	// get data
	netConns, err := ipt.netConnections()
	if err != nil {
		return fmt.Errorf("error getting virtual memory info: %w", err)
	}

	// count every indicator
	for _, netConn := range netConns {
		// walk every single connection
		// result is ipt.netInfos
		ipt.handleNetConn(netConn)
	}

	now := time.Now()

	// Append to the collectCache, the Run() function will handle it
	for i := 0; i < len(ipt.netInfos); i++ {
		netInfo := ipt.netInfos[i]
		if netInfo.typ != "all" {
			// tag of addr+port
			portTags := map[string]string{"addr_port": netInfo.typ, "ip_version": netInfo.ipVersion}
			// tags for these ports
			for k, v := range netInfo.tags {
				portTags[k] = v
			}
			// tags for all
			for k, v := range ipt.Tags {
				portTags[k] = v
			}

			metric := &netStatMeasurement{
				name:   inputNamePort,
				tags:   portTags,
				fields: netInfo.netInfo.toMap(),
				ts:     now,
				ipt:    ipt,
			}
			ipt.collectCachePort = append(ipt.collectCachePort, metric.Point())
		} else { // netstat all
			fields := netInfo.netInfo.toMap()
			delete(fields, "pid")

			// collectCache tags
			tags := map[string]string{}
			for k, v := range ipt.Tags {
				tags[k] = v
			}

			tags["ip_version"] = netInfo.ipVersion

			metric := &netStatMeasurement{
				name: inputName,
				tags: tags,
				// "all" mean count by server
				fields: fields,
				ts:     now,
				ipt:    ipt,
			}
			ipt.collectCache = append(ipt.collectCache, metric.Point())
		}
	}

	return err
}

func (ipt *Input) handleNetConn(netConn net.ConnectionStat) {
	var key string // key word of field Status
	var gotAddrPort string

	// get connection status, like "UDP" "CLOSE_WAIT"
	if netConn.Type == syscall.SOCK_DGRAM {
		key = "UDP"
		// UDP has no status
	} else {
		key = netConn.Status
	}

	// make metric netstat
	ipt.addCounts("all", key, netConn.Laddr.IP, int(netConn.Pid), int(netConn.Laddr.Port), -1)

	// make metric netstat_port
	for confIndex, addrInfo := range ipt.AddrPorts {
		// walk configs
		if len(addrInfo.PortsMatch) > 0 {
			// have like "*:port"
			for _, port := range addrInfo.PortsMatch {
				// walk config.ports_match
				if strings.Index(port, "*:") != 0 {
					continue
				}
				gotPort := strconv.Itoa(int(netConn.Laddr.Port))
				if port[2:] == gotPort {
					// like "*:8080", must show add ip prefix
					gotAddrPort = netConn.Laddr.IP + ":" + gotPort
					if _, err := netip.ParseAddrPort(gotAddrPort); err == nil {
						// use typ make form netConn just now
						ipt.addCounts(gotAddrPort, key, netConn.Laddr.IP, int(netConn.Pid), int(netConn.Laddr.Port), confIndex)
						return
					}
				}
			}
		} else {
			// not have like "*:port"
			for _, port := range addrInfo.Ports {
				// walk config.ports
				// handle like "8080", "1.1.1.1:90"
				gotPort := strconv.Itoa(int(netConn.Laddr.Port))
				gotAddrPort = netConn.Laddr.IP + ":" + gotPort
				if port == gotPort || port == gotAddrPort {
					// use typ from ipt.addrPorts
					ipt.addCounts(port, key, netConn.Laddr.IP, int(netConn.Pid), int(netConn.Laddr.Port), confIndex)
					return
				}
			}
		}
	}
}

// addCounts add counts.
func (ipt *Input) addCounts(typ, key, addr string, pid int, port int, confIndex int) {
	// check key, creat if non-existent
	i := 0
	ipVersion := getIPVersion(addr)
	for i = 0; i < len(ipt.netInfos); i++ {
		if ipt.netInfos[i].typ == typ && ipt.netInfos[i].ipVersion == ipVersion {
			break
		}
	}
	if i >= len(ipt.netInfos) {
		// non-existent, append
		n := &NetInfos{
			typ:       typ,
			ipVersion: ipVersion,
			tags:      map[string]string{},
			netInfo:   newNetInfo(),
		}

		// add tag
		if confIndex > -1 {
			// "all" is netstat, have no tags
			for k, v := range ipt.AddrPorts[confIndex].Tags {
				n.tags[k] = v
			}
		}

		ipt.netInfos = append(ipt.netInfos, n)
	}

	// add pid
	ipt.netInfos[i].netInfo.pid = pid

	// add netInfos data
	switch key {
	case "ESTABLISHED":
		ipt.netInfos[i].netInfo.tcpEstablished++
	case "SYN_SENT":
		ipt.netInfos[i].netInfo.tcpSynSent++
	case "SYN_RECV":
		ipt.netInfos[i].netInfo.tcpSynRecv++
	case "FIN_WAIT1":
		ipt.netInfos[i].netInfo.tcpFinWait1++
	case "FIN_WAIT2":
		ipt.netInfos[i].netInfo.tcpFinWait2++
	case "TIME_WAIT":
		ipt.netInfos[i].netInfo.tcpTimeWait++
	case "CLOSE":
		ipt.netInfos[i].netInfo.tcpClose++
	case "CLOSE_WAIT":
		ipt.netInfos[i].netInfo.tcpCloseWait++
	case "LAST_ACK":
		ipt.netInfos[i].netInfo.tcpLastAck++
	case "LISTEN":
		ipt.netInfos[i].netInfo.tcpListen++
	case "CLOSING":
		ipt.netInfos[i].netInfo.tcpClosing++
	case "NONE":
		ipt.netInfos[i].netInfo.tcpNone++
	case "UDP":
		ipt.netInfos[i].netInfo.udpSocket++
	}
}

// Run Start the process of timing acquisition.
// If this indicator is included in the list to be collected, it will only be called once.
// The for{} loops every tick.
func (ipt *Input) Run() {
	l = logger.SLogger(inputName)
	l.Infof("netStat input started")

	// no election.
	ipt.opt = point.WithExtraTags(dkpt.GlobalHostTags())

	ipt.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval.Duration)
	tick := time.NewTicker(ipt.Interval.Duration)
	defer tick.Stop()

	for {
		start := time.Now()

		// Collect() to get data
		if err := ipt.Collect(); err != nil {
			l.Errorf("Collect: %s", err)
			ipt.feeder.FeedLastError(inputName, err.Error())
		}

		// If there is data in the collectCache, submit it
		if len(ipt.collectCache) > 0 {
			if err := ipt.feeder.Feed(metricName, point.Metric, ipt.collectCache,
				&dkio.Option{CollectCost: time.Since(start)}); err != nil {
				l.Errorf("FeedMeasurement: %s", err)
			}
		}

		// If there is data in the collectCachePort, submit it
		if len(ipt.collectCachePort) > 0 {
			if err := ipt.feeder.Feed(metricNamePort, point.Metric, ipt.collectCachePort,
				&dkio.Option{CollectCost: time.Since(start)}); err != nil {
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

func defaultInput() *Input {
	return &Input{
		netConnections: GetNetConnections,
		platform:       runtime.GOOS,
		Interval:       datakit.Duration{Duration: time.Second * 10},
		semStop:        cliutils.NewSem(),
		Tags:           make(map[string]string),
		netInfos:       []*NetInfos{},
		feeder:         dkio.DefaultFeeder(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
