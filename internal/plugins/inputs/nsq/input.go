// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package nsq collects NSQ metrics.
package nsq

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	dknet "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	timex "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/time"
)

var _ inputs.ElectionInput = (*Input)(nil)

const (
	inputName = "nsq"
	catalog   = "nsq"

	nsqdStatsPattern = "%s/stats?format=json"
	lookupdPattern   = "%s/nodes"

	unknownHost = "unknown"
)

var (
	updateEndpointListInterval = time.Second * 30
	minInterval                = time.Second * 3
	defaultInterval            = time.Second * 10
)

var l = logger.DefaultSLogger(inputName)

type Input struct {
	Lookupd            string            `toml:"lookupd"`
	NSQDs              []string          `toml:"nsqd"`
	Interval           string            `toml:"interval"`
	TLSCA              string            `toml:"tls_ca"`
	TLSCert            string            `toml:"tls_cert"`
	TLSKey             string            `toml:"tls_key"`
	InsecureSkipVerify bool              `toml:"insecure_skip_verify"`
	Tags               map[string]string `toml:"tags"`

	EndpointDeprecated []string `toml:"endpoints,omitempty"`

	lookupdEndpoint  string
	nsqdEndpointList map[string]interface{}

	httpClient *http.Client
	duration   time.Duration

	Election bool `toml:"election"`
	pauseCh  chan bool
	pause    bool

	semStop *cliutils.Sem // start stop signal
	feeder  dkio.Feeder
	Tagger  datakit.GlobalTagger
}

var maxPauseCh = inputs.ElectionPauseChannelLength

func (ipt *Input) ElectionEnabled() bool {
	return ipt.Election
}

func (*Input) SampleConfig() string { return sampleCfg }

func (*Input) Catalog() string { return catalog }

func (*Input) AvailableArchs() []string { return datakit.AllOSWithElection }

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&nsqTopicMeasurement{},
		&nsqNodesMeasurement{},
	}
}

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)

	if ipt.setup() {
		return
	}

	gatherTicker := time.NewTicker(ipt.duration)
	defer gatherTicker.Stop()

	updateListTicker := time.NewTicker(updateEndpointListInterval)
	defer updateListTicker.Stop()

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("nsq exit")
			return

		case <-ipt.semStop.Wait():
			l.Info("nsq return")
			return

		case <-gatherTicker.C:
			if ipt.pause {
				l.Debugf("not leader, skipped")
				continue
			}
			start := time.Now()
			pts, err := ipt.gather()
			if err != nil {
				l.Errorf("gather: %s, ignored", err)
			}

			if len(pts) == 0 {
				continue
			}

			if err := ipt.feeder.Feed(inputName,
				point.Metric,
				pts,
				&dkio.Option{CollectCost: time.Since(start)}); err != nil {
				l.Errorf("io.Feed: %s, ignored", err)
			}

		case <-updateListTicker.C:
			if ipt.pause {
				l.Debugf("not leader, skipped")
				continue
			}
			if ipt.isLookupd() {
				if err := ipt.updateEndpointListByLookupd(ipt.lookupdEndpoint); err != nil {
					l.Error(err)
					continue
				}
				l.Debugf("nsqd endpoint list: %v", ipt.nsqdEndpointList)
			}

		case ipt.pause = <-ipt.pauseCh:
			// nil
		}
	}
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func (ipt *Input) setup() bool {
	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return true
		default:
			// nil
		}
		time.Sleep(time.Second)

		if err := ipt.setupDo(); err != nil {
			continue
		}
		break
	}

	return false
}

func (ipt *Input) setupDo() error {
	if ipt.httpClient == nil {
		ipt.httpClient = &http.Client{Timeout: 5 * time.Second}
	}

	if ipt.TLSCA != "" {
		tlsconfig := &dknet.TLSClientConfig{
			CaCerts:            []string{ipt.TLSCA},
			Cert:               ipt.TLSCert,
			CertKey:            ipt.TLSKey,
			InsecureSkipVerify: ipt.InsecureSkipVerify,
		}

		tc, err := tlsconfig.TLSConfig()
		if err != nil {
			l.Errorf("compose TLS: %s", err)
			return err
		}
		ipt.httpClient.Transport = &http.Transport{TLSClientConfig: tc}
	}

	if ipt.isLookupd() {
		u, err := buildURL(fmt.Sprintf(lookupdPattern, ipt.Lookupd))
		if err != nil {
			l.Errorf("build URL: %s", err)
			return err
		}
		ipt.lookupdEndpoint = u.String()
		if err := ipt.updateEndpointListByLookupd(ipt.lookupdEndpoint); err != nil {
			l.Error(err)
			return err
		}
	} else {
		if len(ipt.NSQDs) == 0 {
			return fmt.Errorf("invalid nsqd endpoints")
		}
		for _, n := range ipt.NSQDs {
			u, err := buildURL(fmt.Sprintf(nsqdStatsPattern, n))
			if err != nil {
				l.Errorf("build URL: %s", err)
				return err
			}
			ipt.nsqdEndpointList[u.String()] = nil
		}
	}

	var err error
	ipt.duration, err = timex.ParseDuration(ipt.Interval)
	if err != nil {
		l.Warnf("parse duration error: %s", err)
	}
	if ipt.duration < minInterval {
		ipt.duration = defaultInterval
		l.Warnf("interval should large %s, got %s, use default interval %s",
			minInterval, ipt.Interval, defaultInterval)
	}

	return nil
}

