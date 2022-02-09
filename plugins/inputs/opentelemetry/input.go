package opentelemetry

/*
	接收opentelemetry发送的 L/T/M 三种数据
		仅支持两种协议方式发送
			HTTP:使用protobuf格式发送 Trace/metric/logging
			grpc:同样使用protobuf格式

	接收到的数据交给trace处理。
	本模块只做数据接收和组装 不做业务处理，并都是在(接收完成、返回客户端statusOK) 之后 再进行组装。

	参考开源项目opentelemetry exports模块， github地址：https://github.com/open-telemetry/opentelemetry-go
*/

import (
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName    = "opentelemetry"
	sampleConfig = `
[[inputs.opentelemetry]]
  ## todo : 语雀文档地址
  ## grpc
  [inputs.opentelemetry.grpc]
  ## trace for grpc
  trace_enable = false

  ## metric for grpc
  metric_enable = false

  ## tcp port
  addr = "127.0.0.1:9550"

  ## http 
  [inputs.opentelemetry.http]
  ## if enable=true  
  ## http path :
  ##	trace : /otel/v11/trace
  ##	metric: /otel/v11/metric
  ## use as : http://127.0.0.1:9529/otel/v11/trace . Method = POST
  enable = false
  [inputs.opentelemetry.http.expectedHeaders]
    ## 如有header配置 则请求中必须要携带 否则返回状态码500
	## 可作为安全检测使用
	# EX_VERSION = xxx
	# EX_NAME = xxx
	# ...
`
)

var (
	l = logger.DefaultSLogger("otel")
)

type Input struct {
	Ogc       *otlpGrpcCollector `toml:"grpc"`
	Otc       *otlpHTTPCollector `toml:"http"`
	inputName string
	semStop   *cliutils.Sem // start stop signal
}

func (i *Input) Catalog() string {
	return inputName
}

func (i *Input) SampleConfig() string {
	return sampleConfig
}

func (i *Input) exit() {
	i.Ogc.stop()
}

func (i *Input) Run() {
	l = logger.SLogger("otlp")
	// 从配置文件 开启
	if i.Otc.Enable {
		go i.Otc.RunHttp()
	}
	if i.Ogc.TraceEnable || i.Ogc.MetricEnable {
		go i.Ogc.run()
	}
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

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			inputName: inputName,
			semStop:   cliutils.NewSem(),
		}
	})
}
