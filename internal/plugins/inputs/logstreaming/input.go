// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package logstreaming handle remote logging data.
package logstreaming

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"google.golang.org/protobuf/proto"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/storage"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/workerpool"
)

var (
	_ inputs.InputV2   = &Input{}
	_ inputs.HTTPInput = &Input{}
	_ inputs.Singleton = (*Input)(nil)
)

const (
	inputName = "logstreaming"
	sampleCfg = `
[inputs.logstreaming]
  ignore_url_tags = false

  ## Threads config controls how many goroutines an agent cloud start to handle HTTP request.
  ## buffer is the size of jobs' buffering of worker channel.
  ## threads is the total number fo goroutines at running time.
  # [inputs.logstreaming.threads]
  #   buffer = 100
  #   threads = 8

  ## Storage config a local storage space in hard dirver to cache trace data.
  ## path is the local file path used to cache data.
  ## capacity is total space size(MB) used to store data.
  # [inputs.logstreaming.storage]
  #   path = "./log_storage"
  #   capacity = 5120
`
)

var (
	log        = logger.DefaultSLogger(inputName)
	wkpool     *workerpool.WorkerPool
	localCache *storage.Storage
)

type Input struct {
	IgnoreURLTags    bool                         `toml:"ignore_url_tags"`
	WPConfig         *workerpool.WorkerPoolConfig `toml:"threads"`
	LocalCacheConfig *storage.StorageConfig       `toml:"storage"`

	ScanBufferSize int `toml:"scan_buffer_size"`
	scanbufPool    *sync.Pool

	feeder  dkio.Feeder
	semStop *cliutils.Sem // start stop signal
}

func (ipt *Input) getScanbuf() []byte {
	b := ipt.scanbufPool.Get()
	if b == nil {
		b = make([]byte, ipt.ScanBufferSize)
	}

	return b.([]byte)
}

func (ipt *Input) putScanbuf(buf []byte) {
	ipt.scanbufPool.Put(buf) //nolint: staticcheck
}

func (*Input) Catalog() string { return "log" }

func (*Input) AvailableArchs() []string { return datakit.AllOS }

func (*Input) SampleConfig() string { return sampleCfg }

func (*Input) Singleton() {}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{&logstreamingMeasurement{}}
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
					ipt.handleLogstreaming(&httpapi.NopResponseWriter{}, req)

					log.Debugf("### process status: buffer-size: %dkb, cost: %dms, err: %v", len(reqpb.Body)>>10, time.Since(start)/time.Millisecond, err)

					return nil
				}
			})
			if err = localCache.RunConsumeWorker(); err != nil {
				log.Errorf("### run local-cache consumer failed: %s", err.Error())
			}
		}
	}

	httpapi.RegHTTPHandler("POST", "/v1/write/logstreaming",
		workerpool.HTTPWrapper(httpStatusRespFunc, wkpool,
			httpapi.HTTPStorageWrapper(storage.HTTP_KEY,
				httpStatusRespFunc,
				localCache,
				httpapi.ProtectedHandlerFunc(ipt.handleLogstreaming, log))))
}

func (ipt *Input) Run() {
	log.Info("### register logstreaming router")

	select {
	case <-datakit.Exit.Wait():
		ipt.exit()
		log.Info(inputName + " exit")
		return
	case <-ipt.semStop.Wait():
		ipt.exit()
		log.Info(inputName + " return")
		return
	}
}

func (*Input) exit() {
	if wkpool != nil {
		wkpool.Shutdown()
		log.Debug("### workerpool closed")
	}
	if localCache != nil {
		if err := localCache.Close(); err != nil {
			log.Error(err.Error())
		}
		log.Debug("### storage closed")
	}
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func defaultInput() *Input {
	return &Input{
		feeder:         dkio.DefaultFeeder(),
		semStop:        cliutils.NewSem(),
		ScanBufferSize: 4 * 1024,
		scanbufPool:    &sync.Pool{},
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
