// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package skywalking handle SkyWalking tracing metrics.
package skywalking

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	agentv3 "github.com/GuanceCloud/tracing-protos/skywalking-gen-go/language/agent/v3"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/storage"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/workerpool"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

var (
	_ inputs.InputV2   = &Input{}
	_ inputs.HTTPInput = &Input{}
)

const (
	inputName     = "skywalking"
	jvmMetricName = "skywalking_jvm"
	sampleConfig  = `
[[inputs.skywalking]]
  ## Skywalking HTTP endpoints for tracing, metric, logging and profiling.
  ## NOTE: DO NOT EDIT.
  endpoints = ["/v3/trace", "/v3/metric", "/v3/logging", "/v3/profiling"]

  ## Skywalking GRPC server listening on address.
  address = "127.0.0.1:11800"

  ## plugins is a list contains all the widgets used in program that want to be regarded as service.
  ## every key words list in plugins represents a plugin defined as special tag by skywalking.
  ## the value of the key word will be used to set the service name.
  # plugins = ["db.type"]

  ## ignore_tags will work as a blacklist to prevent tags send to data center.
  ## Every value in this list is a valid string of regular expression.
  # ignore_tags = ["block1", "block2"]

  ## Keep rare tracing resources list switch.
  ## If some resources are rare enough(not presend in 1 hour), those resource will always send
  ## to data center and do not consider samplers and filters.
  # keep_rare_resource = false

  ## Ignore tracing resources map like service:[resources...].
  ## The service name is the full service name in current application.
  ## The resource list is regular expressions uses to block resource names.
  ## If you want to block some resources universally under all services, you can set the
  ## service name as "*". Note: double quotes "" cannot be omitted.
  # [inputs.skywalking.close_resource]
    # service1 = ["resource1", "resource2", ...]
    # service2 = ["resource1", "resource2", ...]
    # "*" = ["close_resource_under_all_services"]
    # ...

  ## Sampler config uses to set global sampling strategy.
  ## sampling_rate used to set global sampling rate.
  # [inputs.skywalking.sampler]
    # sampling_rate = 1.0

  # [inputs.skywalking.tags]
    # key1 = "value1"
    # key2 = "value2"
    # ...

  ## Threads config controls how many goroutines an agent cloud start to handle HTTP request.
  ## buffer is the size of jobs' buffering of worker channel.
  ## threads is the total number fo goroutines at running time.
  # [inputs.skywalking.threads]
    # buffer = 100
    # threads = 8

  ## Storage config a local storage space in hard dirver to cache trace data.
  ## path is the local file path used to cache data.
  ## capacity is total space size(MB) used to store data.
  # [inputs.skywalking.storage]
    # path = "./skywalking_storage"
    # capacity = 5120
`
)

var (
	log                                       = logger.DefaultSLogger(inputName)
	iptGlobal                                 *Input
	v3trace, v3metric, v3logging, v3profiling = "/v3/trace", "/v3/metric", "/v3/logging", "/v3/profiling"
	address                                   = "localhost:11800"
	plugins                                   []string
	afterGatherRun                            itrace.AfterGatherHandler
	ignoreTags                                []*regexp.Regexp
	tags                                      map[string]string
	wkpool                                    *workerpool.WorkerPool
	localCache                                *storage.Storage
	skySvr                                    *grpc.Server
)

type Input struct {
	V2               interface{}                  `toml:"V2"`            // deprecated *skywalkingConfig
	V3               interface{}                  `toml:"V3"`            // deprecated *skywalkingConfig
	Pipelines        map[string]string            `toml:"pipelines"`     // deprecated
	CustomerTags     []string                     `toml:"customer_tags"` // deprecated
	Endpoints        []string                     `toml:"endpoints"`
	Address          string                       `toml:"address"`
	Plugins          []string                     `toml:"plugins"`
	IgnoreTags       []string                     `toml:"ignore_tags"`
	KeepRareResource bool                         `toml:"keep_rare_resource"`
	CloseResource    map[string][]string          `toml:"close_resource"`
	Sampler          *itrace.Sampler              `toml:"sampler"`
	Tags             map[string]string            `toml:"tags"`
	WPConfig         *workerpool.WorkerPoolConfig `toml:"threads"`
	LocalCacheConfig *storage.StorageConfig       `toml:"storage"`

	feeder  dkio.Feeder
	semStop *cliutils.Sem // start stop signal
	Tagger  datakit.GlobalTagger
}

func (*Input) Catalog() string { return inputName }

func (*Input) AvailableArchs() []string { return datakit.AllOS }

func (*Input) SampleConfig() string { return sampleConfig }

func (ipt *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{&MetricMeasurement{}}
}

