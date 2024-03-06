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
	"net/textproto"
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
	"github.com/GuanceCloud/cliutils/point"
	"github.com/golang/protobuf/proto"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/rum"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
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
	profileIDHeaderKey         = "X-Datakit-ProfileID"
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
	log       = logger.DefaultSLogger(inputName)
	iptGlobal *Input

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

type minHeap struct {
	buckets []*profileBase
	indexes map[*profileBase]int
}

func newMinHeap(initCap int) *minHeap {
	return &minHeap{
		buckets: make([]*profileBase, 0, initCap),
		indexes: make(map[*profileBase]int, initCap),
	}
}

func (mh *minHeap) Swap(i, j int) {
	mh.indexes[mh.buckets[i]], mh.indexes[mh.buckets[j]] = j, i
	mh.buckets[i], mh.buckets[j] = mh.buckets[j], mh.buckets[i]
}

func (mh *minHeap) Less(i, j int) bool {
	return mh.buckets[i].birth.Before(mh.buckets[j].birth)
}

func (mh *minHeap) Len() int {
	return len(mh.buckets)
}

func (mh *minHeap) siftUp(idx int) {
	if idx >= len(mh.buckets) {
		errMsg := fmt.Sprintf("siftUp: index[%d] out of bounds[%d]", idx, len(mh.buckets))
		log.Error(errMsg)
		panic(errMsg)
	}

	for idx > 0 {
		parent := (idx - 1) / 2

		if !mh.Less(idx, parent) {
			break
		}

		// Swap
		mh.Swap(idx, parent)
		idx = parent
	}
}

func (mh *minHeap) siftDown(idx int) {
	for {
		left := idx*2 + 1
		if left >= mh.Len() {
			break
		}

		minIdx := idx
		if mh.Less(left, minIdx) {
			minIdx = left
		}

		right := left + 1
		if right < mh.Len() && mh.Less(right, minIdx) {
			minIdx = right
		}

		if minIdx == idx {
			break
		}

		mh.Swap(idx, minIdx)
		idx = minIdx
	}
}

func (mh *minHeap) push(pb *profileBase) {
	mh.buckets = append(mh.buckets, pb)
	mh.indexes[pb] = mh.Len() - 1
	mh.siftUp(mh.Len() - 1)
}

func (mh *minHeap) pop() *profileBase {
	if mh.Len() == 0 {
		return nil
	}

	top := mh.getTop()
	mh.remove(top)
	return top
}

func (mh *minHeap) remove(pb *profileBase) {
	idx, ok := mh.indexes[pb]
	if !ok {
		log.Errorf("pb not found in the indexes, profileID = %s", pb.profileID)
		return
	}
	if idx >= mh.Len() {
		errMsg := fmt.Sprintf("remove: index[%d] out of bounds [%d]", idx, mh.Len())
		log.Error(errMsg)
		panic(errMsg)
	}

	if mh.buckets[idx] != pb {
		errMsg := fmt.Sprintf("remove: idx of the buckets[%p] not equal the removing target[%p]", mh.buckets[idx], pb)
		log.Error(errMsg)
		panic(errMsg)
	}
	// delete the idx
	mh.buckets[idx] = mh.buckets[mh.Len()-1]
	mh.indexes[mh.buckets[idx]] = idx
	mh.buckets = mh.buckets[:mh.Len()-1]

	if idx < mh.Len() {
		mh.siftDown(idx)
	}
	delete(mh.indexes, pb)
}

func (mh *minHeap) getTop() *profileBase {
	if mh.Len() == 0 {
		return nil
	}
	return mh.buckets[0]
}

type profileBase struct {
	profileID string
	birth     time.Time
	point     *point.Point
}

func defaultDiskCachePath() string {
	return filepath.Join(datakit.CacheDir, defaultDiskCacheFileName)
}

func defaultInput() *Input {
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
		pauseCh:    make(chan bool, inputs.ElectionPauseChannelLength),
		Election:   true,
		semStop:    cliutils.NewSem(),
		feeder:     dkio.DefaultFeeder(),
		Tagger:     datakit.DefaultGlobalTagger(),
		httpClient: http.DefaultClient,
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}

