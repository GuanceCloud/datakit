// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package rum real user monitoring
package rum

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/diskcache"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/gobwas/glob"
	"github.com/prometheus/client_golang/prometheus"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/filter"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/storage"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/workerpool"
	"google.golang.org/protobuf/proto"
)

var (
	_ inputs.HTTPInput = &Input{}
	_ inputs.InputV2   = &Input{}
)

const (
	MiB                      = 1 << 20
	inputName                = "rum"
	ReplayBodyMaxSize        = MiB * 16 // 16Mib
	defaultReplayCacheMaxMib = 20480    // 20 Gib
	sampleConfig             = `
[[inputs.rum]]
  ## profile Agent endpoints register by version respectively.
  ## Endpoints can be skipped listen by remove them from the list.
  ## Default value set as below. DO NOT MODIFY THESE ENDPOINTS if not necessary.
  endpoints = ["/v1/write/rum"]

  ## used to upload rum session replay.
  session_replay_endpoints = ["/v1/write/rum/replay"]

  ## specify which metrics should be captured.
  measurements = ["view", "resource", "action", "long_task", "error", "telemetry"]

  ## Android command-line-tools HOME
  android_cmdline_home = "/usr/local/datakit/data/rum/tools/cmdline-tools"

  ## proguard HOME
  proguard_home = "/usr/local/datakit/data/rum/tools/proguard"

  ## android-ndk HOME
  ndk_home = "/usr/local/datakit/data/rum/tools/android-ndk"

  ## atos or atosl bin path
  ## for macOS datakit use the built-in tool atos default
  ## for Linux there are several tools that can be used to instead of macOS atos partially,
  ## such as https://github.com/everettjf/atosl-rs
  atos_bin_path = "/usr/local/datakit/data/rum/tools/atosl"

  # Provide a list to resolve CDN of your static resource.
  # Below is the Datakit default built-in CDN list, you can uncomment that and change it to your cdn list,
  # it's a JSON array like: [{"domain": "CDN domain", "name": "CDN human readable name", "website": "CDN official website"},...],
  # domain field value can contains '*' as wildcard, for example: "kunlun*.com",
  # it will match "kunluna.com", "kunlunab.com" and "kunlunabc.com" but not "kunlunab.c.com".
  # cdn_map = '''
  # [
  #   {"domain":"15cdn.com","name":"腾正安全加速(原 15CDN)","website":"https://www.15cdn.com"},
  #   {"domain":"tzcdn.cn","name":"腾正安全加速(原 15CDN)","website":"https://www.15cdn.com"}
  # ]
  # '''

  ## Threads config controls how many goroutines an agent cloud start to handle HTTP request.
  ## buffer is the size of jobs' buffering of worker channel.
  ## threads is the total number fo goroutines at running time.
  # [inputs.rum.threads]
  #   buffer = 100
  #   threads = 8

  ## Storage config a local storage space in hard dirver to cache trace data.
  ## path is the local file path used to cache data.
  ## capacity is total space size(MB) used to store data.
  # [inputs.rum.storage]
  #   path = "./rum_storage"
  #   capacity = 5120

  ## session_replay config is used to control Session Replay uploading behavior.
  ## cache_path set the disk directory where temporarily cache session replay data.
  ## cache_capacity_mb specify the max storage space (in MiB) that session replay cache can use.
  ## clear_cache_on_start set whether we should clear all previous session replay cache on restarting Datakit.
  ## upload_workers set the count of session replay uploading workers.
  ## send_timeout specify the http timeout when uploading session replay data to dataway.
  ## send_retry_count set the max retry count when sending every session replay request.
  # [inputs.rum.session_replay]
  #   cache_path = "/usr/local/datakit/cache/session_replay"
  #   cache_capacity_mb = 20480
  #   clear_cache_on_start = false
  #   upload_workers = 16
  #   send_timeout = "75s"
  #   send_retry_count = 3
`
)

