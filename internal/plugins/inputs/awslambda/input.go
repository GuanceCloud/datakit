// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2024-present Guance, Inc.

// Package awslambda receive and process aws lambda api output data.
package awslambda

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"syscall"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/awslambda/lambdaapi/extension"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/awslambda/lambdaapi/model"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/awslambda/lambdaapi/telemetry"
)

const (
	inputName = "awslambda"
)

var l = logger.DefaultSLogger(inputName)

var _ inputs.InputV2 = &Input{}

type Input struct {
	UseNowTimeInstead      bool `toml:"use_local_time_instead"`
	EnableLogCollection    bool `toml:"enable_log_collection"`
	EnableMetricCollection bool `toml:"enable_metric_collection"`

	tags           map[string]string
	lambdaCtxCache *lambdaCtxCache
	feedControl    *FeedControl

	feeder dkio.Feeder

	ctx               context.Context
	cancel            context.CancelFunc
	telemetryListener *telemetry.Listener
	nextEventChan     <-chan *extension.NextEventResponse
	eventDoneChan     chan struct{}

	g *goroutine.Group
}

func (ipt *Input) Catalog() string {
	return "function"
}

func (ipt *Input) Run() {
	if os.Getenv(EnvLambdaFunctionName) == "" {
		l.Warn("the current environment is not aws lambda, awslambda input exit.")
		return
	}
	l.Info("awslambda input started")

	if err := ipt.setup(); err != nil {
		l.Errorf("setup failed: %s", err)
		return
	}

	ipt.g.Go(func(ctx context.Context) error {
		ipt.collect()
		return nil
	})
	for {
		select {
		case eventResponse, ok := <-ipt.nextEventChan:
			if !ok {
				ipt.exit()
				return
			}

			l.Infof("got event: %+#v", eventResponse)
			if eventResponse.EventType == model.Shutdown {
				ipt.exit()
				return
			}
		case <-ipt.ctx.Done():
			l.Infof("input context done")
			return
		}
	}
}

func (ipt *Input) collect() {
	var (
		metricEvents []*telemetry.Event
		logEvents    []*telemetry.LogEvent
	)

	for arr := range ipt.telemetryListener.GetPullChan() {
		arr, delData := telemetry.SeparateEvents(arr)
		logEvents = append(logEvents, delData...)
		metricEvents = append(metricEvents, arr...)

		sizeL, sizeM := len(logEvents), len(metricEvents)
		l.Debugf("size of log events: %d, size of metric events: %d", sizeL, sizeM)

		if slicesContainsWithType(arr, telemetry.TypePlatformRuntimeDone) {
			l.Debugf("size of log events: %d, size of metric events: %d", sizeL, sizeM)

			syncFeed := ipt.feedControl.ShouldFeed()

			ipt.feedMetric(metricEvents, syncFeed)
			ipt.feedLog(logEvents, syncFeed)

			logEvents = make([]*telemetry.LogEvent, 0, sizeL)
			metricEvents = make([]*telemetry.Event, 0, sizeM)

			ipt.eventDoneChan <- struct{}{}
		}
	}

	l.Debugf("tail size of log events: %d, size of metric events: %d", len(logEvents), len(metricEvents))
	ipt.feedLog(logEvents, true)
	ipt.feedMetric(metricEvents, true)
}

func slicesContainsWithType(events []*telemetry.Event, typ string) bool {
	for i := len(events) - 1; i >= 0; i-- {
		if events[i].Record.GetType() == typ {
			return true
		}
	}
	return false
}

func (ipt *Input) feedLog(logEvent []*telemetry.LogEvent, syncSend bool) {
	if !ipt.EnableLogCollection {
		return
	}

	if l.Level() <= -1 {
		for _, event := range logEvent {
			l.Debugf("log event fields: %v", event.Record.GetFields())
		}
	}

	pts := ipt.toLogPointArr(logEvent)

	if err := ipt.feeder.FeedV2(point.Logging, pts,
		dkio.WithSyncSend(syncSend),
		dkio.WithElection(false),
		dkio.WithInputName(inputName)); err != nil {
		ipt.feeder.FeedLastError(err.Error(),
			metrics.WithLastErrorInput(inputName),
			metrics.WithLastErrorCategory(point.Logging),
		)
		l.Errorf("feed measurement: %s", err)
	}
}