func (ipt *Input) isLookupd() bool {
	return ipt.Lookupd != ""
}

func (ipt *Input) gather() ([]*point.Point, error) {
	if len(ipt.nsqdEndpointList) == 0 {
		l.Warn("endpoint list is empty")
		return nil, nil
	}

	st := newStats(ipt)

	for endpoint := range ipt.nsqdEndpointList {
		body, err := ipt.httpGet(endpoint)
		if err != nil {
			l.Errorf("httpGet: %s, ignored", err)
			continue
		}

		if err := st.add(getURLHost(endpoint), body); err != nil {
			l.Errorf("st.add: %s, ignored", err.Error())
			continue
		}
	}

	return st.makePoint(ipt.Tags)
}

func (ipt *Input) updateEndpointListByLookupd(lookupdEndpoint string) error {
	body, err := ipt.httpGet(lookupdEndpoint)
	if err != nil {
		return err
	}

	var endpoints []string
	lk := &LookupNodes{}
	if err := json.Unmarshal(body, lk); err != nil {
		return fmt.Errorf("error parsing response: %w", err)
	}

	for _, p := range lk.Producers {
		// TODO
		// protocol 是否根据 TLS 配置决定使用 https/http ?
		u, err := buildURL(fmt.Sprintf("http://"+nsqdStatsPattern, p.BroadcastAddress+":"+strconv.Itoa((p.HTTPPort))))
		if err != nil {
			l.Warnf("build URL: %s", err)
			continue
		}

		// fix lookupd and producer in the same remote host while Datakit is in another.
		// lookupdURL, err := url.Parse(lookupdEndpoint)
		// if err != nil {
		// 	l.Warnf("url.Parse failed: %v", err)
		// } else {
		// 	lookupdHost, _, err := net.SplitHostPort(lookupdURL.Host)
		// 	if err != nil {
		// 		l.Warnf("net.SplitHostPort failed: %v", err)
		// 	} else {
		// 		uHost, uPort, err := net.SplitHostPort(u.Host)
		// 		if err != nil {
		// 			l.Warnf("net.SplitHostPort failed: %v", err)
		// 		} else {
		// 			if !net.ParseIP(lookupdHost).IsLoopback() && net.ParseIP(uHost).IsLoopback() {
		// 				// replace remote producer host with lookupd host.
		// 				u.Host = net.JoinHostPort(lookupdHost, uPort)
		// 			}
		// 		}
		// 	}
		// }

		endpoints = append(endpoints, u.String())
	}

	for _, endpoint := range endpoints {
		if _, ok := ipt.nsqdEndpointList[endpoint]; !ok {
			ipt.nsqdEndpointList[endpoint] = nil
		}
	}

	return nil
}

func (ipt *Input) httpGet(u string) ([]byte, error) {
	r, err := ipt.httpClient.Get(u)
	if err != nil {
		return nil, fmt.Errorf("error while polling %s: %w", u, err)
	}
	defer r.Body.Close() //nolint:errcheck

	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s returned HTTP status %s", u, r.Status)
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf(`error reading body: %w`, err)
	}

	return body, nil
}

func (ipt *Input) Pause() error {
	tick := time.NewTicker(inputs.ElectionPauseTimeout)
	defer tick.Stop()
	select {
	case ipt.pauseCh <- true:
		return nil
	case <-tick.C:
		return fmt.Errorf("pause %s failed", inputName)
	}
}

func (ipt *Input) Resume() error {
	tick := time.NewTicker(inputs.ElectionResumeTimeout)
	defer tick.Stop()
	select {
	case ipt.pauseCh <- false:
		return nil
	case <-tick.C:
		return fmt.Errorf("resume %s failed", inputName)
	}
}

func buildURL(u string) (*url.URL, error) {
	addr, err := url.Parse(u)
	if err != nil {
		return nil, fmt.Errorf("unable to parse address '%s': %w", u, err)
	}
	return addr, nil
}

func getURLHost(urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		host, _, err := net.SplitHostPort(urlStr)
		if err != nil {
			l.Errorf("url.Parse and net.SplitHostPort failed: %v", err)
			return unknownHost
		}
		return host
	}
	return u.Host
}

func defaultInput() *Input {
	return &Input{
		Tags:             make(map[string]string),
		nsqdEndpointList: make(map[string]interface{}),
		pauseCh:          make(chan bool, maxPauseCh),
		httpClient:       &http.Client{Timeout: 5 * time.Second},
		Election:         true,
		semStop:          cliutils.NewSem(),
		feeder:           dkio.DefaultFeeder(),
		Tagger:           datakit.DefaultGlobalTagger(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
