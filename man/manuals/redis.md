- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}

# 简介

redis监控指标采集，参考datadog提供的指标，具有以下数据收集功能
- 开启AOF数据持久化，会收集相关指标
- RDB数据持久化指标
- Slowlog监控指标
- bigkey scan监控
- 主从replication

备注：
redis开发测试版本为v5.04, v6.0+待支持

## 前置条件
暂无

## 配置

进入 DataKit 安装目录下的 `conf.d/{{.InputName}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```
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
