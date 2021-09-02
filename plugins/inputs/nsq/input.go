package nsq

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
	timex "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/time"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "nsq"
	catalog   = "nsq"

	minInterval     = time.Second * 3
	defaultInterval = time.Second * 10

	requestPattern = `%s/stats?format=json`
)

var l = logger.DefaultSLogger(inputName)

type Input struct {
	Endpoint           string            `toml:"endpoint"`
	Interval           string            `toml:"interval"`
	TLSCA              string            `toml:"tls_ca"`
	TLSCert            string            `toml:"tls_cert"`
	TLSKey             string            `toml:"tls_key"`
	InsecureSkipVerify bool              `toml:"insecure_skip_verify"`
	Tags               map[string]string `toml:"tags"`

	httpClient *http.Client
	duration   time.Duration

	pauseCh chan bool
	pause   bool
}

func newInput() *Input {
	return &Input{
		Tags:       make(map[string]string),
		pauseCh:    make(chan bool, 1),
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func (*Input) SampleConfig() string { return sampleCfg }

func (*Input) Catalog() string { return catalog }

func (*Input) AvailableArchs() []string { return datakit.AllArch }

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&nsqServerMeasurement{},
		&nsqTopicMeasurement{},
		&nsqChannelMeasurement{},
		&nsqClientMeasurement{},
	}
}

func (this *Input) Run() {
	l = logger.SLogger(inputName)

	if this.setup() {
		return
	}

	ticker := time.NewTicker(this.duration)
	defer ticker.Stop()

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return

		case <-ticker.C:
			if this.pause {
				l.Debugf("not leader, skipped")
				continue
			}
			this.gather()

		case this.pause = <-this.pauseCh:
			// nil
		}
	}
}

func (this *Input) setup() bool {
	var err error
	this.duration, err = timex.ParseDuration(this.Interval)
	if err != nil {
		l.Warnf("parse duration error: %s", err)
	}
	if this.duration < minInterval {
		this.duration = defaultInterval
		l.Debugf("interval should large %s, got %s, use default interval %s", minInterval, this.Interval, defaultInterval)
	}

	if this.httpClient == nil {
		this.httpClient = &http.Client{Timeout: 5 * time.Second}
	}

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return true
		default:
			// nil
		}
		time.Sleep(time.Second)

		if this.TLSCA != "" {
			tlsconfig := &net.TLSClientConfig{
				CaCerts:            []string{this.TLSCA},
				Cert:               this.TLSCert,
				CertKey:            this.TLSKey,
				InsecureSkipVerify: this.InsecureSkipVerify,
			}

			tc, err := tlsconfig.TLSConfig()
			if err != nil {
				l.Error(err)
				continue
			}
			this.httpClient.Transport = &http.Transport{TLSClientConfig: tc}
		}

		break
	}

	return false
}

func (this *Input) Pause() error {
	tick := time.NewTicker(time.Second * 5)
	defer tick.Stop()
	select {
	case this.pauseCh <- true:
		return nil
	case <-tick.C:
		return fmt.Errorf("pause %s failed", inputName)
	}
}

func (this *Input) Resume() error {
	tick := time.NewTicker(time.Second * 5)
	defer tick.Stop()
	select {
	case this.pauseCh <- false:
		return nil
	case <-tick.C:
		return fmt.Errorf("resume %s failed", inputName)
	}
}

func (this *Input) gather() {
	start := time.Now()

	pts, err := this.gatherEndpoint(this.Endpoint)
	if err != nil {
		l.Error(err)
	}
	if len(pts) == 0 {
		return
	}

	if err := io.Feed(inputName, datakit.Metric, pts, &io.Option{CollectCost: time.Since(start)}); err != nil {
		l.Error(err)
	}
}

func (this *Input) gatherEndpoint(e string) ([]*io.Point, error) {
	u, err := buildURL(e)
	if err != nil {
		return nil, err
	}

	r, err := this.httpClient.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("error while polling %s: %s", u.String(), err)
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s returned HTTP status %s", u.String(), r.Status)
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf(`error reading body: %s`, err)
	}

	return newStats(u.Host, this.Tags).parse(body)
}

func buildURL(e string) (*url.URL, error) {
	u := fmt.Sprintf(requestPattern, e)
	addr, err := url.Parse(u)
	if err != nil {
		return nil, fmt.Errorf("unable to parse address '%s': %s", u, err)
	}
	return addr, nil
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return newInput()
	})
}
