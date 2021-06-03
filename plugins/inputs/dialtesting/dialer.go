package dialtesting

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	dt "gitlab.jiagouyun.com/cloudcare-tools/kodo/dialtesting"
)

type dialer struct {
	task dt.Task

	ticker *time.Ticker

	initTime time.Time
	testCnt  int64
	class    string

	tags     map[string]string
	updateCh chan dt.Task

	failCnt int
}

func (d *dialer) updateTask(t dt.Task) error {

	select {
	case <-d.updateCh: // if closed?
		l.Warnf("task %s closed", d.task.ID())
		return fmt.Errorf("task exited")
	default:
		d.updateCh <- t
		return nil
	}
}

func (d *dialer) stop() {
	close(d.updateCh)
	if err := d.task.Stop(); err != nil {
		l.Warnf("stop task %s failed: %s", d.task.ID(), err.Error())
	}
}

func newDialer(t dt.Task, ts map[string]string) (*dialer, error) {

	return &dialer{
		task: t,

		updateCh: make(chan dt.Task),
		initTime: time.Now(),
		tags:     ts,
		class:    t.Class(),
	}, nil
}

func (d *dialer) run() error {
	d.ticker = d.task.Ticker()

	l.Debugf("dialer: %+#v", d)

	defer d.ticker.Stop()
	defer close(d.updateCh)

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Infof("dial testing %s exit", d.task.ID())
			return nil

		case <-d.ticker.C:

			d.testCnt++
			//dialtesting start
			//无论成功或失败，都要记录测试结果
			err := d.task.Run()
			if err != nil {
				l.Errorf("task %s failed, %s", d.task.ID(), err.Error())
			}

			err = d.feedIo()
			if err != nil {
				l.Warnf("io feed failed, %s", err.Error())
			}

		case t := <-d.updateCh:
			d.doUpdateTask(t)

			if strings.ToLower(d.task.Status()) == dt.StatusStop {
				d.stop()
				l.Info("task %s stopped", d.task.ID())
				return nil
			}
		}
	}
}

func (d *dialer) feedIo() error {

	// 考虑到推送至不同的dataway地址
	u, err := url.Parse(d.task.PostURLStr())
	if err != nil {
		l.Warn("get invalid url, ignored")
		return err
	}

	u.Path = u.Path + "v1/write/" + datakit.Logging // `/v1/write/logging`

	urlStr := u.String()
	switch d.task.Class() {
	case dt.ClassHTTP:
		return d.pointsFeed(urlStr)
	case dt.ClassHeadless:
		return d.linedataFeed(urlStr, `ms`)
	//TODO other class
	default:
	}

	return nil
}

func (d *dialer) doUpdateTask(t dt.Task) {

	if err := t.Init(); err != nil {
		l.Warn(err)
		return
	}

	if d.task.GetFrequency() != t.GetFrequency() {
		d.ticker = t.Ticker() // update ticker
	}

	d.task = t

}

type httpMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *httpMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *httpMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "http_dial_testing",
		Tags: map[string]interface{}{
			"name":               &inputs.TagInfo{Desc: "示例：拨测名称,百度测试"},
			"url":                &inputs.TagInfo{Desc: "示例 http://wwww.baidu.com"},
			"country":            &inputs.TagInfo{Desc: "示例 中国"},
			"province":           &inputs.TagInfo{Desc: "示例 浙江"},
			"city":               &inputs.TagInfo{Desc: "示例 杭州"},
			"internal":           &inputs.TagInfo{Desc: "示例 true（国内 true /海外 false）"},
			"isp":                &inputs.TagInfo{Desc: "示例 电信/移动/联通"},
			"status":             &inputs.TagInfo{Desc: "示例 OK/FAIL 两种状态 "},
			"status_code_class":  &inputs.TagInfo{Desc: "示例 2xx"},
			"status_code_string": &inputs.TagInfo{Desc: "示例 200 OK"},
			"proto":              &inputs.TagInfo{Desc: "示例 HTTP/1.0"},
		},
		Fields: map[string]interface{}{
			"status_code": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "web page response code",
			},
			"message": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "包括请求头(request_header)/请求体(request_body)/返回头(response_header)/返回体(response_body)/fail_reason 冗余一份",
			},
			"fail_reason": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "拨测失败原因",
			},
			"response_time": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationUS,
				Desc:     "HTTP 相应时间, 单位 ms",
			},
			"response_body_size": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "body 长度",
			},
			"success": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "只有 1/-1 两种状态, 1 表示成功, -1 表示失败",
			},
			"proto": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "示例 HTTP/1.0",
			},
		},
	}
}
