{{.CSS}}

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
- 操作系统支持：全平台

# 准备工作

1. 在您的开发目录下运行 `go get -v github.com/DataDog/dd-trace-go`
2. 复制`Datakit`安装目录下`conf.d/ddtrace/ddtrace.conf.example`并重命名为`ddtrace.conf`
3. 重启`Datakit`

# ddtrace api for golang

```
// The below example illustrates a simple use case using the "tracer" package,
// our native Datadog APM tracing client integration. For thorough documentation
// and further examples, visit its own godoc page.
func Example_datadog() {
// Start the tracer and defer the Stop method.
tracer.Start(tracer.WithAgentAddr("host:port"))
defer tracer.Stop()

    // Start a root span.
    span := tracer.StartSpan("get.data")
    defer span.Finish()

    // Create a child of it, computing the time needed to read a file.
    child := tracer.StartSpan("read.file", tracer.ChildOf(span.Context()))
    child.SetTag(ext.ResourceName, "test.json")

    // Perform an operation.
    _, err := ioutil.ReadFile("~/test.json")

    // We may finish the child span using the returned error. If it's
    // nil, it will be disregarded.
    child.Finish(tracer.WithError(err))
    if err != nil {
    	log.Fatal(err)
    }
}
```

# references

[ddtrace api for go](https://github.com/DataDog/dd-trace-go)
