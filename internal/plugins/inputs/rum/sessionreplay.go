// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package rum

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/GuanceCloud/cliutils/diskcache"
	filter2 "github.com/GuanceCloud/cliutils/filter"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
	"golang.org/x/exp/maps"
	"golang.org/x/time/rate"
	"google.golang.org/protobuf/proto"
)

// logMultiPartBodyRate indicates recording at most 1 body log for every minute.
var logMultiPartBodyRate = rate.NewLimiter(rate.Every(time.Minute), 1)

type SessionReplayCfg struct {
	CachePath         string                    `toml:"cache_path"`
	CacheCapacity     int64                     `toml:"cache_capacity_mb"`
	ClearCacheOnStart bool                      `toml:"clear_cache_on_start"`
	UploadWorkers     int                       `toml:"upload_workers"`
	SendTimeout       time.Duration             `toml:"send_timeout"`
	SendRetryCount    int                       `toml:"send_retry_count"`
	FilterRules       []string                  `toml:"filter_rules"`
	whereConditions   []filter2.WhereConditions `toml:"-"`
}

type ReplayFilterKV map[string]string

func (kv ReplayFilterKV) Get(key string) (any, bool) {
	v, ok := kv[key]
	return v, ok
}

func defaultSessionReplayCfg() *SessionReplayCfg {
	cfg := &SessionReplayCfg{
		CachePath:         filepath.Join(datakit.CacheDir, "session_replay"),
		CacheCapacity:     defaultReplayCacheMaxMib,
		ClearCacheOnStart: false,
		UploadWorkers:     16,
		SendTimeout:       time.Second * 75,
		SendRetryCount:    3,
		FilterRules:       nil,
	}
	return cfg
}

