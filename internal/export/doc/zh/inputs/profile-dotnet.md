# .NET profiling

---

从 Datakit [:octicons-tag-24: Version-1.12.0](../datakit/changelog.md#cl-1.12.0){:target="_blank"} 开始支持使用 [`dd-trace-dotnet`](https://github.com/DataDog/dd-trace-dotnet){:target="_blank"} 作为 `.NET` 平台的应用性能监测工具。

## 前置条件 {#prerequisites}

- `.NET Framework 4.5.2+` / `.NET Core 2.1, 3.1, 5, 6, 7`
- `Linux with glibc 2.17+` / `Windows 10+` / `Windows Server 2012+`

## 安装 `dd-trace-dotnet` {#install-dd-trace-dotnet}

<!-- markdownlint-disable MD046 -->
=== "Linux"

    引入 `Datadog.Trace.Bundle` 包：

    ```shell
    dotnet add package Datadog.Trace.Bundle
    ```

=== "Windows"

    从 [`https://github.com/DataDog/dd-trace-dotnet/releases`](https://github.com/DataDog/dd-trace-dotnet/releases){:target="_blank"} 
    页面下载 `datadog-dotnet-apm-<VERSION>-x64.msi` 文件，如 `datadog-dotnet-apm-2.34.0-x64.msi`，然后使用管理员权限安装。


???+ Note

    当前最高支持到 `dd-trace-dotnet v2.34.0` 版本，更高的版本没有经过系统性测试，兼容性未知，如您在使用中遇到任何问题，可与我们联系。
<!-- markdownlint-enable -->


## 开启 Profiling {#start-profiling}

### Linux 平台 {#on-linux}

进入项目编译或发布后的输出目录，会存在一个 `datadog` 目录，该目录就是存放 `dd-trace-dotnet` 编译后生成的目标文件，不同的子目录适用不同的平台。

<!-- markdownlint-disable MD046 -->
???+ Note "目录说明"

    .NET Core 编译（build）后的输出目录通常位于项目根目录下的 *./bin/<Configuration\>/<Framework\>*，可以用参数指定 `-o|--output <OUTPUT_DIR>`，在本文档中假设为 *./bin/Release/net7.0*。
    .NET Core 发布（publish）后的输出目录默认位于项目根目录下的 *./bin/<Configuration\>/<Framework\>/publish*，同样可以用参数指定 `-o|--output <OUTPUT_DIR>`，在本文档中假设为 *./bin/Release/net7.0/publish*。
<!-- markdownlint-enable -->


```shell
$ ls -l bin/Release/net7.0/datadog/
total 8
-rwxr--r--   1 zy  staff  101  7  3 21:06 createLogPath.sh
drwxr-xr-x   8 zy  staff  256  7 27 16:16 linux-arm64
drwxr-xr-x   8 zy  staff  256  7 27 16:16 linux-musl-x64
drwxr-xr-x   8 zy  staff  256  7 27 16:16 linux-x64
drwxr-xr-x   8 zy  staff  256  7 27 16:16 net461
drwxr-xr-x   7 zy  staff  224  7 27 16:16 net6.0
drwxr-xr-x   7 zy  staff  224  7 27 16:16 netcoreapp3.1
drwxr-xr-x  12 zy  staff  384  7 27 16:16 netstandard2.0
drwxr-xr-x   6 zy  staff  192  7 27 16:16 osx
drwxr-xr-x   7 zy  staff  224  7 27 16:16 win-x64
drwxr-xr-x   7 zy  staff  224  7 27 16:16 win-x86
```

设置 `DDTRACE_HOME` 环境变量：

```shell
DDTRACE_HOME="$(pwd)/bin/Release/net7.0/datadog"
```

检查环境变量设置是否正确：

```shell
ls -l $DDTRACE_HOME
```

设置环境变量并启动项目：

```shell
DD_DOTNET_TRACER_HOME="$DDTRACE_HOME" \
CORECLR_ENABLE_PROFILING=1 \
CORECLR_PROFILER="{846F5F1C-F9AE-4B07-969E-05C26BC060D8}" \
CORECLR_PROFILER_PATH="$DDTRACE_HOME/linux-x64/Datadog.Trace.ClrProfiler.Native.so" \
LD_PRELOAD="$DDTRACE_HOME/linux-x64/Datadog.Linux.ApiWrapper.x64.so" \
DD_PROFILING_ENABLED=1 \
DD_PROFILING_WALLTIME_ENABLED=1 \
DD_PROFILING_CPU_ENABLED=1 \
DD_PROFILING_EXCEPTION_ENABLED=1 \
DD_PROFILING_ALLOCATION_ENABLED=1 \
DD_PROFILING_LOCK_ENABLED=1 \
DD_PROFILING_HEAP_ENABLED=1 \
DD_PROFILING_GC_ENABLED=1 \
DD_SERVICE=dotnet-profiling-demo DD_ENV=testing DD_VERSION=1.2.3 \
DD_AGENT_HOST=127.0.0.1 DD_TRACE_AGENT_PORT=9529 \
dotnet bin/Release/net7.0/<your-project-name>.dll
```

<!-- markdownlint-disable MD046 -->
???+ Note

    如果你当前的架构是 `Linux arm64`，则需要修改设置 `CORECLR_PROFILER_PATH="$DDTRACE_HOME/linux-arm64/Datadog.Trace.ClrProfiler.Native.so"` 和
    `LD_PRELOAD="$DDTRACE_HOME/linux-arm64/Datadog.Linux.ApiWrapper.x64.so"`
<!-- markdownlint-enable -->

稍等几分钟后便可以在 [观测云控制台](https://console.guance.com/tracing/profile){:target="_blank"} 查看相关数据。

### Windows IIS {#on-windows-iis}

- 在服务器上安装 [*datadog-dotnet-apm-2.34.0-x64.msi*](https://github.com/DataDog/dd-trace-dotnet/releases/download/v2.34.0/datadog-dotnet-apm-2.34.0-x64.msi){:target="_blank"} 组件。

- 编辑部署在 `IIS` 上的项目根路径下的 `web.config` 配置文件，在 `<aspNetCore>...</aspNetCore>` 标签内添加 `<environmentVariables></environmentVariables>` 标签（如已存在则忽略此步），加入如下环境变量：

```xml
<?xml version="1.0" encoding="utf-8"?>
<configuration>
    <location path="." inheritInChildApplications="false">
        <system.webServer>
            <handlers>
                <add name="aspNetCore" path="*" verb="*" modules="AspNetCoreModuleV2" resourceType="Unspecified" />
            </handlers>
            <aspNetCore processPath=".\dotnet-profiling-demo.exe" stdoutLogEnabled="false" stdoutLogFile=".\logs\stdout" hostingModel="InProcess">
                
              <environmentVariables>
                  <!-- 此处添加以下环境变量，其中 DD_ENV DD_SERVICE DD_VERSION 值根据你的实际情况调整 -->
                  <!-- DD_AGENT_HOST 和 DD_TRACE_AGENT_PORT 分别为 Datakit 监听的地址和端口 -->
                <environmentVariable name="CORECLR_ENABLE_PROFILING" value="1" />
                <environmentVariable name="CORECLR_PROFILER" value="{846F5F1C-F9AE-4B07-969E-05C26BC060D8}" />
                <environmentVariable name="DD_PROFILING_ENABLED" value="1" />
                <environmentVariable name="DD_PROFILING_CPU_ENABLED" value="1" />
                <environmentVariable name="DD_PROFILING_WALLTIME_ENABLED" value="1" />
                <environmentVariable name="DD_PROFILING_ALLOCATION_ENABLED" value="1" />
                <environmentVariable name="DD_PROFILING_HEAP_ENABLED" value="1" />
                <environmentVariable name="DD_PROFILING_EXCEPTION_ENABLED" value="1" />
                <environmentVariable name="DD_PROFILING_LOCK_ENABLED" value="1" />
                <environmentVariable name="DD_PROFILING_GC_ENABLED" value="1" />
                <environmentVariable name="DD_ENV" value="production" />
                <environmentVariable name="DD_VERSION" value="1.2.3" />
                <environmentVariable name="DD_SERVICE" value="my-dotnet-core-app" />
                <environmentVariable name="DD_AGENT_HOST" value="127.0.0.1" />
                <environmentVariable name="DD_TRACE_AGENT_PORT" value="9529" />
              </environmentVariables>
                
            </aspNetCore>
        </system.webServer>
    </location>
</configuration>
```

- 重启 `IIS` 服务器并访问你的项目。

```shell
net stop /y was
net start w3svc
```

稍等几分钟后便可以在 [观测云控制台](https://console.guance.com/tracing/profile){:target="_blank"} 查看相关数据。
