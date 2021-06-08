{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：全平台

# 简介

某些时候，目标机器没有公网访问出口，按照如下方式可离线安装 DataKit。

## 下载安装包

建议在平台匹配的情况下下载安装包，即用 Linux X86 机器下载 Linux X86 平台的 DataKit的安装程序。在 Linux X86 机器上无法下载 Windows 平台的安装程序。以此类推。

### Windows x86 64 位

```shell
Import-Module bitstransfer; start-bitstransfer -source https://static.dataflux.cn/datakit/installer-windows-amd64.exe -destination .\dk-installer.exe; .\dk-installer.exe -download-only
```

### Windows x86 32 位

```shell
Import-Module bitstransfer; start-bitstransfer -source https://static.dataflux.cn/datakit/installer-windows-386.exe -destination .\dk-installer.exe; .\dk-installer.exe -download-only
```

### Linux x86 64 位

```shell
curl https://static.dataflux.cn/datakit/installer-linux-amd64 -o dk-installer && chmod +x ./dk-installer && ./dk-installer -download-only
```

### Linux x86 32 位

```shell
curl https://static.dataflux.cn/datakit/installer-linux-386 -o dk-installer && chmod +x ./dk-installer && ./dk-installer -download-only
```

### Linux Arm 64 位

```shell
curl https://static.dataflux.cn/datakit/installer-linux-arm64 -o dk-installer && chmod +x ./dk-installer && ./dk-installer -download-only
```

### Linux Arm 32 位

```shell
curl https://static.dataflux.cn/datakit/installer-linux-arm -o dk-installer && chmod +x ./dk-installer && ./dk-installer -download-only
```

### Mac 64 位

```shell
curl https://static.dataflux.cn/datakit/installer-darwin-amd64 -o dk-installer && chmod +x ./dk-installer && ./dk-installer -download-only
```

## 离线安装

> DataKit 安装有两个，一个是程序包（`datakit-xxx.tar.gz`），一个是数据包（`data.tar.gz`）。其中，数据包每个平台都一样。

下载完后，当前目录下会出现 `dk-installer.exe`（或者 `dk-installer`）、DataKit 安装包 `datakit-xxx.tar.gz` 以及数据文件 `data.tar.gz`，按如下方式执行离线安装

- 通过 `scp` 或其它文件传输工具，将安装程序 `dk-installer` （Windows 下文件名为 `dk-installer.exe`）以及安装包（如 `datakit-linux-amd64-1.1.5-rc2.tar.gz`）上传到目标机器。以 Linux 为例：

```shell
scp dk-installer datakit-linux-amd64-<some-version>.tar.gz data.tar.gz user@your-linux-host:~/
```

- 登陆目标机器，在对应目录下，即可执行安装：

```shell
# Windows（需以 administrator 权限运行 Powershell 执行）
.\dk-installer.exe -offline -dataway "https://openway.dataflux.cn?token=<your-token>" -srcs datakit-windows-amd64-<some-version>.tar.gz,data.tar.gz

# Linux（需以 root 权限运行）
./dk-installer -offline -dataway "https://openway.dataflux.cn?token=<your-token>" -srcs datakit-linux-amd64-<some-version>.tar.gz,data.tar.gz

# Mac （需以 root 权限运行）
./dk-installer -offline -dataway "https://openway.dataflux.cn?token=<your-token>" -srcs datakit-darwin-amd64-<some-version>.tar.gz,data.tar.gz
```
