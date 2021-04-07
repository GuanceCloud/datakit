- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}

# 简介

`nginx` 采集器 可以从 `nginx` 实例中采取很多指标，比如 请求总数 连接数，缓存等多种指标，并将指标采集到 `dataflux` ，帮助你监控分析 `nginx` 各种异常情况


## 前置条件

- `nginx` 默认采集开源版本 `nginx` 的 `http_stub_status_module`模块的数据， 开启 `http_stub_status_module`模块参见[这里](http://nginx.org/en/docs/http/ngx_http_stub_status_module.html)，
开启了以后会上报  `nginx` 指标集的数据
- 如果您正在使用 `virtual host traffic status module` 或者想监控更多数据,建议开启 `vts` 相关数据采集,conf 参数 `use_vts` 设置为 `true`，开启 `vts` 参见[这里](https://github.com/vozlt/nginx-module-vts#synopsis)。
开启 `vts` 会产生  `nginx`, `nginx_server_zone`, `nginx_upstream_zone`, `nginx_cache_zone` 等指标集

## 配置

进入 DataKit 安装目录下的 `conf.d/{{.InputName}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```python
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
