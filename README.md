<p align="center">
  <img alt="datakit logo" src="datakit-logo.png" height="150" />
</p>

[![Slack Status](https://img.shields.io/badge/slack-join_chat-orange?logo=slack&style=plastic)](https://app.slack.com/client/T032YB4B6TA/)
[![MIT License](https://img.shields.io/badge/license-MIT-green?style=plastic)](LICENSE)

<h2>
  <a href="https://datakit.tools">Website</a>
  <span> • </span>
  <a href="https://www.yuque.com/dataflux/datakit">Doc</a>
</h2>


## _Read this in other languages._
<kbd>[<img title="中文 (Simplified)" alt="中文 (Simplified)" src="https://cdn.staticaly.com/gh/hjnilsson/country-flags/master/svg/cn.svg" width="22">](README.zh_CN.md)</kbd>

DataKit is an open source, integrated data collection agent, which provides full platform (Linux/Windows/macOS) support and has comprehensive data collection capability, covering various scenarios such as host, container, middleware, tracing, logging and security inspection.

## Key Features

- Support collection of metrics, logging and tracing
- Fully support Kubernetes ecology
- [Pipeline](https://www.yuque.com/dataflux/datakit/pipeline): Simple structured data extraction
- Supports third-party data import:
	- [Telegraf](https://www.yuque.com/dataflux/datakit/telegraf)
	- [Prometheus](https://www.yuque.com/dataflux/datakit/prom)
	- [Statsd](https://www.yuque.com/dataflux/datakit/statsd)
	- [Fluentd](https://www.yuque.com/dataflux/datakit/logstreaming#a653042e)
	- [Function](https://www.yuque.com/dataflux/func/write-data-via-datakit)
	- Tracing related(OpenTelemetry/[DDTrace](https://www.yuque.com/dataflux/datakit/ddtrace)/Zipkin/[Jaeger](https://www.yuque.com/dataflux/datakit/jaeger)/[Skywalking](https://www.yuque.com/dataflux/datakit/skywalking))

## Changelog

All DataKit changelog refers to [here](https://www.yuque.com/dataflux/datakit/changelog).

## Minimal Requirements

| OS | Arch | Install Path |
| --- | --- | --- |
| Linux Kernel 2.6.23+ | amd64/386/arm/arm64 | `/usr/local/datakit` |
| macOS 10.12+([Why](https://github.com/golang/go/issues/25633)) | amd64 | `/usr/local/datakit` |
| Windows 7+/Server 2008R2+ | amd64/386 | 64-bit：`C:\Program Files\datakit`<br />32-nit：`C:\Program Files(32)\datakit` |


## Install DataKit

We can directly obtain the DataKit installation command from [guance cloud](http://guance.com). Most of the installation commands seems like that:

- Linux & Mac
```shell
DK_DATAWAY="https://openway.guance.com?token=<YOUR-TOKEN>" bash -c "$(curl -L https://static.guance.com/datakit/install.sh)"
```

- Windows

```powershell
$env:DK_DATAWAY="https://openway.guance.com?token=<YOUR-TOKEN>";Set-ExecutionPolicy Bypass -scope Process -Force; Import-Module bitstransfer; start-bitstransfer -source https://static.guance.com/datakit/install.ps1 -destination .install.ps1; powershell .install.ps1;
```

- [Kubernetes DaemonSet](https://www.yuque.com/dataflux/datakit/datakit-daemonset-deploy)

For more documentations about DataKit installation, see [here](https://www.yuque.com/dataflux/datakit/datakit-install).

### Install community release

We also released the [community DataKit](https://www.yuque.com/dataflux/datakit/changelog#5a0afc9d), we can install via

- Linux & Mac

```bash
DK_DATAWAY="https://openway.guance.com?token=<YOUR-TOKEN>" bash -c "$(curl -L https://static.guance.com/datakit/community/install.sh)"
```

- Windows

```powershell
$env:DK_DATAWAY="https://openway.guance.com?token=<YOUR-TOKEN>";Set-ExecutionPolicy Bypass -scope Process -Force; Import-Module bitstransfer; start-bitstransfer -source https://static.guance.com/datakit/community/install.ps1 -destination .install.ps1; powershell .install.ps1;
```

- [Kubernetes DaemonSet](https://www.yuque.com/dataflux/datakit/datakit-daemonset-deploy)

```bash
# We should use the community version yaml
wget https://static.guance.com/datakit/community/datakit.yaml
```

## Build From Source

DataKit building relies on some external tools/libs, we must install them all before compile the source code.

> We do not support build DataKit on Windows.


- Go-1.16.4+
- gcc-multilib: Used to build Oracle input(`apt-get install gcc-multilib`)
- tree: After building datakit, `tree` used to show all bianries(`apt-get install tree`)
- packr2: Used to package resources(mainly documents)
- goyacc: Used to build grammar for Pipleine(`go get -u golang.org/x/tools/cmd/goyacc`)
- Docker: Used to build DataKit image
- lint related:
	- gofumpt: Used to format go source code(`go install mvdan.cc/gofumpt@latest`)
	- [golangci-lint 1.42.1](https://github.com/golangci/golangci-lint/releases/tag/v1.42.1)
- eBPF related:
	- clang 10.0+
	- llvm 10.0+
	- `apt install go-bindata`
- Documentation exporting:
	- [waque 1.13.1+](https://github.com/yesmeck/waque)

### Build

1. Clone

```shell
git clone https://github.com/DataFlux-cn/datakit.git
```

2. Building

```shell
cd datakit
make
```

If building ok, all binaries are generated under *dist*:

```
dist/
├── datakit-linux-amd64
│   ├── datakit             # DataKit main binary
│   └── externals      
│       ├── datakit-ebpf    # eBPF collector
│       ├── logfwd          # logfwd collector
│       └── oracle          # Oracle collector
└── local
    ├── installer-linux-amd64 # installer used fo Linux 
    └── version               # version descriptor
```

We can build all platforms(Linux/Mac/Windows) with following command:

```shell
make testing
```

## Basic Usage

We can use `help` command to see more usage of DataKit:

```shell
datakit help

# Or

./dist/datakit-linux-amd64/datakit help
```

## Contributing

Before contributing, check out some guideline of DataKit:

- Read [architecure introduciton](https://www.yuque.com/dataflux/datakit/datakit-arch)
- Read [development guideline](https://www.yuque.com/dataflux/datakit/development)

## Full Documentation

For full documents of DataKit, see

- [DataKit Doc](https://www.yuque.com/dataflux/datakit)
- [DataKit Community Docs](https://www.yuque.com/dataflux/datakit-community)
