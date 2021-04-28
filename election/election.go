package election

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

var (
	defaultConsensusModule = NewConsensusModule()

	l = logger.DefaultSLogger("dk-election")
)

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
		panic("unreachable")
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

const defaultHTTPTimeout = time.Second * 3

func NewConsensusModule() *ConsensusModule {
	return &ConsensusModule{
		state:                Dead,
		electionURL:          datakit.Cfg.MainCfg.DataWay.ElectionURL(),
		electionHeartbeatURL: datakit.Cfg.MainCfg.DataWay.ElectionHeartBeatURL(),
		httpCli: &http.Client{
			Timeout: defaultHTTPTimeout,
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
	tick := time.NewTicker(time.Second * 3)
	defer tick.Stop()

	for {
		select {
		case <-datakit.Exit.Wait():
			return

		case <-tick.C:
			if cm.state.IsCandidate() {
				continue
			}
			res, err := cm.postRequest(cm.electionURL)
			if err != nil {
				l.Error(err)
				cm.SetCandidate()
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

	tick := time.NewTicker(time.Second * 3)
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
			if res.Content.Stauts != statusSuccess {
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
	resp, err := cm.httpCli.Post(url, "", nil)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// ok
	if resp.StatusCode/100 == 2 {
		var e = electionResult{}
		if err := json.Unmarshal(body, &e); err != nil {
			return nil, err
		}
		return &e, nil
	}

	// bad
	return nil, fmt.Errorf("%s", body)
}
