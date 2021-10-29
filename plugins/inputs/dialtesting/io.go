package dialtesting

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

func (d *dialer) pointsFeed(urlStr string) error {
	// 获取此次任务执行的基本信息
	tags, fields := d.task.GetResults()

	for k, v := range d.tags {
		if _, ok := tags[k]; !ok {
			tags[k] = v
		} else {
			l.Warnf("ignore dialer tag %s: %s", k, v)
		}
	}

	data, err := io.MakePoint(d.task.MetricName(), tags, fields, time.Now())
	if err != nil {
		l.Warnf("make metric failed: %s", err.Error)
		return err
	}

	pts := []*io.Point{}
	pts = append(pts, data)

	err = Feed(inputName, datakit.Logging, pts, &io.Option{
		HTTPHost: urlStr,
	})

	l.Debugf(`url:%s, tags: %+#v, fs: %+#v`, urlStr, tags, fields)

	return err
}

func Feed(name, category string, pts []*io.Point, opt *io.Option) error {
	return io.Feed(name, category, pts, opt)
}
