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

  ## customer_tags is a list of keys contains keys set by client code like span.SetTag(key, value)
  ## that want to send to data center. Those keys set by client code will take precedence over
  ## keys in [inputs.ddtrace.tags]. DOT(.) IN KEY WILL BE REPLACED BY DASH(_) WHEN SENDING.
  customer_tags = ["key1", "key2", ...]

  ## Keep rare tracing resources list switch.
  ## If some resources are rare enough(not presend in 1 hour), those resource will always send
  ## to data center and do not consider samplers and filters.
  keep_rare_resource = false

  ## Ignore tracing resources map like service:[resources...].
  ## The service name is the full service name in current application.
  ## The resource list is regular expressions uses to block resource names.
  [inputs.ddtrace.close_resource]
    service1 = ["resource1", "resource2", ...]
    service2 = ["resource1", "resource2", ...]

  ## Sampler config uses to set global sampling strategy.
  ## priority uses to set tracing data propagation level, the valid values are -1, 0, 1
  ##   -1: always reject any tracing data send to datakit
  ##    0: accept tracing data and calculate with sampling_rate
  ##    1: always send to data center and do not consider sampling_rate
  ## sampling_rate used to set global sampling rate
  [inputs.ddtrace.sampler]
    priority = 0
    sampling_rate = 1.0

  [inputs.ddtrace.tags]
    key1 = "value1"
    key2 = "value2"

```

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
