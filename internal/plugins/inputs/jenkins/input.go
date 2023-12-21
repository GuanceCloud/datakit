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
	"github.com/GuanceCloud/cliutils/point"
	"github.com/gin-gonic/gin"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
)

func (*Input) SampleConfig() string {
	return sample
}

func (*Input) Catalog() string {
	return inputName
}

func (*Input) LogExamples() map[string]map[string]string {
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

func (ipt *Input) setup() {
	ipt.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval.Duration)

	client, err := ipt.createHTTPClient()
	if err != nil {
		l.Errorf("[error] jenkins init client err:%s", err.Error())
		return
	}
	ipt.client = client

	if ipt.EnableCIVisibility {
		ipt.setupServer()
		l.Infof("start listening to jenkins CI events at %s", ipt.CIEventPort)
	}
}

func (ipt *Input) setupServer() {
	router := gin.Default()
	router.PUT("/v0.3/traces", gin.WrapH(ipt))
	ipt.srv = &http.Server{
		Addr:        ipt.CIEventPort,
		Handler:     router,
		IdleTimeout: 120 * time.Second,
	}
	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_jenkins"})
	g.Go(func(ctx context.Context) error {
		if err := ipt.srv.ListenAndServe(); err != nil {
			l.Errorf("jenkins CI event server shutdown: %v", err)
		}
		return nil
	})
}

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)
	l.Info("jenkins start")

	ipt.setup()
	if !ipt.EnableCollect {
		l.Info("metric collecting is disabled")
	}
	tick := time.NewTicker(ipt.Interval.Duration)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			if !ipt.EnableCollect {
				continue
			}
			ipt.start = time.Now()
			ipt.getPluginMetric()
			if len(ipt.collectCache) > 0 {
				err := ipt.feeder.Feed(inputName, point.Metric, ipt.collectCache,
					&dkio.Option{CollectCost: time.Since(ipt.start)})
				ipt.collectCache = ipt.collectCache[:0]
				if err != nil {
					ipt.lastErr = err
					l.Errorf(err.Error())
					continue
				}
			}
			if ipt.lastErr != nil {
				ipt.feeder.FeedLastError(ipt.lastErr.Error(),
					dkio.WithLastErrorInput(inputName),
				)
				ipt.lastErr = nil
			}
		case <-datakit.Exit.Wait():
			ipt.exit()
			ipt.shutdownServer()
			l.Info("jenkins exited")
			return

		case <-ipt.semStop.Wait():
			ipt.exit()
			ipt.shutdownServer()
			l.Info("jenkins returned")
			return
		}
	}
}

func (ipt *Input) shutdownServer() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if ipt.srv != nil {
		if err := ipt.srv.Shutdown(ctx); err != nil {
			l.Errorf("jenkins CI event server failed to shutdown: %v", err)
		} else {
			l.Infof("jenkins CI event server is shutdown")
		}
	}
}

func (ipt *Input) exit() {
	if ipt.tail != nil {
		ipt.tail.Close()
		l.Info("jenkins log exit")
	}
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
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
		MultilinePatterns: []string{`^\d{4}-\d{2}-\d{2}`},
		GlobalTags:        inputs.MergeTags(ipt.Tagger.HostTags(), ipt.Tags, ""),
		Done:              ipt.semStop.Wait(),
	}

	var err error
	ipt.tail, err = tailer.NewTailer(ipt.Log.Files, opt)
	if err != nil {
		l.Error(err)
		ipt.feeder.FeedLastError(err.Error(),
			dkio.WithLastErrorInput(inputName),
		)
		return
	}

	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_jenkins"})
	g.Go(func(ctx context.Context) error {
		ipt.tail.Start()
		return nil
	})
}

func (ipt *Input) requestJSON(u string, target interface{}) error {
	u = fmt.Sprintf("%s%s", ipt.URL, u)

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return err
	}

	// req.SetBasicAuth(n.Username, n.Password)

	resp, err := ipt.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != 200 {
		return fmt.Errorf("response err:%+#v", resp)
	}

	return json.NewDecoder(resp.Body).Decode(target)
}

func (ipt *Input) createHTTPClient() (*http.Client, error) {
	tlsCfg, err := ipt.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	if ipt.ResponseTimeout.Duration < time.Second {
		ipt.ResponseTimeout.Duration = time.Second * 5
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
		},
		Timeout: ipt.ResponseTimeout.Duration,
	}

	return client, nil
}

func (*Input) AvailableArchs() []string {
	return datakit.AllOSWithElection
}

func (ipt *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&Measurement{},
		&jenkinsPipelineMeasurement{},
		&jenkinsJobMeasurement{},
	}
}

func defaultInput() *Input {
	return &Input{
		Interval: datakit.Duration{Duration: time.Second * 30},
		semStop:  cliutils.NewSem(),
		feeder:   dkio.DefaultFeeder(),
		Tagger:   datakit.DefaultGlobalTagger(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
