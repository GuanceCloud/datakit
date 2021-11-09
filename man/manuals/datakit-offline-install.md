{{.CSS}}

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
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

## 通过代理安装

### Linux/Mac

- 使用 datakit 代理

增加环境变量 `HTTPS_PROXY="1.2.3.4:9530"`，安装命令如下：

```shell
export HTTPS_PROXY=http://1.2.3.4:9530; DK_DATAWAY=https://openway.guance.com?token=<TOKEN> bash -c "$(curl -L https://static.guance.com/datakit/install.sh)"
```

- 使用 nginx 代理

增加环境变量 `DK_PROXY_TYPE="nginx"; DK_NGINX_IP="1.2.3.4";`，安装命令如下：

```shell
export DK_PROXY_TYPE="nginx"; DK_NGINX_IP="1.2.3.4"; DK_DATAWAY=https://openway.guance.com?token=<TOKEN> bash -c "$(curl -L https://static.guance.com/datakit/install.sh)"
```

### Windows

- 使用 datakit 代理

增加环境变量 `$env:HTTPS_PROXY="1.2.3.4:9530"`，安装命令如下：

```powershell
$env:HTTPS_PROXY="1.2.3.4:9530"; $env:DK_DATAWAY="https://openway.guance.com?token=<TOKEN>"; Set-ExecutionPolicy Bypass -scope Process -Force; Import-Module bitstransfer; start-bitstransfer -ProxyUsage Override -ProxyList $env:HTTPS_PROXY -source https://static.guance.com/datakit/install.ps1 -destination .install.ps1; powershell .install.ps1;
```

- 使用 nginx 代理

增加环境变量 `$env:DK_PROXY_TYPE="nginx"; $env:DK_NGINX_IP="1.2.3.4";`，安装命令如下：

```powershell
$env:DK_PROXY_TYPE="nginx"; $env:DK_NGINX_IP="1.2.3.4"; $env:DK_DATAWAY="https://openway.guance.com?token=<TOKEN>"; Set-ExecutionPolicy Bypass -scope Process -Force; Import-Module bitstransfer; start-bitstransfer -ProxyUsage Override -ProxyList $env:HTTPS_PROXY -source https://static.guance.com/datakit/install.ps1 -destination .install.ps1; powershell .install.ps1;
```

> 注意：其它安装参数设置，跟[正常安装](datakit-install) 无异。

## 全离线安装

当环境完全没有外网的情况下，只能通过移动硬盘（U 盘）等方式。

## 下载安装包

以下文件的地址，可通过 wget 等下载工具，也可以直接在浏览器中输入对应的 URL 下载。

> Safari 浏览器下载时，后缀名可能不同（如将 `.tar.gz` 文件下载成 `.tar`），会导致安装失败。建议用 Chrome 浏览器下载。

先下载数据包，每个平台都一样： https://static.dataflux.cn/datakit/data.tar.gz

然后再下载俩个安装程序：

- Windows 32 位：
  - [Installer](https://static.dataflux.cn/datakit/installer-windows-386.exe)
  - [DataKit](https://static.dataflux.cn/datakit/datakit-windows-386-{{.Version}}.tar.gz)
- Windows 64 位：
  - [Installer](https://static.dataflux.cn/datakit/installer-windows-amd64.exe)
  - [DataKit](https://static.dataflux.cn/datakit/datakit-windows-amd64-{{.Version}}.tar.gz)
- Linux X86 32 位：
  - [Installer](https://static.dataflux.cn/datakit/installer-linux-386)
  - [DataKit](https://static.dataflux.cn/datakit/datakit-linux-386-{{.Version}}.tar.gz)
- Linux X86 64 位
  - [Installer](https://static.dataflux.cn/datakit/installer-linux-amd64)
  - [DataKit](https://static.dataflux.cn/datakit/datakit-linux-amd64-{{.Version}}.tar.gz)
- Linux Arm 32 位
  - [Installer](https://static.dataflux.cn/datakit/installer-linux-arm)
  - [DataKit](https://static.dataflux.cn/datakit/datakit-linux-arm-{{.Version}}.tar.gz)
- Linux Arm 64 位
  - [Installer](https://static.dataflux.cn/datakit/installer-linux-arm64)
  - [DataKit](https://static.dataflux.cn/datakit/datakit-linux-arm64-{{.Version}}.tar.gz)

下载完后，应该有三个文件（此处 `<OS-ARCH>` 指特定平台的安装包）：

- `datakit-<OS-ARCH>.tar.gz`
- `installer-<OS-ARCH>` 或 `installer-<OS-ARCH>.exe`
- `data.tar.gz`

将这些文件拷贝到对应机器上（通过 U 盘或 scp 等命令）。

### 安装

```shell
# Windows（需以 administrator 权限运行 Powershell 执行）
.\installer-windows-amd64.exe --offline --dataway "https://openway.dataflux.cn?token=<YOUR-TOKEN>" --srcs .\datakit-windows-amd64-{{.Version}}.tar.gz,.\data.tar.gz

# Linux（需以 root 权限运行）
chmod +x installer-linux-amd64
./installer-linux-amd64 --offline --dataway "https://openway.dataflux.cn?token=<YOUR-TOKEN>" --srcs datakit-linux-amd64-{{.Version}}.tar.gz,data.tar.gz
```
