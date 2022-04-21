// Package opentelemetry is input for opentelemetry

package opentelemetry

import (
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	dkHTTP "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/trace"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/opentelemetry/collector"
)

var (
	_ inputs.InputV2   = &Input{}
	_ inputs.HTTPInput = &Input{}
)

const (
	inputName    = "opentelemetry"
	sampleConfig = `
[[inputs.opentelemetry]]
  ## 在创建'trace',Span','resource'时，会加入很多标签，这些标签最终都会出现在'Span'中
  ## 当您不希望这些标签太多造成网络上不必要的流量损失时，可选择忽略掉这些标签
  ## 支持正则表达，注意:将所有的'.'替换成'_'
  ## When creating 'trace', 'span' and 'resource', many labels will be added, and these labels will eventually appear in all 'spans'
  ## When you don't want too many labels to cause unnecessary traffic loss on the network, you can choose to ignore these labels
  ## Support regular expression. Note!!!: all '.' Replace with '_'
  # ignore_attribute_keys = ["os_*","process_*"]

  ## Keep rare tracing resources list switch.
  ## If some resources are rare enough(not presend in 1 hour), those resource will always send
  ## to data center and do not consider samplers and filters.
  # keep_rare_resource = false

  ## Ignore tracing resources map like service:[resources...].
  ## The service name is the full service name in current application.
  ## The resource list is regular expressions uses to block resource names.
  # [inputs.opentelemetry.close_resource]
    # service1 = ["resource1", "resource2", ...]
    # service2 = ["resource1", "resource2", ...]
    # ...

  ## Sampler config uses to set global sampling strategy.
  ## priority uses to set tracing data propagation level, the valid values are -1, 0, 1
  ##   -1: always reject any tracing data send to datakit
  ##    0: accept tracing data and calculate with sampling_rate
  ##    1: always send to data center and do not consider sampling_rate
  ## sampling_rate used to set global sampling rate
  # [inputs.opentelemetry.sampler]
    # priority = 0
    # sampling_rate = 1.0

  # [inputs.opentelemetry.tags]
    # key1 = "value1"
    # key2 = "value2"
    # ...

  [inputs.opentelemetry.expectedHeaders]
    ## 如有header配置 则请求中必须要携带 否则返回状态码500
	## 可作为安全检测使用,必须全部小写
	# ex_version = xxx
	# ex_name = xxx
	# ...

  ## grpc
  [inputs.opentelemetry.grpc]
  ## trace for grpc
  trace_enable = false

  ## metric for grpc
  metric_enable = false

  ## grpc listen addr
  addr = "127.0.0.1:9550"

  ## http
  [inputs.opentelemetry.http]
  ## if enable=true
  ## http path (do not edit):
  ##	trace : /otel/v1/trace
  ##	metric: /otel/v1/metric
  ## use as : http://127.0.0.1:9529/otel/v11/trace . Method = POST
  enable = false
  ## return to client status_ok_code :200/202
  http_status_ok = 200
`
)

var l = logger.DefaultSLogger("otel-log")

type Input struct {
	Ogrpc               *otlpGrpcCollector  `toml:"grpc"`
	OHTTPc              *otlpHTTPCollector  `toml:"http"`
	CloseResource       map[string][]string `toml:"close_resource"`
	Sampler             *itrace.Sampler     `toml:"sampler"`
	IgnoreAttributeKeys []string            `toml:"ignore_attribute_keys"`
	Tags                map[string]string   `toml:"tags"`
	ExpectedHeaders     map[string]string   `toml:"expectedHeaders"`

	inputName string
	semStop   *cliutils.Sem // start stop signal
}

func (i *Input) Catalog() string {
	return inputName
}

func (*Input) AvailableArchs() []string {
	return datakit.AllArch
}

func (i *Input) SampleConfig() string {
	return sampleConfig
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{&itrace.TraceMeasurement{Name: inputName}}
}

func (i *Input) RegHTTPHandler() {
	dkHTTP.RegHTTPHandler("POST", "/otel/v1/trace", i.OHTTPc.apiOtlpTrace)
	dkHTTP.RegHTTPHandler("POST", "/otel/v1/metric", i.OHTTPc.apiOtlpMetric)
}

func (i *Input) exit() {
	i.Ogrpc.stop()
}

func (i *Input) Run() {
	l = logger.SLogger("otlp-log")
	storage := collector.NewSpansStorage()
	// add filters: the order append in AfterGather is important!!!
	// add error status penetration
	storage.AfterGather.AppendFilter(itrace.PenetrateErrorTracing)
	// add close resource filter
	if len(i.CloseResource) != 0 {
		closeResource := &itrace.CloseResource{}
		closeResource.UpdateIgnResList(i.CloseResource)
		storage.AfterGather.AppendFilter(closeResource.Close)
	}
	// add sampler
	if i.Sampler != nil {
		defSampler := i.Sampler
		storage.AfterGather.AppendFilter(defSampler.Sample)
	}

	storage.GlobalTags = i.Tags

	if len(i.IgnoreAttributeKeys) > 0 {
		storage.RegexpString = strings.Join(i.IgnoreAttributeKeys, "|")
	}

	open := false
	// 从配置文件 开启
	if i.OHTTPc.Enable {
		// add option
		i.OHTTPc.storage = storage
		i.OHTTPc.ExpectedHeaders = i.ExpectedHeaders
		open = true
	}
	if i.Ogrpc.TraceEnable || i.Ogrpc.MetricEnable {
		open = true
		i.Ogrpc.ExpectedHeaders = i.ExpectedHeaders
		go i.Ogrpc.run(storage)
	}
	if open {
		// add calculators
		storage.AfterGather.AppendCalculator(itrace.StatTracingInfo)
		go storage.Run()
		for {
			select {
			case <-datakit.Exit.Wait():
				i.exit()
				l.Infof("%s exit", i.inputName)
				return

			case <-i.semStop.Wait():
				i.exit()
				l.Infof("%s return", i.inputName)
				return
			}
		}
	}
}

func (i *Input) Terminate() {
	// TODO: 必须写
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			inputName: inputName,
			semStop:   cliutils.NewSem(),
		}
	})
}
