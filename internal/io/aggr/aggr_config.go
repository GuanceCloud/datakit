// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package aggr

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/GuanceCloud/cliutils/aggregate"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

func (ag *Aggregator) getEndpoint(pickKey uint64, route endpointRoute) string {
	if len(ag.Endpoints) > 0 {
		ep := ag.Endpoints[pickKey%uint64(len(ag.Endpoints))]
		u, err := url.Parse(ep)
		if err != nil {
			log.Errorf("parse endpoint url failed: %v", err)
			return ""
		}
		u.Path = routePathByType[route]
		return u.String()
	} else {
		switch route {
		case aggregateRoute:
			return ag.aggrURL
		case tailSamplingRoute:
			return ag.tailSamplingURL
		case tailSamplingConfigRoute:
			return ag.tailSamplingConfigURL
		default:
			return ""
		}
	}
}

func (ag *Aggregator) initHTTP() {
	if len(ag.Endpoints) > 0 {
		t := &http.Transport{
			MaxIdleConnsPerHost:   100,
			MaxIdleConns:          300,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}

		ag.Transport = t
		log.Debugf("init custom transport from configured endpoints: count=%d", len(ag.Endpoints))
	}

	if ag.DW != nil {
		if len(ag.DW.GetEndpoints()) == 0 {
			log.Debugf("skip dataway transport init: no dataway endpoints")
			return
		}
		for _, ep := range ag.DW.GetEndpoints() {
			cats := ep.GetCategoryURL()
			if len(cats) == 0 {
				log.Errorf("dataway endpoint category urls are empty")
				return
			}
			ag.aggrURL = cats[datakit.Aggregate]
			ag.tailSamplingURL = cats[datakit.TailSampling]
			ag.tailSamplingConfigURL = cats[datakit.TailSamplingConfig]
			ag.Transport = ep.Transport()
		}
		log.Debugf("init transport from dataway endpoints: count=%d", len(ag.DW.GetEndpoints()))
	}
}

func (ag *Aggregator) defaultToken() string {
	if ag.DW != nil {
		if ag.DW.Token != "" {
			return ag.DW.Token
		}

		for _, ep := range ag.DW.GetEndpoints() {
			if ep != nil && ep.Token != "" {
				return ep.Token
			}
		}
	}

	for _, ep := range ag.Endpoints {
		u, err := url.Parse(ep)
		if err != nil {
			log.Errorf("parse endpoint token failed: %v", err)
			continue
		}
		if token := u.Query().Get("token"); token != "" {
			return token
		}
	}

	return ""
}

func (ag *Aggregator) applyConfigSourceDefaults() {
	if ag.LocalConfigDir == "" {
		ag.LocalConfigDir = filepath.Join(datakit.TemplateDir, "aggr")
	}
	if ag.LocalMetricConfigFile == "" {
		ag.LocalMetricConfigFile = "aggr.toml"
	}
	if ag.LocalTailSamplingConfigFile == "" {
		ag.LocalTailSamplingConfigFile = "tail-sampling.toml"
	}
}

func (ag *Aggregator) loadMetricConfigFromDataway(bts []byte) {
	config := &aggregate.AggregatorConfigure{}
	err := json.Unmarshal(bts, config)
	if err != nil {
		log.Errorf("unmarshal metric config from dataway failed: %v", err)
		return
	}

	ag.metricConfig = config
	log.Debugf("loaded metric config from dataway: rules=%d", len(config.AggregateRules))
}

func (ag *Aggregator) loadMetricConfigFromFile(file string) {
	bts, err := os.ReadFile(file) //nolint
	if err != nil {
		log.Errorf("read metric config file failed: %v", err)
		return
	}

	ag.metricConfig = &aggregate.AggregatorConfigure{}
	_, err = toml.Decode(string(bts), ag.metricConfig)
	if err != nil {
		log.Errorf("decode metric config file failed: %v", err)
		return
	}
	log.Debugf("loaded metric config from file: file=%s rules=%d", file, len(ag.metricConfig.AggregateRules))
}

func (ag *Aggregator) loadTailSamplingConfigFromFile(file string) {
	bts, err := os.ReadFile(file) //nolint
	if err != nil {
		log.Errorf("read tail sampling config file failed: %v", err)
		return
	}

	tsconfig := &aggregate.TailSamplingConfigs{}
	_, err = toml.Decode(string(bts), tsconfig)
	if err != nil {
		log.Errorf("decode tail sampling config file failed: %v", err)
		return
	}

	ag.tailSamplingConfig = tsconfig
	log.Debugf("loaded tail sampling config from file: file=%s version=%d", file, tsconfig.Version)
}

func (ag *Aggregator) loadTailSamplingConfigFromDataway(bts []byte) {
	tsConfig := &aggregate.TailSamplingConfigs{}
	err := json.Unmarshal(bts, tsConfig)
	if err != nil {
		log.Errorf("unmarshal tail sampling config from dataway failed: %v", err)
		return
	}
	if ag.tailSamplingConfig == nil {
		ag.tailSamplingConfig = tsConfig
	}
	if tsConfig.Version != ag.tailSamplingConfig.Version {
		ag.tailSamplingConfig = tsConfig
	}
	ag.tailSamplingEnabled = true
	log.Debugf("loaded tail sampling config from dataway: version=%d", ag.tailSamplingConfig.Version)
}

func (ag *Aggregator) sendTSConfigToDW() {
	ag.tsConfigSendMu.Lock()
	defer ag.tsConfigSendMu.Unlock()

	ag.stateMu.RLock()
	enabled := ag.tailSamplingEnabled
	cfg := ag.tailSamplingConfig
	transport := ag.Transport
	ag.stateMu.RUnlock()

	if !enabled || cfg == nil {
		log.Debugf("skip sending tail sampling config: disabled")
		return
	}

	now := time.Now()
	if !ag.lastSendTime.IsZero() && now.Sub(ag.lastSendTime) < time.Second*10 {
		log.Debugf("skip sending tail sampling config: throttled")
		return
	}

	body, err := json.Marshal(cfg)
	if err != nil {
		log.Errorf("marshal tail sampling config failed: %v", err)
		ag.recordSendFailed("config", "config", "marshal")
		return
	}

	ep := ag.getEndpoint(1, tailSamplingConfigRoute)
	if ep == "" {
		log.Errorf("tail sampling config endpoint is empty")
		ag.recordSendFailed("config", "config", "transport")
		return
	}

	req, err := http.NewRequest("POST", ep, bytes.NewBuffer(body))
	if err != nil {
		log.Errorf("create tail sampling config request failed: %v", err)
		ag.recordSendFailed("config", "config", "other")
		return
	}
	setPickKeyHeader(req, 1)

	if transport == nil {
		log.Errorf("send tail sampling config failed: transport is nil")
		ag.recordSendFailed("config", "config", "transport")
		return
	}

	ag.lastSendTime = now

	resp, err := transport.RoundTrip(req)
	if err != nil {
		log.Errorf("send tail sampling config failed: %v", err)
		ag.recordSendFailed("config", "config", "network")
		return
	}

	defer resp.Body.Close() //nolint
	if resp.StatusCode/100 != 2 {
		log.Errorf("send tail sampling config got unexpected status=%d", resp.StatusCode)
		ag.recordSendFailed("config", "config", "server")
		return
	}
	log.Debugf("send tail sampling config success: status=%d", resp.StatusCode)
	ag.recordSendSuccess("config", "config")
}
