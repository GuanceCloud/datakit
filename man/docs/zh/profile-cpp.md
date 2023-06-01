# C/C++ Profiling

目前 DataKit 支持 1 种方式来采集 C/C++ profiling 数据，即 [Pyroscope](https://pyroscope.io/){:target="_blank"}。

## Pyroscope {#pyroscope}

[Pyroscope](https://pyroscope.io/){:target="_blank"} 是一款开源的持续 profiling 平台，DataKit 已经支持将其上报的 profiling 数据展示在[观测云](https://www.guance.com/){:target="_blank"}。

Pyroscope 采用 C/S 架构，运行模式分为 [Pyroscope Agent](https://pyroscope.io/docs/agent-overview/){:target="_blank"} 和 [Pyroscope Server](https://pyroscope.io/docs/server-overview/){:target="_blank"}，这两个模式均集成在一个二进制文件中，通过不同的命令行命令来展现。

这里需要的是 Pyroscope Agent 模式。DataKit 已经集成了 Pyroscope Server 功能，通过对外暴露 HTTP 接口的方式，可以接收 Pyroscope Agent 上报的 profiling 数据。

Profiling 数据流向：「Pyroscope Agent 采集 Profiling 数据 -> Datakit -> 观测云」。

### 前置条件 {#pyroscope-requirement}

- 根据 Pyroscope 官方文档 [eBPF Profiling](https://pyroscope.io/docs/ebpf/#prerequisites-for-profiling-with-ebpf){:target="_blank"}，需要 Linux 内核版本 >= 4.9 (due to [BPF_PROG_TYPE_PERF_EVENT](https://lkml.org/lkml/2016/9/1/831){:target="_blank"})。

- 已安装 [DataKit](https://www.guance.com/){:target="_blank"} 并且已开启 [profile](profile.md#config) 采集器，配置参考如下：

```toml
[[inputs.profile]]
  ## profile Agent endpoints register by version respectively.
  ## Endpoints can be skipped listen by remove them from the list.
  ## Default value set as below. DO NOT MODIFY THESE ENDPOINTS if not necessary.
  endpoints = ["/profiling/v1/input"]

  ## set true to enable election.
  election = true

  #  config
  [[inputs.profile.pyroscope]]
    # listen url
    url = "0.0.0.0:4040"

    # service name
    service = "pyroscope-demo"

    # app env
    env = "dev"

    # app version
    version = "0.0.0"

  [inputs.profile.pyroscope.tags]
    tag1 = "val1"
```

- 安装 Pyroscope

这里以 Linux AMD64 平台为例：

```sh
wget https://dl.pyroscope.io/release/pyroscope-0.36.0-linux-amd64.tar.gz
tar -zxvf pyroscope-0.36.0-linux-amd64.tar.gz
```

按照上述方法获取到的是 Pyroscope 的二进制文件，直接运行就可以了，也可以放在 [PATH](http://www.linfo.org/path_env_var.html){:target="_blank"} 下。

其它平台与架构的安装方法，见[下载地址](https://pyroscope.io/downloads/){:target="_blank"}。

### Pyroscope Agent 配置 eBPF 采集模式 {#pyroscope-ebpf}

Pyroscope Agent 的 [eBPF](https://pyroscope.io/docs/ebpf/){:target="_blank"} 模式支持 C/C++ 程序的 profiling 采集。

- 设置环境变量：

```sh
export PYROSCOPE_APPLICATION_NAME='my.ebpf.program{host=server-node-1,region=us-west-1,tag2=val2}'
export PYROSCOPE_SERVER_ADDRESS='http://localhost:4040/' # Datakit profile 配置的 pyroscope listen url.
export PYROSCOPE_SPY_NAME='ebpfspy'
```

- 根据需要 profiling 的目标使用不同的命令：

    - profiling 正在运行的程序(以 PID 为 `1000` 为例)：`sudo -E pyroscope connect --pid 1000`
    - profiling 运行指定程序(以 `mongod` 为例)：`sudo -E pyroscope exec mongod`
    - profiling 整个系统：`sudo -E pyroscope ebpf`

### 查看 Profile {#pyroscope-view}

运行上述 profiling 命令后，Pyroscope Agent 会开始采集指定的 profiling 数据，并将数据上报给观测云。稍等几分钟后就可以在观测云空间[应用性能监测 -> Profile](https://console.guance.com/tracing/profile){:target="_blank"} 查看相应数据。
