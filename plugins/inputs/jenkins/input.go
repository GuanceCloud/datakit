// Package jenkins collects Jenkins metrics.
package jenkins

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func (*Input) SampleConfig() string {
	return sample
}

func (*Input) Catalog() string {
	return inputName
}

func (*Input) PipelineConfig() map[string]string {
	pipelineMap := map[string]string{
		"jenkins": pipelineCfg,
	}
	return pipelineMap
}

func (n *Input) GetPipeline() []*tailer.Option {
	return []*tailer.Option{
		{
			Source:  inputName,
			Service: inputName,
			Pipeline: func() string {
				if n.Log != nil {
					return n.Log.Pipeline
				}
				return ""
			}(),
		},
	}
}

func (n *Input) setup() {
	n.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, n.Interval.Duration)

	client, err := n.createHTTPClient()
	if err != nil {
		l.Errorf("[error] jenkins init client err:%s", err.Error())
		return
	}
	n.client = client
}

func (n *Input) Run() {
	l = logger.SLogger(inputName)
	l.Info("jenkins start")

	n.setup()

	tick := time.NewTicker(n.Interval.Duration)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			n.start = time.Now()
			n.getPluginMetric()
			if len(n.collectCache) > 0 {
				err := inputs.FeedMeasurement(inputName,
					datakit.Metric,
					n.collectCache,
					&io.Option{CollectCost: time.Since(n.start)})
				n.collectCache = n.collectCache[:0]
				if err != nil {
					n.lastErr = err
					l.Errorf(err.Error())
					continue
				}
			}
			if n.lastErr != nil {
				io.FeedLastError(inputName, n.lastErr.Error())
				n.lastErr = nil
			}
		case <-datakit.Exit.Wait():
			n.exit()
			l.Info("jenkins exit")
			return

		case <-n.semStop.Wait():
			n.exit()
			l.Info("jenkins return")
			return
		}
	}
}

func (n *Input) exit() {
	if n.tail != nil {
		n.tail.Close()
		l.Info("jenkins log exit")
	}
}

func (n *Input) Terminate() {
	if n.semStop != nil {
		n.semStop.Close()
	}
}

func (n *Input) RunPipeline() {
	if n.Log == nil || len(n.Log.Files) == 0 {
		return
	}

	if n.Log.Pipeline == "" {
		n.Log.Pipeline = inputName + ".p" // use default
	}

	opt := &tailer.Option{
		Source:            inputName,
		Service:           inputName,
		GlobalTags:        n.Tags,
		IgnoreStatus:      n.Log.IgnoreStatus,
		CharacterEncoding: n.Log.CharacterEncoding,
		MultilineMatch:    `^\d{4}-\d{2}-\d{2}`,
	}

	pl, err := config.GetPipelinePath(n.Log.Pipeline)
	if err != nil {
		l.Error(err)
		io.FeedLastError(inputName, err.Error())
		return
	}
	if _, err := os.Stat(pl); err != nil {
		l.Warn("%s missing: %s", pl, err.Error())
	} else {
		opt.Pipeline = pl
	}

	n.tail, err = tailer.NewTailer(n.Log.Files, opt)
	if err != nil {
		l.Error(err)
		io.FeedLastError(inputName, err.Error())
		return
	}

	go n.tail.Start()
}

func (n *Input) requestJSON(u string, target interface{}) error {
	u = fmt.Sprintf("%s%s", n.URL, u)

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return err
	}

	// req.SetBasicAuth(n.Username, n.Password)

	resp, err := n.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != 200 {
		return fmt.Errorf("response err:%+#v", resp)
	}

	return json.NewDecoder(resp.Body).Decode(target)
}

func (n *Input) createHTTPClient() (*http.Client, error) {
	tlsCfg, err := n.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	if n.ResponseTimeout.Duration < time.Second {
		n.ResponseTimeout.Duration = time.Second * 5
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
		},
		Timeout: n.ResponseTimeout.Duration,
	}

	return client, nil
}

func (*Input) AvailableArchs() []string {
	return datakit.AllArch
}

func (n *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&Measurement{},
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		s := &Input{
			Interval: datakit.Duration{Duration: time.Second * 30},

			semStop: cliutils.NewSem(),
		}
		return s
	})
}
