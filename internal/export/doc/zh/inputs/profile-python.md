# Python profiling

目前 DataKit Python profiling 支持 [dd-trace-py](https://github.com/DataDog/dd-trace-py){:target="_blank"} 和 [py-spy](https://github.com/benfred/py-spy){:target="_blank"} 两种性能采集器。

## 前置条件 {#py-spy-requirement}

已安装 [DataKit](https://www.guance.com){:target="_blank"} 并且已开启 [profile](profile.md#config) 采集器。

## dd-trace-py 接入 {#ddtrace}

`dd-trace-py` 是由 DataDog 推出的链路跟踪和性能分析开源库，能够收集 CPU、内存、阻塞等指标。

- 安装 dd-trace-py 库

<!-- markdownlint-disable MD046 -->
???+ note "版本要求"

    Datakit 目前支持 `dd-trace-py 1.14.x` 及以下版本，更高版本未经系统性测试，兼容性未知。
<!-- markdownlint-enable -->

```shell
pip3 install ddtrace
```

- 无侵入式 profiling

```shell
DD_PROFILING_ENABLED=true \
DD_ENV=dev \
DD_SERVICE=my-web-app \
DD_VERSION=1.0.3 \
DD_TRACE_AGENT_URL=http://127.0.0.1:9529 \
ddtrace-run python app.py
```

- 手动注入代码的方式开启 profiling

```python
import time
import ddtrace
from ddtrace.profiling import Profiler

ddtrace.tracer.configure(
     https=False,
     hostname="localhost",
     port="9529",
)

prof = Profiler()
prof.start(True, True)

# your code here ...
# while True:
#     time.sleep(1)
```

此时启动项目则无需再加 `ddtrace-run` 命令：

```shell
DD_ENV=testing DD_SERVICE=python-profiling-manual DD_VERSION=1.2.3 python3 app.py
```

### 查看 Profile {#view}

程序启动后，DDTrace 会定期（默认 1 分钟上报一次）收集数据并上报给 Datakit，稍等几分钟后就可以在观测云空间[应用性能监测 -> Profile](https://console.guance.com/tracing/profile){:target="_blank"} 查看相应数据。

## `py-spy` 接入 {#py-spy}

### 主机环境下使用 {#py-spy-on-host}

`py-spy` 是由开源社区提供的一款无侵入式的 Python 性能指标采样工具，具有单独运行和对目标程序负载影响低等优点。默认情况下 `py-spy` 会根据指定的参数输出不同格式的采样数据到本地文件，为简化 `py-spy` 和 DataKit 的集成，观测云提供了一个分支版本 [`py-spy-for-datakit`](https://github.com/GuanceCloud/py-spy-for-datakit){:target="_blank"}， 在原版本基础上做了少量修改，支持自动把 profiling
数据发送到 DataKit。

- 安装

推荐使用 pip 安装

```shell
pip3 install py-spy-for-datakit
```

此外，[Github Release](https://github.com/GuanceCloud/py-spy-for-datakit/releases){:target="_blank"} 页面上提供了部分主流平台的预编译版本，你也可以下载之后
用 pip 安装，下面以 Linux x86_64 平台为例（其他平台类似），介绍一下预编译版本的安装步骤

```shell
# 下载对应平台的预编译包
curl -SL https://github.com/GuanceCloud/py-spy-for-datakit/releases/download/v0.3.15/py_spy_for_datakit-0.3.15-py2.py3-none-manylinux_2_5_x86_64.manylinux1_x86_64.whl -O

# 用 pip 安装
pip3 install --force-reinstall --no-index --find-links . py-spy-for-datakit

# 验证安装是否成功
py-spy-for-datakit help
```

如果你的系统安装了 rust 和 cargo，也可以使用 cargo 来安装

```shell
cargo install py-spy-for-datakit
```

- 使用

`py-spy-for-datakit` 在 `py-spy` 原有子命令的基础上增加了 `datakit` 命令，专门用于把采样数据发送给 DataKit，可以输入 `py-spy-for-datakit help datakit` 来查看使用帮助：

| 参数                 | 说明                                             | 默认值                  |
|--------------------|------------------------------------------------|----------------------|
| -H, --host         | 要发往数据的 Datakit 监听的地址                           | 127.0.0.1            |
| -P, --port         | 要发往数据的 Datakit 监听端口                            | 9529                 |
| -S, --service      | 项目名称，可用于后台区分不同的项目，且可以用于筛选和查询，建议设置              | unnamed-service      |
| -E, --env          | 项目所部署的环境，可以用于区分开发、测试和生产环境，也可以用于筛选，建议设置         | unnamed-env          |
| -V, --version      | 项目版本，可以用于后台查询和筛选，建议设置                          | unnamed-version      |
| -p, --pid          | 需要分析的 Python 程序的进程 PID                         | 进程 PID 和项目启动命令必须指定其一 |
| -d, --duration     | 采样的持续时长，每间隔该时间段向 Datakit 发送一次数据，单位秒，最小可以设置为 10 | 60                   |
| -r, --rate         | 采样频率，每秒采样次数                                    | 100                  |
| -s, --subprocesses | 是否分析子进程                                        | false                |
| -i, --idle         | 是否采样非运行状态的线程                                   | false                |

`py-spy-for-datakit` 可以分析当前正在运行的程序，使用 `--pid <PID>` 或 `-p <PID>` 参数把正在运行的 Python 程序的进程 PID 传递给 `py-spy-for-datakit` 即可。

假设你的 Python 应用当前运行的进程 PID 为 12345， Datakit 监听在 127.0.0.1:9529，则使用命令类似如下：

```shell
py-spy-for-datakit datakit \
  --host 127.0.0.1 \
  --port 9529 \
  --service <your-service-name> \
  --env testing \
  --version v0.1 \
  --duration 60 \
  --pid 12345
```

如果提示需要 `sudo` 权限，请在命令前加上 `sudo` 即可。

`py-spy-for-datakit` 同时也支持直接跟 Python 项目的启动命令，这样就无须指定进程 PID，同时程序启动时就会进行数据采样，这时的运行命令类似：

```shell
py-spy-for-datakit datakit \
  --host 127.0.0.1 \
  --port 9529 \
  --service your-service-name \
  --env testing \
  --version v0.1 \
  -d 60 \
  -- python3 server.py  # 注意这里 python3 之前需额外加一个空格
```

如果没有发生错误，稍等一两分钟后即可在观测云平台 [应用性能监测 -> Profile](https://console.guance.com/tracing/profile){:target="_blank"} 页面查看具体的性能指标数据。

### k8s 环境下使用 {#py-spy-on-k8s}

请参考 [使用 `datakit-operator` 注入 `py-spy`](../datakit/datakit-operator.md#inject-py-spy){:target="_blank"}。
