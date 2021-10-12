{{.CSS}}

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# {{.InputName}}

当 Datakit 无法访问外网时，可在内网部署一个代理将流量发送出来。本文提供俩种实现方式：

- 通过 DataKit 内置的正向代理服务
- 通过 Nginx 反向代理服务

## DataKit 代理

挑选网络中的一个能访问外网的 DataKit，作为代理，配置其代理设置。

- 进入 DataKit 安装目录下的 `conf.d/proxy` 目录，复制 `proxy.conf.sample` 并命名为 `proxy.conf`。示例如下：

```toml
{{.InputSample}}
```

配置好后，[重启该代理 DataKit](datakit-how-to#147762ed)。

测试下代理服务是否正常：

- 通过发送 metrics 到工作空间测试

```shell
curl -x <PROXY-IP:PROXY-PORT> -v -X POST https://openway.guance.com/v1/write/metrics?token=<YOUR-TOKEN> -d "proxy_test,name=test c=123i"
```

如果代理服务器工作正常，工作空间将收到指标数据 `proxy_test,name=test c=123i`。

- 设置 _被代理 Datakit_ 的代理模式

进入被代理 DataKit 安装目录下的 `conf.d/` 目录，配置 `datakit.conf` 中的代理服务。如下：

```toml
[dataway]
  urls = ["https://openway.guance.com?token=<YOUR-TOKEN>"]
  http_proxy = "http://<PROXY-IP:PROXY-PORT>"
```

配置好后，[重启 DataKit](datakit-how-to#147762ed)。

## Nginx 反向代理配置

- 配置 `Nginx` 代理服务

```
server {
  listen       8090;

  location / {
    root   /usr/share/nginx/html;
    index  index.html index.htm;

    // 注意：这里不要用 https, 暂不支持
    proxy_pass http://openway.guance.com; # dataway地址
  }

  // ... 其它配置
}
```

- 加载新配置及测试

```shell
nginx -t        # 测试配置
nginx -s reload # reload配置
```

- 配置 `Datakit` 代理服务

进入 DataKit 安装目录下的 `conf.d/` 目录，配置 `datakit.conf` 中的代理服务。如下：

```toml
[dataway]
  urls = ["https://openway.guance.com?token=TOKEN"]
  http_proxy = "http://<NGINX-IP:NGINX-PORT>"
```

在被代理机器上，测试代理是否正常：

```shell
curl -v -x <NGINX-IP:NGINX-PORT> -X POST http://openway.guance.com/v1/write/metrics?token=<YOUR-TOKEN> -d "proxy_test_nginx,name=test c=123i"
```
