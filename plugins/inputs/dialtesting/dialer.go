package dialtesting

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
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
			d.task.Run()
			err := d.feedIo()
			if err != nil {
				l.Warnf("io feed failed, %s", err.Error())
			}

		case t := <-d.updateCh:
			d.doUpdateTask(t)

			if strings.ToLower(d.task.Status()) == dt.StatusStop {
				if err := t.Stop(); err != nil {
					l.Warnf("stop task failed: %s", err.Error())
				}

				l.Info("task %s stopped", d.task.ID())
				return nil
			}
		}
	}

	return nil
}

func (d *dialer) feedIo() error {
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

	// 考虑到推送至不同的dataway地址
	u, err := url.Parse(d.task.PostURLStr())
	if err != nil {
		l.Warn("get invalid url, ignored")
		return err
	}

	u.Path = u.Path + io.Logging // `/v1/write/logging`

	// pts := []*io.Point{}
	// pts = append(pts, data)
	err = Feed(inputName, io.Logging, data, &io.Option{
		HTTPHost: u.String(),
	})

	l.Debugf(`url:%s, tags: %+#v, fs: %+#v`, u.String(), tags, fields)

	return err
}

func (d *dialer) doUpdateTask(t dt.Task) {

	if err := t.Init(); err != nil {
		l.Warn(err)
		return
	}

	d.task = t
	d.ticker = t.Ticker() // update ticker
}
