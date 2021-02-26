package dialtesting

import (
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
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
			l.Info("dial testing %s exit", d.task.ID())

		case <-d.ticker.C:

			d.testCnt++
			if err := d.task.Run(); err != nil {
				// ignore
			} else {
				reasons := d.task.CheckResult()
				// TODO: post result to d.PostURL
				l.Debugf("reasons: %+#v", reasons)
				_ = reasons
			}

		case t := <-d.updateCh:
			d.doUpdateTask(t)

			if d.task.Status() == dt.StatusStop {
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
