// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package profile collector
package profile

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/diskcache"
	filter2 "github.com/GuanceCloud/cliutils/filter"
	"github.com/GuanceCloud/cliutils/logger"
	dkhttp "github.com/GuanceCloud/cliutils/network/http"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/golang/protobuf/proto"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/filter"
	dkMetrics "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/profile/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/rum"
)

const (
	MiB                        = 1 << 20 // 1MB
	inputName                  = "profile"
	defaultProfileMaxSize      = 32    // 32MB
	defaultDiskCacheSize       = 10240 // 10240MB, 10GB
	defaultDiskCacheFileName   = "profile_inputs"
	defaultConsumeWorkersCount = 8
	defaultHTTPClientTimeout   = time.Second * 75
	defaultHTTPRetryCount      = 4
	XDataKitVersionHeader      = "X-Datakit-Version"
	timestampHeaderKey         = "X-Datakit-UnixNano"
	sampleConfig               = `
[[inputs.profile]]
  ## profile Agent endpoints register by version respectively.
  ## Endpoints can be skipped listen by remove them from the list.
  ## Default value set as below. DO NOT MODIFY THESE ENDPOINTS if not necessary.
  endpoints = ["/profiling/v1/input"]

  ## set true to enable election, pull mode only
  election = true

  ## the max allowed size of http request body (of MB), 32MB by default.
  body_size_limit_mb = 32 # MB

  ## set false to stop generating apm metrics from ddtrace output.
  generate_metrics = true

  ## io_config is used to control profiling uploading behavior.
  ## cache_path set the disk directory where temporarily cache profiling data.
  ## cache_capacity_mb specify the max storage space (in MiB) that profiling cache can use.
  ## clear_cache_on_start set whether we should clear all previous profiling cache on restarting Datakit.
  ## upload_workers set the count of profiling uploading workers.
  ## send_timeout specify the http timeout when uploading profiling data to dataway.
  ## send_retry_count set the max retry count when sending every profiling request.
  # [inputs.profile.io_config]
  #   cache_path = "/usr/local/datakit/cache/profile_inputs"  # C:\Program Files\datakit\cache\profile_inputs by default on Windows
  #   cache_capacity_mb = 10240  # 10240MB
  #   clear_cache_on_start = false
  #   upload_workers = 8
  #   send_timeout = "75s"
  #   send_retry_count = 4

  ## set custom tags for profiling data
  # [inputs.profile.tags]
  #   some_tag = "some_value"
  #   more_tag = "some_other_value"

## go pprof config
## collect profiling data in pull mode
#[[inputs.profile.go]]
  ## pprof url
  #url = "http://localhost:6060"

  ## pull interval, should be greater or equal than 10s
  #interval = "10s"

  ## service name
  #service = "go-demo"

  ## app env
  #env = "dev"

  ## app version
  #version = "0.0.0"

  ## types to pull
  ## values: cpu, goroutine, heap, mutex, block
  #enabled_types = ["cpu","goroutine","heap","mutex","block"]

#[inputs.profile.go.tags]
  # tag1 = "val1"

## pyroscope config
#[[inputs.profile.pyroscope]]
  ## listen url
  #url = "0.0.0.0:4040"

  ## service name
  #service = "pyroscope-demo"

  ## app env
  #env = "dev"

  ## app version
  #version = "0.0.0"

#[inputs.profile.pyroscope.tags]
  #tag1 = "val1"
`
)

var (
	log = logger.DefaultSLogger(inputName)

	_ inputs.HTTPInput     = &Input{}
	_ inputs.InputV2       = &Input{}
	_ inputs.ElectionInput = (*Input)(nil)

	diskQueue          *diskcache.DiskCache
	queueConsumerGroup *goroutine.Group
)

type ioConfig struct {
	CachePath         string                    `toml:"cache_path"`
	CacheCapacityMB   int                       `toml:"cache_capacity_mb"`
	ClearCacheOnStart bool                      `toml:"clear_cache_on_start"`
	UploadWorkers     int                       `toml:"upload_workers"`
	SendTimeout       time.Duration             `toml:"send_timeout"`
	SendRetryCount    int                       `toml:"send_retry_count"`
	FilterRules       []string                  `toml:"filter_rules"`
	whereConditions   []filter2.WhereConditions `toml:"-"` // nolint:unused
}