type Input struct {
	Endpoints       []string         `toml:"endpoints"`
	BodySizeLimitMB int              `toml:"body_size_limit_mb"`
	IOConfig        ioConfig         `toml:"io_config"`
	Go              []*GoProfiler    `toml:"go"`
	PyroscopeLists  []*pyroscopeOpts `toml:"pyroscope"`
	Election        bool             `toml:"election"`

	pause   bool
	pauseCh chan bool

	profileSendingAPI *url.URL
	httpClient        *http.Client

	semStop *cliutils.Sem // start stop signal
	feeder  dkio.Feeder
	Tagger  datakit.GlobalTagger
}

func (ipt *Input) getBodySizeLimit() int64 {
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

// uploadResponse {"content":{"profileID":"fa9c3d16-1cfc-4e37-950d-129cbebd1cdb"}}.
type uploadResponse struct {
	Content *struct {
		ProfileID string `json:"profileID"`
	} `json:"content"`
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

func (ipt *Input) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// not a post request
	if req.Body == nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, "profiling request body is nil")
		log.Error("Incoming profiling request body is nil")
		return
	}

	bodyBytes, err := io.ReadAll(http.MaxBytesReader(w, req.Body, ipt.getBodySizeLimit()))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, fmt.Sprintf("Unable to read profile body: %s", err))
		log.Errorf("Unable to read profile body: %s", err)
		return
	}

	headers := make(map[string]string, len(req.Header))

	for k, v := range req.Header {
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
		msg := fmt.Sprintf("unable to marshal request using protobuf: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, msg)
		log.Error(msg)
		return
	}

	if err := diskQueue.Put(pbBytes); err != nil {
		msg := fmt.Sprintf("unable to push request to disk queue: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, msg)
		log.Error(msg)
		return
	}
}

func insertEventFile(req *http.Request, oldBody []byte, metadata *resolvedMetadata) (io.Reader, error) {
	boundary, err := getBoundary(req.Header.Get("Content-Type"))
	if err != nil {
		return nil, fmt.Errorf("unable to get boundary: %w", err)
	}

	buf := &bytes.Buffer{}
	pm, err := newMultipartPrepend(buf, boundary)
	if err != nil {
		return nil, fmt.Errorf("unable add event file to multipart form: %w", err)
	}

	f, err := pm.CreateFormFile(eventJSONFile, eventJSONFileWithSuffix)
	if err != nil {
		return nil, fmt.Errorf("unable to create form file: %w", err)
	}

	md := Metadata{}

	for name, fileHeaders := range req.MultipartForm.File {
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
			md.Format = PPROF
		case ".jfr":
			md.Format = JFR
		}
	}
	if md.Format == "" {
		md.Format = "unknown"
	}
	startTime, err := resolveStartTime(metadata.formValue)
	if err != nil {
		log.Warnf("unable to resolve profile start time: %w", err)
	} else {
		md.Start = rfc3339Time(startTime)
	}

	endTime, err := resolveEndTime(metadata.formValue)
	if err != nil {
		log.Warnf("unable to resolve profile end time: %w", err)
	} else {
		md.End = rfc3339Time(endTime)
	}

	tags := metadata.tags
	lang := resolveLang(metadata.formValue, tags)
	md.Language = lang

	md.TagsProfiler = strings.Join(metadata.formValue[profileTagsKey], ",")

	mdBytes, err := json.Marshal(md)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal data for profiling event file: %w", err)
	}

	if _, err := f.Write(mdBytes); err != nil {
		return nil, fmt.Errorf("unable to write data to multipart file: %w", err)
	}

	if err := pm.Close(); err != nil {
		return nil, fmt.Errorf("unable to close multipart prepend: %w", err)
	}

	buf.Write(oldBody)
	return buf, nil
}

