{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# 简介

Prom 采集器可以获取各种Prometheus Exporters的监控数据，用户只要配置相应的Endpoint，就可以将监控数据接入。支持指标过滤、指标集重命名等。

## 前置条件

- 必须是 Prometheus 的数据格式

## 配置

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}}
```

配置好后，重启 DataKit 即可。
