package election

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
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
	defaultConsensusModule *ConsensusModule

	l = logger.DefaultSLogger("dk-election")

	HTTPTimeout = time.Second * 3

	// 选举请求间隔，和发送心跳间隔都是 3 秒
	electionInterval = time.Second * 3
)

func InitGlobalConsensusModule() error {
	l = logger.SLogger("dk-election")

	if len(datakit.Cfg.DataWay.ElectionURL()) == 0 || len(datakit.Cfg.DataWay.ElectionHeartBeatURL()) == 0 {
		return fmt.Errorf("invalid electionURL or electionHeartbeatURL, is empty")
	}

	electionURL, err := setURLQueryParam(
		datakit.Cfg.DataWay.ElectionURL()[0],
		"id",
		datakit.Cfg.UUID,
	)
	if err != nil {
		return err
	}

	heartbeatURL, err := setURLQueryParam(
		datakit.Cfg.DataWay.ElectionHeartBeatURL()[0],
		"id",
		datakit.Cfg.UUID,
	)
	if err != nil {
		return err
	}

	l.Debugf("election URL: %s", electionURL)
	l.Debugf("election heartbeat URL: %s", heartbeatURL)
	defaultConsensusModule = NewConsensusModule(electionURL, heartbeatURL)
	return nil
}

func StartElection() {
	defaultConsensusModule.StartElection()
}

func SetCandidate() {
	defaultConsensusModule.SetCandidate()
}

func SetLeader() {
	defaultConsensusModule.SetLeader()
}

func CurrentStats() ConsensusState {
	return defaultConsensusModule.CurrentStats()
}

type ConsensusState int

const (
	Candidate ConsensusState = iota + 1
	Leader
	Dead
)

func (s ConsensusState) String() string {
	switch s {
	case Candidate:
		return "Candidate"
	case Leader:
		return "Leader"
	case Dead:
		return "Dead"
	default:
		return "unreachable"
	}
}

func (s ConsensusState) IsCandidate() bool {
	return s == Candidate
}

func (s ConsensusState) IsLeader() bool {
	return s == Leader
}

func (s ConsensusState) IsDead() bool {
	return s == Dead
}

type ConsensusModule struct {
	state                ConsensusState
	electionURL          string
	electionHeartbeatURL string

	httpCli *http.Client
	mu      sync.Mutex
}

// NewCNewConsensusModule 两个参数是配置文件项，为了方便测试将其改为传参方式
func NewConsensusModule(electionURL, heartbeatURL string) *ConsensusModule {
	return &ConsensusModule{
		state:                Dead,
		electionURL:          electionURL,
		electionHeartbeatURL: heartbeatURL,
		httpCli: &http.Client{
			Timeout: HTTPTimeout,
		},
	}
}

func (cm *ConsensusModule) SetCandidate() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.state = Candidate
}

func (cm *ConsensusModule) SetLeader() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.state = Leader
}

func (cm *ConsensusModule) SetDead() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.state = Dead
}

func (cm *ConsensusModule) CurrentStats() ConsensusState {
	return cm.state
}

func (cm *ConsensusModule) StartElection() {
	tick := time.NewTicker(electionInterval)
	defer tick.Stop()

	for {
		select {
		case <-datakit.Exit.Wait():
			return

		case <-tick.C:
			res, err := cm.postRequest(cm.electionURL)
			if err != nil {
				cm.SetCandidate()
				l.Error(err)
				continue
			}

			if res.Content.Stauts == statusSuccess {
				cm.SetLeader()
				// 阻塞在此，成为 leader 之后将不再进行进行选举，而是持续发送心跳
				cm.SendHeartbeat()
			}
		}
	}
}

func (cm *ConsensusModule) SendHeartbeat() {
	defer cm.SetCandidate()

	tick := time.NewTicker(electionInterval)
	defer tick.Stop()

	for {
		select {
		case <-datakit.Exit.Wait():
			return

		case <-tick.C:
			if !cm.state.IsLeader() {
				return
			}
			res, err := cm.postRequest(cm.electionHeartbeatURL)
			if err != nil {
				l.Error(err)
				return
			}
			if res.Content.ErrorMsg != "" {
				l.Debug(res.Content.ErrorMsg)
			}
			if res.Content.Stauts != statusSuccess {
				cm.SetCandidate()
				return
			}
		}
	}
}

type electionResult struct {
	Content struct {
		Stauts   string `json:"status"`
		ErrorMsg string `json:"error_msg"`
	} `json:"content"`
}

const (
	statusSuccess = "success"
	statusFail    = "fail"
)

func (cm *ConsensusModule) postRequest(url string) (*electionResult, error) {
	l.Debugf("election POST URL: %s", url)
	// datakit 数据发送到 dataway，不需要添加一堆 header
	// 简洁发送
	resp, err := cm.httpCli.Post(url, "", nil)
	if err != nil {
		l.Error(err)
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		l.Errorf("request url %s failed: %s", url, err)
		return nil, err
	}

	// ok
	if resp.StatusCode/100 == 2 {
		var e = electionResult{}
		if err := json.Unmarshal(body, &e); err != nil {
			l.Error(err)
			return nil, err
		}
		return &e, nil
	}

	// bad
	return nil, fmt.Errorf("%s", body)
}

func setURLQueryParam(urlStr, paramKey, paramValue string) (string, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Set(paramKey, paramValue)

	u.RawQuery = q.Encode()
	return u.String(), nil
}