const (
	Session   = "session"
	View      = "view"
	Resource  = "resource"
	Action    = "action"
	LongTask  = "long_task"
	Error     = "error"
	Telemetry = "telemetry"
)

var (
	log                = logger.DefaultSLogger(inputName)
	wkpool             *workerpool.WorkerPool
	localCache         *storage.Storage
	replayWorkersGroup *goroutine.Group
)

var kunlunCDNGlob = glob.MustCompile(`*.kunlun*.com`)

type Input struct {
	Endpoints              []string `toml:"endpoints"`
	SessionReplayEndpoints []string `toml:"session_replay_endpoints"`
	Measurements           []string `toml:"measurements"`
	measurementMap         map[string]struct{}
	JavaHome               string                       `toml:"java_home"`
	AndroidCmdLineHome     string                       `toml:"android_cmdline_home"`
	ProguardHome           string                       `toml:"proguard_home"`
	NDKHome                string                       `toml:"ndk_home"`
	AtosBinPath            string                       `toml:"atos_bin_path"`
	WPConfig               *workerpool.WorkerPoolConfig `toml:"threads"`
	LocalCacheConfig       *storage.StorageConfig       `toml:"storage"`
	CDNMap                 string                       `toml:"cdn_map"`
	feeder                 dkio.Feeder
	rumDataDir             string
	SessionReplayCfg       *SessionReplayCfg `toml:"session_replay"`
	replayUploadAPI        string
	replayHTTPClient       *http.Client
	replayDiskQueue        *diskcache.DiskCache
}

type SessionReplayCfg struct {
	CachePath         string        `toml:"cache_path"`
	CacheCapacity     int64         `toml:"cache_capacity_mb"`
	ClearCacheOnStart bool          `toml:"clear_cache_on_start"`
	UploadWorkers     int           `toml:"upload_workers"`
	SendTimeout       time.Duration `toml:"send_timeout"`
	SendRetryCount    int           `toml:"send_retry_count"`
}

func defaultSessionReplayCfg() *SessionReplayCfg {
	cfg := &SessionReplayCfg{
		CachePath:         filepath.Join(datakit.CacheDir, "session_replay"),
		CacheCapacity:     defaultReplayCacheMaxMib,
		ClearCacheOnStart: false,
		UploadWorkers:     16,
		SendTimeout:       time.Second * 75,
		SendRetryCount:    3,
	}
	return cfg
}

type CDN struct {
	Domain  string `json:"domain"`
	Name    string `json:"name"`
	Website string `json:"website"`
}

type CDNPool struct {
	literal map[string]CDN
	glob    map[*glob.Glob]CDN
}

var errLimitReader = errors.New("limit reader err")

type limitReader struct {
	r io.ReadCloser
}

func newLimitReader(r io.ReadCloser, max int64) io.ReadCloser {
	return &limitReader{
		r: http.MaxBytesReader(nil, r, max),
	}
}

func (l *limitReader) Read(p []byte) (int, error) {
	n, err := l.r.Read(p)
	if err != nil {
		if err == io.EOF { //nolint:errorlint
			return n, err
		}
		// wrap the errLimitReader
		return n, fmt.Errorf("%w: %s", errLimitReader, err)
	}
	return n, nil
}

func (l *limitReader) Close() error {
	return l.r.Close()
}

func (*Input) Catalog() string { return inputName }

func (*Input) AvailableArchs() []string { return datakit.AllOS }

