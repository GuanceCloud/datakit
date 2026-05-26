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
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	lambdaextsrv "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/awslambda/extension"
	lambdaextapi "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/awslambda/lambdaapi/extension"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/awslambda/lambdaapi/model"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/awslambda/lambdaapi/telemetry"
	lambdatrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/awslambda/trace"
)

const (
	inputName = "awslambda"

	invocationDrainSafetyMargin = 100 * time.Millisecond
	shutdownDrainTimeout        = 300 * time.Millisecond
	shutdownDrainSafetyMargin   = 100 * time.Millisecond
	inputAPI                    = "/awslambda"
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
	nextEventChan     <-chan *lambdaextapi.NextEventResponse
	eventDoneChan     chan struct{}
	runtimeDoneChan   chan string
	traceProcessor    *lambdatrace.Processor

	g            *goroutine.Group
	lambdaServer *http.Server
}

func (ipt *Input) Catalog() string {
	return "function"
}

func (ipt *Input) Run() {
	if !IsLambdaEnvironment() {
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

			l.Infof("got event: type=%s request_id=%s deadline_ms=%d shutdown_reason=%s",
				eventResponse.EventType,
				eventResponse.RequestID,
				eventResponse.DeadlineMs,
				eventResponse.ShutdownReason,
			)
			if ipt.traceProcessor != nil && eventResponse.EventType == model.Invoke {
				ipt.traceProcessor.OnInvokeEvent(eventResponse.RequestID)
				ipt.waitRuntimeDone(eventResponse.RequestID, eventResponse.DeadlineMs)
				ipt.eventDoneChan <- struct{}{}
				continue
			}
			if eventResponse.EventType == model.Shutdown {
				if ipt.traceProcessor != nil {
					ipt.drainTelemetryBeforeShutdown(eventResponse.DeadlineMs)
					_ = ipt.traceProcessor.OnShutdown()
				}
				ipt.exit()
				return
			}
			ipt.eventDoneChan <- struct{}{}
		case <-ipt.ctx.Done():
			l.Infof("input context done")
			return
		}
	}
}

func (ipt *Input) waitRuntimeDone(requestID string, deadlineMs int64) {
	if requestID == "" {
		return
	}
	if deadlineMs <= 0 {
		return
	}

	wait := time.Until(time.UnixMilli(deadlineMs)) - invocationDrainSafetyMargin
	if wait <= 0 {
		l.Debugf("skip invocation telemetry drain request_id=%s: deadline too close", requestID)
		return
	}

	timer := time.NewTimer(wait)
	defer timer.Stop()

	for {
		select {
		case doneRequestID := <-ipt.runtimeDoneChan:
			if doneRequestID == requestID {
				l.Debugf("runtime done received before next event request_id=%s", requestID)
				return
			}
			l.Debugf("skip runtime done for another request while waiting request_id=%s done_request_id=%s", requestID, doneRequestID)
		case <-timer.C:
			l.Debugf("wait runtime done timeout request_id=%s wait=%s", requestID, wait)
			return
		case <-ipt.ctx.Done():
			return
		}
	}
}

func (ipt *Input) drainTelemetryBeforeShutdown(deadlineMs int64) {
	wait := shutdownDrainTimeout
	if deadlineMs > 0 {
		remaining := time.Until(time.UnixMilli(deadlineMs)) - shutdownDrainSafetyMargin
		if remaining <= 0 {
			return
		}
		if remaining < wait {
			wait = remaining
		}
	}

	l.Debugf("drain telemetry before shutdown for %s", wait)
	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case <-timer.C:
	case <-ipt.ctx.Done():
	}
}

func (ipt *Input) collect() {
	var (
		metricEvents []*telemetry.Event
		logEvents    []*telemetry.LogEvent
	)

	defer func() {
		l.Debugf("tail size of log events: %d, size of metric events: %d", len(logEvents), len(metricEvents))
		ipt.feedLog(logEvents, true)
		ipt.feedMetric(metricEvents, true)
	}()

	for {
		select {
		case arr := <-ipt.telemetryListener.GetPullChan():
			arr, delData := telemetry.SeparateEvents(arr)
			if ipt.traceProcessor != nil {
				for _, event := range arr {
					_ = ipt.traceProcessor.OnTelemetryEvent(event)
				}
			}
			logEvents = append(logEvents, delData...)
			metricEvents = append(metricEvents, arr...)

			runtimeDoneIDs := runtimeDoneRequestIDs(arr)
			if len(runtimeDoneIDs) > 0 {
				syncFeed := ipt.feedControl.ShouldFeed()

				ipt.feedMetric(metricEvents, syncFeed)
				ipt.feedLog(logEvents, syncFeed)

				sizeL, sizeM := len(logEvents), len(metricEvents)
				logEvents = make([]*telemetry.LogEvent, 0, sizeL)
				metricEvents = make([]*telemetry.Event, 0, sizeM)

				ipt.notifyRuntimeDone(runtimeDoneIDs)
			}
		case <-ipt.ctx.Done():
			l.Infof("collect loop context done")
			return
		case <-datakit.Exit.Wait():
			l.Infof("collect loop datakit exit")
			return
		}
	}
}

