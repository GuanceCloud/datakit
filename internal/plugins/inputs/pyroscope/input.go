// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package pyroscope continuous profiling collector
package pyroscope

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
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
	"github.com/google/pprof/profile"
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
	timeutils "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/time"
)

const (
	inputName                  = "pyroscope"
	MiB                        = 1 << 20 // 1MB
	defaultPyroscopeMaxSize    = 32      // 32MB
	defaultDiskCacheSize       = 10240   // 10240MB, 10GB
	defaultDiskCacheFileName   = "pyroscope_inputs"
	defaultConsumeWorkersCount = 8
	defaultHTTPClientTimeout   = time.Second * 75
	defaultHTTPRetryCount      = 4
	XDataKitVersionHeader      = "X-Datakit-Version"
	timestampHeaderKey         = "X-Datakit-UnixNano"
	sampleConfig               = `
[[inputs.pyroscope]]
  ## pyroscope Agent endpoints register by version respectively.
  ## Endpoints can be skipped listen by remove them from the list.
  ## Default value set as below. DO NOT MODIFY THESE ENDPOINTS if not necessary.
  endpoints = ["/ingest"]

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
  # [inputs.pyroscope.io_config]
  #   cache_path = "/usr/local/datakit/cache/pyroscope_inputs"  # C:\Program Files\datakit\cache\pyroscope_inputs by default on Windows
  #   cache_capacity_mb = 10240  # 10240MB
  #   clear_cache_on_start = false
  #   upload_workers = 8
  #   send_timeout = "75s"
  #   send_retry_count = 4

  ## set custom tags for profiling data
  # [inputs.pyroscope.tags]
  #   some_tag = "some_value"
  #   more_tag = "some_other_value"
`
)

var (
	log = logger.DefaultSLogger(inputName)

	_ inputs.HTTPInput     = (*Input)(nil)
	_ inputs.InputV2       = (*Input)(nil)
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
		BodySizeLimitMB: defaultPyroscopeMaxSize,
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
	Election        bool              `toml:"election"`
	GenerateMetrics bool              `toml:"generate_metrics"`

	pauseCh           chan bool
	profileSendingAPI *url.URL
	httpClient        *http.Client
	semStop           *cliutils.Sem // start stop signal
	feeder            dkio.Feeder
	Tagger            datakit.GlobalTagger
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
		return dkhttp.NewErr(fmt.Errorf("unable to read pyroscope body: %w", err), http.StatusBadRequest)
	}

	headers := make(map[string]string, len(r.Header))

	for k, v := range r.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		} else {
			headers[k] = ""
		}
	}

	queries := QueryToPBForms(r.URL.Query())

	reqPB := &rum.RequestPB{
		Header:     headers,
		Body:       bodyBytes,
		FormValues: queries,
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
		io.WriteString(w, err.Error()) //nolint:errcheck,gosec
		return
	}
}

func getBoundary(contentType string) (string, error) {
	if contentType == "" {
		return "", fmt.Errorf("empty content-type")
	}
	_, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return "", fmt.Errorf("unable to parse mediatype: %w", err)
	}
	if _, ok := params["boundary"]; !ok {
		return "", fmt.Errorf("boundray not found")
	}
	return params["boundary"], nil
}

func insertEventFile(r *http.Request, metadata *metrics.Metadata) ([]byte, error) {
	boundary, err := getBoundary(r.Header.Get("Content-Type"))
	if err != nil {
		return nil, fmt.Errorf("unable to get multipart boundary: %w", err)
	}

	out := new(bytes.Buffer)
	mw := multipart.NewWriter(out)

	if err = mw.SetBoundary(boundary); err != nil {
		return nil, fmt.Errorf("unable to set multipart form boundary: %w", err)
	}

	f, err := mw.CreateFormFile(metrics.EventJSONFile, metrics.EventJSONFile)
	if err != nil {
		return nil, fmt.Errorf("unable to create form file: %w", err)
	}

	if err = json.NewEncoder(f).Encode(metadata); err != nil {
		return nil, fmt.Errorf("unable to marshal data for profiling event file: %w", err)
	}

	out.WriteString("\r\n")

	return out.Bytes(), nil
}

