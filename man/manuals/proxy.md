{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：全平台

# 简介

为了解决 Datakit 部署在特定环境下存在无法连接请求访问外网，提供了以下代理解决方案，用来解决该场景下的使用

- Nginx 反向代理服务
- Datakit 内置正向代理服务

## Nginx 反向代理配置

- 配置 Nginx 代理服务

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

curl -v -X POST http://127.0.0.1:8090/v1/write/metrics?token=tkn_76d2d1efd3ffxxxxxxxxxx -d "proxy_test_nginx,name=test c=123i"
```

- 配置 Datakit

进入 DataKit 安装目录下的 `conf.d/` 目录，配置  `datakit.conf` 中的代理服务。如下：

```
[dataway]
  urls = ["http://127.0.0.1:8090/v1/write/metrics?token=tkn_76d2d1efdxxxxxxxxxxxxxxxxxxxxxxx"] # ip和port为代理服务的配置信息
```

## Datakit 代理

- 配置 Datakit 提供的代理服务

进入 DataKit 安装目录下的 `conf.d/proxy` 目录，复制 `proxy.conf.sample` 并命名为 `proxy.conf`。示例如下：

```toml
[[inputs.proxy]]
    bind = "0.0.0.0"
    port = 9530
```

配置好后，重启 DataKit 即可。

- 测试代理服务

```
curl -x http://127.0.0.1:9530 -v -X POST https://openway.dataflux.cn/v1/write/metrics?token=tkn_76d2d1efd3ff43db984xxxxxx -d "proxy_test,name=test c=123i"
```

- 开启 Datakit 代理模式

进入 DataKit 安装目录下的 `conf.d/` 目录，配置  `datakit.conf` 中的代理服务。如下：

```toml
[dataway]
  urls = ["https://openway.dataflux.cn/v1/write/metrics?token=tkn_76d2d1efdxxxxxxxxxxxxxxxxxxxxxxx"]
  http_proxy = "http://xxx.xxx.xxx.xxx:9530" # Datakit 启动代理服务的 IP 和 Port
```
