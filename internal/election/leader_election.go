// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package election

import (
	"encoding/json"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
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
	status  electionStatus
	plugins []inputs.ElectionInput
}

func newLeaderElection(opt *option, plugins map[string][]inputs.ElectionInput) *leaderElection {
	x := &leaderElection{
		option: opt,
		status: statusFail,
	}
	for _, v := range plugins {
		x.plugins = append(x.plugins, v...)
	}
	return x
}

func (x *leaderElection) Run() {
	defer func() {
		electionStatusVec.WithLabelValues(
			CurrentElected,
			x.id,
			x.namespace,
			x.status.String(),
		).Set(float64(x.status))
	}()

	x.pausePlugins()
	tick := time.NewTicker(time.Second * time.Duration(electionIntervalDefault))
	defer tick.Stop()

	for {
		select {
		case <-datakit.Exit.Wait():
			electionInputs.WithLabelValues(x.namespace).Set(float64(len(x.plugins)))
			return

		case <-tick.C:
			electionInterval := x.runOnce()
			if electionInterval != electionIntervalDefault {
				tick.Reset(time.Second * time.Duration(electionInterval))
				electionIntervalDefault = electionInterval
			}
		}
	}
}

func (x *leaderElection) runOnce() int {
	var (
		elecIntv int
		err      error
	)

	switch x.status {
	case statusSuccess:
		elecIntv, err = x.keepalive()
	case statusFail:
		elecIntv, err = x.tryElection()
	case statusDisabled: // pass
		return electionIntervalDefault
	}

	if err != nil {
		io.FeedLastError("election", err.Error())
	}

	return elecIntv
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
	var (
		electedTime int64
		start       = time.Now()
	)

	body, err := x.puller.Election(x.namespace, x.id, nil)
	if err != nil {
		log.Errorf("puller.Election: %s", err)
		return electionIntervalDefault, err
	}

	defer func() {
		electionVec.WithLabelValues(
			x.namespace,
			x.status.String(),
		).Observe(float64(time.Since(start) / time.Millisecond))

		electionStatusVec.WithLabelValues(
			CurrentElected,
			x.id,
			x.namespace,
			x.status.String(),
		).Set(float64(electedTime))
	}()

	e := leaderElectionResult{}
	if err := json.Unmarshal(body, &e); err != nil {
		log.Error(err)

		return electionIntervalDefault, nil
	}

	log.Debugf("result body: %s", body)

	if CurrentElected != e.Content.IncumbencyID {
		CurrentElected = e.Content.IncumbencyID
		electionStatusVec.Reset() // cleanup election status metrics
	}

	switch e.Content.Status {
	case statusFail.String():
		electionStatusVec.Reset() // cleanup election status if election failed

		x.status = statusFail

	case statusSuccess.String():
		electionStatusVec.Reset() // cleanup election status if election ok

		x.status = statusSuccess
		x.resumePlugins()
		electedTime = time.Now().Unix()

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

	CurrentElected = e.Content.IncumbencyID

	switch e.Content.Status {
	case statusFail.String():
		electionStatusVec.Reset() // cleanup election status if election fail
		x.status = statusFail
		x.pausePlugins()

	case statusSuccess.String():
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
	for i, p := range x.plugins {
		log.Debugf("pause %dth inputs...", i)
		if err := p.Pause(); err != nil {
			log.Warn(err)
		}
	}
}

func (x *leaderElection) resumePlugins() {
	defer func() {
		inputsResumeVec.WithLabelValues(x.id, x.namespace).Add(float64(len(x.plugins)))
	}()
	for i, p := range x.plugins {
		log.Debugf("resume %dth inputs...", i)
		if err := p.Resume(); err != nil {
			log.Warn(err)
		}
	}
}
