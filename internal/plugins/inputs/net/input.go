// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package net collects host network metrics.
package net

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	psNet "github.com/shirou/gopsutil/net"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/export/doc"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	minInterval = time.Second
	maxInterval = time.Minute
	inputName   = "net"
	metricName  = inputName
)

var (
	_ inputs.ReadEnv   = (*Input)(nil)
	_ inputs.Singleton = (*Input)(nil)
	l                  = logger.DefaultSLogger(inputName)

	linuxProtoRate = map[string]bool{
		"insegs":       true,
		"outsegs":      true,
		"indatagrams":  true,
		"outdatagrams": true,
	}
)

type Input struct {
	Interval                time.Duration
	Interfaces              []string
	EnableVirtualInterfaces bool
	IgnoreProtocolStats     bool
	Tags                    map[string]string

	semStop        *cliutils.Sem
	collectCache   []*point.Point
	feeder         dkio.Feeder
	lastStats      map[string]psNet.IOCountersStat
	lastProtoStats []psNet.ProtoCountersStat
	mergedTags     map[string]string
	tagger         datakit.GlobalTagger

	ptsTime,
	lastTime time.Time
	netIO            NetIO
	netProto         NetProto
	netVirtualIfaces NetVirtualIfaces
}

func (ipt *Input) Run() {
	ipt.setup()

	tick := time.NewTicker(ipt.Interval)
	defer tick.Stop()

	ipt.ptsTime = ntp.Now()

	for {
		start := time.Now()
		if err := ipt.collect(); err != nil {
			l.Errorf("collect: %s", err)
			ipt.feeder.FeedLastError(err.Error(),
				metrics.WithLastErrorInput(inputName),
				metrics.WithLastErrorCategory(point.Metric),
			)
		}

		if len(ipt.collectCache) > 0 {
			if err := ipt.feeder.Feed(point.Metric, ipt.collectCache,
				dkio.WithCollectCost(time.Since(start)),
				dkio.WithSource(metricName)); err != nil {
				ipt.feeder.FeedLastError(err.Error(),
					metrics.WithLastErrorInput(inputName),
					metrics.WithLastErrorCategory(point.Metric),
				)
				l.Errorf("feed measurement: %s", err)
			}
		}

		select {
		case tt := <-tick.C:
			ipt.ptsTime = inputs.AlignTime(tt, ipt.ptsTime, ipt.Interval)
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
	ipt.collectCache = make([]*point.Point, 0)
	ts := time.Now()
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(ipt.ptsTime))

	netio, err := NetIOCounters()
	if err != nil {
		return fmt.Errorf("error getting net io info: %w", err)
	}
	interfaces, err := net.Interfaces()
	if err != nil {
		return fmt.Errorf("error getting net interfaces info: %w", err)
	}

	filteredInterface, err := FilterInterface(netio,
		interfaces,
		ipt.Interfaces,
		ipt.EnableVirtualInterfaces,
		ipt.netVirtualIfaces)

	for name, ioStat := range filteredInterface {
		var kvs point.KVs

		kvs = kvs.SetTag("interface", ioStat.Name)
		kvs = kvs.Set("bytes_sent", ioStat.BytesSent)
		kvs = kvs.Set("bytes_recv", ioStat.BytesRecv)
		kvs = kvs.Set("packets_sent", ioStat.PacketsSent)
		kvs = kvs.Set("packets_recv", ioStat.PacketsRecv)
		kvs = kvs.Set("err_in", ioStat.Errin)
		kvs = kvs.Set("err_out", ioStat.Errout)
		kvs = kvs.Set("drop_in", ioStat.Dropin)
		kvs = kvs.Set("drop_out", ioStat.Dropout)

		if ipt.lastStats != nil {
			if lastIOStat, ok := ipt.lastStats[name]; ok {
				if ioStat.BytesSent >= lastIOStat.BytesSent && ts.Unix() > ipt.lastTime.Unix() {
					kvs = kvs.Set("bytes_sent/sec", int64(ioStat.BytesSent-lastIOStat.BytesSent)/(ts.Unix()-ipt.lastTime.Unix()))
					kvs = kvs.Set("bytes_recv/sec", int64(ioStat.BytesRecv-lastIOStat.BytesRecv)/(ts.Unix()-ipt.lastTime.Unix()))
					kvs = kvs.Set("packets_sent/sec", int64(ioStat.PacketsSent-lastIOStat.PacketsSent)/(ts.Unix()-ipt.lastTime.Unix()))
					kvs = kvs.Set("packets_recv/sec", int64(ioStat.PacketsRecv-lastIOStat.PacketsRecv)/(ts.Unix()-ipt.lastTime.Unix()))
				}
			}
		}

		for k, v := range ipt.mergedTags {
			kvs = kvs.AddTag(k, v)
		}

		ipt.collectCache = append(ipt.collectCache, point.NewPoint(inputName, kvs, opts...))
	}

	// Get system wide stats for network protocols TCP and UDP(Only supports linux)
	if !ipt.IgnoreProtocolStats {
		netprotos, _ := ipt.netProto([]string{"tcp", "udp"}) // TCP/UDP only
		fields := make(map[string]interface{})
		for _, proto := range netprotos {
			for stat, value := range proto.Stats {
				name := fmt.Sprintf("%s_%s", strings.ToLower(proto.Protocol), strings.ToLower(stat))
				fields[name] = value
			}
		}

		for _, proto := range ipt.lastProtoStats {
			pname := strings.ToLower(proto.Protocol)
			for stat, value := range proto.Stats {
				sname := strings.ToLower(stat)
				if _, ok := linuxProtoRate[sname]; ok {
					if v, ok := fields[pname+"_"+sname]; ok && v.(int64) >= value && ts.Unix() > ipt.lastTime.Unix() {
						fields[pname+"_"+sname+"/sec"] = (v.(int64) - value) / (ts.Unix() - ipt.lastTime.Unix())
					}
				}
			}
		}

		if len(fields) > 0 {
			var kvs point.KVs

			for k, v := range fields {
				kvs = kvs.Set(k, v)
			}

			kvs = kvs.SetTag("interface", "all")
			for k, v := range ipt.mergedTags {
				kvs = kvs.AddTag(k, v)
			}

			ipt.collectCache = append(ipt.collectCache, point.NewPoint(inputName, kvs))
		}

		ipt.lastProtoStats = netprotos
	}
	ipt.lastStats = filteredInterface
	ipt.lastTime = ts
	return err
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
		{FieldName: "IgnoreProtocolStats", Type: doc.Boolean, Default: `false`, Desc: "Ignore reporting of protocol metrics", DescZh: "跳过协议度量的报告"},
		{FieldName: "EnableVirtualInterfaces", Type: doc.Boolean, Default: `false`, Desc: "Enable collect virtual interfaces stats for Linux", DescZh: "采集 Linux 的虚拟网卡"},
		{FieldName: "Interfaces", Type: doc.List, Example: `eth[\w-]+,lo`, Desc: "Expected interfaces (regular)", DescZh: "期望采集的网卡（正则）"},
		{FieldName: "Tags"},
	}

	return doc.SetENVDoc("ENV_INPUT_NET_", infos)
}

