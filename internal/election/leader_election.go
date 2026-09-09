// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package election

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

var errInvalidElectionResponse = errors.New("invalid election response")

/*
 * DataKit 中心选举说明
 *
 * 流程：
 *      1. DataKit 开启 cfg.EnableElection（booler）配置
 *      2. 当运行对应的采集器（采集器列表在 config/inputcfg.go）时，程序会创建一个 goroutine 向已选定的 provider 发送选举请求，并携带 token、namespace（若存在）以及 id
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

	leaseDeadline time.Time
	pollInterval  time.Duration
	epoch         int64
	status        ElectionStatus
	plugins       []inputs.ElectionInput
}

func (x *leaderElection) reportStatus(status string) {
	electionStatusVec.WithLabelValues(
		CurrentElected,
		x.id,
		x.namespace,
		status,
	).Set(float64(x.clock.Now().Unix()))
}

func newLeaderElection(opt *option, plugins map[string][]inputs.ElectionInput) *leaderElection {
	if opt.provider == "" {
		opt.provider = ProviderDataway
	}
	if opt.clock == nil {
		opt.clock = realClock{}
	}
	x := &leaderElection{
		option: opt,
		status: StatusFail,
	}
	electionProviderInfoVec.WithLabelValues(string(opt.provider), opt.namespace).Set(1)
	for _, v := range plugins {
		x.plugins = append(x.plugins, v...)
	}
	return x
}

func (x *leaderElection) Run() {
	x.pausePlugins()
	x.reportStatus(x.status.String())
	interval := x.defaultInterval()
	tickerInterval := x.nextCheckInterval(interval)
	tick := time.NewTicker(tickerInterval)
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

				x.reportStatus(x.status.String())
			}

		case <-tick.C:
			if electionInterval, err := x.runOnce(); err == nil {
				if electionInterval <= 0 {
					electionInterval = x.defaultInterval()
				}
				interval = electionInterval
			}
			nextCheckInterval := x.nextCheckInterval(interval)
			if nextCheckInterval != tickerInterval {
				tick.Reset(nextCheckInterval)
				tickerInterval = nextCheckInterval
			}
		}
	}
}

func (x *leaderElection) runOnce() (int, error) {
	var (
		elecIntv int
		err      error
	)
	x.updateLeaseRemainingMetric()
	x.expireOperatorLease()
	if x.shouldWaitForOperatorLeaseExpiry() {
		return int(x.pollInterval / time.Second), nil
	}

	switch x.status {
	case StatusSuccess:
		elecIntv, err = x.keepalive()
		if err != nil {
			x.recordRequestError("heartbeat", err)
			log.Errorf("keepalive: %s", err)
		}
	case StatusFail:
		elecIntv, err = x.tryElection()
		if err != nil {
			x.recordRequestError("campaign", err)
			log.Errorf("tryElection: %s", err)
		}

	case StatusDisabled, StatusBanned, StatusImpeached: // pass
		return x.defaultInterval(), nil
	}

	return elecIntv, err
}

type leaderElectionResult struct {
	Content struct {
		Status        string `json:"status"`
		Namespace     string `json:"namespace,omitempty"`
		ID            string `json:"id"`
		IncumbencyID  string `json:"incumbency_id,omitempty"`
		ErrorMsg      string `json:"error_msg,omitempty"`
		Interval      int    `json:"interval"`
		LeaseDuration int    `json:"lease_duration,omitempty"`
		Epoch         int64  `json:"epoch,omitempty"`
	} `json:"content"`
}

func (x *leaderElection) tryElection() (int, error) {
	body, err := x.puller.Election(x.namespace, x.id, nil)
	if err != nil {
		log.Errorf("puller.Election: %s", err)
		return x.defaultInterval(), err
	}

	e := leaderElectionResult{}
	if err := json.Unmarshal(body, &e); err != nil {
		log.Error(err)
		if x.provider != ProviderOperator {
			return x.defaultInterval(), nil
		}
		return x.defaultInterval(), fmt.Errorf("%w: decode campaign response", errInvalidElectionResponse)
	}

	if err := x.validateOperatorResult(&e); err != nil {
		return x.defaultInterval(), err
	}
	if x.provider == ProviderOperator && e.Content.Status == StatusSuccess.String() &&
		e.Content.IncumbencyID != x.id {
		x.dropLeadership("holder_changed", e.Content.IncumbencyID)
		return x.defaultInterval(), fmt.Errorf("%w: Operator holder", errInvalidElectionResponse)
	}
	responseTime := x.clock.Now()

	electedChanged := CurrentElected != e.Content.IncumbencyID
	if electedChanged {
		CurrentElected = e.Content.IncumbencyID
	}

	statusChanged := e.Content.Status != x.status.String()
	if statusChanged {
		electionStatusSwitched.WithLabelValues(
			x.namespace,
			e.Content.Status,
		).Inc()
	}
	if statusChanged || electedChanged {
		x.reportStatus(e.Content.Status)
	}

	switch e.Content.Status {
	case StatusFail.String():

		x.status = StatusFail

	case StatusSuccess.String():
		previousStatus := x.status
		x.status = StatusSuccess
		x.lastElected = responseTime
		x.recordSuccessfulResponse(&e, responseTime)
		x.recordTransition(previousStatus, StatusSuccess, "acquired")
		x.logTransition(previousStatus, StatusSuccess, "acquired", e.Content.IncumbencyID)
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
		x.expireOperatorLease()
		return x.defaultInterval(), err
	}

	e := leaderElectionResult{}
	if err := json.Unmarshal(body, &e); err != nil {
		log.Error(err)
		x.expireOperatorLease()
		return x.defaultInterval(), fmt.Errorf("%w: decode heartbeat response", errInvalidElectionResponse)
	}

	if err := x.validateOperatorResult(&e); err != nil {
		x.expireOperatorLease()
		return x.defaultInterval(), err
	}
	if x.provider == ProviderOperator && e.Content.Status == StatusSuccess.String() &&
		e.Content.IncumbencyID != x.id {
		x.dropLeadership("holder_changed", e.Content.IncumbencyID)
		return x.defaultInterval(), fmt.Errorf("%w: Operator holder", errInvalidElectionResponse)
	}
	responseTime := x.clock.Now()

	statusChanged := e.Content.Status != x.status.String()
	if statusChanged {
		electionStatusSwitched.WithLabelValues(
			x.namespace,
			e.Content.Status,
		).Inc()
	}

	electedChanged := CurrentElected != e.Content.IncumbencyID
	CurrentElected = e.Content.IncumbencyID
	if statusChanged || electedChanged {
		x.reportStatus(e.Content.Status)
	}

	switch e.Content.Status {
	case StatusFail.String():
		x.recordTransition(x.status, StatusFail, "defeated")
		x.logTransition(x.status, StatusFail, "defeated", e.Content.IncumbencyID)
		x.status = StatusFail
		x.clearOperatorLease()
		x.pausePlugins()

	case StatusSuccess.String():
		epochChanged := x.provider == ProviderOperator && x.epoch != 0 && x.epoch != e.Content.Epoch
		previousEpoch := x.epoch
		if epochChanged {
			x.recordTransition(StatusSuccess, StatusSuccess, "epoch_changed")
			x.lastElected = responseTime
		}
		x.recordSuccessfulResponse(&e, responseTime)
		if epochChanged {
			x.pausePlugins()
			x.resumePlugins()
		}
		if epochChanged {
			log.Infof("election_transition provider=%s namespace=%q id=%q from=%s to=%s reason=epoch_changed holder=%q previous_epoch=%d epoch=%d lease_remaining_seconds=%.0f",
				x.provider, x.namespace, x.id, StatusSuccess, StatusSuccess, e.Content.IncumbencyID,
				previousEpoch, x.epoch, x.leaseRemainingSeconds())
		}
		log.Debugf("%s election keepalive ok", x.id)

	default:
		log.Warnf("unknown election status: %s", e.Content.Status)
	}
	return e.Content.Interval, nil
}

func (x *leaderElection) validateOperatorResult(result *leaderElectionResult) error {
	if x.provider != ProviderOperator {
		return nil
	}
	pollInterval, ok := durationFromSeconds(result.Content.Interval)
	if !ok || pollInterval <= operatorRequestTimeout {
		return fmt.Errorf("%w: Operator interval", errInvalidElectionResponse)
	}
	if result.Content.Namespace != x.namespace || result.Content.ID != x.id {
		return fmt.Errorf("%w: Operator election scope", errInvalidElectionResponse)
	}
	if result.Content.Status != StatusSuccess.String() && result.Content.Status != StatusFail.String() {
		return fmt.Errorf("%w: Operator status", errInvalidElectionResponse)
	}
	if result.Content.Status == StatusFail.String() {
		return nil
	}
	if result.Content.IncumbencyID == "" {
		return fmt.Errorf("%w: Operator holder", errInvalidElectionResponse)
	}
	if result.Content.Epoch <= 0 {
		return fmt.Errorf("%w: Operator epoch", errInvalidElectionResponse)
	}
	leaseDuration, ok := durationFromSeconds(result.Content.LeaseDuration)
	if !ok || leaseDuration <= pollInterval {
		return fmt.Errorf("%w: Operator lease duration", errInvalidElectionResponse)
	}
	return nil
}

func (x *leaderElection) recordSuccessfulResponse(result *leaderElectionResult, responseTime time.Time) {
	if result.Content.Status != StatusSuccess.String() {
		return
	}
	electionLastSuccessVec.WithLabelValues(string(x.provider), x.namespace).Set(float64(responseTime.Unix()))
	if x.provider != ProviderOperator {
		return
	}
	leaseDuration, leaseOK := durationFromSeconds(result.Content.LeaseDuration)
	pollInterval, intervalOK := durationFromSeconds(result.Content.Interval)
	if !leaseOK || !intervalOK {
		return
	}
	x.leaseDeadline = responseTime.Add(leaseDuration - pollInterval)
	x.pollInterval = pollInterval
	x.epoch = result.Content.Epoch
	electionEpochVec.WithLabelValues(string(x.provider), x.namespace).Set(float64(x.epoch))
	x.updateLeaseRemainingMetric()
}

func (x *leaderElection) expireOperatorLease() {
	if x.provider != ProviderOperator || x.status != StatusSuccess || x.leaseDeadline.IsZero() ||
		x.clock.Now().Before(x.leaseDeadline) {
		return
	}

	x.dropLeadership("lease_expired", CurrentElected)
}

func (x *leaderElection) defaultInterval() int {
	if x.provider == ProviderOperator {
		return operatorElectionIntervalDefault
	}
	return datawayElectionIntervalDefault
}

func (x *leaderElection) nextCheckInterval(pollIntervalSeconds int) time.Duration {
	pollInterval, ok := durationFromSeconds(pollIntervalSeconds)
	if !ok {
		pollInterval = time.Duration(x.defaultInterval()) * time.Second
	}
	if x.provider != ProviderOperator || x.status != StatusSuccess || x.leaseDeadline.IsZero() {
		return pollInterval
	}

	remaining := x.leaseDeadline.Sub(x.clock.Now())
	if remaining <= 0 {
		return time.Nanosecond
	}
	if remaining < pollInterval {
		return remaining
	}
	return pollInterval
}

func (x *leaderElection) shouldWaitForOperatorLeaseExpiry() bool {
	if x.provider != ProviderOperator || x.status != StatusSuccess || x.leaseDeadline.IsZero() ||
		x.pollInterval <= 0 {
		return false
	}
	remaining := x.leaseDeadline.Sub(x.clock.Now())
	return remaining > 0 && remaining < x.pollInterval
}

func durationFromSeconds(seconds int) (time.Duration, bool) {
	const maxDurationSeconds = (1<<63 - 1) / int64(time.Second)

	if seconds <= 0 || int64(seconds) > maxDurationSeconds {
		return 0, false
	}
	return time.Duration(seconds) * time.Second, true
}

func (x *leaderElection) dropLeadership(reason, holder string) {
	if x.status != StatusSuccess {
		return
	}

	x.recordTransition(x.status, StatusFail, reason)
	x.status = StatusFail
	CurrentElected = holder
	x.logTransition(StatusSuccess, StatusFail, reason, holder)
	x.clearOperatorLease()
	x.pausePlugins()
	x.reportStatus(x.status.String())
}

func (x *leaderElection) clearOperatorLease() {
	if x.provider != ProviderOperator {
		return
	}
	x.leaseDeadline = time.Time{}
	x.pollInterval = 0
	electionLeaseRemainingVec.WithLabelValues(string(x.provider), x.namespace).Set(0)
}

func (x *leaderElection) updateLeaseRemainingMetric() {
	if x.provider != ProviderOperator || x.leaseDeadline.IsZero() {
		return
	}
	electionLeaseRemainingVec.WithLabelValues(string(x.provider), x.namespace).Set(x.leaseRemainingSeconds())
}

func (x *leaderElection) leaseRemainingSeconds() float64 {
	if x.leaseDeadline.IsZero() {
		return 0
	}
	return max(0, x.leaseDeadline.Sub(x.clock.Now()).Seconds())
}

func (x *leaderElection) recordRequestError(operation string, err error) {
	reason := electionRequestErrorReason(err)
	electionRequestErrorsVec.WithLabelValues(
		string(x.provider), x.namespace, operation, reason,
	).Inc()
	log.Warnf("election_request_error provider=%s namespace=%q id=%q operation=%s reason=%s",
		x.provider, x.namespace, x.id, operation, reason)
}

func (x *leaderElection) recordTransition(from, to ElectionStatus, reason string) {
	electionTransitionsVec.WithLabelValues(
		string(x.provider), x.namespace, from.String(), to.String(), reason,
	).Inc()
}

func (x *leaderElection) logTransition(from, to ElectionStatus, reason, holder string) {
	leaderDuration := 0.0
	if from == StatusSuccess && !x.lastElected.IsZero() {
		leaderDuration = max(0, x.clock.Now().Sub(x.lastElected).Seconds())
	}
	log.Infof("election_transition provider=%s namespace=%q id=%q from=%s to=%s reason=%s holder=%q epoch=%d lease_remaining_seconds=%.0f leader_duration_seconds=%.0f",
		x.provider, x.namespace, x.id, from, to, reason, holder, x.epoch,
		x.leaseRemainingSeconds(), leaderDuration)
}

func electionRequestErrorReason(err error) string {
	var requestErr *operatorRequestError
	if errors.As(err, &requestErr) {
		if requestErr.statusCode != 0 {
			switch requestErr.statusCode / 100 {
			case 4:
				return "http_4xx"
			case 5:
				return "http_5xx"
			default:
				return "http_other"
			}
		}
		switch requestErr.kind {
		case "timeout":
			return "timeout"
		case "transport_error":
			return "transport"
		default:
			return "client"
		}
	}
	if errors.Is(err, errInvalidElectionResponse) {
		return "invalid_response"
	}
	return "other"
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
