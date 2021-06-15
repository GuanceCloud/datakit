package logfilter

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
)

var (
	defReqUrl     = "/v1/logfilter/pull"
	defInterval   = 10 * time.Second
	defReqTimeout = 3 * time.Second
	log           = logger.DefaultSLogger("logfilter")
)

type rules struct {
	content []string `json:"content"`
}

type LogFilter struct {
	clnt  *http.Client
	rules rules
	sync.Mutex
}

func NewLogFilter() *LogFilter {
	return &LogFilter{clnt: &http.Client{Timeout: defReqTimeout}}
}

func (this *LogFilter) Run() {

}

func (this *LogFilter) refreshRules() error {
	req, err := http.NewRequest(http.MethodGet, defReqUrl, nil)
	if err != nil {
		return err
	}

	resp, err := this.clnt.Do(req)
	if err != nil {
		return err
	}

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var rules rules
	if err = json.Unmarshal(buf, &rules); err != nil {
		return err
	}

	this.Lock()
	defer this.Unlock()

	this.rules = rules

	return nil
}

func init() {
	log = logger.SLogger("logfilter")
}