func (ipt *Input) sendRequestToDW(ctx context.Context, pbBytes []byte) error {
	var reqPB rum.RequestPB

	if err := proto.Unmarshal(pbBytes, &reqPB); err != nil {
		return fmt.Errorf("unable to unmarshal profiling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx,
		http.MethodPost, ipt.profileSendingAPI.String(), bytes.NewReader(reqPB.Body))
	if err != nil {
		return fmt.Errorf("unable to create http request: %w", err)
	}

	for k, v := range reqPB.Header {
		// ignore header "Host" and "Content-Length"
		if !strings.EqualFold("Host", k) && !strings.EqualFold("Content-Length", k) {
			req.Header.Set(k, v)
		}
	}

	if err := req.ParseMultipartForm(ipt.getBodySizeLimit()); err != nil {
		return fmt.Errorf("unable to parse multipart/formdata: %w", err)
	}

	metadata, _, err := parseMetadata(req)
	if err != nil {
		return fmt.Errorf("unable to resolve profiling tags: %w", err)
	}

	bodyReader := io.Reader(bytes.NewReader(reqPB.Body))

	// Add event form file to multipartForm if it doesn't exist
	_, ok1 := req.MultipartForm.File[eventJSONFile]
	_, ok2 := req.MultipartForm.File[eventJSONFileWithSuffix]

	if !ok1 && !ok2 {
		if reader, err := insertEventFile(req, reqPB.Body, metadata); err != nil {
			log.Warnf("unable to insert event form file to profiling request: %s", err)
		} else {
			bodyReader = reader
		}
	}

	req.Header.Set(XDataKitVersionHeader, datakit.Version)

	xGlobalTag := dataway.SinkHeaderValueFromTags(metadata.tags,
		config.Cfg.Dataway.GlobalTags(),
		config.Cfg.Dataway.CustomTagKeys())
	if xGlobalTag == "" {
		xGlobalTag = config.Cfg.Dataway.GlobalTagsHTTPHeaderValue()
	}

	req.Header.Set(dataway.HeaderXGlobalTags, xGlobalTag)

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
			_ = req.Body.Close()
			req.Body = io.NopCloser(bodyReader)

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

	metricName := inputName + "/" + resolveLang(metadata.formValue, metadata.tags).String()
	if sendErr == nil && resp.StatusCode/100 == 2 {
		dkio.InputsFeedVec().WithLabelValues(metricName, point.Profiling.String()).Inc()
		dkio.InputsFeedPtsVec().WithLabelValues(metricName, point.Profiling.String()).Inc()
		dkio.InputsLastFeedVec().WithLabelValues(metricName, point.Profiling.String()).Set(float64(time.Now().Unix()))
		dkio.InputsCollectLatencyVec().WithLabelValues(metricName, point.Profiling.String()).Observe(reqCost.Seconds())
	} else {
		feedErr := sendErr
		if feedErr == nil {
			feedErr = fmt.Errorf("error status code %d", resp.StatusCode)
		}
		ipt.feeder.FeedLastError(feedErr.Error(),
			dkio.WithLastErrorInput(metricName),
			dkio.WithLastErrorCategory(point.Profiling),
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
				default:
					func() {
						var reqData []byte
						if err := diskQueue.Get(func(msg []byte) error {
							reqData = msg
							return nil
						}); err != nil {
							if errors.Is(err, diskcache.ErrEOF) {
								log.Debugf("disk queue is empty: %s", err)
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
	iptGlobal = ipt
	log.Infof("the input %s is running...", inputName)

	if err := ipt.InitDiskQueueIO(); err != nil {
		log.Errorf("unable to start IO process for profiling: %s", err)
	}

	group := goroutine.NewGroup(goroutine.Option{
		Name: "profile",
		PanicCb: func(b []byte) bool {
			log.Error(string(b))
			return false
		},
	})

	for _, g := range ipt.Go {
		func(g *GoProfiler) {
			group.Go(func(ctx context.Context) error {
				if err := g.run(ipt); err != nil {
					log.Errorf("go profile collect error: %s", err.Error())
				}
				return nil
			})
		}(g)
	}

	for _, g := range ipt.PyroscopeLists {
		func(g *pyroscopeOpts) {
			group.Go(func(ctx context.Context) error {
				if err := g.run(ipt); err != nil {
					log.Errorf("pyroscope profile collect error: %s", err.Error())
				}
				return nil
			})
		}(g)
	}

	if err := group.Wait(); err != nil {
		log.Errorf("profile collect err: %s", err.Error())
	}
}

func (ipt *Input) SampleConfig() string {
	return sampleConfig
}

func (ipt *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{&trace.TraceMeasurement{Name: inputName}}
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

type eventOpts struct {
	Family   string `json:"family"`
	Format   string `json:"format"`
	Profiler string `json:"profiler"`
}

func pushProfileData(opt *pushProfileDataOpt, event *eventOpts) error {
	b := new(bytes.Buffer)
	mw := multipart.NewWriter(b)

	for _, profileData := range opt.profiledatas {
		if ff, err := mw.CreateFormFile(profileData.fileName, profileData.fileName); err != nil {
			continue
		} else {
			if _, err = io.Copy(ff, profileData.buf); err != nil {
				continue
			}
		}
	}

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", "form-data; name=\"event\"; filename=\"event.json\"")
	h.Set("Content-Type", "application/json")
	f, err := mw.CreatePart(h)
	if err != nil {
		return err
	}

	eventJSONString, err := json.Marshal(event)
	if err != nil {
		return err
	}

	if _, err := io.Copy(f, bytes.NewReader(eventJSONString)); err != nil {
		return err
	}
	if err := mw.Close(); err != nil {
		return err
	}

	profileID := randomProfileID()

	URL, transport, err := profilingProxyURL()
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", URL.String(), b)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set(profileIDHeaderKey, profileID)
	req.Header.Set(XDataKitVersionHeader, datakit.Version)
	req.Header.Set(timestampHeaderKey, strconv.FormatInt(opt.startTime.UnixNano(), 10))

	client := &http.Client{
		Transport: transport,
		Timeout:   15 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close() //nolint:errcheck

	bo, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusOK {
		var resp uploadResponse

		if err := json.Unmarshal(bo, &resp); err != nil {
			return fmt.Errorf("json unmarshal upload profile binary response err: %w", err)
		}

		if resp.Content == nil || resp.Content.ProfileID == "" {
			return fmt.Errorf("fetch profile upload response profileID fail")
		}

		if err := writeProfilePoint(&writeProfilePointOpt{
			profileID:       profileID,
			startTime:       opt.startTime,
			endTime:         opt.endTime,
			reportFamily:    event.Family,
			reportFormat:    event.Format,
			endPoint:        opt.endPoint,
			inputTags:       opt.inputTags,
			inputNameSuffix: opt.inputNameSuffix,
			Input:           opt.Input,
		}); err != nil {
			return fmt.Errorf("write profile point failed: %w", err)
		}
	} else {
		return fmt.Errorf("push profile data failed, response status: %s", resp.Status)
	}
	return nil
}

type writeProfilePointOpt struct {
	profileID       string
	startTime       time.Time
	endTime         time.Time
	reportFamily    string
	reportFormat    string
	endPoint        string
	inputTags       map[string]string
	inputNameSuffix string
	Input           *Input
}

func writeProfilePoint(opt *writeProfilePointOpt) error {
	pointTags := map[string]string{
		TagEndPoint: opt.endPoint,
		TagLanguage: opt.reportFamily,
	}

	// extend custom tags
	for k, v := range opt.inputTags {
		if _, ok := pointTags[k]; !ok {
			pointTags[k] = v
		}
	}

	//nolint:lll
	pointFields := map[string]interface{}{
		FieldProfileID:  opt.profileID,
		FieldFormat:     opt.reportFormat,
		FieldDatakitVer: datakit.Version,
		FieldStart:      opt.startTime.UnixNano(),
		FieldEnd:        opt.endTime.UnixNano(),
		FieldDuration:   opt.endTime.Sub(opt.startTime).Nanoseconds(),
	}

	opts := point.CommonLoggingOptions()
	opts = append(opts, point.WithTime(opt.startTime))

	if opt.Input.Election {
		pointTags = inputs.MergeTagsWrapper(pointTags, opt.Input.Tagger.ElectionTags(), opt.inputTags, "")
	} else {
		pointTags = inputs.MergeTagsWrapper(pointTags, opt.Input.Tagger.HostTags(), opt.inputTags, "")
	}

	pt := point.NewPointV2(inputName, append(point.NewTags(pointTags), point.NewKVs(pointFields)...), opts...)

	if err := iptGlobal.feeder.FeedV2(point.Profiling, []*point.Point{pt},
		dkio.WithCollectCost(time.Since(pt.Time())),
		dkio.WithElection(opt.Input.Election),
		dkio.WithInputName(inputName+opt.inputNameSuffix),
	); err != nil {
		return err
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
