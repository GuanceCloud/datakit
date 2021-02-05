package dialtesting

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

type dialer struct {
	task     Task
	taskMd5  string
	taskJson []byte

	ticker *time.Ticker

	lastUpdate time.Time
	initTime   time.Time
	testCnt    int64

	updateCh chan Task
}

func (d *dialer) updateTask(t Task) error {

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

func newDialer(t Task) (*dialer, error) {

	j, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}

	return &dialer{
		task:     t,
		taskJson: j,
		taskMd5:  fmt.Sprintf("%x", md5.Sum(j)),

		updateCh: make(chan Task),
		initTime: time.Now(),
	}, nil
}

func (d *dialer) run() error {
	d.ticker = d.task.Ticker()

	defer d.ticker.Stop()

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
				_ = reasons
			}

		case t := <-d.updateCh:
			d.doUpdateTask(t)
		}
	}

	return nil
}

func (d *dialer) doUpdateTask(t Task) {

	j, err := json.Marshal(t)
	if err != nil {
		l.Warn(err)
		return
	}

	taskMd5 := fmt.Sprintf("%x", md5.Sum(j))
	if taskMd5 == d.taskMd5 {
		return
	}

	if err := t.Init(); err != nil {
		l.Warn(err)
		return
	}

	if t.Status() == StatusStop {
		d.stop()
		return
	}

	if err := d.task.Stop(); err != nil {
		l.Warn(err)
	}

	d.lastUpdate = time.Now()
	d.task = t
	d.ticker = t.Ticker() // update ticker
}