func (ipt *Input) RegHTTPHandler() {
	log = logger.SLogger(inputName)
	iptGlobal = ipt

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
					handleSkyTraceV3(&httpapi.NopResponseWriter{}, req)

					log.Debugf("### process status: buffer-size: %dkb, cost: %dms, err: %v", len(reqpb.Body)>>10, time.Since(start)/time.Millisecond, err)

					return nil
				}
			})
			localCache.RegisterConsumer(storage.SKY_WALKING_GRPC_KEY, func(buf []byte) error {
				start := time.Now()
				segobj := &agentv3.SegmentObject{}
				if err := proto.Unmarshal(buf, segobj); err != nil {
					return err
				}
				dktrace := parseSegmentObjectV3(segobj)
				if len(dktrace) != 0 && afterGatherRun != nil {
					afterGatherRun.Run(inputName, itrace.DatakitTraces{dktrace})
				}

				log.Debugf("### process status: buffer-size: %dkb, cost: %dms, err: %v", len(buf)>>10, time.Since(start)/time.Millisecond, err)

				return nil
			})
			if err = localCache.RunConsumeWorker(); err != nil {
				log.Errorf("### run local-cache consumer failed: %s", err.Error())
			}
		}
	}

	var afterGather *itrace.AfterGather
	if localCache != nil && localCache.Enabled() {
		afterGather = itrace.NewAfterGather(
			itrace.WithLogger(log),
			itrace.WithRetry(100*time.Millisecond),
			itrace.WithIOBlockingMode(true),
			itrace.WithPointOptions(point.WithExtraTags(ipt.Tagger.HostTags())),
			itrace.WithFeeder(ipt.feeder),
		)
	} else {
		afterGather = itrace.NewAfterGather(itrace.WithLogger(log),
			itrace.WithPointOptions(point.WithExtraTags(ipt.Tagger.HostTags())), itrace.WithFeeder(ipt.feeder))
	}
	afterGatherRun = afterGather

	// add filters: the order of appending filters into AfterGather is important!!!
	// the order of appending represents the order of that filter executes.
	// add close resource filter
	if len(ipt.CloseResource) != 0 {
		closeResource := &itrace.CloseResource{}
		closeResource.UpdateIgnResList(ipt.CloseResource)
		afterGather.AppendFilter(closeResource.Close)
	}
	// add error status penetration
	afterGather.AppendFilter(itrace.PenetrateErrorTracing)
	// add rare resource keeper
	if ipt.KeepRareResource {
		keepRareResource := &itrace.KeepRareResource{}
		keepRareResource.UpdateStatus(ipt.KeepRareResource, time.Hour)
		afterGather.AppendFilter(keepRareResource.Keep)
	}
	// add sampler
	if ipt.Sampler != nil && (ipt.Sampler.SamplingRateGlobal >= 0 && ipt.Sampler.SamplingRateGlobal <= 1) {
		sampler := ipt.Sampler.Init()
		afterGather.AppendFilter(sampler.Sample)
	}

	for _, v := range ipt.Endpoints {
		log.Debugf("### register skywalking http v3: %s", v)
		switch v {
		case v3trace:
			httpapi.RegHTTPHandler(http.MethodPost, v,
				workerpool.HTTPWrapper(httpStatusRespFunc, wkpool,
					httpapi.HTTPStorageWrapper(storage.HTTP_KEY, httpStatusRespFunc, localCache, handleSkyTraceV3)))
		case v3metric:
			httpapi.RegHTTPHandler(http.MethodPost, v, ipt.handleSkyMetricV3)
		case v3logging:
			httpapi.RegHTTPHandler(http.MethodPost, v, handleSkyLoggingV3)
			httpapi.RegHTTPHandler(http.MethodPost, "/v3/logs", handleSkyLoggingV3)
		case v3profiling:
			httpapi.RegHTTPHandler(http.MethodPost, v, handleProfilingV3)
		}
	}
}

func (ipt *Input) Run() {
	for _, v := range ipt.IgnoreTags {
		if rexp, err := regexp.Compile(v); err != nil {
			log.Debug(err.Error())
		} else {
			ignoreTags = append(ignoreTags, rexp)
		}
	}
	tags = ipt.Tags

	// start up grpc v3 routine
	if len(ipt.Address) == 0 {
		ipt.Address = address
	}
	log.Debug("start skywalking grpc v3 server")
	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_skywalking"})
	g.Go(func(ctx context.Context) error {
		runGRPCV3(ipt)

		return nil
	})

	select {
	case <-datakit.Exit.Wait():
		ipt.exit()
		log.Info("skywalking exit")
		return
	case <-ipt.semStop.Wait():
		ipt.exit()
		log.Info("skywalking return")
		return
	}
}

func (ipt *Input) exit() {
	if skySvr != nil {
		skySvr.Stop()
	}
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

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func defaultInput() *Input {
	return &Input{
		feeder:  dkio.DefaultFeeder(),
		semStop: cliutils.NewSem(),
		Tagger:  datakit.DefaultGlobalTagger(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
