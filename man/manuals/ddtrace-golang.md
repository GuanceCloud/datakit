{{.CSS}}

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
- 操作系统支持：全平台

# 准备工作

1. 在您的开发目录下运行 `go get -v github.com/DataDog/dd-trace-go`
2. 复制`Datakit`安装目录下`conf.d/ddtrace/ddtrace.conf.example`并重命名为`ddtrace.conf`
3. 重启`Datakit`

# Datakit Config

```toml
[[inputs.ddtrace]]
  ## DDTrace Agent endpoints register by version respectively.
  ## Endpoints can be skipped listen by remove them from the list.
  ## Default value set as below. DO NOT MODIFY THESE ENDPOINTS if not necessary.
  endpoints = ["/v0.3/traces", "/v0.4/traces", "/v0.5/traces"]
```

- 打开 Datakit 配置中需要的 endpoints 目前支持的 DDTrace 版本有 v0.3 v0.4 v0.5
- 配置 DDTrace Client 中的 Agent Host:Port 为 Datakit 的 Host:Port

# ddtrace api for golang

```golang
// The below example illustrates a simple use case using the "tracer" package,
// our native Datadog APM tracing client integration. For thorough documentation
// and further examples, visit its own godoc page.
func example_for_golang_tracing() {
	// Start the tracer and defer the Stop method.
	tracer.Start(tracer.WithAgentAddr("{datakit_host}:{datakit_port}"))
	defer tracer.Stop()

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

# references

[ddtrace api for go](https://github.com/DataDog/dd-trace-go)
