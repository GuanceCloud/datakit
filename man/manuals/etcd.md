{{.CSS}}

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# etcd

etcd 采集器可以从 etcd 实例中采取很多指标，比如etcd服务器状态和网络的状态等多种指标，并将指标采集到 DataFlux ，帮助你监控分析 etcd 各种异常情况

## 前置条件

- etcd 版本  >=3

- 开启etcd，默认的metrics接口是http://localhost:2379/metrics，也可以自己去配置文件中修改。

## 配置

进入 DataKit 安装目录下的 `conf.d/etcd` 目录，复制如下示例 并命名为 `etcd.conf`。示例如下：

```toml
{{.InputSample}}
```

配置好后，重启 DataKit 即可。

## 指标集

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

-  标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}
