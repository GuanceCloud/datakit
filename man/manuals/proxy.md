{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# 简介

解决`Datakit`部署在无法访问`Internet`的内部网络环境需要使用代理服务器访问`Internet`。

- Nginx 反向代理服务
- Datakit 内置正向代理服务

## Nginx 反向代理配置

- 配置 `Nginx` 代理服务

```
server {
  listen       8090;

  location / {
    root   /usr/share/nginx/html;
    index  index.html index.htm;
    proxy_pass https://openway.dataflux.cn; # dataway地址
  }
}
```

- 加载新配置及测试

```
nginx -t # 测试配置
nginx -s reload # reload配置

curl -v -X POST http://127.0.0.1:8090/v1/write/metrics?token=TOKEN -d "proxy_test_nginx,name=test c=123i"
```

- 配置 `Datakit` 代理服务

进入 DataKit 安装目录下的 `conf.d/` 目录，配置 `datakit.conf` 中的代理服务。如下：

```
[dataway]
	# IP 和 Port 为 Nginx 代理服务的配置信息
  urls = ["http://127.0.0.1:8090?token=TOKEN"]
```

> 注意：Nginx 代理的情况下，到此即可，无需进行以下步骤。

## Datakit 代理

挑选网络中的一个能访问外网的 DataKit，作为代理，配置其代理设置。

- 进入 DataKit 安装目录下的 `conf.d/proxy` 目录，复制 `proxy.conf.sample` 并命名为 `proxy.conf`。示例如下：

```toml
{{.InputSample}}
```

配置好后，[重启该代理 DataKit](datakit-how-to#147762ed)。

测试下代理服务是否正常：

```shell
curl -x http://127.0.0.1:9530 -v -X POST https://openway.dataflux.cn/v1/write/metrics?token=TOKEN -d "proxy_test,name=test c=123i"
```

- 设置 _被代理 Datakit_ 的代理模式

进入被代理 DataKit 安装目录下的 `conf.d/` 目录，配置 `datakit.conf` 中的代理服务。如下：

```toml
[dataway]
  urls = ["https://openway.dataflux.cn?token=TOKEN"]
  http_proxy = "http://<代理 DataKit IP>:9530"
```

配置好后，[重启 DataKit](datakit-how-to#147762ed)。

## Datakit 代理指标集

以下所有指标集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

```toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

- 标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}
