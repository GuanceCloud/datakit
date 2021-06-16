package io

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/parser"
)

const (
	filter_release uint8 = iota + 1
	filter_refreshed
)

var (
	defInterval   = 10 * time.Second
	defReqTimeout = 3 * time.Second
	defLogfilter  *logFilter
	log           = logger.DefaultSLogger("logfilter")
)

type rules struct {
	content []string `json:"content"`
}

type logFilter struct {
	clnt   *http.Client
	status uint8
	rules  string
	conds  parser.WhereConditions
	sync.Mutex
}

func newLogFilter() *logFilter {
	return &logFilter{
		clnt:   &http.Client{Timeout: defReqTimeout},
		status: filter_refreshed,
	}
}

func (this *logFilter) filter(pts []*Point) []*Point {
	if this.status == filter_release {
		return pts
	}

	var after []*Point
	for _, pt := range pts {
		fields, err := pt.Fields()
		if err != nil {
			log.Error(err)
			continue
		}
		if !this.conds.Eval(pt.Name(), pt.Tags(), fields) {
			after = append(after, pt)
		}
	}

	return after
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
	// req, err := http.NewRequest(http.MethodGet, defReqUrl, nil)
	// if err != nil {
	// 	return err
	// }

	// resp, err := this.clnt.Do(req)
	// if err != nil {
	// 	return err
	// }

	// buf, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	return err
	// }

	body, err := defLogFilterMock.GetLogFilter()
	if err != nil {
		return err
	}

	var rules rules
	if err = json.Unmarshal(body, &rules); err != nil {
		return err
	}

	if len(rules.content) == 0 {
		this.status = filter_release

		return nil
	}

	this.Lock()
	defer this.Unlock()

	// compare and refresh
	if newRules := strings.Join(rules.content, ";"); newRules != this.rules {
		this.rules = newRules
		this.conds = parser.GetConds(this.rules)
		this.status = filter_refreshed
	}

	return nil
}

func init() {
	log = logger.SLogger("logfilter")

	defLogfilter = newLogFilter()
	defLogfilter.start()
}
