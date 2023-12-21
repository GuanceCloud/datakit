// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package apache collects Apache metrics.
package apache

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
)

var _ inputs.ElectionInput = (*Input)(nil)

var (
	l = logger.DefaultSLogger(inputName)
	g = datakit.G("inputs_apache")
)

type Input struct {
	URLsDeprecated []string `toml:"urls,omitempty"`

	URL      string            `toml:"url"`
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
	host string

	start  time.Time
	tail   *tailer.Tailer
	client *http.Client

	Election bool `toml:"election"`
	pause    bool
	pauseCh  chan bool

	feeder  dkio.Feeder
	semStop *cliutils.Sem // start stop signal
	Tagger  datakit.GlobalTagger
}

func (ipt *Input) ElectionEnabled() bool {
	return ipt.Election
}

//nolint:lll
func (ipt *Input) LogExamples() map[string]map[string]string {
	return map[string]map[string]string{
		inputName: {
			"Apache error log":  `[Tue May 19 18:39:45.272121 2021] [access_compat:error] [pid 9802] [client ::1:50547] AH01797: client denied by server configuration: /Library/WebServer/Documents/server-status`,
			"Apache access log": `127.0.0.1 - - [17/May/2021:14:51:09 +0800] "GET /server-status?auto HTTP/1.1" 200 917`,
		},
	}
}

var maxPauseCh = inputs.ElectionPauseChannelLength

func (*Input) SampleConfig() string { return sample }

func (*Input) Catalog() string { return inputName }

func (*Input) AvailableArchs() []string { return datakit.AllOSWithElection }

func (*Input) SampleMeasurement() []inputs.Measurement { return []inputs.Measurement{&Measurement{}} }

func (*Input) PipelineConfig() map[string]string { return map[string]string{"apache": pipeline} }

func (ipt *Input) GetPipeline() []*tailer.Option {
	return []*tailer.Option{
		{
			Source:  inputName,
			Service: inputName,
			Pipeline: func() string {
				if ipt.Log != nil {
					return ipt.Log.Pipeline
				}
				return ""
			}(),
		},
	}
}

func (ipt *Input) RunPipeline() {
	if ipt.Log == nil || len(ipt.Log.Files) == 0 {
		return
	}

	opt := &tailer.Option{
		Source:            inputName,
		Service:           inputName,
		Pipeline:          ipt.Log.Pipeline,
		IgnoreStatus:      ipt.Log.IgnoreStatus,
		CharacterEncoding: ipt.Log.CharacterEncoding,
		MultilinePatterns: []string{`^\[\w+ \w+ \d+`},
		GlobalTags:        inputs.MergeTags(ipt.Tagger.HostTags(), ipt.Tags, ""),
		Done:              ipt.semStop.Wait(),
	}

	var err error
	ipt.tail, err = tailer.NewTailer(ipt.Log.Files, opt)
	if err != nil {
		l.Error(err)
		ipt.feeder.FeedLastError(err.Error(),
			dkio.WithLastErrorInput(inputName),
			dkio.WithLastErrorCategory(point.Metric),
		)
		return
	}

	g.Go(func(ctx context.Context) error {
		ipt.tail.Start()
		return nil
	})
}

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)
	l.Info("apache start")
	ipt.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval.Duration)

	client, err := ipt.createHTTPClient()
	if err != nil {
		l.Errorf("[error] apache init client err:%s", err.Error())
		return
	}
	ipt.client = client

	if err := ipt.setHost(); err != nil {
		l.Errorf("failed to set host: %v", err)
	}

	tick := time.NewTicker(ipt.Interval.Duration)
	defer tick.Stop()

	for {
		select {
		case <-datakit.Exit.Wait():
			ipt.exit()
			l.Info("apache exit")
			return

		case <-ipt.semStop.Wait():
			ipt.exit()
			l.Info("apache return")
			return

		case <-tick.C:
			if ipt.pause {
				l.Debugf("not leader, skipped")
				continue
			}

			m, err := ipt.getMetric()
			if err != nil {
				ipt.feeder.FeedLastError(err.Error(),
					dkio.WithLastErrorInput(inputName),
					dkio.WithLastErrorCategory(point.Metric),
				)
			}

			if m != nil {
				if err := ipt.feeder.Feed(inputName, point.Metric, []*point.Point{m}, &dkio.Option{CollectCost: time.Since(ipt.start)}); err != nil {
					l.Warnf("Feed failed: %s, ignored", err.Error())
				}
			}

		case ipt.pause = <-ipt.pauseCh:
			// nil
		}
	}
}

