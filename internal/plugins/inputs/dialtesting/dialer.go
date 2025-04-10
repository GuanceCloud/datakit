// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !windows
// +build !windows

package dialtesting

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"

	dt "github.com/GuanceCloud/cliutils/dialtesting"
	_ "github.com/go-ping/ping"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const LabelDF = "df_label"

type dialer struct {
	task                 dt.ITask
	ipt                  *Input
	ticker               *time.Ticker
	initTime             time.Time
	dialingTime          time.Time // time to run dialtesting test
	taskExecTimeInterval time.Duration
	testCnt              int64
	class                string
	tags                 map[string]string
	dfTags               map[string]string // tags from df_label
	category             string
	regionName           string
	measurementInfo      *inputs.MeasurementInfo
	seqNumber            int64 // the number of test has been executed
	failCnt              int
	variablePos          int64

	updateCh chan dt.ITask
	done     <-chan interface{} // input exit signal
	stopCh   chan interface{}   // dialer stop signal
}

func (d *dialer) updateTask(t dt.ITask) error {
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
	d.task.Stop()
}

// exit stop the dialer.
func (d *dialer) exit() {
	d.stop()
	close(d.stopCh)
}

// populateDFLabelTags populate df_label tags.
//
// label format: ["v1","k1:v2","v2"].
//
// or old version format: "v1,v2", which is deprecated.
func populateDFLabelTags(label string, tags map[string]string) {
	if tags == nil {
		return
	}

	// treat empty label as []
	if label == "" {
		tags[LabelDF] = "[]"
		return
	}

	isOldLabel := true
	labels := []string{}

	if strings.HasPrefix(label, "[") && strings.HasSuffix(label, "]") {
		isOldLabel = false
	}

	if isOldLabel {
		labels = strings.Split(label, ",")
		if jsonLabel, err := json.Marshal(labels); err != nil {
			l.Warnf("failed to marshal label %s to json: %s", label, err.Error())
		} else {
			label = string(jsonLabel)
		}
	} else if err := json.Unmarshal([]byte(label), &labels); err != nil {
		l.Warnf("failed to unmarshal label %s to json: %s", label, err.Error())
	}

	tags[LabelDF] = label

	for _, l := range labels {
		ls := strings.SplitN(l, ":", 2)
		if len(ls) == 2 {
			k := strings.TrimSpace(ls[0])
			v := strings.TrimSpace(ls[1])
			if k == "" || k == LabelDF || v == "" {
				continue
			}

			tags[k] = v
		}
	}
}

func newDialer(t dt.ITask, ipt *Input) *dialer {
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
	for k, v := range ipt.RegionTags {
		tags[k] = v
	}

	dfTags := make(map[string]string)
	populateDFLabelTags(t.GetDFLabel(), dfTags)

	return &dialer{
		task:                 t,
		updateCh:             make(chan dt.ITask),
		initTime:             time.Now(),
		tags:                 tags,
		dfTags:               dfTags,
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
	_, vars := d.ipt.variables.getVariables(d.task.GetGlobalVars())
	if err := d.task.RenderTemplateAndInit(vars); err != nil {
		return fmt.Errorf("task render template error: %w", err)
	}

	taskInterval, err := time.ParseDuration(d.task.GetFrequency())
	if err != nil {
		return fmt.Errorf("invalid task frequency(%s): %w", d.task.GetFrequency(), err)
	}

	if err = d.resetTicker(d.task.GetFrequency()); err != nil {
		return fmt.Errorf("get ticker error: %w", err)
	}

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
		taskDatawaySendFailedGauge.WithLabelValues(d.regionName, d.class).Set(float64(failCount))

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

			// run task
			// variable task and variable pos changed
			pos, vars := d.ipt.variables.getVariables(d.task.GetGlobalVars())
			if pos > d.variablePos {
				if err := d.task.RenderTemplateAndInit(vars); err != nil {
					l.Warnf("task reset and run error: %s", err.Error())
				} else {
					d.variablePos = pos
				}
			}

			if err := d.task.Run(); err != nil {
				l.Warnf("task run error: %s", err.Error())
				goto wait
			}

			// update global variables
			if d.ipt.isServerMode {
				vars := d.ipt.variables.getVariablesByTask(d.task)

				for _, v := range vars {
					if value, err := d.task.GetVariableValue(v); err != nil {
						l.Warnf("get variable value failed: %s", err.Error())
						d.ipt.variables.updateVariableValue(v, value, 1)
					} else {
						l.Debugf("set variable %s value: %s", v.Name, value)
						d.ipt.variables.updateVariableValue(v, value, 0)
					}
				}
			}

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
			if err := d.doUpdateTask(t); err != nil {
				d.stop()
				l.Errorf("update task %s failed: %s, stopped", d.task.ID(), err.Error())
			}

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

			d.dfTags = make(map[string]string)
			// update df_label
			populateDFLabelTags(t.GetDFLabel(), d.dfTags)

			if err := d.checkInternalNetwork(); err != nil {
				return err
			}
		}
	}
}

// checkInternalNetwork check whether the host is allowed to be tested.
func (d *dialer) checkInternalNetwork() error {
	hostNames, err := d.task.GetHostName()
	for _, hostName := range hostNames {
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
	}

	return nil
}

func (d *dialer) feedIO() error {
	u, err := url.Parse(d.task.PostURLStr())
	if err != nil {
		l.Warn("get invalid url, ignored")
		return err
	}
	u.Path = path.Join(u.Path, datakit.Logging) // `/v1/write/logging`
	urlStr := u.String()

	switch d.task.Class() {
	case dt.ClassHTTP, dt.ClassTCP, dt.ClassICMP, dt.ClassWebsocket, dt.ClassMulti:
		d.category = urlStr
		d.pointsFeed(urlStr)
	case dt.ClassHeadless:
		return fmt.Errorf("headless task deprecated")
	default:
		// TODO other class
	}

	return nil
}

func (d *dialer) doUpdateTask(t dt.ITask) error {
	_, vars := d.ipt.variables.getVariables(d.task.GetGlobalVars())
	if err := t.RenderTemplateAndInit(vars); err != nil {
		return fmt.Errorf("render template and init error: %w", err)
	}

	if d.task.GetFrequency() != t.GetFrequency() {
		if err := d.resetTicker(t.GetFrequency()); err != nil {
			return fmt.Errorf("reset ticker error: %w", err)
		}
	}

	d.task = t
	return nil
}

func (d *dialer) resetTicker(frequency string) error {
	du, err := time.ParseDuration(frequency)
	if err != nil {
		return err
	}

	if d.ticker != nil {
		d.ticker.Stop()
	}

	d.ticker = time.NewTicker(du)

	return nil
}
