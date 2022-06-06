// Package socket collect socket metrics
package socket

import (
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/system/rtpanic"
	"runtime"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func (i *Input) SampleConfig() string {
	return sample
}

func (i *Input) appendTCPMeasurement(name string, tags map[string]string, fields map[string]interface{}, ts time.Time) {
	tmp := &TCPMeasurement{name: name, tags: tags, fields: fields, ts: ts}
	i.collectCache = append(i.collectCache, tmp)
}

func (i *Input) appendUDPMeasurement(name string, tags map[string]string, fields map[string]interface{}, ts time.Time) {
	tmp := &UDPMeasurement{name: name, tags: tags, fields: fields, ts: ts}
	i.collectCache = append(i.collectCache, tmp)
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
	return datakit.AllArch
}

func (i *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&TCPMeasurement{},
		&UDPMeasurement{},
	}
}

func (i *Input) Collect() error {
	for _, cont := range i.DestUrl {
		kv := strings.Split(cont, ":")
		if len(kv) != 3 {
			io.FeedLastError(inputName, "input socket desturl error:"+cont)
		}
		if kv[0] == TCP {
			err := i.dispatchTasks(kv[1], kv[2])
			if err != nil {
				return err
			}
		} else {
			err := i.CollectUdp(kv[1], kv[2])
			if err != nil {
				return err
			}
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
			Interval: datakit.Duration{Duration: time.Second * 900},
			semStop:  cliutils.NewSem(),
			curTasks: map[string]*dialer{},
			platform: runtime.GOOS,
		}
	})
}

func (i *Input) dispatchTasks(destHost string, destPort string) error {
	t := &TcpTask{}
	t.Host = destHost
	t.Port = destPort
	t.Timeout = "10s"
	t.Frequency = "10s"
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
