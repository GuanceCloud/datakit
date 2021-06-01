package net

import (
	"fmt"
	"net"
	"strings"
	"time"

	psNet "github.com/shirou/gopsutil/net"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	minInterval = time.Second
	maxInterval = time.Minute
)

var (
	inputName     = "net"
	netMetricName = "net"
	l             = logger.DefaultSLogger(inputName)

	linuxProtoRate = map[string]bool{
		"insegs":       true,
		"outsegs":      true,
		"indatagrams":  true,
		"outdatagrams": true,
	}
	sampleCfg = `
[[inputs.net]]
  ##(optional) collect interval, default is 10 seconds
  interval = '10s'
  ##
  ## By default, gathers stats from any up interface, but Linux does not contain virtual interfaces.
  ## Setting interfaces using regular expressions will collect these expected interfaces.
  ##
  # interfaces = ['''eth[\w-]+''', '''lo''', ]
  ##
  ## Datakit does not collect network virtual interfaces under the linux system.
  ## Setting enable_virtual_interfaces to true will collect virtual interfaces stats for linux.
  ##
  # enable_virtual_interfaces = true
  ##
  ## On linux systems also collects protocol stats.
  ## Setting ignore_protocol_stats to true will skip reporting of protocol metrics.
  ##
  # ignore_protocol_stats = false
  ##

[inputs.net.tags]
# some_tag = "some_value"
# more_tag = "some_other_value"
`
)

type netMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

// https://tools.ietf.org/html/rfc1213#page-48
// https://www.kernel.org/doc/html/latest/networking/snmp_counter.html
// https://sourceforge.net/p/net-tools/code/ci/master/tree/statistics.c#l178