func (ipt *Input) feedMetric(metricEvent []*telemetry.Event, syncSend bool) {
	if !ipt.EnableMetricCollection {
		return
	}
	pts := ipt.toMetricPointArr(metricEvent)

	if err := ipt.feeder.FeedV2(point.Metric, pts,
		dkio.WithSyncSend(syncSend),
		dkio.WithElection(false),
		dkio.WithInputName(inputName)); err != nil {
		ipt.feeder.FeedLastError(err.Error(),
			metrics.WithLastErrorInput(inputName),
			metrics.WithLastErrorCategory(point.Metric),
		)
		l.Errorf("feed measurement: %s", err)
	}
}

func (ipt *Input) SampleConfig() string {
	return `
[[inputs.awslambda]]
  ## Enable log collection
  # enable_log_collection = true
  
  ## Enable metric collection
  # enable_metric_collection = true
`
}

func (ipt *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&metricMeasurement{},
		&logMeasurement{},
	}
}

func (ipt *Input) AvailableArchs() []string {
	return []string{
		datakit.OSLabelLinux, datakit.LabelK8s, datakit.LabelDocker,
	}
}

func (ipt *Input) Terminate() {
	ipt.telemetryListener.Shutdown()
	ipt.cancel()
	httpapi.RemoveHTTPRoute(http.MethodPost, "/awslambda")
}

func (ipt *Input) setup() error {
	resetLog()
	ipt.ctx, ipt.cancel = context.WithCancel(context.Background())
	ipt.g.Go(func(ctx context.Context) error {
		select {
		case <-datakit.Exit.Wait():
			ipt.Terminate()
		case <-ipt.ctx.Done():
		}
		return nil
	})

	ipt.feedControl = NewFeedControl(20)
	ipt.telemetryListener = telemetry.NewTelemetryListener()
	ipt.lambdaCtxCache = newLambdaCtxCache()
	ipt.eventDoneChan = make(chan struct{})

	extensionClient := extension.NewClient(extension.GetAwsLambdaRuntimeAPI())
	r, err := extensionClient.Register(ipt.ctx, path.Base(os.Args[0]))
	if err != nil {
		l.Errorf("register extension client failed: %s", err)
		ipt.exit()
		return fmt.Errorf("extensionClient.Register: %w", err)
	}
	if r.AccountID != "" {
		ipt.tags[AccountID] = r.AccountID
	}

	telemetryClient := telemetry.NewTelemetryClient(extension.GetAwsLambdaRuntimeAPI(),
		extensionClient.ExtensionID,
		strings.Split(config.Cfg.HTTPAPI.Listen, ":")[1],
		"awslambda")
	err = telemetryClient.Subscribe(ipt.ctx)
	if err != nil {
		l.Errorf("subscribe telemetry client failed: %s", err)
		ipt.exit()
		return err
	}

	ipt.nextEventChan, err = extensionClient.AsyncNextEventLoop(ipt.ctx, ipt.eventDoneChan)
	if err != nil {
		l.Errorf("create async next event loop failed: %s", err)
		ipt.exit()
		return err
	}
	ipt.eventDoneChan <- struct{}{}

	l.Infof("setup ok")
	return nil
}

func resetLog() {
	l = logger.SLogger(inputName)
	telemetry.SetLogger(l)
	extension.SetLogger(l)
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		ipt := &Input{
			EnableMetricCollection: true,
			EnableLogCollection:    true,
			feeder:                 dkio.DefaultFeeder(),
			tags:                   make(map[string]string),
		}
		ipt.initTags()
		ipt.g = datakit.G(inputName)
		return ipt
	})
}

func (ipt *Input) exit() {
	ipt.Terminate()
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		l.Error("Error finding process:", err)
		return
	}

	l.Info("Sending SIGTERM to self")
	if err := p.Signal(syscall.SIGTERM); err != nil {
		l.Error("Error sending signal:", err)
	}
}

func (ipt *Input) RegHTTPHandler() {
	h := func(w http.ResponseWriter, r *http.Request, _ ...interface{}) (interface{}, error) {
		err := ipt.telemetryListener.HandlerTelemetry(w, r)
		return nil, err
	}
	httpapi.RegHTTPRoute(http.MethodPost, "/awslambda", h)
}