func (*Input) SampleConfig() string { return sampleConfig }

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{&trace.TraceMeasurement{Name: inputName}}
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
			replayDroppedPointCount.WithLabelValues(appID, env, version, service, statusCode).Inc()
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

	if err := req.ParseMultipartForm(ReplayBodyMaxSize); err != nil {
		return fmt.Errorf("unable to parse multipart form from session replay request: %w", err)
	}

	globalTags := config.Cfg.Dataway.GlobalTags()
	customTagKeys := config.Cfg.Dataway.CustomTagKeys()

	tags := map[string]string{
		"category": "session_replay",
	}

	for k, v := range req.MultipartForm.Value {
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

	headerValue := dataway.HTTPHeaderGlobalTagValue(filter.NewTFDataFromMap(tags), globalTags, customTagKeys)
	if headerValue == "" {
		headerValue = config.Cfg.Dataway.GlobalTagsHTTPHeaderValue()
	}
	req.Header.Set(dataway.HeaderXGlobalTags, headerValue)

	startTime := time.Now()
	defer func() {
		replayUploadingDurationSummary.WithLabelValues(appID, env, version, service, statusCode).Observe(time.Since(startTime).Seconds())
	}()

	for i := 0; i < ipt.SessionReplayCfg.SendRetryCount; i++ {
		if lastErr = func() error {
			req.Body = io.NopCloser(bytes.NewReader(reqPB.Body))

			resp, err = ipt.replayHTTPClient.Do(req)
			if err != nil {
				return fmt.Errorf("at #%d try: unable to send session replay data to dataway: %w", i+1, err)
			}
			defer resp.Body.Close() // nolint:errcheck

			statusCode = strconv.Itoa(resp.StatusCode)

			errMsg := []byte(nil)
			if resp.StatusCode/100 != 2 {
				errMsg, _ = io.ReadAll(resp.Body)
			}

			switch resp.StatusCode / 100 {
			case 5:
				return fmt.Errorf("at #%d try: unable to send session replay data to dataway, http Status: %s, response: %s",
					i+1, resp.Status, string(errMsg))
			case 2:
				// ignore
			default:
				log.Errorf("at #%d try: unable to send session replay data to dataway, http status: %s, response: %s",
					i+1, resp.Status, string(errMsg))
			}

			return nil
		}(); lastErr == nil {
			return nil
		}
	}
	return lastErr
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

func (ipt *Input) initDiskQueue() error {
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

func (ipt *Input) sessionReplayHandler() (f http.HandlerFunc, err error) {
	if err := ipt.initReplayHTTPClient(); err != nil {
		return nil, fmt.Errorf("unable to init session replay http client: %w", err)
	}

	if err := ipt.initDiskQueue(); err != nil {
		return nil, fmt.Errorf("unable to init diskqueue: %w", err)
	}

	if err := ipt.initSessionReplayWorkers(); err != nil {
		return nil, fmt.Errorf("unable to start session replay uploading workers: %w", err)
	}

	return func(w http.ResponseWriter, req *http.Request) {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			log.Errorf("unable to read request body: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if len(body) > ReplayBodyMaxSize {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = io.WriteString(w, fmt.Sprintf("request body size [%d] exceeds the limit", req.ContentLength))
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

		reqPB := &RequestPB{
			Header: headers,
			Body:   body,
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

func (ipt *Input) RegHTTPHandler() {
	log = logger.SLogger(inputName)

	var err error
	if ipt.WPConfig != nil {
		if wkpool, err = workerpool.NewWorkerPool(ipt.WPConfig, log); err != nil {
			log.Errorf("### new worker-pool failed: %s", err.Error())
		} else if err = wkpool.Start(); err != nil {
			log.Errorf("### start worker-pool failed: %s", err.Error())
		}
	}
	if ipt.LocalCacheConfig != nil {
		if localCache, err = storage.NewStorage(ipt.LocalCacheConfig, log); err != nil {
			log.Errorf("### new local-cache failed: %s", err.Error())
		} else {
			localCache.RegisterConsumer(storage.HTTP_KEY, func(buf []byte) error {
				start := time.Now()
				reqpb := &storage.Request{}
				if err := proto.Unmarshal(buf, reqpb); err != nil {
					return err
				} else {
					req := &http.Request{
						Method:           reqpb.Method,
						Proto:            reqpb.Proto,
						ProtoMajor:       int(reqpb.ProtoMajor),
						ProtoMinor:       int(reqpb.ProtoMinor),
						Header:           storage.ConvertMapEntriesToMap(reqpb.Header),
						Body:             io.NopCloser(bytes.NewBuffer(reqpb.Body)),
						ContentLength:    reqpb.ContentLength,
						TransferEncoding: reqpb.TransferEncoding,
						Close:            reqpb.Close,
						Host:             reqpb.Host,
						Form:             storage.ConvertMapEntriesToMap(reqpb.Form),
						PostForm:         storage.ConvertMapEntriesToMap(reqpb.PostForm),
						RemoteAddr:       reqpb.RemoteAddr,
						RequestURI:       reqpb.RequestUri,
					}
					if req.URL, err = url.Parse(reqpb.Url); err != nil {
						log.Errorf("### parse raw URL: %s failed: %s", reqpb.Url, err.Error())
					}
					ipt.handleRUM(&httpapi.NopResponseWriter{}, req)

					log.Debugf("### process status: buffer-size: %dkb, cost: %dms, err: %v", len(reqpb.Body)>>10, time.Since(start)/time.Millisecond, err)

					return nil
				}
			})
			if err = localCache.RunConsumeWorker(); err != nil {
				log.Errorf("### run local-cache consumer failed: %s", err.Error())
			}
		}
	}

	for _, endpoint := range ipt.Endpoints {
		httpapi.RegHTTPHandler(http.MethodPost, endpoint,
			workerpool.HTTPWrapper(httpStatusRespFunc, wkpool,
				httpapi.HTTPStorageWrapper(storage.HTTP_KEY, httpStatusRespFunc, localCache, ipt.handleRUM)))

		log.Infof("### register RUM endpoint: %s", endpoint)
	}

	handler, err := ipt.sessionReplayHandler()
	if err != nil {
		log.Errorf("register rum replay upload proxy fail: %s", err)
	} else {
		for _, endpoint := range ipt.SessionReplayEndpoints {
			httpapi.RegHTTPHandler(http.MethodPost, endpoint, handler)
			log.Infof("register RUM replay upload endpoint: %s", endpoint)
		}
	}

	// add handler for sourcemap related api
	httpapi.RegHTTPRoute(http.MethodGet, "/v1/sourcemap/check", ipt.handleSourcemapCheck)
	httpapi.RegHTTPRoute(http.MethodPut, "/v1/sourcemap", ipt.handleSourcemapUpload)
	httpapi.RegHTTPRoute(http.MethodDelete, "/v1/sourcemap", ipt.handleSourcemapDelete)
}

func (ipt *Input) loadCDNListConf() error {
	var cdnVector []CDN
	if err := json.Unmarshal([]byte(ipt.CDNMap), &cdnVector); err != nil {
		return fmt.Errorf("json unmarshal cdn_map config fail: %w", err)
	}

	if len(cdnVector) == 0 {
		return fmt.Errorf("cdn_map resolved length is 0")
	}

	literalCDNMap := make(map[string]CDN, len(cdnVector))
	globCDNMap := make(map[*glob.Glob]CDN, 0)
	for _, cdn := range cdnVector {
		cdn.Domain = strings.TrimSpace(cdn.Domain)
		if cdn.Domain == "" {
			continue
		}
		if strings.ContainsRune(cdn.Domain, '*') {
			domain := cdn.Domain
			// Prepend prefix `*.` to domain, if the domain is `kunlun*.com`, the result will be `*.kunlun*.com`
			if domain[0] != '*' {
				if domain[0] == '.' {
					domain = "*" + domain
				} else {
					domain = "*." + domain
				}
			}
			g, err := glob.Compile(domain)
			if err == nil {
				globCDNMap[&g] = cdn
				continue
			}
		}
		literalCDNMap[strings.ToLower(cdn.Domain)] = cdn
	}
	CDNList.literal = literalCDNMap
	CDNList.glob = globCDNMap
	return nil
}

func (ipt *Input) initMeasurementMap() {
	if ipt.measurementMap == nil {
		ipt.measurementMap = make(map[string]struct{}, len(ipt.Measurements))
		for _, measure := range ipt.Measurements {
			ipt.measurementMap[measure] = struct{}{}
		}
	}
}

func (ipt *Input) Run() {
	log.Infof("### RUM agent serving on: %+#v", ipt.Endpoints)

	for _, m := range []prometheus.Collector{
		ClientRealIPCounter,
		sourceMapCount,
		loadedZipGauge,
		sourceMapDurationSummary,
		replayUploadingDurationSummary,
		replayDroppedPointCount,
	} {
		if err := metrics.Register(m); err != nil {
			log.Warnf("regist metrics failed: %s, ignored", err)
		}
	}

	ipt.initMeasurementMap()
	log.Infof("captured measurements are: %s", strings.Join(ipt.Measurements, ","))

	if err := ipt.extractArchives(true); err != nil {
		log.Errorf("init extract zip archives encounter err: %s", err)
	}

	if err := ipt.loadSourcemapFile(); err != nil {
		log.Warnf("load source map file failed: %s", err.Error())
	}

	group := goroutine.NewGroup(goroutine.Option{
		Name: "rum",
		PanicCb: func(b []byte) bool {
			log.Error(string(b))
			return false
		},
		PanicTimes: 3,
	})
	group.Go(func(ctx context.Context) error {
		tick := time.NewTicker(time.Minute * 3)
		defer tick.Stop()
		for {
			select {
			case <-datakit.Exit.Wait():
				return nil
			case <-tick.C:
				if err := ipt.extractArchives(false); err != nil {
					log.Errorf("extract zip archives encounter err: %s", err)
				}

				sourceMapDirs := ipt.getWebSourceMapDirs()

				var webSourcemapCacheFile map[string]struct{}
				func() {
					webSourcemapLock.RLock()
					defer webSourcemapLock.RUnlock()
					webSourcemapCacheFile = make(map[string]struct{}, len(webSourcemapCache))
					for file := range webSourcemapCache {
						webSourcemapCacheFile[file] = struct{}{}
					}
				}()

				func() {
					for webDir := range sourceMapDirs {
						archives, err := scanArchives(webDir)
						if err != nil {
							log.Warnf("unable to find zip archive in dir [%s]: %s", webDir, err)
							return
						}

						for _, archive := range archives {
							delete(webSourcemapCacheFile, filepath.Base(archive.Filepath))
						}
					}

					// delete removed zip archive from cache
					if len(webSourcemapCacheFile) > 0 {
						removedFiles := make([]string, 0, len(webSourcemapCacheFile))
						for file := range webSourcemapCacheFile {
							removedFiles = append(removedFiles, file)
						}
						deleteSourcemapCache(removedFiles...)
					}
					webSourcemapCacheFile = nil
				}()
			}
		}
	})

	if ipt.CDNMap != "" {
		if err := ipt.loadCDNListConf(); err != nil {
			log.Errorf("load cdn map config err: %s", err)
		}
	}

	<-datakit.Exit.Wait()
	ipt.Terminate()
}

func (ipt *Input) Terminate() {
	if wkpool != nil {
		wkpool.Shutdown()
		log.Debug("### workerpool closed")
	}
	if localCache != nil {
		if err := localCache.Close(); err != nil {
			log.Error(err.Error())
		}
		log.Debug("### local storage closed")
	}
	if ipt.replayDiskQueue != nil {
		if err := ipt.replayDiskQueue.Close(); err != nil {
			log.Errorf("unable to close session replay disk queue: %s", err)
		}
	}
	if replayWorkersGroup != nil {
		if err := replayWorkersGroup.Wait(); err != nil {
			log.Errorf("goroutine [%s] exit abnormal: %s", replayWorkersGroup.Name(), err)
		}
	}
}

func defaultInput() *Input {
	return &Input{
		feeder:     dkio.DefaultFeeder(),
		rumDataDir: datakit.DataRUMDir,
		Measurements: []string{
			View,
			Resource,
			Error,
			LongTask,
			Action,
			Telemetry,
		},
		SessionReplayCfg: defaultSessionReplayCfg(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
