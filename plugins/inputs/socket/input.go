// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package socket collect socket metrics
package socket

import (
	"runtime"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/system/rtpanic"
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
			io.FeedLastError(inputName, err.Error())
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
	return []string{"linux", "darwin"}
}

func (i *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&TCPMeasurement{},
		&UDPMeasurement{},
	}
}

func (i *Input) Collect() error {
	for _, cont := range i.DestURL {
		parseURL := strings.Split(cont, "://")
		if len(parseURL) != 2 {
			io.FeedLastError(inputName, "input socket desturl error:"+cont)
		}
		kv := strings.Split(parseURL[1], ":")
		protoType := parseURL[0]
		if len(kv) != 2 {
			io.FeedLastError(inputName, "input socket desturl error:"+cont)
		}

		switch protoType {
		case TCP:
			err := i.dispatchTasks(kv[0], kv[1])
			if err != nil {
				return err
			}
		case UDP:
			err := i.CollectUDP(kv[0], kv[1])
			if err != nil {
				return err
			}
		default:
			l.Warnf("input socket can not support proto : %s", kv[0])
		}
	}

	for _, c := range i.curTasks {
		i.collectCache = append(i.collectCache, c.collectCache)
	}

	return nil
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			Interval:   datakit.Duration{Duration: time.Second * 30},
			semStop:    cliutils.NewSem(),
			curTasks:   map[string]*dialer{},
			platform:   runtime.GOOS,
			UDPTimeOut: datakit.Duration{Duration: time.Second * 10},
			TCPTimeOut: datakit.Duration{Duration: time.Second * 10},
		}
	})
}

func (i *Input) dispatchTasks(destHost string, destPort string) error {
	t := &TCPTask{}
	t.Host = destHost
	t.Port = destPort
	t.timeout = 10 * time.Second
	dialer, err := i.newTaskRun(t)
	if err != nil {
		l.Errorf(`%s, ignore`, err.Error())
	} else {
		i.curTasks[t.ID()] = dialer
	}
	return nil
}

func (i *Input) newTaskRun(t Task) (*dialer, error) {
	if err := t.Init(); err != nil {
		l.Errorf(`%s`, err.Error())
		return nil, err
	}
	dialer := newDialer(t, i.Tags)

	i.wg.Add(1)
	go func(id string) {
		defer i.wg.Done()
		protectedRun(dialer)
		l.Infof("input %s exited", id)
	}(t.ID())
	return dialer, nil
}

var maxCrashCnt = 6

func protectedRun(d *dialer) {
	crashcnt := 0
	var f rtpanic.RecoverCallback

	l.Infof("task %s(%s) starting...", d.task.ID(), d.class)

	f = func(trace []byte, err error) {
		defer rtpanic.Recover(f, nil)
		if trace != nil {
			l.Warnf("input socket task %s panic: %+#v, trace: %s", d.task.ID(), err, string(trace))

			crashcnt++
			if crashcnt > maxCrashCnt {
				l.Warnf("input socket task %s crashed %d times, exit now", d.task.ID(), crashcnt)
				return
			}
		}
		d.run()
	}

	f(nil, nil)
}