func rebuildMultipartForm(r *http.Request, metadata *metrics.Metadata, newFilename string) ([]byte, error) {
	boundary, err := getBoundary(r.Header.Get("Content-Type"))
	if err != nil {
		return nil, fmt.Errorf("unable to get multipart boundary: %w", err)
	}

	out := &bytes.Buffer{}
	mw := multipart.NewWriter(out)

	if err = mw.SetBoundary(boundary); err != nil {
		return nil, fmt.Errorf("unable to set multipart form boundary: %w", err)
	}

	for k, vv := range r.MultipartForm.Value {
		for _, v := range vv {
			if err = mw.WriteField(k, v); err != nil {
				return nil, fmt.Errorf("unable to write form field: %w", err)
			}
		}
	}

	f, err := mw.CreateFormFile(metrics.EventJSONFile, metrics.EventJSONFile)
	if err != nil {
		return nil, fmt.Errorf("unable to create form file: %w", err)
	}

	if err = json.NewEncoder(f).Encode(metadata); err != nil {
		return nil, fmt.Errorf("unable to marshal data for profiling event file: %w", err)
	}

	cp := func(h *multipart.FileHeader, fieldName string) error {
		src, err := h.Open()
		if err != nil {
			return fmt.Errorf("unable to open form file: %w", err)
		}
		defer src.Close() // nolint:errcheck,gosec

		dst, err := mw.CreateFormFile(fieldName, h.Filename)
		if err != nil {
			return fmt.Errorf("unable to create form file: %w", err)
		}
		if _, err := io.Copy(dst, src); err != nil {
			return fmt.Errorf("unable to copy form file: %w", err)
		}
		return nil
	}

	for name, files := range r.MultipartForm.File {
		if name == metrics.EventFile || name == metrics.EventJSONFile {
			continue
		}

		for _, fileHeader := range files {
			if fileHeader.Filename == metrics.EventFile || fileHeader.Filename == metrics.EventJSONFile {
				continue
			}
			if name == "profile" || len(r.MultipartForm.File) == 1 {
				name = newFilename
				fileHeader.Filename = newFilename
			}
			if err = cp(fileHeader, name); err != nil {
				return nil, fmt.Errorf("copy form fileHeader: %w", err)
			}
		}
	}

	if err = mw.Close(); err != nil {
		return nil, fmt.Errorf("unable to close multipart form: %w", err)
	}

	return out.Bytes(), nil
}

