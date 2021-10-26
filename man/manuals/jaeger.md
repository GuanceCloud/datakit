{{.CSS}}

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# {{.InputName}}

## Jaeger 文档

- [getting started](https://www.jaegertracing.io/docs/1.27/getting-started/)
- [source code](https://github.com/jaegertracing/jaeger)
- [client download](https://github.com/jaegertracing/jaeger-client-go/releases)

！！！Golang 客户端 lib 用户注意

```code
import "github.com/uber/jaeger-client-go"
```

## Sample Code for Golang

```code
package main

import (
	"log"
	"time"

	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
)

func main() {
	jgcfg := jaegercfg.Configuration{
		ServiceName: "jaeger_sample_code",
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaegercfg.ReporterConfig{
			BufferFlushInterval: time.Second,
			LocalAgentHostPort:  "127.0.0.1:6831",
			// CollectorEndpoint:   "http://localhost:9529/jaeger/traces",
			LogSpans: true,
			// HTTPHeaders:         map[string]string{"Content-Type": "application/vnd.apache.thrift.binary"},
		},
	}

	tracer, closer, err := jgcfg.NewTracer(jaegercfg.Logger(jaegerlog.StdLogger))
	defer func() {
		if err := closer.Close(); err != nil {
			log.Println(err.Error())
		}
	}()
	if err != nil {
		log.Panicln(err.Error())
	}

	for {
		span := tracer.StartSpan("test_start_span")
		log.Println("start new span")
		span.SetTag("key", "value")
		span.Finish()
		log.Println("new span finished")

		time.Sleep(time.Second)
	}
}
```

## 配置

！！！注意不要修改配置文件中的 endpoint 除非 Jaeger 客户端中的配置也做了对应配置

```toml
endpoint = "/apis/traces"
```

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}}
```