var (
	ErrDatakitExiting     = errors.New("datakit is exiting, request canceled")
	ErrRequestCtxCanceled = errors.New("request context timeout or canceled")
)

type retryError struct {
	error
}

func newRetryError(err error) *retryError {
	return &retryError{
		err,
	}
}

//nolint:unused
type pyroscopeOpts struct {
	URL     string            `toml:"url"`
	Service string            `toml:"service"`
	Env     string            `toml:"env"`
	Version string            `toml:"version"`
	Tags    map[string]string `toml:"tags"`

	tags      map[string]string
	input     *Input
	cacheData sync.Map // key: name, value: *cacheDetail
}

func defaultDiskCachePath() string {
	return filepath.Join(datakit.CacheDir, defaultDiskCacheFileName)
}

func DefaultInput() *Input {
	return &Input{
		BodySizeLimitMB: defaultProfileMaxSize,
		IOConfig: ioConfig{
			CachePath:         defaultDiskCachePath(),
			CacheCapacityMB:   defaultDiskCacheSize,
			ClearCacheOnStart: false,
			UploadWorkers:     defaultConsumeWorkersCount,
			SendTimeout:       defaultHTTPClientTimeout,
			SendRetryCount:    defaultHTTPRetryCount,
		},
		GenerateMetrics: true,
		pauseCh:         make(chan bool, inputs.ElectionPauseChannelLength),
		Election:        true,
		semStop:         cliutils.NewSem(),
		feeder:          dkio.DefaultFeeder(),
		Tagger:          datakit.DefaultGlobalTagger(),
		httpClient:      http.DefaultClient,
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return DefaultInput()
	})
}

type Input struct {
	Endpoints       []string          `toml:"endpoints"`
	BodySizeLimitMB int               `toml:"body_size_limit_mb"`
	IOConfig        ioConfig          `toml:"io_config"`
	Tags            map[string]string `toml:"tags"`
	Go              []*GoProfiler     `toml:"go"`
	PyroscopeLists  []*pyroscopeOpts  `toml:"pyroscope"`
	Election        bool              `toml:"election"`
	GenerateMetrics bool              `toml:"generate_metrics"`

	pause   bool
	pauseCh chan bool

	profileSendingAPI *url.URL
	httpClient        *http.Client

	semStop *cliutils.Sem // start stop signal
	feeder  dkio.Feeder
	Tagger  datakit.GlobalTagger
}

func (ipt *Input) GetBodySizeLimit() int64 {
	return int64(ipt.BodySizeLimitMB) * MiB
}

func (ipt *Input) getDiskCacheCapacity() int64 {
	return int64(ipt.IOConfig.CacheCapacityMB) * MiB
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

func (ipt *Input) ElectionEnabled() bool {
	return ipt.Election
}

func profilingProxyURL() (*url.URL, *http.Transport, error) {
	lastErr := fmt.Errorf("no dataway endpoint available now")

	endpoints := config.Cfg.Dataway.GetEndpoints()

	if len(endpoints) == 0 {
		return nil, nil, lastErr
	}

	for _, ep := range endpoints {
		rawURL, ok := ep.GetCategoryURL()[datakit.ProfilingUpload]
		if !ok || rawURL == "" {
			lastErr = fmt.Errorf("profiling upload url empty")
			continue
		}

		URL, err := url.Parse(rawURL)
		if err != nil {
			lastErr = fmt.Errorf("profiling upload url [%s] parse err:%w", rawURL, err)
			continue
		}
		return URL, ep.Transport(), nil
	}
	return nil, nil, lastErr
}

func cacheRequest(w http.ResponseWriter, r *http.Request, bodySizeLimit int64) *dkhttp.HttpError {
	if r.Body == nil {
		return dkhttp.NewErr(fmt.Errorf("incoming profiling request body is nil"), http.StatusBadRequest)
	}

	bodyBytes, err := io.ReadAll(http.MaxBytesReader(w, r.Body, bodySizeLimit))
	if err != nil {
		return dkhttp.NewErr(fmt.Errorf("unable to read profile body: %w", err), http.StatusBadRequest)
	}

	headers := make(map[string]string, len(r.Header))

	for k, v := range r.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		} else {
			headers[k] = ""
		}
	}

	reqPB := &rum.RequestPB{
		Header: headers,
		Body:   bodyBytes,
	}

	pbBytes, err := proto.Marshal(reqPB)
	if err != nil {
		return dkhttp.NewErr(fmt.Errorf("unable to marshal request using protobuf: %w", err), http.StatusBadRequest)
	}

	if err = diskQueue.Put(pbBytes); err != nil {
		return dkhttp.NewErr(fmt.Errorf("unable to push request to disk queue: %w", err), http.StatusInternalServerError)
	}
	return nil
}

