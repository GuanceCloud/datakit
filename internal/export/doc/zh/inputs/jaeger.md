---
title     : 'Jaeger'
summary   : '接收 Jaeger APM 数据'
__int_icon      : 'icon/jaeger'
tags      :
  - 'JAEGER'
  - '链路追踪'
dashboard :
  - desc  : '暂无'
    path  : '-'
monitor   :
  - desc  : '暂无'
    path  : '-'
---

{{.AvailableArchs}}

---

DataKit 内嵌的 Jaeger Agent 用于接收，运算，分析 Jaeger Tracing 协议数据。

## 配置 {#config}

<!-- markdownlint-disable MD046 -->
???+ info

    当前 Jaeger 版本支持 HTTP 和 UDP 通信协议和 Apache Thrift 编码规范

=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/samples` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    可通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting) 或 [配置 ENV_DATAKIT_INPUTS](../datakit/datakit-daemonset-deploy.md#env-setting) 开启采集器。

    也支持以环境变量的方式修改配置参数（需要在 ENV_DEFAULT_ENABLED_INPUTS 中加为默认采集器）：

{{ CodeBlock .InputENVSampleZh 4 }}

<!-- markdownlint-enable -->

在使用 UDP 协议的时候，注意协议中的数据格式，默认情况下使用 6831 端口使用的是 `thrift CompactProtocol` 格式，使用 6832 端口时的协议为 `thrift BinaryProtocol` 。
Jaeger 默认情况下使用的是 6831 端口中的协议，所以 当您不使用 6832 端口时，请不要打开注释。

### 配置 Jaeger HTTP Agent {#config-http-agent}

endpoint 代表 Jaeger HTTP Agent 路由

```toml
[[inputs.{{.InputName}}]]
  # Jaeger endpoint for receiving tracing span over HTTP.
  # Default value set as below. DO NOT MODIFY THE ENDPOINT if not necessary.
  endpoint = "/apis/traces"
```

- 修改 Jaeger Client 的 Agent Host Port 为 DataKit Port（默认为 9529）
- 修改 Jaeger Client 的 Agent endpoint 为上面配置中指定的 endpoint

### 配置 Jaeger UDP Agent {#config-udp-agent}

修改 Jaeger Client 的 Agent UDP Host:Port 为下面配置中指定的 address：

```toml
[[inputs.{{.InputName}}]]
  # Jaeger agent host:port address for UDP transport.
  address = "127.0.0.1:6831"
```

有关数据采样，数据过滤，关闭资源等配置请参考[DataKit Tracing](datakit-tracing.md)

## 示例 {#demo}

### Golang 示例 {#go-http}

以下是一个 HTTP Agent 示例：

```golang
package main

import (
  "fmt"
  "io"
  "log"
  "net/http"
  "net/http/httptest"
  "time"

  "github.com/opentracing/opentracing-go"
  "github.com/opentracing/opentracing-go/ext"
  "github.com/uber/jaeger-client-go"
  jaegercfg "github.com/uber/jaeger-client-go/config"
  jaegerlog "github.com/uber/jaeger-client-go/log"
)

var tracer opentracing.Tracer

func main() {
  jgcfg := jaegercfg.Configuration{
    ServiceName: "jaeger_sample_http",
    Sampler: &jaegercfg.SamplerConfig{
      Type:  jaeger.SamplerTypeConst,
      Param: 1,
    },
    Reporter: &jaegercfg.ReporterConfig{
      CollectorEndpoint:   "http://localhost:9529/apis/traces",
      HTTPHeaders:         map[string]string{"Content-Type": "application/x-thrift"},
      BufferFlushInterval: time.Second,
      LogSpans:            true,
    },
  }

  var (
    closer io.Closer
    err    error
  )
  tracer, closer, err = jgcfg.NewTracer(jaegercfg.Logger(jaegerlog.StdLogger))
  defer func() {
    if err := closer.Close(); err != nil {
      log.Println(err.Error())
    }
  }()
  if err != nil {
    log.Panicln(err.Error())
  }

  srv := httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
    spctx, err := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
    var span opentracing.Span
    if err != nil {
      log.Println(err.Error())
      span = tracer.StartSpan(req.RequestURI)
    } else {
      span = tracer.StartSpan(req.RequestURI, ext.RPCServerOption(spctx))
    }
    defer span.Finish()

    span.SetTag("finish_ts", time.Now())

    resp.Write([]byte("hello, world"))
  }))

  for i := 0; i < 100; i++ {
    send(srv.URL, i)

    time.Sleep(time.Second)
  }
}

func send(urlstr string, i int) {
  span := tracer.StartSpan(fmt.Sprintf("main_loop->send(%d)", i))
  defer span.Finish()

  req, err := http.NewRequest(http.MethodGet, urlstr, nil)
  if err != nil {
    log.Println(err.Error())

    return
  }

  if err = tracer.Inject(span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header)); err != nil {
    log.Panicln(err.Error())

    return
  }

  span.SetTag(fmt.Sprintf("send_%d_finish", i), time.Now())
}
```

### Golang UDP 示例 {#go-udp}

以下是一个 UDP Agent 示例：

```golang
package main

import (
  "io"
  "log"
  "time"

  "github.com/opentracing/opentracing-go"
  "github.com/uber/jaeger-client-go"
  jaegercfg "github.com/uber/jaeger-client-go/config"
  jaegerlog "github.com/uber/jaeger-client-go/log"
)

var tracer opentracing.Tracer

func main() {
  jgcfg := jaegercfg.Configuration{
    ServiceName: "jaeger_sample_app",
    Sampler: &jaegercfg.SamplerConfig{
      Type:  jaeger.SamplerTypeConst,
      Param: 1,
    },
    Reporter: &jaegercfg.ReporterConfig{
      LocalAgentHostPort:  "127.0.0.1:6831",
      BufferFlushInterval: time.Second,
      LogSpans:            true,
    },
  }

  var (
    closer io.Closer
    err    error
  )
  tracer, closer, err = jgcfg.NewTracer(jaegercfg.Logger(jaegerlog.StdLogger))
  defer func() {
    if err := closer.Close(); err != nil {
      log.Println(err.Error())
    }
  }()
  if err != nil {
    log.Panicln(err.Error())
  }

  for i := 0; i < 10; i++ {
    foo()

    time.Sleep(time.Second)
  }
}

func foo() {
  span := tracer.StartSpan("foo")
  defer span.Finish()

  span.SetTag("finish_ts", time.Now())
}
```

## 指标 {#metric}

{{range $i, $m := .Measurements}}

{{if eq $m.Type "tracing"}}

### `{{$m.Name}}`

{{$m.Desc}}

{{$m.MarkdownTable}}
{{end}}

{{ end }}

## Jaeger 官方文档 {#doc}

- [Quick Start](https://www.jaegertracing.io/docs/1.27/getting-started/){:target="_blank"}
- [Docs](https://www.jaegertracing.io/docs/){:target="_blank"}
- [Clients Download](https://www.jaegertracing.io/download/){:target="_blank"}
- [Source Code](https://github.com/jaegertracing/jaeger){:target="_blank"}
