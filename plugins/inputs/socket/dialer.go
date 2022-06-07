package socket

import (
	"time"

	_ "github.com/go-ping/ping"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/dataway"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type dialer struct {
	task Task

	initTime time.Time
	class    string

	tags map[string]string

	collectCache inputs.Measurement
}

func newDialer(t Task, ts map[string]string) *dialer {
	return &dialer{
		task:     t,
		initTime: time.Now(),
		tags:     ts,
		class:    t.Class(),
	}
}

func (d *dialer) run() {
	_ = d.task.Run() //nolint:errcheck
	// 无论成功或失败，都要记录测试结果
	d.feedMeasurement()
}

func (d *dialer) feedMeasurement() {
	tags, fields := d.task.GetResults()
	ts := time.Now()
	for k, v := range d.tags {
		if _, ok := tags[k]; !ok {
			tags[k] = v
		} else {
			l.Warnf("ignore input socket tag %s: %s", k, v)
		}
	}
	tmp := &TCPMeasurement{name: "tcp", tags: tags, fields: fields, ts: ts}
	d.collectCache = tmp
}
