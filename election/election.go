package election

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	// "gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

var defaultConsensusModule = NewConsensusModule()

func StartElection() {
	defaultConsensusModule.StartElection()
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

func (cm *ConsensusModule) setCandidate() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.state = Candidate
}

func (cm *ConsensusModule) setLeader() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.state = Leader
}

func (cm *ConsensusModule) setDead() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.state = Dead
}

func (cm *ConsensusModule) CurrentStats() ConsensusState {
	return cm.state
}

func (cm *ConsensusModule) StartElection() {
	cm.setCandidate()

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
				cm.setCandidate()
			}

			if res.Stauts == statusSuccess {
				cm.setLeader()
			}
		}
	}
}

func (cm *ConsensusModule) SendHeartbeat() {
	defer cm.setCandidate()

	tick := time.NewTicker(time.Second * 3)
	defer tick.Stop()

	for {
		select {
		case <-datakit.Exit.Wait():
			return

		case <-tick.C:
			if cm.state.IsLeader() {
				continue
			}
			res, err := cm.postRequest(cm.electionHeartbeatURL)
			if err != nil {
				return
			}
			if res.Stauts != statusSuccess {
				return
			}
		}
	}
}

type electionResult struct {
	Stauts   string `json:"status"`
	ErrorMsg string `json:"error_msg"`
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

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var e = electionResult{}
	if err := json.Unmarshal(body, &e); err != nil {
		return nil, err
	}

	return &e, nil
}