func (ipt *Input) sessionReplayHandler() (f http.HandlerFunc, err error) {
	if err := ipt.initReplayHTTPClient(); err != nil {
		return nil, fmt.Errorf("unable to init session replay http client: %w", err)
	}

	if err := ipt.initReplayDiskQueue(); err != nil {
		return nil, fmt.Errorf("unable to init diskqueue: %w", err)
	}

	if err := ipt.initSessionReplayWorkers(); err != nil {
		return nil, fmt.Errorf("unable to start session replay uploading workers: %w", err)
	}

	return func(w http.ResponseWriter, req *http.Request) {
		readAt := time.Now()
		body, err := io.ReadAll(req.Body)
		if err != nil {
			log.Errorf("unable to read request body, already read data length [%d] : %s", len(body), err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		readDuration := time.Since(readAt)

		appID := StatusUnknown
		env := StatusUnknown
		version := StatusUnknown
		service := StatusUnknown

		defer func() {
			replayReadBodyDelaySeconds.
				WithLabelValues(appID, env, version, service).
				Observe(readDuration.Seconds())
		}()

		if len(body) > ReplayBodyMaxSize {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = io.WriteString(w, fmt.Sprintf("request body size [%d] exceeds the limit", req.ContentLength))
			return
		}

		req.Body = io.NopCloser(bytes.NewReader(body))

		if err := req.ParseMultipartForm(ReplayBodyMaxSize); err != nil {
			if logMultiPartBodyRate.Allow() {
				log.Warnf("unable to parse session replay multipart form: %s, the malformed body is: %v", err, body)
			} else {
				log.Errorf("unable to parse session replay multipart form: %s, body length: %d", err, len(body))
			}
			w.WriteHeader(http.StatusBadRequest)
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

		filterKV := maps.Clone[ReplayFilterKV, string, string](headers)
		formValues := make(map[string]*ValuesSlice, len(req.MultipartForm.Value))

		for k, v := range req.MultipartForm.Value {
			formValues[k] = &ValuesSlice{
				Values: v,
			}
			if len(v) > 0 {
				filterKV[k] = v[0]
			} else {
				filterKV[k] = ""
			}
		}

		appID = filterKV["app_id"]
		env = filterKV["env"]
		version = filterKV["version"]
		service = filterKV["service"]

		if len(ipt.SessionReplayCfg.whereConditions) > 0 {
			// multi rules are of relationship OR.
			for _, cond := range ipt.SessionReplayCfg.whereConditions {
				if cond.Eval(filterKV) >= 0 {
					// drop this data if match the rule
					log.Infof("session replay data is dropped as it matches the filter rules")
					replayFilteredTotalCount.WithLabelValues(appID, env, version, service).Inc()

					replayFilteredTotalBytes.WithLabelValues(appID, env, version, service).Add(float64(len(body)))
					w.WriteHeader(http.StatusAccepted)
					return
				}
			}
		}

		reqPB := &RequestPB{
			Header:     headers,
			Body:       body,
			FormValues: formValues,
		}

		pbData, err := proto.Marshal(reqPB)
		if err != nil {
			log.Errorf("unable to marshal request to protobuf: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := ipt.replayDiskQueue.Put(pbData); err != nil {
			log.Errorf("unable to cache request: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}, nil
}

func (ipt *Input) initReplayHTTPClient() error {
	endpoints := config.Cfg.Dataway.GetEndpoints()
	if len(endpoints) == 0 {
		return fmt.Errorf("no available dataway endpoint")
	}

	ep := endpoints[0]

	ipt.replayHTTPClient = &http.Client{
		Timeout:   ipt.SessionReplayCfg.SendTimeout,
		Transport: ep.Transport(),
	}

	ipt.replayUploadAPI = ep.GetCategoryURL()[datakit.SessionReplayUpload]
	return nil
}

func (ipt *Input) initSessionReplayWorkers() error {
	replayWorkersGroup = goroutine.NewGroup(goroutine.Option{Name: "session_replay_uploading"})

	for i := 0; i < ipt.SessionReplayCfg.UploadWorkers; i++ {
		replayWorkersGroup.Go(func(ctx context.Context) error {
			for {
				select {
				case <-datakit.Exit.Wait():
					log.Infof("session replay uploading worker exit now...")
					return nil
				case <-ctx.Done():
					log.Infof("context canceld...")
					return nil
				default:
					func() {
						var msgData []byte
						if err := ipt.replayDiskQueue.Get(func(msg []byte) error {
							msgData = msg
							return nil
						}); err != nil {
							if errors.Is(err, diskcache.ErrEOF) {
								log.Debugf("disk queue is empty: %s", err)
								time.Sleep(time.Millisecond * 1500)
							} else {
								log.Errorf("unable to get msg from disk cache: %s", err)
								time.Sleep(time.Millisecond * 100)
							}
							return
						}
						if err := ipt.uploadSessionReplay(msgData); err != nil {
							log.Errorf("fail to send session replay: %s", err)
						}
					}()
				}
			}
		})
	}
	return nil
}

func (ipt *Input) uploadSessionReplay(msg []byte) (err error) {
	var (
		reqPB      RequestPB
		resp       *http.Response
		lastErr    error
		appID      string
		env        string
		version    string
		service    string
		statusCode = "unknown"
	)

	defer func() {
		if err != nil || (resp != nil && resp.StatusCode/100 != 2) {
			replayFailureTotalCount.WithLabelValues(appID, env, version, service, statusCode).Inc()
			replayFailureTotalBytes.WithLabelValues(appID, env, version, service, statusCode).Add(float64(len(msg)))
		}
	}()

	if len(msg) == 0 {
		return fmt.Errorf("empty session replay cache data")
	}

	if err := proto.Unmarshal(msg, &reqPB); err != nil {
		return fmt.Errorf("unable to unmarshal protobuf msg [%v] from disk queue: %w", msg, err)
	}

	req, err := http.NewRequest(http.MethodPost, ipt.replayUploadAPI, bytes.NewReader(reqPB.Body))
	if err != nil {
		return fmt.Errorf("unbale to create http request: %w", err)
	}

	for k, v := range reqPB.Header {
		req.Header.Set(k, v)
	}

	var formValues map[string][]string
	if reqPB.FormValues == nil || len(reqPB.FormValues) == 0 {
		if err := req.ParseMultipartForm(ReplayBodyMaxSize); err != nil {
			return fmt.Errorf("unable to parse multipart form from session replay request: %w", err)
		}

		formValues = req.MultipartForm.Value
	} else {
		formValues = make(map[string][]string, len(reqPB.FormValues))

		for k, v := range reqPB.FormValues {
			formValues[k] = v.Values
		}
	}

	globalTags := config.Cfg.Dataway.GlobalTags()
	customTagKeys := config.Cfg.Dataway.CustomTagKeys()

	tags := map[string]string{
		"category": "session_replay",
	}

	for k, v := range formValues {
		if _, ok := tags[k]; !ok {
			if len(v) == 0 {
				tags[k] = ""
			} else {
				tags[k] = v[0]
			}
		}
	}

	appID = tags["app_id"]
	env = tags["env"]
	version = tags["version"]
	service = tags["service"]

	headerValue := dataway.SinkHeaderValueFromTags(tags,
		globalTags,
		customTagKeys)
	if headerValue == "" {
		headerValue = config.Cfg.Dataway.GlobalTagsHTTPHeaderValue()
	}
	req.Header.Set(dataway.HeaderXGlobalTags, headerValue)

	startTime := time.Now()
	defer func() {
		reqCost := time.Since(startTime).Seconds()
		replayUploadingDurationSummary.WithLabelValues(appID, env, version, service, statusCode).Observe(reqCost)
		dataway.APISumVec().WithLabelValues(req.URL.Path, statusCode).Observe(reqCost)
	}()

	for i := 0; i < ipt.SessionReplayCfg.SendRetryCount; i++ {
		lastErr = func(idx int) error {
			req.Body = io.NopCloser(bytes.NewReader(reqPB.Body))

			resp, err = ipt.replayHTTPClient.Do(req)
			if err != nil {
				statusCode = "unknown"
				return fmt.Errorf("at #%d try: unable to send session replay data to dataway: %w", idx+1, err)
			}
			defer resp.Body.Close() // nolint:errcheck

			statusCode = http.StatusText(resp.StatusCode)

			errMsg := []byte(nil)
			if resp.StatusCode/100 != 2 {
				errMsg, _ = io.ReadAll(resp.Body)
			}

			switch resp.StatusCode / 100 {
			case 5:
				return fmt.Errorf("at #%d try: unable to send session replay data to dataway, http Status: %s, response: %s",
					idx+1, resp.Status, string(errMsg))
			case 2:
				// ignore
			default:
				log.Errorf("at #%d try: unable to send session replay data to dataway, http status: %s, response: %s",
					idx+1, resp.Status, string(errMsg))
			}

			return nil
		}(i)

		// Log IO retry metrics
		if i > 0 {
			dataway.HTTPRetry().WithLabelValues(req.URL.Path, statusCode).Inc()
		}

		if lastErr == nil {
			return nil
		}
	}
	return lastErr
}

func (ipt *Input) initReplayDiskQueue() error {
	if ipt.SessionReplayCfg.ClearCacheOnStart {
		if err := os.RemoveAll(ipt.SessionReplayCfg.CachePath); err != nil {
			return fmt.Errorf("unable to clear previous session replay cache: %w", err)
		}
	}

	queue, err := diskcache.Open(
		diskcache.WithPath(ipt.SessionReplayCfg.CachePath),
		diskcache.WithCapacity(ipt.SessionReplayCfg.CacheCapacity*MiB),
		diskcache.WithNoFallbackOnError(true),
	)
	if err != nil {
		return fmt.Errorf("unable to init disk queue: %w", err)
	}
	ipt.replayDiskQueue = queue

	return nil
}
