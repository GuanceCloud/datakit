package dialtesting

import (
	"fmt"
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

func newDialer(t dt.Task) (*dialer, error) {

	return &dialer{
		task: t,

		updateCh: make(chan dt.Task),
		initTime: time.Now(),
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

			// 获取此次任务执行的基本信息
			tags := map[string]string{}
			fields := map[string]interface{}{}
			tags, fields = d.task.GetResults()

			reasons := d.task.CheckResult()
			if len(reasons) != 0 {
				fields[`failed_reason`] = strings.Join(reasons, `;`)
			}

			if _, ok := fields[`failed_reason`]; !ok {
				tags["result"] = "OK"
				fields["success"] = int64(1)
			}

			pt, err := io.MakePoint(d.task.MetricName(), tags, fields, time.Now())
			if err != nil {
				l.Warnf("io feed failed, %s", err.Error())
			} else {
				if err := io.Feed(inputName,
					io.Metric,
					[]*io.Point{pt},
					&io.Option{HTTPHost: d.task.PostURLStr()}); err != nil {
					l.Warnf("io feed failed, %s", err.Error())
				}
			}

			l.Debugf(`url:%s, tags: %+#v, fs: %+#v`, d.task.PostURLStr(), tags, fields)

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

func (d *dialer) doUpdateTask(t dt.Task) {

	if err := t.Init(); err != nil {
		l.Warn(err)
		return
	}

	d.task = t
	d.ticker = t.Ticker() // update ticker
}
