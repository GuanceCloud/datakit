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
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
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

func (ipt *Input) GetPipeline() []tailer.Option {
	opts := []tailer.Option{
		tailer.WithSource(inputName),
		tailer.WithService(inputName),
	}
	if ipt.Log != nil {
		opts = append(opts, tailer.WithPipeline(ipt.Log.Pipeline))
	}
	return opts
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
	router.GET("/info", gin.WrapH(ipt))
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

	ipt.ptsTime = ntp.Now()
	for {
		if ipt.pause {
			l.Debug("%s election paused", inputName)
		} else if ipt.EnableCollect {
			collectStart := time.Now()
			ipt.getPluginMetric()

			if len(ipt.collectCache) > 0 {
				if err := ipt.feeder.Feed(point.Metric, ipt.collectCache,
					dkio.WithCollectCost(time.Since(collectStart)),
					dkio.WithElection(ipt.Election),
					dkio.WithSource(inputName),
				); err != nil {
					ipt.lastErr = err
					l.Errorf(err.Error())
				}

				ipt.collectCache = ipt.collectCache[:0]
			}

			if ipt.lastErr != nil {
				ipt.feeder.FeedLastError(ipt.lastErr.Error(),
					metrics.WithLastErrorInput(inputName),
				)
				ipt.lastErr = nil
			}
		}

		select {
		case tt := <-tick.C:
			ipt.ptsTime = inputs.AlignTime(tt, ipt.ptsTime, ipt.Interval.Duration)

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

	opts := []tailer.Option{
		tailer.WithSource(inputName),
		tailer.WithService(inputName),
		tailer.WithPipeline(ipt.Log.Pipeline),
		tailer.WithIgnoreStatus(ipt.Log.IgnoreStatus),
		tailer.WithCharacterEncoding(ipt.Log.CharacterEncoding),
		tailer.EnableMultiline(true),
		tailer.WithMaxMultilineLength(int64(float64(config.Cfg.Dataway.MaxRawBodySize) * 0.8)),
		tailer.WithMultilinePatterns([]string{`^\d{4}-\d{2}-\d{2}`}),
		tailer.WithGlobalTags(inputs.MergeTags(ipt.Tagger.HostTags(), ipt.Tags, "")),
		tailer.EnableDebugFields(config.Cfg.EnableDebugFields),
	}

	var err error
	ipt.tail, err = tailer.NewTailer(ipt.Log.Files, opts...)
	if err != nil {
		l.Error(err)
		ipt.feeder.FeedLastError(err.Error(),
			metrics.WithLastErrorInput(inputName),
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
		&metricMeasurement{},
		&jenkinsPipelineMeasurement{},
		&jenkinsJobMeasurement{},
	}
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
		Interval:   datakit.Duration{Duration: time.Second * 30},
		semStop:    cliutils.NewSem(),
		feeder:     dkio.DefaultFeeder(),
		Tagger:     datakit.DefaultGlobalTagger(),
		DDInfoResp: `{"endpoints": ["/v0.3/traces"]}`,

		pauseCh:  make(chan bool, inputs.ElectionPauseChannelLength),
		Election: true, // default enable election
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