func (ipt *Input) exit() {
	if ipt.tail != nil {
		ipt.tail.Close()
		l.Info("apache log exit")
	}
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func (ipt *Input) createHTTPClient() (*http.Client, error) {
	tlsCfg, err := ipt.ClientConfig.TLSConfig()
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

func (ipt *Input) getMetric() (*point.Point, error) {
	ipt.start = time.Now()
	req, err := http.NewRequest("GET", ipt.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("error on new request to %s : %w", ipt.URL, err)
	}

	if len(ipt.Username) != 0 && len(ipt.Password) != 0 {
		req.SetBasicAuth(ipt.Username, ipt.Password)
	}

	resp, err := ipt.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error on request to %s : %w", ipt.URL, err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s returned HTTP status %s", ipt.URL, resp.Status)
	}
	return ipt.parse(resp.Body)
}

func (ipt *Input) parse(body io.Reader) (*point.Point, error) {
	sc := bufio.NewScanner(body)

	tags := map[string]string{
		"url": ipt.URL,
	}
	if ipt.host != "" {
		tags["host"] = ipt.host
	}
	for k, v := range ipt.Tags {
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
			key, part := strings.ReplaceAll(parts[0], " ", ""), strings.TrimSpace(parts[1])
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
					waitingForConnection: 0,
					startingUp:           0,
					readingRequest:       0,
					sendingReply:         0,
					keepAlive:            0,
					dnsLookup:            0,
					closingConnection:    0,
					logging:              0,
					gracefullyFinishing:  0,
					idleCleanup:          0,
					openSlot:             0,
				}
				for _, c := range part {
					switch c {
					case '_':
						scoreboard[waitingForConnection]++
					case 'S':
						scoreboard[startingUp]++
					case 'R':
						scoreboard[readingRequest]++
					case 'W':
						scoreboard[sendingReply]++
					case 'K':
						scoreboard[keepAlive]++
					case 'D':
						scoreboard[dnsLookup]++
					case 'C':
						scoreboard[closingConnection]++
					case 'L':
						scoreboard[logging]++
					case 'G':
						scoreboard[gracefullyFinishing]++
					case 'I':
						scoreboard[idleCleanup]++
					case '.':
						scoreboard[openSlot]++
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

	if ipt.Election {
		tags = inputs.MergeTags(ipt.Tagger.ElectionTags(), tags, ipt.URL)
	} else {
		tags = inputs.MergeTags(ipt.Tagger.HostTags(), tags, ipt.URL)
	}

	metric.tags = tags

	return metric.Point(), nil
}

func (ipt *Input) setHost() error {
	u, err := url.Parse(ipt.URL)
	if err != nil {
		return err
	}
	var host string
	h, _, err := net.SplitHostPort(u.Host)
	if err != nil {
		host = u.Host
	} else {
		host = h
	}
	if host != "localhost" && !net.ParseIP(host).IsLoopback() {
		ipt.host = host
	}
	return nil
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

func defaultInput() *Input {
	return &Input{
		Interval: datakit.Duration{Duration: time.Second * 30},
		pauseCh:  make(chan bool, maxPauseCh),
		Election: true,
		feeder:   dkio.DefaultFeeder(),
		semStop:  cliutils.NewSem(),
		Tagger:   datakit.DefaultGlobalTagger(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