func (m *netMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: netMetricName,
		Fields: map[string]interface{}{
			"bytes_sent":       NewFieldsInfoIByte("The number of bytes sent by the interface ."),
			"bytes_sent/sec":   NewFieldsInfoIBytePerSec("The number of bytes sent by the interface per second."),
			"bytes_recv":       NewFieldsInfoIByte("The number of bytes received by the interface."),
			"bytes_recv/sec":   NewFieldsInfoIBytePerSec("The number of bytes received by the interface per second."),
			"packets_sent":     NewFieldsInfoCount("The number of packets sent by the interface."),
			"packets_sent/sec": NewFieldsInfoCountPerSec("The number of packets sent by the interface per second."),
			"packets_recv":     NewFieldsInfoCount("The number of packets received by the interface."),
			"packets_recv/sec": NewFieldsInfoCountPerSec("The number of packets received by the interface per second."),
			"err_in":           NewFieldsInfoCount("The number of receive errors detected by the interface."),
			"err_out":          NewFieldsInfoCount("The number of transmit errors detected by the interface."),
			"drop_in":          NewFieldsInfoCount("The number of received packets dropped by the interface."),
			"drop_out":         NewFieldsInfoCount("The number of transmitted packets dropped by the interface."),
			// linux only
			"tcp_insegs":       NewFieldsInfoCount("The number of packets received by the TCP layer."),
			"tcp_insegs/sec":   NewFieldsInfoCountPerSec("The number of packets received by the TCP layer per second."),
			"tcp_outsegs":      NewFieldsInfoCount("The number of packets sent by the TCP layer. "),
			"tcp_outsegs/sec":  NewFieldsInfoCountPerSec("The number of packets sent by the TCP layer per second."),
			"tcp_activeopens":  NewFieldsInfoCount("It means the TCP layer sends a SYN, and come into the SYN-SENT state. "),
			"tcp_passiveopens": NewFieldsInfoCount("It means the TCP layer receives a SYN, replies a SYN+ACK, come into the SYN-RCVD state."),
			"tcp_estabresets": NewFieldsInfoCount("The number of times TCP connections have made a " +
				"direct transition to the CLOSED state from either " +
				"the ESTABLISHED state or the CLOSE-WAIT state."),
			"tcp_attemptfails": NewFieldsInfoCount("The number of times TCP connections have made a " +
				"direct transition to the CLOSED state from either " +
				"the SYN-SENT state or the SYN-RCVD state, plus the " +
				"number of times TCP connections have made a direct " +
				"transition to the LISTEN state from the SYN-RCVD " +
				"state."),
			"tcp_outrsts": NewFieldsInfoCount("The number of TCP segments sent containing the RST flag."),
			"tcp_retranssegs": NewFieldsInfoCount("The total number of segments retransmitted - that " +
				"is, the number of TCP segments transmitted " +
				"containing one or more previously transmitted" +
				"octets."),
			"tcp_inerrs":           NewFieldsInfoCount("The number of incoming TCP segments in error"),
			"tcp_incsumerrors":     NewFieldsInfoCount("The number of incoming TCP segments in checksum error"),
			"tcp_rtoalgorithm":     NewFieldsInfoCount("The algorithm used to determine the timeout value used for retransmitting unacknowledged octets."),
			"tcp_rtomin":           NewFieldsInfoMS("The minimum value permitted by a TCP implementation for the retransmission timeout, measured in milliseconds."),
			"tcp_rtomax":           NewFieldsInfoMS("The maximum value permitted by a TCP implementation for the retransmission timeout, measured in milliseconds."),
			"tcp_maxconn":          NewFieldsInfoCount("The limit on the total number of TCP connections the entity can support."),
			"tcp_currestab":        NewFieldsInfoCount("The number of TCP connections for which the current state is either ESTABLISHED or CLOSE-WAIT."),
			"udp_incsumerrors":     NewFieldsInfoCount("The number of incoming UDP datagrams in checksum error"),
			"udp_indatagrams":      NewFieldsInfoCount("The number of UDP datagrams delivered to UDP users."),
			"udp_indatagrams/sec":  NewFieldsInfoCountPerSec("The number of UDP datagrams delivered to UDP users per second."),
			"udp_outdatagrams":     NewFieldsInfoCount("The number of UDP datagrams sent from this entity."),
			"udp_outdatagrams/sec": NewFieldsInfoCountPerSec("The number of UDP datagrams sent from this entity per second."),
			"udp_rcvbuferrors":     NewFieldsInfoCount("The number of receive buffer errors."),
			"udp_noports":          NewFieldsInfoCount("The number of packets to unknown port received."),
			"udp_sndbuferrors":     NewFieldsInfoCount("The number of send buffer errors."),
			"udp_inerrors":         NewFieldsInfoCount("The number of packet receive errors"),
			"udp_ignoredmulti":     NewFieldsInfoCount("IgnoredMulti"),
		},
		Tags: map[string]interface{}{
			"host":      &inputs.TagInfo{Desc: "主机名"},
			"interface": &inputs.TagInfo{Desc: "网络接口名"},
		},
	}
}

func (m *netMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

type Input struct {
	Interval                datakit.Duration
	IgnoreProtocolStats     bool
	Interfaces              []string
	EnableVirtualInterfaces bool
	Tags                    map[string]string

	collectCache     []inputs.Measurement
	lastStats        map[string]psNet.IOCountersStat
	lastProtoStats   []psNet.ProtoCountersStat
	lastTime         time.Time
	netIO            NetIO
	netProto         NetProto
	netVirtualIfaces NetVirtualIfaces
}

func (i *Input) appendMeasurement(name string, tags map[string]string, fields map[string]interface{}, ts time.Time) {
	tmp := &netMeasurement{name: name, tags: tags, fields: fields, ts: ts}
	i.collectCache = append(i.collectCache, tmp)
}

func (i *Input) AvailableArchs() []string {
	return datakit.AllArch
}

func (i *Input) Catalog() string {
	return "host"
}

func (i *Input) SampleConfig() string {
	return sampleCfg
}

func (i *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&netMeasurement{},
	}
}

