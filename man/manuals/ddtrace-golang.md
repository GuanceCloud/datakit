{{.CSS}}

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
- 操作系统支持：全平台

# Tracing Golang Application

## Install Libarary & Dependence

安装 ddtrace golang library 在开发目录下运行

```shell
go get -v github.com/DataDog/dd-trace-go
```

## Golang Code Example

**Start Global Tracer**

```golang
func main_func(){
	// Start the global tracer in main function.
	startOpts := []tracer.StartOption{
		tracer.WithAgentAddr("localhost:9529"),
		tracer.WithService("go-application"),
		tracer.WithServiceVersion("v1.0.0"),
		tracer.WithLogStartup(true),
		tracer.WithDebugMode(true),
		tracer.WithEnv("name=test,stage=dev"),
	}
	tracer.Start(startOpts...)
	defer tracer.Stop()
}
```

**Tracing API**

```golang
// The below example illustrates a simple use case using the "tracer" package,
// our native Datadog APM tracing client integration. For thorough documentation
// and further examples, visit its own godoc page.
func example_for_golang_tracing() {
	var err error
	// Start a root span.
	span := tracer.StartSpan("get.data")
	defer span.Finish(tracer.WithError(err))

	// Create a child of it, computing the time needed to read a file.
	child := tracer.StartSpan("read.file", tracer.ChildOf(span.Context()))
	child.SetTag(ext.ResourceName, "test.json")

	// Perform an operation.
	var bts []byte
	bts, err = ioutil.ReadFile("~/test.json")
	span.SetTag("file_len", len(bts))
}
```

## Run Golang Code With DDTrace

```shell
go run your-go-code.go
```

## Start Options For Tracing Golang Code

- WithEnv: 为服务设置环境变量，对应环境变量 DD_ENV。
- WithServiceVersion: APP 版本号，对应环境变量 DD_VERSION。
- WithService: 用于设置应用程序的服务名称，对应环境变量 DD_SERVICE。
- WithLogStartup: 开启启动配置和诊断日志，对应环境变量 DD_TRACE_STARTUP_LOGS。
- WithDebugMode: 开启 debug 日志，对应环境变量 DD_TRACE_DEBUG。
- WithAgentAddr: Datakit 监听的地址和端口号，默认 localhost:9529。
