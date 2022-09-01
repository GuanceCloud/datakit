// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2022-present Guance, Inc.

// Package netstat collects host netstat metrics.
package netstat

import (
	"fmt"
	"runtime"
	"syscall"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var _ inputs.ReadEnv = (*Input)(nil)

const (
	minInterval = time.Second
	maxInterval = time.Minute
)

const (
	inputName  = "netstat"
	metricName = inputName
	// conf File samples, reflected in the document.
	sampleCfg = `
[[inputs.netstat]]
  ##(optional) collect interval, default is 10 seconds
  interval = '10s'

[inputs.netstat.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"`
)

var l = logger.DefaultSLogger(inputName)

type Input struct {
	Interval       datakit.Duration
	Tags           map[string]string // Indicator name
	collectCache   []inputs.Measurement
	platform       string
	netConnections NetConnections // A function Type, the instance of Input calls the function that implements this function type through this property.
	semStop        *cliutils.Sem  // start stop signal
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
		},

		Tags: map[string]interface{}{
			"host": &inputs.TagInfo{Desc: "主机名"},
		},
	}
}

// Collect Get, Aggregate, Calculate Data.
func (ipt *Input) Collect() error {
	ipt.collectCache = make([]inputs.Measurement, 0)
	// get data
	netConns, err := ipt.netConnections()
	if err != nil {
		return fmt.Errorf("error getting virtual memory info: %w", err)
	}
	counts := make(map[string]int) // sum of every indicator
	counts["UDP"] = 0

	// count every indicator
	for _, netConn := range netConns {
		if netConn.Type == syscall.SOCK_DGRAM {
			counts["UDP"]++
			continue // UDP has no status
		}
		c, ok := counts[netConn.Status]
		if !ok {
			counts[netConn.Status] = 0
		}
		counts[netConn.Status] = c + 1
	}

	fields := map[string]interface{}{
		"tcp_established": counts["ESTABLISHED"],
		"tcp_syn_sent":    counts["SYN_SENT"],
		"tcp_syn_recv":    counts["SYN_RECV"],
		"tcp_fin_wait1":   counts["FIN_WAIT1"],
		"tcp_fin_wait2":   counts["FIN_WAIT2"],
		"tcp_time_wait":   counts["TIME_WAIT"],
		"tcp_close":       counts["CLOSE"],
		"tcp_close_wait":  counts["CLOSE_WAIT"],
		"tcp_last_ack":    counts["LAST_ACK"],
		"tcp_listen":      counts["LISTEN"],
		"tcp_closing":     counts["CLOSING"],
		"tcp_none":        counts["NONE"],
		"udp_socket":      counts["UDP"],
	}

	tags := map[string]string{}
	for k, v := range ipt.Tags {
		tags[k] = v
	}

	// Append to the cache, the Run() function will handle it
	ipt.collectCache = append(ipt.collectCache, &netStatMeasurement{
		name:   inputName,
		tags:   tags,
		fields: fields,
	})
	return err
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

		// If there is data in the cache, submit it
		if len(ipt.collectCache) > 0 {
			if err := inputs.FeedMeasurement(metricName, datakit.Metric, ipt.collectCache,
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

// ReadEnv support envs：only for K8S
//   ENV_INPUT_NETSTAT_TAGS : "a=b,c=d"
//   ENV_INPUT_NETSTAT_INTERVAL : datakit.Duration
func (ipt *Input) ReadEnv(envs map[string]string) {
	if tagsStr, ok := envs["ENV_INPUT_NETSTAT_TAGS"]; ok {
		tags := config.ParseGlobalTags(tagsStr)
		for k, v := range tags {
			ipt.Tags[k] = v
		}
	}

	//   ENV_INPUT_NETSTAT_INTERVAL : datakit.Duration
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
}

func newDefaultInput() *Input {
	ipt := &Input{
		netConnections: GetNetConnections,
		platform:       runtime.GOOS,
		Interval:       datakit.Duration{Duration: time.Second * 10},
		semStop:        cliutils.NewSem(),
		Tags:           make(map[string]string),
	}
	return ipt
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return newDefaultInput()
	})
}
