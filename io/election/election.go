// Package election implements DataFlux central election client.
package election

import (
	"context"
	"encoding/json"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/dataway"
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

var (
	defaultCandidate        = &candidate{status: statusDisabled}
	log                     = logger.DefaultSLogger("dk-election")
	HTTPTimeout             = time.Second * 3
	electionIntervalDefault = 4
)

const (
	statusSuccess  = "success"
	statusFail     = "defeat"
	statusDisabled = "disabled"
)

type candidate struct {
	status                         string
	id, namespace                  string
	dw                             *dataway.DataWayCfg
	plugins                        []inputs.ElectionInput
	ElectedTime                    time.Time
	nElected, nHeartbeat, nOffline int
}

func Start(namespace, id string, dw *dataway.DataWayCfg) {
	log = logger.SLogger("dk-election")
	defaultCandidate.run(namespace, id, dw)
}

func (x *candidate) run(namespace, id string, dw *dataway.DataWayCfg) {
	x.id = id
	x.namespace = namespace
	x.dw = dw
	x.plugins = inputs.GetElectionInputs()

	log.Debugf("namespace: %s id: %s", x.namespace, x.id)
	log.Infof("get %d election inputs", len(x.plugins))

	x.startElection()
}

func (x *candidate) startElection() {
	g := datakit.G("election")
	g.Go(func(ctx context.Context) error {
		x.pausePlugins()
		tick := time.NewTicker(time.Second * time.Duration(electionIntervalDefault))
		defer tick.Stop()

		for {
			select {
			case <-datakit.Exit.Wait():
				return nil
			case <-tick.C:
				log.Debugf("###run once...")
				electionInterval := x.runOnce()
				if electionInterval != electionIntervalDefault {
					tick.Reset(time.Second * time.Duration(electionInterval))
					electionIntervalDefault = electionInterval
				}
			}
		}
	})
}

// Elected 此处暂不考虑互斥性，只用于状态展示.
func Elected() (string, string) {
	return defaultCandidate.status, defaultCandidate.namespace
}

func GetElectedTime() time.Time {
	return defaultCandidate.ElectedTime
}

func (x *candidate) runOnce() int {
	var (
		elecIntv int
		err      error
	)
	switch x.status {
	case statusSuccess:
		elecIntv, err = x.keepalive()
	default:
		elecIntv, err = x.tryElection()
	}

	if err != nil {
		io.FeedLastError("election", err.Error())
	}

	return elecIntv
}

func (x *candidate) pausePlugins() {
	for i, p := range x.plugins {
		log.Debugf("pause %dth inputs...", i)
		if err := p.Pause(); err != nil {
			log.Warn(err)
		}
	}
}

func (x *candidate) resumePlugins() {
	for i, p := range x.plugins {
		log.Debugf("resume %dth inputs...", i)
		if err := p.Resume(); err != nil {
			log.Warn(err)
		}
	}
}

func (x *candidate) keepalive() (int, error) {
	body, err := x.dw.ElectionHeartbeat(x.namespace, x.id)
	if err != nil {
		log.Error(err)
		return electionIntervalDefault, err
	}

	e := electionResult{}
	if err := json.Unmarshal(body, &e); err != nil {
		log.Error(err)
		return electionIntervalDefault, err
	}

	log.Debugf("result body: %s", body)

	switch e.Content.Status {
	case statusFail:
		x.status = statusFail
		x.nOffline++
		x.pausePlugins()
	case statusSuccess:
		x.nHeartbeat++
		log.Debugf("%s HB %d", x.id, x.nHeartbeat)
	default:
		log.Warnf("unknown election status: %s", e.Content.Status)
	}
	return e.Content.Interval, nil
}

type electionResult struct {
	Content struct {
		Status       string `json:"status"`
		Namespace    string `json:"namespace,omitempty"`
		ID           string `json:"id"`
		IncumbencyID string `json:"incumbency_id,omitempty"`
		ErrorMsg     string `json:"error_msg,omitempty"`
		Interval     int    `json:"interval"`
	} `json:"content"`
}

func (x *candidate) tryElection() (int, error) {
	body, err := x.dw.Election(x.namespace, x.id)
	if err != nil {
		log.Error(err)

		return electionIntervalDefault, err
	}

	e := electionResult{}
	if err := json.Unmarshal(body, &e); err != nil {
		log.Error(err)

		return electionIntervalDefault, nil
	}

	log.Debugf("result body: %s", body)

	switch e.Content.Status {
	case statusFail:
		if x.status != statusFail {
			io.FeedEventLog(&io.Reporter{Message: "election fail", Logtype: "event"})
		}
		x.status = statusFail
	case statusSuccess:
		if x.status != statusSuccess {
			io.FeedEventLog(&io.Reporter{Message: "election success", Logtype: "event"})
		}
		x.status = statusSuccess
		x.resumePlugins()
		x.nElected++
		x.nHeartbeat = 0
		x.ElectedTime = time.Now()
	default:
		log.Warnf("unknown election status: %s", e.Content.Status)
	}

	return e.Content.Interval, nil
}
