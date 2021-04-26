package election

import (
	"encoding/json"
	"fmt"
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

func SendHeartbeat() error {
	return defaultConsensusModule.SendHeartbeat()
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

type electionResult struct {
	Stauts string `json:"election_status"`
}

const (
	electionSuccess      = "success"
	electionUnsuccessful = "unsuccessful"
)

func (cm *ConsensusModule) StartElection() {
	cm.setCandidate()

	tick := time.NewTicker(time.Second * 3)
	defer tick.Stop()

	for {
		select {
		case <-datakit.Exit.Wait():
			return

		case <-tick.C:
			func() {
				resp, err := cm.httpCli.Post(cm.electionURL, "", nil)
				if err != nil {
					cm.setCandidate()
				}

				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					cm.setCandidate()
				}
				defer resp.Body.Close()

				var e = electionResult{}

				if err := json.Unmarshal(body, &e); err != nil {

				}

				if e.Stauts == electionSuccess {
					cm.setLeader()
				}
			}()
		}
	}
}

type heartbeatResult struct {
}

func (cm *ConsensusModule) SendHeartbeat() error {
	if cm.state != Leader {
		return fmt.Errorf("")
	}

	tick := time.NewTicker(time.Second * 3)
	defer tick.Stop()

	for {
		select {
		case <-datakit.Exit.Wait():
			return nil

		case <-tick.C:
			resp, err := cm.httpCli.Post(cm.electionURL, "", nil)
			if err != nil {

			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {

			}

			_ = body
		}
	}
}
