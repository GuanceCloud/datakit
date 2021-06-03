package election

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/system/rtpanic"
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
 *      2. 当运行对应的采集器（采集器列表在 config/inputcfg.go）时，程序会创建一个 goroutine 向 DataWay 发送选举请求，并携带此 Datakit 的 token 和 UUID
 *      3. 选举成功担任 leader 后会持续发送心跳，心跳间隔过长或选举失败，会恢复 candidate 状态并继续发送选举请求
 *      4. 采集器端只要在采集数据时，判断当前是否为 leader 状态，具体使用见下
 *
 * 使用方式：
 *      1. 在 config/inputcfg.go 的 electionInputs 中添加需要选举的采集器（目前使用此方式后续会优化）
 *      2. 采集器中 import "gitlab.jiagouyun.com/cloudcare-tools/datakit/election"
 *      3. 在采集入口处，调用 election.CurrentStats().IsLeader() 进行判断，并决定是否执行采集
 *      4. 详见 demo 采集器
 */

var (
	defaultCandidate = &candidate{
		status: statusFail,
	}

	l = logger.DefaultSLogger("dk-election")

	HTTPTimeout = time.Second * 3

	// 选举请求间隔，和发送心跳间隔都是 3 秒
	electionInterval = time.Second * 3
)

const (
	statusSuccess = "success"
	statusFail    = "fail"
)

type candidate struct {
	status  string
	id      string
	dw      *dataway.DataWayCfg
	plugins []inputs.ElectionInput

	nElected, nHeartbeat, nOffline int
}

// NewCNewConsensusModule 两个参数是配置文件项，为了方便测试将其改为传参方式
func Setup(id string, dw *dataway.DataWayCfg) {
	l = logger.SLogger("dk-election")
	defaultCandidate.setup(id, dw)
}

func (x *candidate) setup(id string, dw *dataway.DataWayCfg) {
	x.id = id
	x.dw = dw
	x.plugins = inputs.GetElectionInputs()

	x.startElection()
}

func (x *candidate) startElection() {

	var f rtpanic.RecoverCallback
	crashTime := []string{}
	f = func(trace []byte, err error) {

		defer rtpanic.Recover(f, nil)
		if trace != nil {
			l.Warnf("election panic:\n%s", string(trace))
			crashTime = append(crashTime, fmt.Sprintf("%v", time.Now()))
			if len(crashTime) > 6 {
				io.FeedLastError("Election", fmt.Sprintf("election crashed %d times, exited", len(crashTime)))
				l.Errorf("election crashed %d times(at %s), exit now", len(crashTime), strings.Join(crashTime, "\n"))
				return
			}
		}

		tick := time.NewTicker(electionInterval)
		defer tick.Stop()

		for {
			select {
			case <-datakit.Exit.Wait():
				return

			case <-tick.C:
				x.runOnce()
			}
		}
	}

	// 先暂停采集，待选举成功再恢复运行
	x.pausePlugins()
	f(nil, nil)
}

func (x *candidate) runOnce() {
	switch x.status {
	case statusSuccess:
		x.keepalive()
	default:
		x.tryElection()
	}
}

func (x *candidate) pausePlugins() {
	for _, p := range x.plugins {
		if err := p.Pause(); err != nil {
			l.Warn(err)
		}
	}
}

func (x *candidate) resumePlugins() {
	for _, p := range x.plugins {
		if err := p.Resume(); err != nil {
			l.Warn(err)
		}
	}
}

func (x *candidate) keepalive() {
	body, err := x.dw.ElectionHeartbeat(x.id)
	if err != nil {
		l.Error(err)
		return
	}

	var e = electionResult{}
	if err := json.Unmarshal(body, &e); err != nil {
		l.Error(err)
		return
	}

	switch e.Content.Status {
	case statusFail:
		x.status = statusFail
		x.nOffline++
		l.Debugf("%s offline", x.id)
	case statusSuccess:
		x.nHeartbeat++
	default:
		l.Warnf("unknown election status: %s", e.Content.Status)
	}
	return
}

type electionResult struct {
	Content struct {
		Status   string `json:"status"`
		ErrorMsg string `json:"error_msg"`
	} `json:"content"`
}

func (x *candidate) tryElection() error {

	body, err := x.dw.Election(x.id)
	if err != nil {
		l.Error(err)
		return err
	}

	var e = electionResult{}
	if err := json.Unmarshal(body, &e); err != nil {
		l.Error(err)
		return nil
	}

	switch e.Content.Status {
	case statusFail:
		x.status = e.Content.Status
	case statusSuccess:
		x.status = e.Content.Status
		x.nElected++
	default:
		l.Warnf("unknown election status: %s", e.Content.Status)
	}
	return nil
}
