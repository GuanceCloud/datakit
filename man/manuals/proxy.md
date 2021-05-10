# 简介

为了解决 Datakit 部署在特定环境下存在无法连接请求访问外网，提供了以下代理解决方案，用来解决该场景下的使用
- Nginx代理
- Datakit提供的代理

## Nginx代理配置

- 配置 Nginx 代理服务

```
server {
    resolver 114.114.114.114;       #指定DNS服务器IP地址 
    listen 80;
    location / {
        proxy_pass http://$host$request_uri;     #设定代理服务器的协议和地址 
            proxy_set_header HOST $host;
            proxy_buffers 256 4k;
            proxy_max_temp_file_size 0k;
            proxy_connect_timeout 30;
            proxy_send_timeout 60;
            proxy_read_timeout 60;
            proxy_next_upstream error timeout invalid_header http_502;
    }
}
```

- 开启 Datakit 代理模式
```
[dataway]
  urls = ["https://openway.dataflux.cn/v1/write/metrics?token=tkn_76d2d1efd3ff43db984497bfb4f3c25a"]
  http_proxy = "http://xxx.xxx.xxx.xxx:9530" # nginx 启动代理服务的ip和port
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

- 开启 Datakit 代理模式

进入 DataKit 安装目录下的 `conf.d/` 目录，配置  `datakit.conf` 中的代理服务。如下：

```
[dataway]
  urls = ["https://openway.dataflux.cn/v1/write/metrics?token=xxxxxx"]
  http_proxy = "http://xxx.xxx.xxx.xxx:9530" # Datakit 启动代理服务的ip和port
```
