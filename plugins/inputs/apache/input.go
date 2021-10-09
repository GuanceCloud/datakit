package apache

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf/plugins/common/tls"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
	iod "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var _ inputs.ElectionInput = (*Input)(nil)

var l = logger.DefaultSLogger(inputName)

type Input struct {
	Url      string            `toml:"url"`
	Username string            `toml:"username,omitempty"`
	Password string            `toml:"password,omitempty"`
	Interval datakit.Duration  `toml:"interval,omitempty"`
	Tags     map[string]string `toml:"tags,omitempty"`
	Log      *struct {
		Files             []string `toml:"files"`
		Pipeline          string   `toml:"pipeline"`
		IgnoreStatus      []string `toml:"ignore"`
		CharacterEncoding string   `toml:"character_encoding"`
	} `toml:"log"`

	tls.ClientConfig

	start        time.Time
	tail         *tailer.Tailer
	collectCache []inputs.Measurement
	client       *http.Client
	lastErr      error

	pause   bool
	pauseCh chan bool
}

var maxPauseCh = inputs.ElectionPauseChannelLength

func newInput() *Input {
	return &Input{
		Interval: datakit.Duration{Duration: time.Second * 30},
		pauseCh:  make(chan bool, maxPauseCh),
	}
}

func (*Input) SampleConfig() string { return sample }

func (*Input) Catalog() string { return inputName }

func (*Input) AvailableArchs() []string { return datakit.AllArch }

func (*Input) SampleMeasurement() []inputs.Measurement { return []inputs.Measurement{&Measurement{}} }

func (*Input) PipelineConfig() map[string]string { return map[string]string{"apache": pipeline} }

func (n *Input) RunPipeline() {
	if n.Log == nil || len(n.Log.Files) == 0 {
		return
	}

	if n.Log.Pipeline == "" {
		n.Log.Pipeline = "apache.p" // use default
	}

	opt := &tailer.Option{
		Source:            inputName,
		Service:           inputName,
		GlobalTags:        n.Tags,
		IgnoreStatus:      n.Log.IgnoreStatus,
		CharacterEncoding: n.Log.CharacterEncoding,
		MultilineMatch:    `^\[\w+ \w+ \d+`,
	}

	pl := filepath.Join(datakit.PipelineDir, n.Log.Pipeline)
	if _, err := os.Stat(pl); err != nil {
		l.Warn("%s missing: %s", pl, err.Error())
	} else {
		opt.Pipeline = pl
	}

	var err error
	n.tail, err = tailer.NewTailer(n.Log.Files, opt)
	if err != nil {
		l.Error(err)
		iod.FeedLastError(inputName, err.Error())
		return
	}

	go n.tail.Start()
}

func (n *Input) Run() {
	l = logger.SLogger(inputName)
	l.Info("apache start")
	n.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, n.Interval.Duration)

	client, err := n.createHttpClient()
	if err != nil {
		l.Errorf("[error] apache init client err:%s", err.Error())
		return
	}
	n.client = client

	tick := time.NewTicker(n.Interval.Duration)
	defer tick.Stop()

	for {
		select {
		case <-datakit.Exit.Wait():
			if n.tail != nil {
				n.tail.Close()
				l.Info("apache log exit")
			}
			l.Info("apache exit")
			return

		case <-tick.C:
			if n.pause {
				l.Debugf("not leader, skipped")
				continue
			}

			m, err := n.getMetric()
			if err != nil {
				iod.FeedLastError(inputName, err.Error())
			}

			if m != nil {
				if err := inputs.FeedMeasurement(inputName,
					datakit.Metric,
					[]inputs.Measurement{m},
					&iod.Option{CollectCost: time.Since(n.start)}); err != nil {
					l.Warnf("inputs.FeedMeasurement: %s, ignored", err)
				}
			}

		case n.pause = <-n.pauseCh:
			// nil
		}
	}
}

func (n *Input) createHttpClient() (*http.Client, error) {
	tlsCfg, err := n.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
		},
		Timeout: time.Second * 5,
	}

	return client, nil
}

func (n *Input) getMetric() (*Measurement, error) {
	n.start = time.Now()
	req, err := http.NewRequest("GET", n.Url, nil)
	if err != nil {
		return nil, fmt.Errorf("error on new request to %s : %s", n.Url, err)
	}

	if len(n.Username) != 0 && len(n.Password) != 0 {
		req.SetBasicAuth(n.Username, n.Password)
	}

	resp, err := n.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error on request to %s : %s", n.Url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s returned HTTP status %s", n.Url, resp.Status)
	}
	return n.parse(resp.Body)
}

func (n *Input) parse(body io.Reader) (*Measurement, error) {
	sc := bufio.NewScanner(body)

	tags := map[string]string{
		"url": n.Url,
	}
	for k, v := range n.Tags {
		tags[k] = v
	}
	metric := &Measurement{
		name:   inputName,
		fields: map[string]interface{}{},
		ts:     time.Now(),
	}

	for sc.Scan() {
		line := sc.Text()
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			key, part := strings.Replace(parts[0], " ", "", -1), strings.TrimSpace(parts[1])
			if tagKey, ok := tagMap[key]; ok {
				tags[tagKey] = part
			}

			fieldKey, ok := filedMap[key]
			if !ok {
				continue
			}
			switch key {
			case "CPULoad":
				value, err := strconv.ParseFloat(part, 64)
				if err != nil {
					l.Error(err.Error())
					continue
				}
				metric.fields[fieldKey] = value
			case "Scoreboard":
				scoreboard := map[string]int{
					WaitingForConnection: 0,
					StartingUp:           0,
					ReadingRequest:       0,
					SendingReply:         0,
					KeepAlive:            0,
					DNSLookup:            0,
					ClosingConnection:    0,
					Logging:              0,
					GracefullyFinishing:  0,
					IdleCleanup:          0,
					OpenSlot:             0,
				}
				for _, c := range part {
					switch c {
					case '_':
						scoreboard[WaitingForConnection]++
					case 'S':
						scoreboard[StartingUp]++
					case 'R':
						scoreboard[ReadingRequest]++
					case 'W':
						scoreboard[SendingReply]++
					case 'K':
						scoreboard[KeepAlive]++
					case 'D':
						scoreboard[DNSLookup]++
					case 'C':
						scoreboard[ClosingConnection]++
					case 'L':
						scoreboard[Logging]++
					case 'G':
						scoreboard[GracefullyFinishing]++
					case 'I':
						scoreboard[IdleCleanup]++
					case '.':
						scoreboard[OpenSlot]++
					}
				}
				for k, v := range scoreboard {
					metric.fields[k] = v
				}
			default:
				value, err := strconv.ParseInt(part, 10, 64)
				if err != nil {
					l.Error(err.Error())
					continue
				}
				if fieldKey == "Total kBytes" {
					// kbyte to byte
					metric.fields[fieldKey] = value * 1024
					continue
				}
				metric.fields[fieldKey] = value
			}
		}
	}

	metric.tags = tags

	return metric, nil
}

func (n *Input) Pause() error {
	tick := time.NewTicker(inputs.ElectionPauseTimeout)
	defer tick.Stop()
	select {
	case n.pauseCh <- true:
		return nil
	case <-tick.C:
		return fmt.Errorf("pause %s failed", inputName)
	}
}

func (n *Input) Resume() error {
	tick := time.NewTicker(inputs.ElectionResumeTimeout)
	defer tick.Stop()
	select {
	case n.pauseCh <- false:
		return nil
	case <-tick.C:
		return fmt.Errorf("resume %s failed", inputName)
	}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return newInput()
	})
}
