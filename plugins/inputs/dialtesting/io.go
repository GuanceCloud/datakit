package dialtesting

import (
	"time"

	lp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/lineproto"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

func (d *dialer) linedataFeed(urlStr string, precision string) error {
	data := d.task.GetLineData()
	if len(data) == 0 {
		l.Warnf("no any data for task %s", d.task.ID())
		return nil
	}

	l.Debugf(`task %s, `, d.task.ID())

	tags, _ := d.task.GetResults()
	for k, v := range d.tags {
		tags[k] = v
	}

	pts, err := lp.ParsePoints([]byte(data), &lp.Option{
		Time:      time.Now(),
		Precision: precision,
		ExtraTags: tags,
		Strict:    true,
	})
	if err != nil {
		return err
	}

	x := []*io.Point{}
	for _, pt := range pts {
		x = append(x, &io.Point{Point: pt})
	}

	return Feed(inputName, datakit.Logging, x, &io.Option{
		HTTPHost: urlStr,
	})

}

func (d *dialer) pointsFeed(urlStr string) error {
	// 获取此次任务执行的基本信息
	tags := map[string]string{}
	fields := map[string]interface{}{}
	tags, fields = d.task.GetResults()

	for k, v := range d.tags {
		tags[k] = v
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
