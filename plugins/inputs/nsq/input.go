// Package nsq collects NSQ metrics.
package nsq

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
	timex "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/time"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var _ inputs.ElectionInput = (*Input)(nil)

const (
	inputName = "nsq"
	catalog   = "nsq"

	nsqdStatsPattern = "%s/stats?format=json"
	lookupdPattern   = "%s/nodes"
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

	pauseCh chan bool
	pause   bool

	semStop          *cliutils.Sem // start stop signal
	semStopCompleted *cliutils.Sem // stop completed signal
}

var maxPauseCh = inputs.ElectionPauseChannelLength

func newInput() *Input {
	return &Input{
		Tags:             make(map[string]string),
		nsqdEndpointList: make(map[string]interface{}),
		pauseCh:          make(chan bool, maxPauseCh),
		httpClient:       &http.Client{Timeout: 5 * time.Second},

		semStop:          cliutils.NewSem(),
		semStopCompleted: cliutils.NewSem(),
	}
}

func (*Input) SampleConfig() string { return sampleCfg }

func (*Input) Catalog() string { return catalog }

func (*Input) AvailableArchs() []string { return datakit.AllArch }

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

			if ipt.semStopCompleted != nil {
				ipt.semStopCompleted.Close()
			}
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

			if err := io.Feed(inputName, datakit.Metric, pts, &io.Option{CollectCost: time.Since(start)}); err != nil {
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

		// wait stop completed
		if ipt.semStopCompleted != nil {
			for range ipt.semStopCompleted.Wait() {
				return
			}
		}
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
		tlsconfig := &net.TLSClientConfig{
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

func (ipt *Input) gather() ([]*io.Point, error) {
	if len(ipt.nsqdEndpointList) == 0 {
		l.Warn("endpoint list is empty")
		return nil, nil
	}

	st := newStats()

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
		return "unknown"
	}
	return u.Host
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return newInput()
	})
}