func (ipt *Input) sendRequestToDW(ctx context.Context, pbBytes []byte) error {
	var reqPB rum.RequestPB

	if err := proto.Unmarshal(pbBytes, &reqPB); err != nil {
		return fmt.Errorf("unable to unmarshal profiling request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ipt.profileSendingAPI.String(), nil)
	if err != nil {
		return fmt.Errorf("unable to create http request: %w", err)
	}

	for k, v := range reqPB.Header {
		if strings.EqualFold("Host", k) || strings.EqualFold("Content-Length", k) ||
			strings.EqualFold("Content-Encoding", k) {
			continue
		}
		req.Header.Set(k, v)
	}

	tags := ParseTags(reqPB.FormValues)
	language := ResolveLanguage(tags)
	tags["language"] = language.String()

	for tk, tv := range ipt.Tags {
		tags[tk] = tv
	}

	var filename string
	format := metrics.PPROF
	switch tags["format"] {
	case "jfr":
		filename = "auto.jfr"
		format = metrics.JFR
	case "pprof":
		filename = "auto.pprof"
	default:
		if language == metrics.Java {
			filename = "auto.jfr"
			format = metrics.JFR
		} else {
			filename = "auto.pprof"
			format = metrics.PPROF
		}
	}

	fromTime, err := strconv.ParseInt(tags["from"], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid profile start time [%q]: %w", tags["from"], err)
	}
	toTime, err := strconv.ParseInt(tags["until"], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid profile end time [%q]: %w", tags["until"], err)
	}

	pprofEnd := timeutils.UnixStampToTime(toTime, time.UTC)

	metadata := metrics.Metadata{
		Format:        format,
		Profiler:      metrics.Pyroscope,
		Attachments:   []string{filename},
		Language:      language,
		Tags:          tags,
		TagsProfiler:  metrics.JoinTags(tags),
		SubCustomTags: tags[metrics.SubCustomTagsKey],
		Start:         metrics.NewRFC3339Time(timeutils.UnixStampToTime(fromTime, time.UTC)),
		End:           metrics.NewRFC3339Time(pprofEnd),
	}

	contentType := req.Header.Get("Content-Type")
	switch {
	case strings.Contains(contentType, "multipart/form-data"):
		if language == metrics.Golang {
			req.Body = io.NopCloser(bytes.NewReader(reqPB.Body))
			if err = req.ParseMultipartForm(ipt.GetBodySizeLimit()); err != nil {
				return fmt.Errorf("unable to parse multipart form: %w", err)
			}

			fileHeaders, ok := req.MultipartForm.File["sample_type_config"]
			if ok {
				if len(fileHeaders) == 0 {
					return fmt.Errorf("invalid data: no sample_type_config.json file")
				}
				fh := fileHeaders[0]
				f, err := fh.Open()
				if err != nil {
					return fmt.Errorf("unable to open sample_type_config.json file: %w", err)
				}
				defer f.Close() // nolint:errcheck,gosec

				var pprofConfig map[string]map[string]string
				if err := json.NewDecoder(f).Decode(&pprofConfig); err != nil {
					return fmt.Errorf("unable to json decode sample_type_config.json: %w", err)
				}

				containsKey := func(m map[string]map[string]string, k string) bool {
					if _, exists := m[k]; exists {
						return true
					}
					return false
				}

				switch {
				case containsKey(pprofConfig, "goroutine"):
					filename = "goroutines.pprof"
				case containsKey(pprofConfig, "alloc_objects") || containsKey(pprofConfig, "alloc_space") ||
					containsKey(pprofConfig, "inuse_objects") || containsKey(pprofConfig, "inuse_space"):
					filename = "delta-heap.pprof"
				case containsKey(pprofConfig, "contentions") || containsKey(pprofConfig, "delay"):
					if displayName := pprofConfig["contentions"]["display-name"]; displayName != "" {
						if strings.Contains(displayName, "mutex") {
							filename = "delta-mutex.pprof"
							break
						} else if strings.Contains(displayName, "block") {
							filename = "delta-block.pprof"
							break
						}
					}
					if displayName := pprofConfig["delay"]["display-name"]; displayName != "" {
						if strings.Contains(displayName, "mutex") {
							filename = "delta-mutex.pprof"
							break
						} else if strings.Contains(displayName, "block") {
							filename = "delta-block.pprof"
							break
						}
					}
				}
			} else {
				filename = "cpu.pprof"
			}

			metadata.Attachments = []string{filename}

			// Pack pprof files that start at the same time to a bundle.
			if sessionID := tags["__session_id__"]; sessionID != "" && len(req.MultipartForm.File["profile"]) > 0 {
				f, err := req.MultipartForm.File["profile"][0].Open()
				if err != nil {
					return fmt.Errorf("unable to open profile file: %w", err)
				}
				defer f.Close() // nolint:errcheck

				fileBody, err := io.ReadAll(f)
				if err != nil {
					return fmt.Errorf("unable to read profile file: %w", err)
				}

				pf := &pprofFile{
					name:    filename,
					payload: fileBody,
				}
				ipt.addPProfToCachePool(sessionID, pprofEnd, pf, metadata, reqPB.Header)
				return nil
			}
			newBody, err := rebuildMultipartForm(req, &metadata, filename)
			if err != nil {
				return fmt.Errorf("unable to rebuild multipart form: %w", err)
			}
			reqPB.Body = newBody
		} else {
			// Add an event form file to multipartForm if it doesn't exist.
			if eventBytes, err := insertEventFile(req, &metadata); err != nil {
				log.Warnf("unable to insert event form file to profiling request: %s", err)
			} else {
				reqPB.Body = append(eventBytes, reqPB.Body...)
			}
		}
	case strings.Contains(contentType, "binary/octet-stream"):
		// find user defined tags in .pprof
		if format == metrics.PPROF {
			pprof, err := profile.Parse(bytes.NewReader(reqPB.Body))
			if err != nil {
				log.Warnf("fialed to parse pprof file: %v", err)
			}
			newTags := false
			for _, sample := range pprof.Sample {
				for k, v := range sample.Label {
					if k == "span_id" || k == "thread_id" || k == "thread_name" {
						continue
					}
					if _, ok := tags[k]; !ok {
						tags[k] = strings.Join(v, ",")
						newTags = true
					}
				}
			}
			if tags["runtime_id"] != "" && tags["runtime-id"] == "" {
				tags["runtime-id"] = tags["runtime_id"]
				newTags = true
			}
			if newTags {
				metadata.TagsProfiler = metrics.JoinTags(tags)
			}
		}

		// build multipart/form-data body
		multiBody := new(bytes.Buffer)
		mw := multipart.NewWriter(multiBody)
		ff, err := mw.CreateFormFile(filename, filename)
		if err != nil {
			return fmt.Errorf("unable to create form file: %w", err)
		}
		if _, err = ff.Write(reqPB.Body); err != nil {
			return fmt.Errorf("unable to write data to multipart file: %w", err)
		}

		ef, err := mw.CreateFormFile(metrics.EventJSONFile, metrics.EventJSONFile)
		if err != nil {
			return fmt.Errorf("unable to create event file: %w", err)
		}

		if err = json.NewEncoder(ef).Encode(metadata); err != nil {
			return fmt.Errorf("unable to marshal event file: %w", err)
		}

		if err = mw.Close(); err != nil {
			return fmt.Errorf("unable to close multipart builder: %w", err)
		}

		reqPB.Body = multiBody.Bytes()
		req.Header.Set("Content-Type", mw.FormDataContentType())

	default:
		return fmt.Errorf("unsupported Content-Type: %s", contentType)
	}

	return ipt.doSend(req, reqPB.Body, tags)
}

func (ipt *Input) doSend(req *http.Request, body []byte, tags map[string]string) error {
	// apply remote or local filter
	pt := point.NewPoint(inputName, point.NewTags(tags), point.WithTime(time.Now()))
	if len(filter.FilterPts(point.Profiling, []*point.Point{pt})) == 0 {
		log.Infof("the profiling data matched the remote or local blacklist and was dropped")
		return nil
	}

	req.Header.Set(XDataKitVersionHeader, datakit.Version)
	if config.Cfg.Dataway.EnableSinker {
		xGlobalTag := dataway.SinkHeaderValueFromTags(tags,
			config.Cfg.Dataway.GlobalTags(),
			config.Cfg.Dataway.CustomTagKeys())
		if xGlobalTag == "" {
			xGlobalTag = config.Cfg.Dataway.GlobalTagsHTTPHeaderValue()
		}

		req.Header.Set(dataway.HeaderXGlobalTags, xGlobalTag)
	}

	var (
		sendErr    error
		err        error
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
				req.Body.Close() //nolint:errcheck,gosec
			}
			req.Body = io.NopCloser(bytes.NewReader(body))
			req.GetBody = func() (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader(body)), nil
			}
			req.ContentLength = int64(len(body))

			resp, err = ipt.httpClient.Do(req)
			if err != nil {
				statusCode = "unknown"
				time.Sleep(time.Second)
				e := fmt.Errorf("at #%d try: unable to send profiling request to dataway: %w", idx+1, err)
				return newRetryError(e)
			}
			defer resp.Body.Close() // nolint:errcheck,gosec
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

	metricName := inputName + "/" + tags["language"]
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

	pyroscopeURL, transport, err := profilingProxyURL()
	if err != nil {
		return fmt.Errorf("no dataway endpoint available: %w", err)
	}

	ipt.profileSendingAPI = pyroscopeURL

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
	log.Infof("the input %s is running...", inputName)

	metrics.InitLog()

	if err := ipt.InitDiskQueueIO(); err != nil {
		log.Errorf("unable to start IO process for profiling: %s", err)
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
}

func (*Input) Dashboard(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		return map[string]string{
			//nolint:lll
		}
	case inputs.I18nEn:
		return map[string]string{
			//nolint:lll
		}
	default:
		return nil
	}
}

func (*Input) Monitor(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		return map[string]string{
			//nolint:lll
		}
	case inputs.I18nEn:
		return map[string]string{
			//nolint:lll
		}
	default:
		return nil
	}
}

type pushPyroscopeDataOpt struct {
	startTime       time.Time
	endTime         time.Time
	pyroscopeData   []*pyroscopeData
	endPoint        string
	inputTags       map[string]string
	inputNameSuffix string
	Input           *Input
}

func pushPyroscopeData(opt *pushPyroscopeDataOpt, event *metrics.Metadata, bodySizeLimit int64) error {
	b := new(bytes.Buffer)
	mw := multipart.NewWriter(b)

	for _, pd := range opt.pyroscopeData {
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
