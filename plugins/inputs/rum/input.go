// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package rum real user monitoring
package rum

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
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
	inputName    = "rum"
	sampleConfig = `
[[inputs.rum]]
  ## profile Agent endpoints register by version respectively.
  ## Endpoints can be skipped listen by remove them from the list.
  ## Default value set as below. DO NOT MODIFY THESE ENDPOINTS if not necessary.
  endpoints = ["/v1/write/rum"]

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
	Endpoints          []string                     `toml:"endpoints"`
	JavaHome           string                       `toml:"java_home"`
	AndroidCmdLineHome string                       `toml:"android_cmdline_home"`
	ProguardHome       string                       `toml:"proguard_home"`
	NDKHome            string                       `toml:"ndk_home"`
	AtosBinPath        string                       `toml:"atos_bin_path"`
	WPConfig           *workerpool.WorkerPoolConfig `toml:"threads"`
	LocalCacheConfig   *storage.StorageConfig       `toml:"storage"`
}

func (*Input) Catalog() string { return inputName }

func (*Input) AvailableArchs() []string { return datakit.AllOS }

func (*Input) SampleConfig() string { return sampleConfig }

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{&trace.TraceMeasurement{Name: inputName}}
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
		} else if err = localCache.RunConsumeWorker(); err != nil {
			log.Errorf("### run local-cache consumer failed: %s", err.Error())
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
		}
	}

	for _, endpoint := range ipt.Endpoints {
		dkhttp.RegHTTPHandler(http.MethodPost, endpoint, workerpool.HTTPWrapper(wkpool, storage.HTTPWrapper(storage.HTTP_KEY, localCache, ipt.handleRUM)))
		log.Infof("### register RUM endpoint: %s", endpoint)
	}
}

func (ipt *Input) Run() {
	log.Infof("### RUM agent serving on: %+#v", ipt.Endpoints)

	if err := loadSourcemapFile(); err != nil {
		log.Errorf("load source map file failed: %s", err.Error())
	}
}

func (*Input) Terminate() {
	log.Info("### RUM exits")
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
