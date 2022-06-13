{{.CSS}}
# Zipkin
---

- DataKit 版本：{{.Version}}
- 操作系统支持：`{{.AvailableArchs}}`

Datakit 内嵌的 Zipkin Agent 用于接收，运算，分析 Zipkin Tracing 协议数据。

## Zipkin 文档

- [Quickstart](https://zipkin.io/pages/quickstart.html)
- [Docs](https://zipkin.io/pages/instrumenting.html)
- [Souce Code](https://github.com/openzipkin/zipkin)

## 配置 Zipkin Agent

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}}
```
