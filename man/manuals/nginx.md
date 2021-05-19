{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# {{.InputName}}

NGINX 采集器可以从 NGINX 实例中采取很多指标，比如请求总数连接数、缓存等多种指标，并将指标采集到 DataFlux ，帮助你监控分析 NGINX 各种异常情况

## 前置条件

- NGINX 默认采集 `http_stub_status_module` 模块的数据，开启 `http_stub_status_module` 模块参见[这里](http://nginx.org/en/docs/http/ngx_http_stub_status_module.html)，开启了以后会上报 NGINX 指标集的数据

- 如果您正在使用 [VTS](https://github.com/vozlt/nginx-module-vts) 或者想监控更多数据，建议开启 VTS 相关数据采集，可在 `{{.InputName}}.conf` 中将选项 `use_vts` 设置为 `true`。如何开启 VTS 参见[这里](https://github.com/vozlt/nginx-module-vts#synopsis)。

- 开启 VTS 功能后，能产生如下指标集：

    - `nginx`
    - `nginx_server_zone`
    - `nginx_upstream_zone` (NGINX 需配置 `upstream` 相关配置)
    - `nginx_cache_zone`    (NGINX 需配置 `cache` 相关配置)

- 以产生 `nginx_upstream_zone` 指标集为例，NGINX 相关配置示例如下：

```
    ...
    http {
       ...
       upstream your-upstreamname {
         server upstream-ip:upstream-port;
      }
       server {
       ...
       location / {
       root  html;
       index  index.html index.htm;
       proxy_pass http://yourupstreamname;
     }}}

```

- 已经开启了 VTS 功能以后，不必再去采集 `http_stub_status_module` 模块的数据，因为 VTS 模块的数据会包括 `http_stub_status_module` 模块的数据

## 配置

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}}
```

配置好后，重启 DataKit 即可。

## 指标集

以下所有指标集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[[inputs.{{.InputName}}.tags]]` 另择 host 来命名。

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

-  标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }} 


## 日志采集

如需采集 NGINX 的日志，可在 {{.InputName}}.conf 中 将 `files` 打开，并写入 NGINX 日志文件的绝对路径。比如：

```
    [[inputs.nginx]]
      ...
      [inputs.nginx.log]
		files = ["/usr/local/var/log/nginx/error.log","/usr/local/var/log/nginx/access.log"]
```


开启日志采集以后，默认会产生日志来源（`source`）为 `nginx` 的日志。

>注意：必须将 DataKit 安装在 NGINX 所在主机才能采集 NGINX 日志
