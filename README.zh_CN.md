# DataKit

<p align="center">
  <img alt="golangci-lint logo" src="datakit-logo.png" height="150" />
</p>

[![Slack Status](https://img.shields.io/badge/slack-join__chat-orange)](https://app.slack.com/client/T032YB4B6TA/)
[![MIT License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

<h2>
  <a href="https://pyroscope.io/">Website</a>
  <span> • </span>
  <a href="https://pyroscope.io/docs">Docs</a>
  <span> • </span>
</h2>

DataKit 是一款开源、一体式的数据集成 Agent，它提供全平台操作系统（Linux/Windows/macOS）支持，拥有全面数据采集能力，涵盖主机、容器、中间件、Tracing、日志以及安全巡检等各种场景：

## 主要功能点

- 支持[主机]()、[中间件]()、[日志]()、[APM]() 等领域的指标、日志以及 Tracing 几大类数据采集
- 完整支持 [Kubernates]() 生态
- [Pipeline]()：简便的结构化数据提取
- 支持接入其它第三方数据采集
	- [Telegraf]()
	- [Prometheus]()
	- [Statsd]()
	- [Fluentd]()
	- [Function]()
	- Tracing 相关（[OpenTelemetry]()/[DDTrace]()/[Zipkin]()/[Jaeger]()/[Skywalking]()）

## 操作系统最低要求

| 操作系统 | 架构 | 安装路径 |
| --- | --- | --- |
| Linux 内核 2.6.23 或更高版本 | amd64/386/arm/arm64 | `/usr/local/datakit` |
| macOS 10.12 或更高版本([原因](https://github.com/golang/go/issues/25633)) | amd64 | `/usr/local/datakit` |
| Windows 7, Server 2008R2 或更高版本 | amd64/386 | 64位：`C:\Program Files\datakit`<br />32位：`C:\Program Files(32)\datakit` |

## 观测云 DataKit 安装

我们可以直接在观测云平台获取 DataKit 安装命令，主流平台的安装命令大概如下：

- Linux & Mac
```shell
DK_DATAWAY="https://openway.guance.com?token=<YOUR-TOKEN>" bash -c "$(curl -L https://static.guance.com/datakit/install.sh)"
```

- Windows

```powershell
$env:DK_DATAWAY="https://openway.guance.com?token=tkn_2af4b19d7f5a489fa81f0fff7e63b588";Set-ExecutionPolicy Bypass -scope Process -Force; Import-Module bitstransfer; start-bitstransfer -source https://static.guance.com/datakit/install.ps1 -destination .install.ps1; powershell .install.ps1;
```

## 源码编译

### 外部依赖

以下依赖（库/工具）主要用于 DataKit 自身的编译、打包以及发布流程。其中，**不建议在 Windows 上编译 DataKit**。

- Go-1.16.4 及以上版本
- `apt-get install gcc-multilib`: 用于编译 Oracle 采集器
- `apt-get install tree`: Makefile 中用于显示编译结果
- `packr2`: 用于打包一些资源文件
- `go get -u golang.org/x/tools/cmd/goyacc`: 用于生成 Pipeline 语法代码
- Docker 用于生成 DataKit 镜像
- lint 相关
	- `go install mvdan.cc/gofumpt@latest` 用于规范化 Golang 代码格式
	- [golangci-lint 1.42.1](https://github.com/golangci/golangci-lint/releases/tag/v1.42.1)
- eBPF 相关
	- clang 10.0+
	- llvm 10.0+
	- `apt install go-bindata`
- 文档相关
	- [waque 1.13.1+](https://github.com/yesmeck/waque)

### 编译

1. 拉取代码：

```shell
git clone https://github.com/DataFlux-cn/datakit.git
```

2. 编译：

```shell
cd datakit
make
```

如果编译通过，将在当前目录的 *disk* 目录下生成如下文件：

```
dist/
├── datakit-linux-amd64
│   ├── datakit             # DataKit 主程序
│   └── externals      
│       ├── datakit-ebpf    # eBPF 相关采集器
│       ├── logfwd          # logfwd 采集器
│       └── oracle          # Oracle 采集器
└── local
    ├── installer-linux-amd64 # Linux 平台安装程序
    └── version               # version 信息描述文件
```

如果要编译全平台版本，执行：

```shell
make testing
```

## DataKit 基本使用

可通过如下命令查看更多使用方法：

```shell
datakit help
```

## 文档

DataKit 文档，参见[这里](https://www.yuque.com/dataflux/datakit)。