func (ipt *Input) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if err := cacheRequest(w, req, ipt.GetBodySizeLimit()); err != nil {
		log.Errorf("unable to cache profiling request: %v", err)
		w.WriteHeader(err.HttpCode)
		_, _ = io.WriteString(w, err.Error())
		return
	}
}

func insertEventFormFile(form *multipart.Form, mw *multipart.Writer, metadata map[string]string) error {
	f, err := mw.CreateFormFile(metrics.EventFile, metrics.EventJSONFile)
	if err != nil {
		return fmt.Errorf("unable to create form file: %w", err)
	}

	md := metrics.Metadata{}

	for name, fileHeaders := range form.File {
		extName := filepath.Ext(name)
		if extName == "" {
			// try to fetch binary extname from multipart.FileHeader.Filename
			for _, fh := range fileHeaders {
				extName = filepath.Ext(fh.Filename)
				if extName != "" {
					name = fh.Filename
					break
				}
			}
		}

		md.Attachments = append(md.Attachments, name)
		switch strings.ToLower(extName) {
		case ".pprof":
			md.Format = metrics.PPROF
		case ".jfr":
			md.Format = metrics.JFR
		}
	}
	if md.Format == "" {
		md.Format = "unknown"
	}
	startTime, err := metrics.ResolveStartTime(metadata)
	if err != nil {
		log.Warnf("unable to resolve profile start time: %w", err)
	} else {
		md.Start = metrics.NewRFC3339Time(startTime)
	}

	endTime, err := metrics.ResolveEndTime(metadata)
	if err != nil {
		log.Warnf("unable to resolve profile end time: %w", err)
	} else {
		md.End = metrics.NewRFC3339Time(endTime)
	}

	lang := metrics.ResolveLang(metadata)
	md.Language = lang

	md.TagsProfiler = metrics.JoinTags(metadata)

	mdBytes, err := json.Marshal(md)
	if err != nil {
		return fmt.Errorf("unable to marshal data for profiling event file: %w", err)
	}

	if _, err = f.Write(mdBytes); err != nil {
		return fmt.Errorf("unable to write data to multipart file: %w", err)
	}

	return nil
}