func runtimeDoneRequestIDs(events []*telemetry.Event) []string {
	var requestIDs []string
	for _, event := range events {
		if record, ok := event.Record.(*telemetry.PlatformRuntimeDone); ok && record.RequestID != "" {
			requestIDs = append(requestIDs, record.RequestID)
		}
	}
	return requestIDs
}

func (ipt *Input) notifyRuntimeDone(requestIDs []string) {
	for _, requestID := range requestIDs {
		select {
		case ipt.runtimeDoneChan <- requestID:
		default:
			l.Debugf("runtime done notify channel full, drop request_id=%s", requestID)
		}
	}
}

func (ipt *Input) feedLog(logEvent []*telemetry.LogEvent, syncSend bool) {
	if !ipt.EnableLogCollection {
		return
	}

	pts := ipt.toLogPointArr(logEvent)
	l.Debugf("feed %d log events as %d logging points", len(logEvent), len(pts))

	if err := ipt.feeder.Feed(point.Logging, pts,
		dkio.WithSyncSend(syncSend),
		dkio.WithElection(false),
		dkio.WithSource(inputName)); err != nil {
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

	if err := ipt.feeder.Feed(point.Metric, pts,
		dkio.WithSyncSend(syncSend),
		dkio.WithElection(false),
		dkio.WithSource(inputName)); err != nil {
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
		&traceMeasurement{},
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
	if ipt.lambdaServer != nil {
		if err := ipt.lambdaServer.Shutdown(ipt.ctx); err != nil {
			l.Errorf("lambda server shutdown failed: %s", err.Error())
		}
	}
	httpapi.RemoveHTTPRoute(http.MethodPost, inputAPI)
}

func (ipt *Input) startDDLambdaExtensionService() {
	if tracingEnabled, err := strconv.ParseBool(os.Getenv("DD_TRACE_ENABLED")); err == nil {
		if !tracingEnabled {
			l.Debugf("DD_TRACE_ENABLED is false, skip lambda extesion service")
			return
		}
	} else {
		l.Warnf("parse DD_TRACE_ENABLED failed: %s", err.Error())
		return
	}

	ipt.g.Go(func(ctx context.Context) error {
		server, err := lambdaextsrv.StartLifecycleServer(":8124", ipt.traceProcessor)
		if err != nil {
			l.Errorf("start lambda extension server failed: %s", err.Error())
			return nil
		}
		ipt.lambdaServer = server
		l.Infof("start lambda extension server at addr: %s", ipt.lambdaServer.Addr)

		return nil
	})
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
	ipt.runtimeDoneChan = make(chan string, 32)
	managed := strings.EqualFold(os.Getenv(EnvLambdaInitializationType), "lambda-managed-instances")
	ipt.traceProcessor = lambdatrace.NewProcessor(lambdatrace.NewPointSink(inputName, ipt.feeder, ipt.tags), managed)
	lambdatrace.SetActiveProcessor(ipt.traceProcessor)

	extensionClient := lambdaextapi.NewClient(lambdaextapi.GetAwsLambdaRuntimeAPI())
	r, err := extensionClient.Register(ipt.ctx, path.Base(os.Args[0]))
	if err != nil {
		l.Errorf("register extension client failed: %s", err)
		ipt.exit()
		return fmt.Errorf("extensionClient.Register: %w", err)
	}
	if r.AccountID != "" {
		ipt.tags[AccountID] = r.AccountID
	}

	telemetryClient := telemetry.NewTelemetryClient(lambdaextapi.GetAwsLambdaRuntimeAPI(),
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

	ipt.startDDLambdaExtensionService()

	l.Infof("setup ok")
	return nil
}

func resetLog() {
	l = logger.SLogger(inputName)
	telemetry.SetLogger(l)
	lambdaextapi.SetLogger(l)
	lambdatrace.SetLogger(l)
}

func init() { //nolint:gochecknoinits
	httpapi.RegInputHTTPRouteMatcher(func(method, path string) (string, bool) {
		if method == http.MethodPost && path == inputAPI {
			return inputName, true
		}
		return "", false
	})

	inputs.Add(inputName, func() inputs.Input {
		ipt := &Input{
			EnableMetricCollection: true,
			EnableLogCollection:    true,
			feeder:                 dkio.DefaultFeeder(),
			tags:                   make(map[string]string),
		}
		ipt.initTags()
		ipt.g = goroutine.G(inputName)
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
	httpapi.RegHTTPRoute(http.MethodPost, inputAPI, h)
}