// ReadEnv , support envs：
//
//	ENV_INPUT_NET_IGNORE_PROTOCOL_STATS : booler
//	ENV_INPUT_NET_ENABLE_VIRTUAL_INTERFACES : booler
//	ENV_INPUT_NET_TAGS : "a=b,c=d"
//	ENV_INPUT_NET_INTERVAL : time.Duration
//	ENV_INPUT_NET_INTERFACES : []string
func (ipt *Input) ReadEnv(envs map[string]string) {
	if ignore, ok := envs["ENV_INPUT_NET_IGNORE_PROTOCOL_STATS"]; ok {
		b, err := strconv.ParseBool(ignore)
		if err != nil {
			l.Warnf("parse ENV_INPUT_NET_IGNORE_PROTOCOL_STATS to bool: %s, ignore", err)
		} else {
			ipt.IgnoreProtocolStats = b
		}
	}

	if enable, ok := envs["ENV_INPUT_NET_ENABLE_VIRTUAL_INTERFACES"]; ok {
		b, err := strconv.ParseBool(enable)
		if err != nil {
			l.Warnf("parse ENV_INPUT_NET_ENABLE_VIRTUAL_INTERFACES to bool: %s, ignore", err)
		} else {
			ipt.EnableVirtualInterfaces = b
		}
	}

	if tagsStr, ok := envs["ENV_INPUT_NET_TAGS"]; ok {
		tags := config.ParseGlobalTags(tagsStr)
		for k, v := range tags {
			ipt.Tags[k] = v
		}
	}

	//   ENV_INPUT_NET_INTERVAL : time.Duration
	//   ENV_INPUT_NET_INTERFACES : []string
	if str, ok := envs["ENV_INPUT_NET_INTERVAL"]; ok {
		da, err := time.ParseDuration(str)
		if err != nil {
			l.Warnf("parse ENV_INPUT_NET_INTERVAL to time.Duration: %s, ignore", err)
		} else {
			ipt.Interval = config.ProtectedInterval(minInterval,
				maxInterval,
				da)
		}
	}

	if str, ok := envs["ENV_INPUT_NET_INTERFACES"]; ok {
		arrays := strings.Split(str, ",")
		l.Debugf("add ENV_INPUT_NET_INTERFACES from ENV: %v", arrays)
		ipt.Interfaces = append(ipt.Interfaces, arrays...)
	}
}

func defaultInput() *Input {
	return &Input{
		netIO:            NetIOCounters,
		netProto:         psNet.ProtoCounters,
		netVirtualIfaces: NetVirtualInterfaces,
		Interval:         time.Second * 10,

		semStop: cliutils.NewSem(),
		Tags:    make(map[string]string),
		feeder:  dkio.DefaultFeeder(),
		tagger:  datakit.DefaultGlobalTagger(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
