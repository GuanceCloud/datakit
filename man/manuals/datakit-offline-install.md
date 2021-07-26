{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：全平台

# 简介

某些时候，目标机器没有公网访问出口，按照如下方式可离线安装 DataKit。

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
- Darwin(Mac) 64 位：
	- [Installer](https://static.dataflux.cn/datakit/installer-darwin-amd64)
	- [DataKit](https://static.dataflux.cn/datakit/datakit-darwin-amd64-{{.Version}}.tar.gz)
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

## 离线安装

下载完后，应该有三个文件：

- `datakit-OS-ARCH.tar.gz`
- `installer-OS-ARCH` 或 `installer-OS-ARCH.exe`
- `data.tar.gz`

通过 `scp` 或其它文件传输工具，将安装程序 `installer-OS-ARCH` （Windows 下文件名为 `installer-OS-ARCH.exe`）以及安装包（`data.tar.gz` 以及 `datakit-OS-ARCH-{{.Version}}.tar.gz`）上传到目标机器。以 Linux 为例：

```shell
scp installer-linux-amd64 datakit-linux-amd64-{{.Version}}.tar.gz data.tar.gz USER-NAME@YOUR-LINUX-HOST:~/
```

登陆目标机器，在对应目录下，*将以下命令中的 `TOKEN` 替换成工作空间的 Token*，即可执行安装（以 64 位 X86 为例）：

```shell
# Windows（需以 administrator 权限运行 Powershell 执行）
installer-windows-amd64.exe -offline -dataway "https://openway.dataflux.cn?token=<TOKEN>" -srcs .\datakit-windows-amd64-{{.Version}}.tar.gz,.\data.tar.gz

# Linux（需以 root 权限运行）
chmod +x installer-linux-amd64
./installer-linux-amd64 -offline -dataway "https://openway.dataflux.cn?token=<TOKEN>" -srcs datakit-linux-amd64-{{.Version}}.tar.gz,data.tar.gz

# Mac （需以 root 权限运行）
chmod +x installer-darwin-amd64
./installer-darwin-amd64 -offline -dataway "https://openway.dataflux.cn?token=<TOKEN>" -srcs datakit-darwin-amd64-{{.Version}}.tar.gz,data.tar.gz
```
