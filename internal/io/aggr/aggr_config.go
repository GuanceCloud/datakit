// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package aggr

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/GuanceCloud/cliutils/aggregate"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/endpoint"
)

func (ag *Aggregator) endpointForPickKey(pickKey uint64) *endpoint.EndPoint {
	if len(ag.eps) == 0 {
		return nil
	}

	return ag.eps[pickKey%uint64(len(ag.eps))]
}

func (ag *Aggregator) endpointsForPickKey(pickKey uint64) []*endpoint.EndPoint {
	if len(ag.eps) == 0 {
		return nil
	}

	if len(ag.Endpoints) > 0 {
		ep := ag.endpointForPickKey(pickKey)
		if ep == nil {
			return nil
		}
		return []*endpoint.EndPoint{ep}
	}

	return ag.eps
}

func (ag *Aggregator) configuredHTTPTimeout() time.Duration {
	if ag.Timeout > 0 {
		return ag.Timeout
	}
	if ag.DW != nil && ag.DW.HTTPTimeout > 0 {
		return ag.DW.HTTPTimeout
	}

	return 30 * time.Second
}

func (ag *Aggregator) newCustomEndpoint(rawURL string) (*endpoint.EndPoint, error) {
	apis := []string{
		datakit.Aggregate,
		datakit.TailSampling,
		datakit.TailSamplingConfig,
	}

	if ag.DW == nil {
		return endpoint.NewEndpoint(rawURL,
			endpoint.WithOwner("aggr"),
			endpoint.WithAPIs(apis),
			endpoint.WithHTTPTimeout(ag.configuredHTTPTimeout()),
		)
	}

	return endpoint.NewEndpoint(rawURL,
		endpoint.WithOwner("aggr"),
		endpoint.WithAPIs(apis),
		endpoint.WithHTTPTimeout(ag.configuredHTTPTimeout()),
		endpoint.WithProxy(ag.DW.HTTPProxy),
		endpoint.WithInsecureSkipVerify(ag.DW.InsecureSkipVerify),
		endpoint.WithHTTPTrace(ag.DW.EnableHTTPTrace),
		endpoint.WithMaxHTTPIdleConnectionPerHost(ag.DW.MaxIdleConnsPerHost),
		endpoint.WithMaxHTTPConnections(ag.DW.MaxIdleConns),
		endpoint.WithHTTPIdleTimeout(ag.DW.IdleTimeout),
		endpoint.WithMaxRetryCount(ag.DW.MaxRetryCount),
		endpoint.WithRetryDelay(ag.DW.RetryDelay),
		endpoint.WithHTTPHeaders(map[string]string{
			"User-Agent": fmt.Sprintf("datakit-%s-%s/%s/%s",
				runtime.GOOS, runtime.GOARCH, git.Version, datakit.DKHost),
			"Referer": "DataKit",
		}),
	)
}

func (ag *Aggregator) initHTTP() {
	ag.eps = ag.eps[:0]

	if len(ag.Endpoints) > 0 {
		for _, rawURL := range ag.Endpoints {
			ep, err := ag.newCustomEndpoint(rawURL)
			if err != nil {
				log.Errorf("init aggr endpoint %s failed: %v", rawURL, err)
				continue
			}
			ag.eps = append(ag.eps, ep)
		}
		log.Debugf("init aggr endpoints from configured endpoints: count=%d", len(ag.eps))
		return
	}

	if ag.DW != nil {
		if len(ag.DW.GetEndpoints()) == 0 {
			log.Debugf("skip dataway endpoint init: no dataway endpoints")
			return
		}
		ag.eps = append(ag.eps, ag.DW.GetEndpoints()...)
		log.Debugf("init aggr endpoints from dataway endpoints: count=%d", len(ag.eps))
	}
}

func (ag *Aggregator) defaultToken() string {
	if len(ag.Endpoints) > 0 {
		for _, ep := range ag.eps {
			if ep != nil && ep.Token != "" {
				return ep.Token
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
	}

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
	ag.stateMu.RUnlock()

	if !enabled || cfg == nil {
		log.Debugf("skip sending tail sampling config: disabled")
		return
	}
	if len(ag.eps) == 0 {
		log.Errorf("skip sending tail sampling config: no endpoints available")
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
		return
	}

	headers := map[string]string{
		aggregate.GuancePickKey: "1",
	}
	success := false

	for _, ep := range ag.eps {
		if ep == nil {
			continue
		}

		resp, _, err := ep.WriteAggrData(&endpoint.AggrData{
			API:             datakit.TailSamplingConfig,
			Category:        "config",
			ContentType:     "application/json",
			ContentEncoding: identityContentEncoding,
			Body:            body,
			RawLen:          len(body),
			Headers:         headers,
		})
		if err != nil {
			log.Errorf("send tail sampling config failed: %v", err)
			continue
		}
		if resp == nil {
			log.Errorf("send tail sampling config failed: response terminated")
			continue
		}
		if resp.StatusCode/100 != 2 {
			log.Errorf("send tail sampling config got unexpected status=%d", resp.StatusCode)
			continue
		}

		success = true
		log.Debugf("send tail sampling config success: status=%d host=%s", resp.StatusCode, ep.Host)
	}

	if success {
		ag.lastSendTime = now
	}
}
