# DDTrace profiling

## 安装运行 Profiling Agent {#install}

目前支持 java 和 python 语言使用相应语言的profiling 库。

### Java {#java}

下载最新的 ddtrace agent dd-java-agent.jar

```shell
# java版本要求：java8版本需要高于8u262+，或者使用java11及以上版本
wget -O dd-java-agent.jar 'https://github.com/DataDog/dd-trace-java/releases/download/v0.107.0/dd-java-agent-0.107.0.jar'
```

运行 Java Code

```shell
java -javaagent:/<your-path>/dd-java-agent.jar \
-Ddd.service=profiling-demo \
-Ddd.env=dev \
-Ddd.version=1.2.3  \
-Ddd.profiling.enabled=true  \
-XX:FlightRecorderOptions=stackdepth=256 \
-Ddd.trace.agent.port=9529 \
-jar your-app.jar 
```

### Python {#python}

安装 DDTrace Python 函数库

```shell
pip install ddtrace
```

- 自动profiling

```shell
DD_PROFILING_ENABLED=true \
DD_ENV=dev \
DD_SERVICE=my-web-app \
DD_VERSION=1.0.3 \
DD_TRACE_AGENT_URL=http://127.0.0.1:9529 \
ddtrace-run python app.py
```

- 通过代码的方式开启profiling

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
#while True:
#    time.sleep(1)

```

此时运行则不需要再用 ddtrace-run 命令

```shell
DD_ENV=testing DD_SERVICE=python-profiling-manual DD_VERSION=7.8.9 python3 app.py
```

## 查看 Profile {#view}

上述程序启动后，会定期（默认 1 分钟上报一次）收集 profiling 数据并上报给 DataKit，稍等片刻后就可以在观测云工作空间看到 profiling 数据。
