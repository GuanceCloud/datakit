// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package rum real user monitoring
package rum

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"

	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	dkhttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	ihttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/storage"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/workerpool"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"google.golang.org/protobuf/proto"
)

var (
	_ inputs.HTTPInput = &Input{}
	_ inputs.InputV2   = &Input{}
)

const (
	inputName         = "rum"
	ReplayFileMaxSize = 1 << 22 // 1024 * 1024 * 4  4Mib
	ProxyErrorHeader  = "X-Proxy-Error"
	sampleConfig      = `
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
`
)

var (
	log        = logger.DefaultSLogger(inputName)
	wkpool     *workerpool.WorkerPool
	localCache *storage.Storage
)

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
	endpoints := config.Cfg.DataWay.GetAvailableEndpoints()

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
			req.Body = newLimitReader(req.Body, ReplayFileMaxSize)
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
					ipt.handleRUM(&ihttp.NopResponseWriter{}, req)

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
		dkhttp.RegHTTPHandler(http.MethodPost, endpoint,
			workerpool.HTTPWrapper(httpStatusRespFunc, wkpool,
				storage.HTTPWrapper(storage.HTTP_KEY, httpStatusRespFunc, localCache, ipt.handleRUM)))

		log.Infof("### register RUM endpoint: %s", endpoint)
	}

	proxy, err := replayUploadHandler()
	if err != nil {
		log.Errorf("register rum replay upload proxy fail: %s", err)
	} else {
		for _, endpoint := range ipt.SessionReplayEndpoints {
			dkhttp.RegHTTPHandler(http.MethodPost, endpoint, proxy.ServeHTTP)
			log.Infof("register RUM replay upload endpoint: %s", endpoint)
		}
	}
}

func (ipt *Input) Run() {
	log.Infof("### RUM agent serving on: %+#v", ipt.Endpoints)

	if err := loadSourcemapFile(); err != nil {
		log.Errorf("load source map file failed: %s", err.Error())
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
		return &Input{}
	})
}
