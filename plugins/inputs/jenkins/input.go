// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package jenkins collects Jenkins metrics.
package jenkins

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/gin-gonic/gin"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
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

func (n *Input) LogExamples() map[string]map[string]string {
	return map[string]map[string]string{
		"jenkins": {
			"Jenkins log": `2021-05-18 03:08:58.053+0000 [id=32]	INFO	jenkins.InitReactorRunner$1#onAttained: Started all plugins`,
		},
	}
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

	if n.EnableCIVisibility {
		n.setupServer()
		l.Infof("start listening to jenkins CI events at %s", n.CIEventPort)
	}
}

func (n *Input) setupServer() {
	router := gin.Default()
	router.PUT("/v0.3/traces", gin.WrapH(n))
	n.srv = &http.Server{
		Addr:        n.CIEventPort,
		Handler:     router,
		IdleTimeout: 120 * time.Second,
	}
	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_jenkins"})
	g.Go(func(ctx context.Context) error {
		if err := n.srv.ListenAndServe(); err != nil {
			l.Errorf("jenkins CI event server shutdown: %v", err)
		}
		return nil
	})
}

func (n *Input) Run() {
	l = logger.SLogger(inputName)
	l.Info("jenkins start")

	n.setup()
	if !n.EnableCollect {
		l.Info("metric collecting is disabled")
	}
	tick := time.NewTicker(n.Interval.Duration)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			if !n.EnableCollect {
				continue
			}
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
			n.shutdownServer()
			l.Info("jenkins exited")
			return

		case <-n.semStop.Wait():
			n.exit()
			n.shutdownServer()
			l.Info("jenkins returned")
			return
		}
	}
}

func (n *Input) shutdownServer() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := n.srv.Shutdown(ctx); err != nil {
		l.Errorf("jenkins CI event server failed to shutdown: %v", err)
	} else {
		l.Infof("jenkins CI event server is shutdown")
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
		Pipeline:          n.Log.Pipeline,
		GlobalTags:        n.Tags,
		IgnoreStatus:      n.Log.IgnoreStatus,
		CharacterEncoding: n.Log.CharacterEncoding,
		MultilinePatterns: []string{`^\d{4}-\d{2}-\d{2}`},
		Done:              n.semStop.Wait(),
	}

	var err error
	n.tail, err = tailer.NewTailer(n.Log.Files, opt)
	if err != nil {
		l.Error(err)
		io.FeedLastError(inputName, err.Error())
		return
	}

	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_jenkins"})
	g.Go(func(ctx context.Context) error {
		n.tail.Start()
		return nil
	})
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
	return datakit.AllOSWithElection
}

func (n *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&Measurement{},
		&jenkinsPipelineMeasurement{},
		&jenkinsJobMeasurement{},
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
