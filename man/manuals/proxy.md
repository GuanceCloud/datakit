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

配置好后，[重启该代理 DataKit](datakit-service-how-to#147762ed)。

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

配置好后，[重启 DataKit](datakit-service-how-to#147762ed)。

## Nginx 正向代理配置

代理 HTTPS 流量这里 nginx 采用 4 层的透明代理方式，即需要:

- 一台可以访问外网的 nginx 的透明代理服务器
- datakit 所在的客户机使用 hosts 文件进行域名配置

### 配置 `Nginx` 代理服务

```
# 代理 HTTPS
stream {
    # resolver 114.114.114.114;
    # resolver_timeout 30s;
    server {
        listen 443;
        ssl_preread on;
        proxy_connect_timeout 10s;
        proxy_pass $ssl_preread_server_name:$server_port;
    }
}

http {
    ...
}
```

代理 HTTP 流量这里 nginx 采用 7 层的透明代理方式(如果不需要代理 HTTP 这段可以跳过):

```
# 代理 HTTP
http {
    # resolver 114.114.114.114;
    # resolver_timeout 30s;
    server {
        listen 80;
        location / {
            proxy_pass http://$http_host$request_uri;    # 配置正向代理参数
            proxy_set_header Host $http_host;            # 解决如果 URL 中带 "." 后 nginx 503 错误
            proxy_buffers 256 4k;                        # 配置缓存大小
            proxy_max_temp_file_size 0;                  # 关闭磁盘缓存读写减少 I/O
            proxy_connect_timeout 30;                    # 代理连接超时时间
            proxy_cache_valid 200 302 10m;
            proxy_cache_valid 301 1h;
            proxy_cache_valid any 1m;                    # 配置代理服务器缓存时间
            proxy_send_timeout 60;
            proxy_read_timeout 60;
        }
    }

    // ... 其它配置
}
```

### 加载新配置及测试

```shell
nginx -t        # 测试配置
nginx -s reload # reload配置
```

### 配置 `Datakit` 被代理机器上的域名

下面假设 `192.168.1.66` 是 nginx 透明代理服务器的 IP 地址。

```sh
$ sudo vi /etc/hosts

192.168.1.66 static.guance.com
192.168.1.66 openway.guance.com
192.168.1.66 dflux-dial.guance.com

192.168.1.66 static.dataflux.cn
192.168.1.66 openway.dataflux.cn
192.168.1.66 dflux-dial.dataflux.cn

192.168.1.66 zhuyun-static-files-production.oss-cn-hangzhou.aliyuncs.com
```

在被代理机器上，测试代理是否正常：

Linux/Unix:

```shell
curl -H "application/x-www-form-urlencoded; param=value" \
  -d 'proxy_test_nginx,name=test c=123i' \
  "https://openway.guance.com/v1/write/metrics?token=<YOUR-TOKEN>"
```

PowerShell:

```PowerShell
curl -uri 'https://openway.guance.com/v1/write/metrics?token=<YOUR-TOKEN>' -Headers @{"param"="value"} -ContentType 'application/x-www-form-urlencoded' -body 'proxy_test_nginx,name=test c=123i' -method 'POST'
```

注意: PowerShell 有的机器上会报 `curl : 请求被中止: 未能创建 SSL/TLS 安全通道。` 的错误，这是因为服务端证书加密版本号在本地默认不被支持造成的，可以通过命令 `[Net.ServicePointManager]::SecurityProtocol` 查看支持的协议。如果想要本地支持可以做以下操作:

```PowerShell
# 64 bit PowerShell
Set-ItemProperty -Path 'HKLM:\SOFTWARE\Wow6432Node\Microsoft\.NetFramework\v4.0.30319' -Name 'SchUseStrongCrypto' -Value '1' -Type DWord

# 32 bit PowerShell
Set-ItemProperty -Path 'HKLM:\SOFTWARE\Microsoft\.NetFramework\v4.0.30319' -Name 'SchUseStrongCrypto' -Value '1' -Type DWord
```

关闭 PowerShell 窗口，打开一个新的 PowerShell 窗口，执行以下代码查看支持的协议:

```PowerShell
[Net.ServicePointManager]::SecurityProtocol
```