func (ipt *Input) sendRequestToDW(ctx context.Context, pbBytes []byte) error {
	var reqPB rum.RequestPB

	if err := proto.Unmarshal(pbBytes, &reqPB); err != nil {
		return fmt.Errorf("unable to unmarshal profiling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ipt.profileSendingAPI.String(), bytes.NewReader(reqPB.Body))
	if err != nil {
		return fmt.Errorf("unable to create http request: %w", err)
	}

	for k, v := range reqPB.Header {
		// ignore header "Host" and "Content-Length"
		if !strings.EqualFold("Host", k) && !strings.EqualFold("Content-Length", k) {
			req.Header.Set(k, v)
		}
	}

	if err := req.ParseMultipartForm(ipt.GetBodySizeLimit()); err != nil {
		return fmt.Errorf("unable to parse multipart/formdata: %w", err)
	}

	metadata, _, err := metrics.ParseMetadata(req)
	if err != nil {
		return fmt.Errorf("unable to resolve profiling tags: %w", err)
	}

	var subCustomTags map[string]string
	if metadata[metrics.SubCustomTagsKey] != "" {
		subCustomTags = metrics.NewTags(strings.Split(metadata[metrics.SubCustomTagsKey], ","))
	}

	globalHostTags := datakit.GlobalHostTags()
	language := metrics.ResolveLang(metadata)
	if ipt.GenerateMetrics {
		allCustomTags := make(map[string]string, len(ipt.Tags)+len(subCustomTags)+len(globalHostTags))
		// global host tags have the lowest priority.
		for k, v := range globalHostTags {
			allCustomTags[k] = v
		}
		for k, v := range ipt.Tags {
			allCustomTags[k] = v
		}
		for k, v := range subCustomTags {
			allCustomTags[k] = v
		}
		switch language { // nolint:exhaustive
		case metrics.Java:
			if err = metrics.ExportJVMMetrics(req.MultipartForm.File, metadata, allCustomTags); err != nil {
				log.Errorf("unable to export java ddtrace profiling metrics: %v", err)
			}
		case metrics.Golang:
			if err = metrics.ExportGoMetrics(req.MultipartForm.File, metadata, allCustomTags); err != nil {
				log.Errorf("unable to export golang ddtrace profiling metrics: %v", err)
			}
		case metrics.Python:
			if err = metrics.ExportPythonMetrics(req.MultipartForm.File, metadata, allCustomTags); err != nil {
				log.Errorf("unable to export python ddtrace profiling metrics: %v", err)
			}
		}
	}

	customTagsDefined := false
	for tk, tv := range ipt.Tags {
		if _, ok := subCustomTags[tk]; ok {
			// has set tags in sub settings, ignore
			continue
		}
		if old, ok := metadata[tk]; !ok || old != tv {
			customTagsDefined = true
			metadata[tk] = tv
		}
	}
	for k, v := range globalHostTags {
		if _, ok := metadata[k]; !ok {
			customTagsDefined = true
			metadata[k] = v
		}
	}

	// apply remote or local filter
	pt := point.NewPoint(inputName, point.NewTags(metadata), point.WithTime(time.Now()))
	if len(filter.FilterPts(point.Profiling, []*point.Point{pt})) == 0 {
		log.Infof("the profiling data matched the remote or local blacklist and was dropped")
		return nil
	}

	// Add event form file to multipartForm if it doesn't exist
	_, ok1 := req.MultipartForm.File[metrics.EventFile]
	_, ok2 := req.MultipartForm.File[metrics.EventJSONFile]

	if (!ok1 && !ok2) || customTagsDefined {
		if newBody, err := modifyMultipartForm(req, req.MultipartForm, metadata); err != nil {
			log.Warnf("unable to insert event form file to profiling request: %s", err)
		} else {
			reqPB.Body = newBody
		}
	}

	req.Header.Set(XDataKitVersionHeader, datakit.Version)
	if config.Cfg.Dataway.EnableSinker {
		xGlobalTag := dataway.SinkHeaderValueFromTags(metadata,
			config.Cfg.Dataway.GlobalTags(),
			config.Cfg.Dataway.CustomTagKeys())
		if xGlobalTag == "" {
			xGlobalTag = config.Cfg.Dataway.GlobalTagsHTTPHeaderValue()
		}

		req.Header.Set(dataway.HeaderXGlobalTags, xGlobalTag)
	}

	var (
		sendErr    error
		resp       *http.Response
		reqCost    time.Duration
		statusCode = "unknown"
	)

	defer func() {
		dataway.APISumVec().WithLabelValues(req.URL.Path, statusCode).Observe(reqCost.Seconds())
	}()

	reqStart := time.Now()
	for i := 0; i < ipt.IOConfig.SendRetryCount; i++ {
		select {
		case <-datakit.Exit.Wait():
			return ErrDatakitExiting
		case <-req.Context().Done():
			return ErrRequestCtxCanceled
		default:
		}

		sendErr = func(idx int) error {
			if req.Body != nil {
				_ = req.Body.Close()
			}
			req.Body = io.NopCloser(bytes.NewReader(reqPB.Body))
			req.GetBody = func() (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader(reqPB.Body)), nil
			}
			req.ContentLength = int64(len(reqPB.Body))

			resp, err = ipt.httpClient.Do(req)
			if err != nil {
				statusCode = "unknown"
				time.Sleep(time.Second)
				e := fmt.Errorf("at #%d try: unable to send profiling request to dataway: %w", idx+1, err)
				return newRetryError(e)
			}
			defer resp.Body.Close() // nolint:errcheck
			statusCode = http.StatusText(resp.StatusCode)

			errMsg := []byte(nil)
			if resp.StatusCode/100 != 2 {
				errMsg, _ = io.ReadAll(resp.Body)
			}

			switch resp.StatusCode / 100 {
			case 5:
				// Retry only when the status code is "5XX".
				e := fmt.Errorf("at #%d try: unable to send profiling data to dataway, http Status: %s, response: %q",
					idx+1, resp.Status, string(errMsg))
				return newRetryError(e)
			case 2:
				// ignore
				return nil
			default:
				return fmt.Errorf("at #%d try: unable to send profiling data to dataway, http status: %s, response: %q",
					idx+1, resp.Status, string(errMsg))
			}
		}(i)

		// Log IO retry metrics
		if i > 0 {
			dataway.HTTPRetry().WithLabelValues(req.URL.Path, statusCode).Inc()
		}

		if sendErr != nil {
			log.Warnf("fail to send http request: %s", sendErr)
		}
		var re *retryError
		if !errors.As(sendErr, &re) {
			break
		}
	}
	reqCost = time.Since(reqStart)

	metricName := inputName + "/" + language.String()
	if sendErr == nil && resp.StatusCode/100 == 2 {
		dkio.InputsFeedVec().WithLabelValues(metricName, point.Profiling.String()).Inc()
		dkio.InputsFeedPtsVec().WithLabelValues(metricName, point.Profiling.String()).Observe(float64(1))
		dkio.InputsLastFeedVec().WithLabelValues(metricName, point.Profiling.String()).Set(float64(time.Now().Unix()))
		dkio.InputsCollectLatencyVec().WithLabelValues(metricName, point.Profiling.String()).Observe(reqCost.Seconds())
	} else {
		feedErr := sendErr
		if feedErr == nil {
			feedErr = fmt.Errorf("error status code %d", resp.StatusCode)
		}
		ipt.feeder.FeedLastError(feedErr.Error(),
			dkMetrics.WithLastErrorInput(metricName),
			dkMetrics.WithLastErrorCategory(point.Profiling),
		)
	}

	return sendErr
}

// RegHTTPHandler simply proxy profiling request to dataway.
func (ipt *Input) RegHTTPHandler() {
	for _, endpoint := range ipt.Endpoints {
		httpapi.RegHTTPHandler(http.MethodPost, endpoint, ipt.ServeHTTP)
		log.Infof("pattern: %s registered", endpoint)
	}
}

func (ipt *Input) Catalog() string {
	return inputName
}

func (ipt *Input) InitDiskQueueIO() error {
	if ipt.IOConfig.ClearCacheOnStart {
		if err := os.RemoveAll(ipt.IOConfig.CachePath); err != nil {
			return fmt.Errorf("unable to clear previous profiling cache: %w", err)
		}
	}

	dc, err := diskcache.Open(
		diskcache.WithPath(ipt.IOConfig.CachePath),
		diskcache.WithCapacity(ipt.getDiskCacheCapacity()),
		diskcache.WithNoFallbackOnError(true),
	)
	if err != nil {
		return fmt.Errorf("unable to start disk cache of profiling: %w", err)
	}

	diskQueue = dc

	profileURL, transport, err := profilingProxyURL()
	if err != nil {
		return fmt.Errorf("no dataway endpoint available: %w", err)
	}

	ipt.profileSendingAPI = profileURL

	ipt.httpClient = &http.Client{
		Transport: transport,
		Timeout:   ipt.IOConfig.SendTimeout,
	}

	queueConsumerGroup = goroutine.NewGroup(goroutine.Option{
		Name: "profiling_disk_cache_consumer",
	})

	for i := 0; i < ipt.IOConfig.UploadWorkers; i++ {
		queueConsumerGroup.Go(func(ctx context.Context) error {
			for {
				select {
				case <-datakit.Exit.Wait():
					log.Infof("profiling uploading worker exit now...")
					return nil
				case <-ctx.Done():
					log.Infof("context canceld...")
					return nil
				case <-ipt.semStop.Wait():
					log.Infof("profiling uploading worker exit now...")
					return nil
				default:
					func() {
						var reqData []byte
						if err := diskQueue.Get(func(msg []byte) error {
							reqData = msg
							return nil
						}); err != nil {
							if errors.Is(err, diskcache.ErrNoData) {
								time.Sleep(time.Second * 3)
							} else {
								log.Errorf("unable to get msg from disk cache: %s", err)
								time.Sleep(time.Millisecond * 100)
							}
							return
						}
						if err := ipt.sendRequestToDW(ctx, reqData); err != nil {
							log.Errorf("fail to send profiling data: %s", err)
						}
					}()
				}
			}
		})
	}
	return nil
}

func (ipt *Input) Run() {
	log = logger.SLogger(inputName)
	log.Infof("the input %s is running...", inputName)

	metrics.InitLog()

	if err := ipt.InitDiskQueueIO(); err != nil {
		log.Errorf("unable to start IO process for profiling: %s", err)
	}

	groupPull := goroutine.NewGroup(goroutine.Option{
		Name: PullInputMode,
		PanicCb: func(b []byte) bool {
			log.Errorf("goroutine profile-pull-mode panic: %s", string(b))
			return false
		},
	})

	groupPyroscope := goroutine.NewGroup(goroutine.Option{
		Name: "profile-pyroscope",
		PanicCb: func(b []byte) bool {
			log.Errorf("goroutine profile-pyroscope panic: %s", b)
			return false
		},
	})

	for _, g := range ipt.Go {
		func(g *GoProfiler) {
			groupPull.Go(func(ctx context.Context) error {
				if err := g.run(ipt); err != nil {
					log.Errorf("profile-pull-mode collect error: %s", err.Error())
				}
				return nil
			})
		}(g)
	}

	for _, g := range ipt.PyroscopeLists {
		func(g *pyroscopeOpts) {
			groupPyroscope.Go(func(ctx context.Context) error {
				if err := g.run(ipt); err != nil {
					log.Errorf("pyroscope profile collect error: %s", err.Error())
				}
				return nil
			})
		}(g)
	}

	if err := groupPull.Wait(); err != nil {
		log.Errorf("profile-pull-mode collect err: %s", err.Error())
	}

	if err := groupPyroscope.Wait(); err != nil {
		log.Errorf("profile-pyroscope collect err: %s", err.Error())
	}
}

func (ipt *Input) SampleConfig() string {
	return sampleConfig
}

func (ipt *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{inputs.DefaultEmptyMeasurement}
}

func (ipt *Input) AvailableArchs() []string {
	return datakit.AllOS
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}

	if queueConsumerGroup != nil {
		if err := queueConsumerGroup.Wait(); err != nil {
			log.Errorf("goroutine group [%s] abnormally exit: %s", queueConsumerGroup.Name(), err)
		}
	}

	if diskQueue != nil {
		if err := diskQueue.Close(); err != nil {
			log.Errorf("unable to close disk queue: %s", err)
		}
	}

	for _, endpoint := range ipt.Endpoints {
		httpapi.RemoveHTTPRoute(http.MethodPost, endpoint)
	}
}

