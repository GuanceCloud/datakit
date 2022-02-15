// Package opentelemetry is input for opentelemetry

package opentelemetry

/*
	接收从 opentelemetry 发送的 L/T/M 三种数据
		仅支持两种协议方式发送
			HTTP:使用 protobuf 格式发送 Trace/metric/logging
			grpc:同样使用 protobuf 格式

	接收到的数据交给trace处理。
	本模块只做数据接收和组装 不做业务处理，并都是在(接收完成、返回客户端statusOK) 之后 再进行组装。

	参考开源项目 opentelemetry exports 模块， github地址：https://github.com/open-telemetry/opentelemetry-go

	接收到原生trace 组装成dktrace对象后存储 每隔5秒 或者长度超过100条之后 发送到IO
*/

import (
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	dkHTTP "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/trace"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName    = "opentelemetry"
	sampleConfig = `
[[inputs.opentelemetry]]
  ## customer_tags is a list of keys contains keys set by client code like span.SetTag(key, value)
  ## that want to send to data center. These keys will take precedence over keys in 
  # customer_tags = ["key1", "key2", ...]

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

  [inputs.opentelemetry.http.expectedHeaders]
    ## 如有header配置 则请求中必须要携带 否则返回状态码500
	## 可作为安全检测使用
	# EX_VERSION = xxx
	# EX_NAME = xxx
	# ...
`
)

var (
	l             = logger.DefaultSLogger("otel")
	closeResource *itrace.CloseResource
	afterGather   = itrace.NewAfterGather()
	defSampler    *itrace.Sampler
	storage       = NewSpansStorage()
	maxSend       = 100
	interval      = 10
)

type Input struct {
	Ogc           *otlpGrpcCollector  `toml:"grpc"`
	Otc           *otlpHTTPCollector  `toml:"http"`
	CloseResource map[string][]string `toml:"close_resource"`
	Sampler       *itrace.Sampler     `toml:"sampler"`
	CustomerTags  []string            `toml:"customer_tags"`
	Tags          map[string]string   `toml:"tags"`
	inputName     string
	semStop       *cliutils.Sem // start stop signal
}

func (i *Input) Catalog() string {
	return inputName
}

func (i *Input) SampleConfig() string {
	return sampleConfig
}

func (i *Input) RegHTTPHandler() {
	dkHTTP.RegHTTPHandler("POST", "/otel/v11/trace", i.Otc.apiOtlpTrace)
	dkHTTP.RegHTTPHandler("GET", "/otel/v11/trace", i.Otc.apiOtlpTrace)
}

func (i *Input) exit() {
	i.Ogc.stop()
}

func (i *Input) Run() {
	l = logger.SLogger("otlp")
	// add filters: the order append in AfterGather is important!!!
	// add close resource filter
	if len(i.CloseResource) != 0 {
		closeResource = &itrace.CloseResource{}
		closeResource.UpdateIgnResList(i.CloseResource)
		afterGather.AppendFilter(closeResource.Close)
	}
	// add sampler
	if i.Sampler != nil {
		defSampler = i.Sampler
		afterGather.AppendFilter(defSampler.Sample)
	}
	open := false
	// 从配置文件 开启
	if i.Otc.Enable {
		open = true
		go i.Otc.RunHTTP()
	}
	if i.Ogc.TraceEnable || i.Ogc.MetricEnable {
		go i.Ogc.run()
	}
	if open {
		// add calculators
		afterGather.AppendCalculator(itrace.StatTracingInfo)
		go storage.run()
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

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			inputName: inputName,
			semStop:   cliutils.NewSem(),
		}
	})
}
