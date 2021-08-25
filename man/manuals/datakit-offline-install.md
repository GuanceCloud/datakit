{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：全平台

# 简介

某些时候，目标机器没有公网访问出口，按照如下方式可离线安装 DataKit。

## 前置条件

- 通过[正常安装方式](datakit-install)，在有公网出口的机器上安装一个 DataKit
- 开通该 DataKit 上的 [proxy](proxy) 采集器，假定 proxy 采集器所在 Datakit IP 为 1.2.3.4，有如下配置：

```toml
[[inputs.proxy]]
  ## default bind ip address
  bind = "0.0.0.0" 
  ## default bind port
  port = 9530
```

## 离线安装

### Linux/Mac

增加环境变量 `HTTPS_PROXY="1.2.3.4:9530"`，安装命令如下：

```
HTTPS_PROXY="http://1.2.3.4:9530" DK_DATAWAY=https://openway.dataflux.cn?token=<TOKEN> bash -c "$(curl -L https://static.dataflux.cn/datakit/install.sh)"
```

### Windows

增加环境变量 `$env:HTTPS_PROXY="1.2.3.4:9530"`，安装命令如下：

```shell
$env:HTTPS_PROXY="1.2.3.4:9530"; $env:DK_DATAWAY="https://openway.dataflux.cn?token=<TOKEN>"; Set-ExecutionPolicy Bypass -scope Process -Force; Import-Module bitstransfer; start-bitstransfer -ProxyUsage Override -ProxyList $env:HTTPS_PROXY -source https://static.dataflux.cn/datakit/install.ps1 -destination .install.ps1; powershell .install.ps1;
```

> 注意：其它安装参数设置，跟[正常安装](datakit-install) 无异。
