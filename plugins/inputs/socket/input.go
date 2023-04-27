// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package socket collect socket metrics
package socket

import (
	"fmt"
	"net/url"
	"runtime"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	clipt "github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func (i *Input) SampleConfig() string {
	return sample
}

func (i *Input) Catalog() string {
	return "socket"
}

func (i *Input) Run() {
	l = logger.SLogger(inputName)
	l.Infof("socket input started")
	i.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, i.Interval.Duration)
	tick := time.NewTicker(i.Interval.Duration)
	defer tick.Stop()

	for {
		i.collectCache = make([]inputs.Measurement, 0)

		start := time.Now()
		if err := i.Collect(); err != nil {
			l.Errorf("Collect: %s", err)
			io.FeedLastError(inputName, err.Error(), clipt.Metric)
		}

		if len(i.collectCache) > 0 {
			if err := inputs.FeedMeasurement(metricName,
				datakit.Metric,
				i.collectCache,
				&io.Option{CollectCost: time.Since((start))}); err != nil {
				l.Errorf("FeedMeasurement: %s", err)
			}
		}

		select {
		case <-tick.C:
		case <-datakit.Exit.Wait():
			l.Infof("socket input exit")
			return

		case <-i.semStop.Wait():
			l.Infof("socket input return")
			return
		}
	}
}

func (i *Input) Terminate() {
	if i.semStop != nil {
		i.semStop.Close()
	}
}

func (*Input) AvailableArchs() []string {
	return []string{datakit.OSLabelLinux, datakit.OSLabelMac}
}

func (i *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&TCPMeasurement{},
		&UDPMeasurement{},
	}
}

func (i *Input) Collect() error {
	if len(i.DestURL) == 0 {
		l.Warnf("input socket have no desturl")
	}
	for _, cont := range i.DestURL {
		resURL, err := url.Parse(cont)
		if err != nil {
			return fmt.Errorf("inpust socket parse dest_url error %w", err)
		}

		switch resURL.Scheme {
		case TCP:
			err := i.CollectTCP(resURL.Hostname(), resURL.Port())
			if err != nil {
				return err
			}
		case UDP:
			err := i.CollectUDP(resURL.Hostname(), resURL.Port())
			if err != nil {
				return err
			}
		default:
			l.Warnf("input socket can not support proto : %s", resURL.Scheme)
		}
	}
	return nil
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			Interval:   datakit.Duration{Duration: time.Second * 30},
			semStop:    cliutils.NewSem(),
			platform:   runtime.GOOS,
			UDPTimeOut: datakit.Duration{Duration: time.Second * 10},
			TCPTimeOut: datakit.Duration{Duration: time.Second * 10},
		}
	})
}

func (i *Input) CollectTCP(destHost string, destPort string) error {
	t := &TCPTask{}
	t.Host = destHost
	t.Port = destPort
	t.timeout = i.TCPTimeOut.Duration
	if err := i.runTCP(t); err != nil {
		return err
	}
	return nil
}
