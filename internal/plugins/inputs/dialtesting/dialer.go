// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !windows
// +build !windows

package dialtesting

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	dt "github.com/GuanceCloud/cliutils/dialtesting"
	_ "github.com/go-ping/ping"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type dialer struct {
	task                 dt.Task
	ipt                  *Input
	ticker               *time.Ticker
	initTime             time.Time
	dialingTime          time.Time // time to run dialtesting test
	taskExecTimeInterval time.Duration
	testCnt              int64
	class                string
	tags                 map[string]string
	category             string
	regionName           string
	measurementInfo      *inputs.MeasurementInfo
	seqNumber            int64 // the number of test has been executed
	failCnt              int

	updateCh chan dt.Task
	done     <-chan interface{} // input exit signal
	stopCh   chan interface{}   // dialer stop signal
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
	if err := d.task.Stop(); err != nil {
		l.Warnf("stop task %s failed: %s", d.task.ID(), err.Error())
	}
}

// exit stop the dialer.
func (d *dialer) exit() {
	d.stop()
	close(d.stopCh)
}

func newDialer(t dt.Task, ipt *Input) *dialer {
	var info *inputs.MeasurementInfo
	switch t.Class() {
	case dt.ClassHTTP:
		info = (&httpMeasurement{}).Info()
	case dt.ClassTCP:
		info = (&tcpMeasurement{}).Info()
	case dt.ClassICMP:
		info = (&icmpMeasurement{}).Info()
	case dt.ClassWebsocket:
		info = (&websocketMeasurement{}).Info()
	}

	tags := make(map[string]string)
	for k, v := range ipt.Tags {
		tags[k] = v
	}
	if t.GetTagsInfo() != "" {
		tags["tags_info"] = t.GetTagsInfo()
	}

	return &dialer{
		task:                 t,
		updateCh:             make(chan dt.Task),
		initTime:             time.Now(),
		tags:                 tags,
		measurementInfo:      info,
		class:                t.Class(),
		taskExecTimeInterval: ipt.taskExecTimeInterval,
		ipt:                  ipt,
		stopCh:               make(chan interface{}),
	}
}

func (d *dialer) getSendFailCount() int {
	if d.category != "" {
		return dialWorker.getFailCount(d.category)
	}
	return 0
}