func (i *Input) Collect() error {
	i.collectCache = make([]inputs.Measurement, 0)
	ts := time.Now()
	netio, err := NetIOCounters()
	if err != nil {
		return fmt.Errorf("error getting net io info: %s", err)
	}
	interfaces, err := net.Interfaces()
	if err != nil {
		return fmt.Errorf("error getting net interfaces info: %s", err)
	}

	filteredInterface, err := FilterInterface(netio, interfaces, i.Interfaces, i.EnableVirtualInterfaces, i.netVirtualIfaces)

	for name, ioStat := range filteredInterface {

		tags := map[string]string{
			"interface": ioStat.Name,
		}
		for k, v := range i.Tags {
			tags[k] = v
		}
		fields := map[string]interface{}{
			"bytes_sent":   ioStat.BytesSent,
			"bytes_recv":   ioStat.BytesRecv,
			"packets_sent": ioStat.PacketsSent,
			"packets_recv": ioStat.PacketsRecv,
			"err_in":       ioStat.Errin,
			"err_out":      ioStat.Errout,
			"drop_in":      ioStat.Dropin,
			"drop_out":     ioStat.Dropout,
		}
		if i.lastStats != nil {
			if lastIOStat, ok := i.lastStats[name]; ok {
				if ioStat.BytesSent >= lastIOStat.BytesSent && ts.Unix() > i.lastTime.Unix() {
					fields["bytes_sent/sec"] = int64(ioStat.BytesSent-lastIOStat.BytesSent) / (ts.Unix() - i.lastTime.Unix())
					fields["bytes_recv/sec"] = int64(ioStat.BytesRecv-lastIOStat.BytesRecv) / (ts.Unix() - i.lastTime.Unix())
					fields["packets_sent/sec"] = int64(ioStat.PacketsSent-lastIOStat.PacketsSent) / (ts.Unix() - i.lastTime.Unix())
					fields["packets_recv/sec"] = int64(ioStat.PacketsRecv-lastIOStat.PacketsRecv) / (ts.Unix() - i.lastTime.Unix())
				}
			}
		}

		i.appendMeasurement(netMetricName, tags, fields, ts)
	}
	// Get system wide stats for network protocols tcp and udp
	// Only supports linux
	if !i.IgnoreProtocolStats {
		netprotos, _ := i.netProto([]string{"tcp", "udp"}) // tcp udp only
		fields := make(map[string]interface{})
		for _, proto := range netprotos {
			for stat, value := range proto.Stats {
				name := fmt.Sprintf("%s_%s", strings.ToLower(proto.Protocol),
					strings.ToLower(stat))
				fields[name] = value
			}
		}
		for _, proto := range i.lastProtoStats {
			pname := strings.ToLower(proto.Protocol)
			for stat, value := range proto.Stats {
				sname := strings.ToLower(stat)
				if _, ok := linuxProtoRate[sname]; ok {
					if v, ok := fields[pname+"_"+sname]; ok && v.(int64) >= value && ts.Unix() > i.lastTime.Unix() {
						fields[pname+"_"+sname+"/sec"] = (v.(int64) - value) / (ts.Unix() - i.lastTime.Unix())
					}
				}

			}
		}
		tags := map[string]string{
			"interface": "all",
		}
		for k, v := range i.Tags {
			tags[k] = v
		}
		if len(fields) > 0 {
			i.appendMeasurement(netMetricName, tags, fields, ts)
		}
		i.lastProtoStats = netprotos
	}
	i.lastStats = filteredInterface
	i.lastTime = ts
	return err
}

func (i *Input) Run() {
	l = logger.SLogger(inputName)
	l.Infof("net input started")
	i.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, i.Interval.Duration)
	tick := time.NewTicker(i.Interval.Duration)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			start := time.Now()
			if err := i.Collect(); err == nil {
				if errFeed := inputs.FeedMeasurement(netMetricName, datakit.Metric, i.collectCache,
					&io.Option{CollectCost: time.Since(start)}); errFeed != nil {
					io.FeedLastError(inputName, errFeed.Error())
					l.Error(errFeed)
				}
			} else {
				io.FeedLastError(inputName, err.Error())
				l.Error(err)
			}
		case <-datakit.Exit.Wait():
			l.Infof("net input exit")
			return
		}
	}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			netIO:            NetIOCounters,
			netProto:         psNet.ProtoCounters,
			netVirtualIfaces: NetVirtualInterfaces,
			Interval:         datakit.Duration{Duration: time.Second * 10},
		}
	})
}
