---
title     : 'DDTrace C++'
summary   : 'DDTrace C++ 集成'
tags      :
  - '链路追踪'
  - 'C/C++'
__int_icon: 'icon/ddtrace'
---

本页描述 `dd-opentracing-cpp` 的兼容接入方式。该 SDK 需要在业务代码中手动创建和结束 span；常见 C++ 中间件没有可直接复用的自动插桩，因此应先为入口请求和关键外部调用设计少量、稳定的埋点。

<!-- markdownlint-disable MD046 -->
???+ warning

    上游已将 `dd-opentracing-cpp` 标记为废弃，并推荐迁移到 `dd-trace-cpp`。本页保留它是为了维护已接入的应用。新项目应优先评估新的 SDK，并在接入 DataKit 前做协议与功能兼容性测试；不要把旧 SDK 示例当作长期技术选型建议。
<!-- markdownlint-enable MD046 -->

## 安装库和依赖 {#dependence}

<!-- markdownlint-disable MD046 -->
=== "Linux/macOS"

    - 下载 DDTrace-C++ SDK
    
    ```shell
    git clone https://github.com/DataDog/dd-opentracing-cpp
    ```
    
    - 编译、安装 SDK
    
    ```shell
    # 安装依赖
    cd dd-opentracing-cpp && sudo scripts/install_dependencies.sh
    
    # 编译并安装
    mkdir .build && cd .build && cmake .. && make && make install
    ```

    编译完成后，先在测试环境运行 SDK 自带测试，再将应用连接到 DataKit。若仅为定位构建问题，可临时使用<<<custom_key.brand_name>>>准备的[头文件][5]和[动态库][6]；生产环境仍应使用可追溯、与运行时匹配的构建产物。

=== "Windows"

    上游仓库提供早期 Windows 构建说明，但本页未提供经过 DataKit 验证的 Windows 组合。请在目标编译器、运行时库和 DataKit 网络环境中自行完成构建与链路验证后再使用。

???+ note "构建工具"

    使用发行版支持的 CMake、C++14 编译器和构建工具即可。安装前先阅读上游仓库当前的依赖说明；不要为了本示例在生产主机上随意替换系统级编译器或 CMake。
<!-- markdownlint-enable MD046 -->

## C++ 代码示例 {#simple-example}

以下示例演示最小的手动埋点。关键点是：创建 tracer 时设置服务和 DataKit 目标；每个 span 都必须 `Finish()`；进程退出前调用 `Close()` 刷新缓冲数据。

```cpp linenums="1" title="demo.cc"
#include <datadog/opentracing.h>

#include <chrono>
#include <thread>

int main() {
    datadog::opentracing::TracerOptions options;
    options.service = "cpp-demo";
    options.environment = "test";
    options.agent_host = "127.0.0.1";
    options.agent_port = 9529;

    auto tracer = datadog::opentracing::makeTracer(options);

    {
        auto span = tracer->StartSpan("read_file");
        span->SetTag("file.path", "example.txt");
        std::this_thread::sleep_for(std::chrono::milliseconds(20));
        span->Finish();
    }

    tracer->Close(); // Flush buffered spans before process exit.
    return 0;
}
```

### 编译运行 {#build-run}

```shell
c++ -std=c++14 -Wall -Wextra demo.cc -o demo \
  -I/usr/local/include -L/usr/local/lib \
  -Wl,-rpath,/usr/local/lib -ldd_opentracing

DD_AGENT_HOST=localhost \
DD_TRACE_AGENT_PORT=9529 \
DD_SERVICE=cpp-demo \
./demo
```

上例假设 `make install` 将头文件和动态库安装在 `/usr/local/include`、`/usr/local/lib`。部分发行版会使用 `lib64`；请同步调整 `-L` 与 rpath。环境变量会覆盖 `TracerOptions` 中的同名设置，适合在不重新编译的情况下切换 DataKit 地址。

程序运行一段时间后，即可在<<<custom_key.brand_name>>>看到类似如下 trace 数据：

<figure markdown>
  ![](https://static.<<<custom_key.brand_main_domain>>>/images/datakit/cpp-ddtrace-example.png){ width="800"}
  <figcaption>C++ trace 数据展示</figcaption>
</figure>

## 环境变量支持 {#envs}

### 支持的环境变量 {#start-options}

以下环境变量支持在启动程序的时候指定 DDTrace 的一些配置参数，其基本形式为：

```shell
DD_XXX=<env-value> DD_YYY=<env-value> ./demo
```

常用 ENV 如下。`DD_TRACE_AGENT_URL`（如 `http://datakit:9529`）优先于 `DD_AGENT_HOST` 和 `DD_TRACE_AGENT_PORT`。更多旧 SDK 配置见 [DDTrace 文档][7]{:target="_blank"}。

- **`DD_VERSION`**

    设置应用程序版本，如 `1.2.3`、`2022.02.13`

- **`DD_AGENT_HOST`**

    **默认值**：`localhost`

    设置 DataKit 地址；上游默认通常为 `localhost`。

- **`DD_TRACE_AGENT_PORT`**

    设置 trace 接收端口。上游默认通常是 `8126`，接入 DataKit 时需指定 [DataKit HTTP 端口][4]（通常为 `9529`）。

- **`DD_ENV`**

    设置应用当前的环境，如 prod、pre-prod 等

- **`DD_SERVICE`**

    设置应用服务名

- **`DD_TRACE_SAMPLING_RULES`**

    使用 JSON 数组定义规则，按数组顺序匹配，`sample_rate` 取值范围为 `[0.0, 1.0]`。

    **示例一**：所有 trace 采样率为 20%：`DD_TRACE_SAMPLING_RULES='[{"sample_rate":0.2}]' ./demo`

    **示例二**：服务名匹配 `app1.*` 且 span 名为 `abc` 时采样 10%，其他 trace 采样 20%：`DD_TRACE_SAMPLING_RULES='[{"service":"app1.*","name":"abc","sample_rate":0.1},{"sample_rate":0.2}]' ./demo`

- **`DD_TAGS`**

    这里可注入一组全局 tag，这些 tag 会出现在每个 span 和 profile 数据中。多个 tag 之间可以用空格和英文逗号分割，例如 `layer:api,team:intake`、`layer:api team:intake`

<!-- markdownlint-disable MD053 -->
[1]: https://static.<<<custom_key.brand_main_domain>>>/gfw/cmake-3.24.2.tar.gz
[2]: https://static.<<<custom_key.brand_main_domain>>>/gfw/cmake-3.24.2-windows-x86_64.msi
[3]: https://cmake.org/download/
[4]: ../datakit/datakit-conf.md#config-http-server
[5]: https://static.<<<custom_key.brand_main_domain>>>/gfw/dd-cpp-include.tar.gz
[6]: https://static.<<<custom_key.brand_main_domain>>>/gfw/libdd_opentracing.so
[7]: https://docs.datadoghq.com/tracing/trace_collection/library_config/cpp/
<!-- markdownlint-enable MD053 -->
