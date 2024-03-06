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

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/export/doc"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	minInterval    = time.Second
	maxInterval    = time.Minute
	inputName      = "netstat"
	metricName     = inputName
	inputNamePort  = "netstat_port"
	metricNamePort = inputNamePort
)

var (
	_ inputs.ReadEnv   = (*Input)(nil)
	_ inputs.Singleton = (*Input)(nil)
	l                  = logger.DefaultSLogger(inputName)
)

type portConf struct {
	Ports      []string          `toml:"ports"`       // monitor addr:port or port
	PortsMatch []string          `toml:"ports_match"` // monitor *:port, will shield Ports
	Tags       map[string]string `toml:"tags"`
}

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

type NetInfos struct {
	typ       string
	tags      map[string]string
	ipVersion string // ip version
	netInfo   *netInfo
}

type Input struct {
	Interval  time.Duration
	Tags      map[string]string // Indicator name
	AddrPorts []*portConf       `toml:"addr_ports"` // the ip and port that must show

	semStop          *cliutils.Sem // start stop signal
	collectCache     []*point.Point
	collectCachePort []*point.Point
	platform         string
	netConnections   NetConnections // A function Type, the instance of Input calls the function
	netInfos         []*NetInfos    // cache metric,
	feeder           dkio.Feeder
	mergedTags       map[string]string
	tagger           datakit.GlobalTagger
}

// Run Start the process of timing acquisition.
// If this indicator is included in the list to be collected, it will only be called once.
// The for{} loops every tick.
func (ipt *Input) Run() {
	ipt.setup()

	tick := time.NewTicker(ipt.Interval)
	defer tick.Stop()

	for {
		start := time.Now()

		if err := ipt.collect(); err != nil {
			l.Errorf("collect: %s", err)
			ipt.feeder.FeedLastError(err.Error(),
				dkio.WithLastErrorInput(inputName),
			)
		}

		// If there is data in the collectCache, submit it
		if len(ipt.collectCache) > 0 {
			if err := ipt.feeder.FeedV2(point.Metric, ipt.collectCache,
				dkio.WithCollectCost(time.Since(start)),
				dkio.WithInputName(metricName)); err != nil {
				ipt.feeder.FeedLastError(err.Error(),
					dkio.WithLastErrorInput(inputName),
					dkio.WithLastErrorCategory(point.Metric),
				)
				l.Errorf("feed measurement: %s", err)
			}
		}

		// If there is data in the collectCachePort, submit it
		if len(ipt.collectCachePort) > 0 {
			if err := ipt.feeder.FeedV2(point.Metric, ipt.collectCachePort,
				dkio.WithCollectCost(time.Since(start)),
				dkio.WithInputName(metricNamePort)); err != nil {
				ipt.feeder.FeedLastError(err.Error(),
					dkio.WithLastErrorInput(inputName),
					dkio.WithLastErrorCategory(point.Metric),
				)
				l.Errorf("feed measurement: %s", err)
			}
		}

		select {
		case <-tick.C:
		case <-datakit.Exit.Wait():
			l.Infof("%s input exit", inputName)
			return
		case <-ipt.semStop.Wait():
			l.Infof("%s input return", inputName)
			return
		}
	}
}

func (ipt *Input) setup() {
	l = logger.SLogger(inputName)

	l.Infof("%s input started", inputName)
	ipt.Interval = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval)
	ipt.mergedTags = inputs.MergeTags(ipt.tagger.HostTags(), ipt.Tags, "")
	l.Debugf("merged tags: %+#v", ipt.mergedTags)
}

func (ipt *Input) collect() error {
	ipt.netInfos = make([]*NetInfos, 0)
	ipt.collectCache = make([]*point.Point, 0)
	ipt.collectCachePort = make([]*point.Point, 0)
	ts := time.Now()
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(ts))

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

	// Append to the collectCache, the Run() function will handle it
	for i := 0; i < len(ipt.netInfos); i++ {
		netInfo := ipt.netInfos[i]
		if netInfo.typ != "all" {
			var kvs point.KVs

			for k, v := range netInfo.netInfo.toMap() {
				kvs = kvs.Add(k, v, false, true)
			}

			kvs = kvs.Add("addr_port", netInfo.typ, true, true)
			kvs = kvs.Add("ip_version", netInfo.ipVersion, true, true)
			for k, v := range netInfo.tags {
				kvs = kvs.AddTag(k, v)
			}
			for k, v := range ipt.mergedTags {
				kvs = kvs.AddTag(k, v)
			}

			ipt.collectCachePort = append(ipt.collectCachePort, point.NewPointV2(inputNamePort, kvs, opts...))
		} else { // netstat all
			fields := netInfo.netInfo.toMap()
			delete(fields, "pid")

			var kvs point.KVs

			for k, v := range fields {
				kvs = kvs.Add(k, v, false, true)
			}

			kvs = kvs.Add("ip_version", netInfo.ipVersion, true, true)
			for k, v := range ipt.mergedTags {
				kvs = kvs.AddTag(k, v)
			}

			ipt.collectCache = append(ipt.collectCache, point.NewPointV2(inputName, kvs, opts...))
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

func (*Input) Singleton() {}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}
func (*Input) Catalog() string          { return "host" }
func (*Input) SampleConfig() string     { return sampleCfg }
func (*Input) AvailableArchs() []string { return datakit.AllOS }
func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&docMeasurement{},
	}
}

func (ipt *Input) GetENVDoc() []*inputs.ENVInfo {
	// nolint:lll
	infos := []*inputs.ENVInfo{
		{FieldName: "Interval"},
		{FieldName: "AddrPorts", Type: doc.JSON, Example: `["1.1.1.1:80","443"]`, Desc: "Groups of ports and add different tags to facilitate statistics", DescZh: "端口分组并添加不同的标签以便于统计"},
		{FieldName: "Tags"},
	}

	return doc.SetENVDoc("ENV_INPUT_NETSTAT_", infos)
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

	// ENV_INPUT_NETSTAT_INTERVAL : time.Duration
	if str, ok := envs["ENV_INPUT_NETSTAT_INTERVAL"]; ok {
		da, err := time.ParseDuration(str)
		if err != nil {
			l.Warnf("parse ENV_INPUT_NETSTAT_INTERVAL to time.Duration: %s, ignore", err)
		} else {
			ipt.Interval = config.ProtectedInterval(minInterval,
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
		Interval:       time.Second * 10,
		semStop:        cliutils.NewSem(),
		Tags:           make(map[string]string),
		netInfos:       []*NetInfos{},
		feeder:         dkio.DefaultFeeder(),
		tagger:         datakit.DefaultGlobalTagger(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
