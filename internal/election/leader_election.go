// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package election

import (
	"encoding/json"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

/*
 * DataKit 选举说明文档
 *
 * 流程：
 *      1. DataKit 开启 cfg.EnableElection（booler）配置
 *      2. 当运行对应的采集器（采集器列表在 config/inputcfg.go）时，程序会创建一个 goroutine 向 DataWay 发送选举请求，并携带 token 和 namespace（若存在）以及 id
 *      3. 选举成功担任 leader 后会持续发送心跳，心跳间隔过长或选举失败，会恢复 candidate 状态并继续发送选举请求
 *      4. 采集器端只要在采集数据时，判断当前是否为 leader 状态，具体使用见下
 *
 * 使用方式：
 *      1. 在 config/inputcfg.go 的 electionInputs 中添加需要选举的采集器（目前使用此方式后续会优化）
 *      2. 采集器中 import "gitlab.jiagouyun.com/cloudcare-tools/datakit/election"
 *      4. 详见 demo 采集器
 */

type leaderElection struct {
	*option

	lastElected time.Time
	status      ElectionStatus
	plugins     []inputs.ElectionInput
}

func newLeaderElection(opt *option, plugins map[string][]inputs.ElectionInput) *leaderElection {
	x := &leaderElection{
		option: opt,
		status: StatusFail,
	}
	for _, v := range plugins {
		x.plugins = append(x.plugins, v...)
	}
	return x
}

func (x *leaderElection) Run() {
	x.pausePlugins()
	tick := time.NewTicker(time.Second * time.Duration(electionIntervalDefault))
	defer tick.Stop()

	for {
		select {
		case <-datakit.Exit.Wait():
			electionInputs.WithLabelValues(x.namespace).Set(float64(len(x.plugins)))
			return

		case s := <-chStatus:
			if s != x.status {
				log.Infof("switched from %s to %s", x.status, s)

				x.status = s
				electionStatusSwitched.WithLabelValues(
					x.namespace,
					x.status.String(),
				).Inc()

				electionStatusVec.WithLabelValues(
					CurrentElected,
					x.id,
					x.namespace,
					x.status.String(),
				).Set(float64(time.Now().Unix()))
			}

		case <-tick.C:
			if electionInterval, err := x.runOnce(); err == nil {
				if electionInterval != electionIntervalDefault {
					tick.Reset(time.Second * time.Duration(electionInterval))
					electionIntervalDefault = electionInterval
				}
			}
		}
	}
}

func (x *leaderElection) runOnce() (int, error) {
	var (
		elecIntv int
		err      error
	)

	switch x.status {
	case StatusSuccess:
		elecIntv, err = x.keepalive()
		if err != nil {
			log.Errorf("keepalive: %s", err)
		}
	case StatusFail:
		elecIntv, err = x.tryElection()
		if err != nil {
			log.Errorf("tryElection: %s", err)
		}

	case StatusDisabled, StatusBanned, StatusImpeached: // pass
		return electionIntervalDefault, nil
	}

	return elecIntv, err
}

type leaderElectionResult struct {
	Content struct {
		Status       string `json:"status"`
		Namespace    string `json:"namespace,omitempty"`
		ID           string `json:"id"`
		IncumbencyID string `json:"incumbency_id,omitempty"`
		ErrorMsg     string `json:"error_msg,omitempty"`
		Interval     int    `json:"interval"`
	} `json:"content"`
}

func (x *leaderElection) tryElection() (int, error) {
	body, err := x.puller.Election(x.namespace, x.id, nil)
	if err != nil {
		log.Errorf("puller.Election: %s", err)
		return electionIntervalDefault, err
	}

	e := leaderElectionResult{}
	if err := json.Unmarshal(body, &e); err != nil {
		log.Error(err)

		return electionIntervalDefault, nil
	}

	log.Debugf("result body: %s", body)

	if CurrentElected != e.Content.IncumbencyID {
		CurrentElected = e.Content.IncumbencyID
	}

	if e.Content.Status != x.status.String() {
		electionStatusSwitched.WithLabelValues(
			x.namespace,
			e.Content.Status,
		).Inc()

		electionStatusVec.WithLabelValues(
			CurrentElected,
			x.id,
			x.namespace,
			e.Content.Status,
		).Set(float64(time.Now().Unix()))
	}

	switch e.Content.Status {
	case StatusFail.String():

		x.status = StatusFail

	case StatusSuccess.String():
		log.Info("elected as leader")

		x.status = StatusSuccess
		x.lastElected = time.Now()
		x.resumePlugins()

	default:
		log.Warnf("unknown election status: %s", e.Content.Status)
	}

	return e.Content.Interval, nil
}

func (x *leaderElection) keepalive() (int, error) {
	body, err := x.puller.ElectionHeartbeat(x.namespace, x.id, nil)
	if err != nil {
		log.Error(err)
		return electionIntervalDefault, err
	}

	e := leaderElectionResult{}
	if err := json.Unmarshal(body, &e); err != nil {
		log.Error(err)
		return electionIntervalDefault, err
	}

	log.Debugf("result body: %s", body)

	if e.Content.Status != x.status.String() {
		electionStatusSwitched.WithLabelValues(
			x.namespace,
			e.Content.Status,
		).Inc()
	}

	CurrentElected = e.Content.IncumbencyID

	switch e.Content.Status {
	case StatusFail.String():
		log.Errorf("election keepalive failed, leader dropped(live: %s)", time.Since(x.lastElected))

		x.status = StatusFail
		x.pausePlugins()

	case StatusSuccess.String():
		log.Debugf("%s election keepalive ok", x.id)

	default:
		log.Warnf("unknown election status: %s", e.Content.Status)
	}
	return e.Content.Interval, nil
}

func (x *leaderElection) pausePlugins() {
	defer func() {
		inputsPauseVec.WithLabelValues(x.id, x.namespace).Add(float64(len(x.plugins)))
	}()

	log.Infof("pause %d inputs...", len(x.plugins))
	for _, p := range x.plugins {
		if err := p.Pause(); err != nil {
			log.Warnf("pause: %s", err)
		}
	}
}

func (x *leaderElection) resumePlugins() {
	defer func() {
		inputsResumeVec.WithLabelValues(x.id, x.namespace).Add(float64(len(x.plugins)))
	}()

	log.Infof("resume %d inputs...", len(x.plugins))
	for _, p := range x.plugins {
		if err := p.Resume(); err != nil {
			log.Warnf("resume: %s", err)
		}
	}
}
