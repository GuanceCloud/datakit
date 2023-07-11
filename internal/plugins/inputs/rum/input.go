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
	"net/http/httputil"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/gobwas/glob"
	"google.golang.org/protobuf/proto"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/storage"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/workerpool"
)

var (
	_ inputs.HTTPInput = &Input{}
	_ inputs.InputV2   = &Input{}
)

const (
	inputName         = "rum"
	ReplayFileMaxSize = 1 << 22 // 1024 * 1024 * 4  4Mib
	ProxyErrorHeader  = "X-Proxy-Error"
	// nolint: lll
	sampleConfig = `
[[inputs.rum]]
  ## profile Agent endpoints register by version respectively.
  ## Endpoints can be skipped listen by remove them from the list.
  ## Default value set as below. DO NOT MODIFY THESE ENDPOINTS if not necessary.
  endpoints = ["/v1/write/rum"]

  ## use to upload rum screenshot,html,etc...
  session_replay_endpoints = ["/v1/write/rum/replay"]

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

  ## Threads config controls how many goroutines an agent cloud start to handle HTTP request.
  ## buffer is the size of jobs' buffering of worker channel.
  ## threads is the total number fo goroutines at running time.
  # [inputs.rum.threads]
    # buffer = 100
    # threads = 8

  ## Storage config a local storage space in hard dirver to cache trace data.
  ## path is the local file path used to cache data.
  ## capacity is total space size(MB) used to store data.
  # [inputs.rum.storage]
    # path = "./rum_storage"
    # capacity = 5120

  # Provide a list to resolve CDN of your static resource.
  # Below is the Datakit default built-in CDN list, you can uncomment that and change it to your cdn list,
  # it's a JSON array like: [{"domain": "CDN domain", "name": "CDN human readable name", "website": "CDN official website"},...],
  # domain field value can contains '*' as wildcard, for example: "kunlun*.com",
  # it will match "kunluna.com", "kunlunab.com" and "kunlunabc.com" but not "kunlunab.c.com".
  # cdn_map = '[{"domain":"15cdn.com","name":"腾正安全加速(原 15CDN)","website":"https://www.15cdn.com"},{"domain":"tzcdn.cn","name":"腾正安全加速(原 15CDN)","website":"https://www.15cdn.com"},...]'
`
)

const (
	Session  = "session"
	View     = "view"
	Resource = "resource"
	Action   = "action"
	LongTask = "long task"
	Error    = "error"
)

var (
	log        = logger.DefaultSLogger(inputName)
	wkpool     *workerpool.WorkerPool
	localCache *storage.Storage
)

var kunlunCDNGlob = glob.MustCompile(`*.kunlun*.com`)

type Input struct {
	Endpoints              []string                     `toml:"endpoints"`
	SessionReplayEndpoints []string                     `toml:"session_replay_endpoints"`
	JavaHome               string                       `toml:"java_home"`
	AndroidCmdLineHome     string                       `toml:"android_cmdline_home"`
	ProguardHome           string                       `toml:"proguard_home"`
	NDKHome                string                       `toml:"ndk_home"`
	AtosBinPath            string                       `toml:"atos_bin_path"`
	WPConfig               *workerpool.WorkerPoolConfig `toml:"threads"`
	LocalCacheConfig       *storage.StorageConfig       `toml:"storage"`
	CDNMap                 string                       `toml:"cdn_map"`
	feeder                 dkio.Feeder
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

func replayUploadHandler() (*httputil.ReverseProxy, error) {
	endpoints := config.Cfg.Dataway.GetAvailableEndpoints()

	if len(endpoints) == 0 {
		return nil, fmt.Errorf("no available dataway endpoint now")
	}

	var (
		validURL *url.URL
		lastErr  error
	)
	for _, ep := range endpoints {
		replayURL := ep.GetCategoryURL()[datakit.SessionReplayUpload]
		if replayURL == "" {
			lastErr = fmt.Errorf("empty category url")
			continue
		}
		parsedURL, err := url.Parse(replayURL)
		if err == nil {
			validURL = parsedURL
			break
		}
		lastErr = err
	}

	if validURL == nil {
		if lastErr != nil {
			return nil, lastErr
		}
		return nil, fmt.Errorf("no available dataway endpoint")
	}

	return &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			if req.ContentLength > ReplayFileMaxSize {
				req.URL = nil // this will trigger a proxy err, and let request complete earlier
				req.Header.Set(ProxyErrorHeader, fmt.Sprintf("request body size [%d] exceeds the limit", req.ContentLength))
				return
			}

			req.URL = validURL
			req.Host = validURL.Host
			if req.Body != nil {
				req.Body = newLimitReader(req.Body, ReplayFileMaxSize)
			}
		},

		ErrorHandler: func(w http.ResponseWriter, req *http.Request, err error) {
			proxyErr := req.Header.Get(ProxyErrorHeader)
			if proxyErr != "" {
				log.Errorf("proxy error: %s", proxyErr)
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(proxyErr))
				return
			}
			if errors.Is(err, errLimitReader) {
				log.Errorf("request body is too large: %s", err)
				w.WriteHeader(http.StatusBadRequest)
			} else {
				log.Errorf("other rum replay err: %s", err)
				w.WriteHeader(http.StatusBadGateway)
			}
			_, _ = w.Write([]byte(err.Error()))
		},
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

	proxy, err := replayUploadHandler()
	if err != nil {
		log.Errorf("register rum replay upload proxy fail: %s", err)
	} else {
		for _, endpoint := range ipt.SessionReplayEndpoints {
			httpapi.RegHTTPHandler(http.MethodPost, endpoint, proxy.ServeHTTP)
			log.Infof("register RUM replay upload endpoint: %s", endpoint)
		}
	}
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

func (ipt *Input) Run() {
	log.Infof("### RUM agent serving on: %+#v", ipt.Endpoints)

	metrics.MustRegister(ClientRealIPCounter, sourceMapCount, loadedZipGauge, sourceMapDurationSummary)

	if err := extractArchives(true); err != nil {
		log.Errorf("init extract zip archives encounter err: %s", err)
	}

	if err := loadSourcemapFile(); err != nil {
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
				if err := extractArchives(false); err != nil {
					log.Errorf("extract zip archives encounter err: %s", err)
				}

				sourceMapDirs := getWebSourceMapDirs()

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

func (*Input) Terminate() {
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
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			feeder: dkio.DefaultFeeder(),
		}
	})
}