type pushProfileDataOpt struct {
	startTime       time.Time
	endTime         time.Time
	profiledatas    []*profileData
	endPoint        string
	inputTags       map[string]string
	inputNameSuffix string
	Input           *Input
}

func pushProfileData(opt *pushProfileDataOpt, event *metrics.Metadata, bodySizeLimit int64) error {
	b := new(bytes.Buffer)
	mw := multipart.NewWriter(b)

	for _, pd := range opt.profiledatas {
		ff, err := mw.CreateFormFile(pd.fileName, pd.fileName)
		if err != nil {
			return fmt.Errorf("unable to create profiling form file: %w", err)
		}

		if _, err = io.Copy(ff, pd.buf); err != nil {
			return fmt.Errorf("unable to copy porfiling data: %w", err)
		}
	}

	f, err := mw.CreateFormFile(metrics.EventFile, metrics.EventJSONFile)
	if err != nil {
		return err
	}

	eventJSONString, err := json.Marshal(event)
	if err != nil {
		return err
	}

	if _, err = io.Copy(f, bytes.NewReader(eventJSONString)); err != nil {
		return err
	}
	if err = mw.Close(); err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, "/", b)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set(XDataKitVersionHeader, datakit.Version)
	req.Header.Set(timestampHeaderKey, strconv.FormatInt(opt.startTime.UnixNano(), 10))

	if err := cacheRequest(&httpapi.NopResponseWriter{}, req, bodySizeLimit); err != nil {
		return fmt.Errorf("unbale to cache profiling request: %w", err)
	}
	return nil
}

func originAddTagsSafe(originTags map[string]string, newKey, newVal string) {
	if len(newKey) > 0 && len(newVal) > 0 {
		if _, ok := originTags[newKey]; !ok {
			originTags[newKey] = newVal
		}
	}
}

const (
	pyroscopeReservedPrefix = "__"
)

func getPyroscopeTagFromLabels(labels map[string]string) map[string]string {
	out := make(map[string]string, len(labels)-1) // exclude '__name__'.
	for k, v := range labels {
		if strings.HasPrefix(k, pyroscopeReservedPrefix) {
			continue
		}
		out[k] = v
	}
	return out
}