func (d *dialer) run() error {
	var maskURL string

	if err := d.task.Init(); err != nil {
		l.Errorf(`task init error: %s`, err.Error())
		return err
	}

	taskInterval, err := time.ParseDuration(d.task.GetFrequency())
	if err != nil {
		return fmt.Errorf("invalid task frequency(%s): %w", d.task.GetFrequency(), err)
	}

	d.ticker = d.task.Ticker()

	taskGauge.WithLabelValues(d.regionName, d.class).Inc()

	defer func() {
		taskGauge.WithLabelValues(d.regionName, d.class).Dec()
	}()

	l.Debugf("dialer: %+#v", d)

	defer d.ticker.Stop()
	defer close(d.updateCh)

	if parts, err := url.Parse(d.task.PostURLStr()); err != nil {
		taskInvalidCounter.WithLabelValues(d.regionName, d.class, "invalid_post_url").Inc()
		return fmt.Errorf("invalid post url")
	} else {
		params := parts.Query()
		if tokens, ok := params["token"]; ok {
			// check token
			if len(tokens) >= 1 {
				maskURL = getMaskURL(parts.String())
				if isValid, err := dialWorker.sender.checkToken(tokens[0], parts.Scheme, parts.Host); err != nil {
					l.Warnf("check token error: %s", err.Error())
				} else if !isValid {
					taskInvalidCounter.WithLabelValues(d.regionName, d.class, "invalid_token").Inc()
					return fmt.Errorf("invalid token")
				}
			} else {
				taskInvalidCounter.WithLabelValues(d.regionName, d.class, "token_empty").Inc()
				return fmt.Errorf("token is required")
			}
		} else {
			taskInvalidCounter.WithLabelValues(d.regionName, d.class, "token_empty").Inc()
			return fmt.Errorf("token is required")
		}
	}

	if err := d.checkInternalNetwork(); err != nil {
		return err
	}

	isSleep := false
	// init sleep timer
	sleepTimer := time.NewTimer(0)
	sleepTimer.Stop()
	defer sleepTimer.Stop()

	for {
		failCount := d.getSendFailCount()
		taskDatawaySendFailedGauge.WithLabelValues(d.regionName, d.class, maskURL).Set(float64(failCount))

		// exceed max send fail count, sleep for MaxSendFailSleepTime
		if failCount > MaxSendFailCount {
			if isSleep {
				goto wait
			}
			isSleep = true
			sleepTimer.Reset(d.ipt.MaxSendFailSleepTime.Duration)
			taskInvalidCounter.WithLabelValues(d.regionName, d.class, "exceed_max_failure_count").Inc()
			l.Warnf("dial testing %s send data failed %d times", d.task.ID(), failCount)
		}

		l.Debugf(`dialer run %+#v, fail count: %d`, d, failCount)
		d.testCnt++

		switch d.task.Class() {
		case dt.ClassHeadless:
			return fmt.Errorf("headless task deprecated")
		default:
			now := time.Now()
			if !d.dialingTime.IsZero() {
				lastDialingDuration := now.Sub(d.dialingTime)
				interval := lastDialingDuration - taskInterval
				if interval > d.taskExecTimeInterval {
					taskExecTimeIntervalSummary.WithLabelValues(d.regionName, d.class).Observe(float64(interval) / float64(time.Second))
				}
			}
			d.dialingTime = now
			_ = d.task.Run() //nolint:errcheck
			taskRunCostSummary.WithLabelValues(d.regionName, d.class).Observe(float64(time.Since(d.dialingTime)) / float64(time.Second))
			// dialtesting start
			err := d.feedIO()
			if err != nil {
				l.Warnf("io feed failed, %s", err.Error())
			}
		}

	wait:
		select {
		case <-datakit.Exit.Wait():
			l.Infof("dial testing %s exit", d.task.ID())
			return nil

		case <-d.done:
			l.Infof("dial testing %s exit", d.task.ID())
			return nil

		case <-d.ticker.C:

		case <-d.stopCh:
			l.Infof("stop dial testing %s, exit", d.task.ID())
			return nil
		case <-sleepTimer.C:
			isSleep = false
			sleepTimer.Stop()
			goto wait
		case t := <-d.updateCh:
			d.doUpdateTask(t)

			if strings.ToLower(d.task.Status()) == dt.StatusStop {
				d.stop()
				l.Info("task %s stopped", d.task.ID())
				return nil
			}
			// update regionName
			if t.GetWorkspaceLanguage() == "en" && d.ipt.regionNameEn != "" {
				d.regionName = d.ipt.regionNameEn
			} else {
				d.regionName = d.ipt.regionName
			}

			// update tags_info
			d.tags["tags_info"] = t.GetTagsInfo()

			if err := d.checkInternalNetwork(); err != nil {
				return err
			}
		}
	}
}

// checkInternalNetwork check whether the host is allowed to be tested.
func (d *dialer) checkInternalNetwork() error {
	hostName, err := d.task.GetHostName()
	if err != nil {
		l.Warnf("get host name error: %s", err.Error())
		return fmt.Errorf("get host name error: %w", err)
	} else if d.ipt.DisableInternalNetworkTask {
		if isInternal, err := httpapi.IsInternalHost(hostName, d.ipt.DisabledInternalNetworkCIDRList); err != nil {
			taskInvalidCounter.WithLabelValues(d.regionName, d.class, "host_not_valid").Inc()
			return fmt.Errorf("dest host is not valid: %w", err)
		} else if isInternal {
			taskInvalidCounter.WithLabelValues(d.regionName, d.class, "host_not_allowed").Inc()
			return fmt.Errorf("dest host [%s] is not allowed to be tested", hostName)
		}
	}

	return nil
}

func (d *dialer) feedIO() error {
	u, err := url.Parse(d.task.PostURLStr())
	if err != nil {
		l.Warn("get invalid url, ignored")
		return err
	}

	u.Path += datakit.Logging // `/v1/write/logging`

	urlStr := u.String()
	switch d.task.Class() {
	case dt.ClassHTTP, dt.ClassTCP, dt.ClassICMP, dt.ClassWebsocket:
		d.category = urlStr
		d.pointsFeed(urlStr)
	case dt.ClassHeadless:
		return fmt.Errorf("headless task deprecated")
	default:
		// TODO other class
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
