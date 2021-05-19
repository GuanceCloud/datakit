{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# {{.InputName}}

Apache 采集器可以从 Apache 服务中采集请求数、连接数等，并将指标采集到 DataFlux ，帮助你监控分析 Apache 各种异常情况

## 前置条件

- 一般发行版 Linux 会自带 Apache,如需下载[参见](https://httpd.apache.org/download.cgi)

- 默认配置路径: `/etc/apache2/apache2.conf`,`/etc/apache2/httpd.conf`

- 开启 Apache `mod_status`,在 Apache 配置文件中添加:

```
<Location /server-status>
SetHandler server-status

Order Deny,Allow
Deny from all
Allow from your_ip
</Location>
```

- 重启 Apache

```sudo apachectl restart```


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

如需采集 Apache 的日志，可在 {{.InputName}}.conf 中 将 `files` 打开，并写入 Apache 日志文件的绝对路径。比如：

```
    [[inputs.apache]]
      ...
      [inputs.apache.log]
		files = ["/var/log/apache2/error.log","/var/log/apache2/access.log"]
```


开启日志采集以后，默认会产生日志来源（`source`）为 `apache` 的日志。

>注意：必须将 DataKit 安装在 Apache 所在主机才能采集 Apache 日志
