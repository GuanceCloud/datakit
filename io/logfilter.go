package io

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

var (
	defReqUrl     = "/v1/logfilter/pull"
	defInterval   = 10 * time.Second
	defReqTimeout = 3 * time.Second
	log           = logger.DefaultSLogger("logfilter")
)

var defLogfilter *logFilter

type rules struct {
	content []string `json:"content"`
}

type logFilter struct {
	clnt  *http.Client
	rules rules
	sync.Mutex
}

func newLogFilter() *logFilter {
	return &logFilter{clnt: &http.Client{Timeout: defReqTimeout}}
}

func (this *logFilter) check(point *Point) bool {
	return false
}

func (this *logFilter) start() {
	log.Info("log filter engaged")

	go func() {
		tick := time.NewTicker(defInterval)
		for {
			select {
			case <-datakit.Exit.Wait():
				log.Info("log filter exits")
			case <-tick.C:
				if err := this.refreshRules(); err != nil {
					log.Error(err.Error())
				}
			}
		}
	}()
}

func (this *logFilter) refreshRules() error {
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

func (this *logFilter) getRules() []string {
	return this.rules.content
}

func init() {
	log = logger.SLogger("logfilter")

	defLogfilter = newLogFilter()
	defLogfilter.start()
}
