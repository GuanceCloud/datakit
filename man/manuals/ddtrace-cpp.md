{{.CSS}}
# C++ 示例
---

## 安装库和依赖 {#dependence}

Datadog tracing 可以通过以下两种方式启动：

- 通过编译 dd-opentracing-cpp，完成对 Datadog lib 的编译与配置。
- 在运行时动态加载 Datadog OpenTracing library 并通过 JSON 进行配置。

### 编译 {#compile}

```shell
# Gets the latest release version number from GitHub.
get_latest_release() {
  wget -qO- "https://api.github.com/repos/$1/releases/latest" |
    grep '"tag_name":' |
    sed -E 's/.*"([^"]+)".*/\1/';
}
DD_OPENTRACING_CPP_VERSION="$(get_latest_release DataDog/dd-opentracing-cpp)"
# Download and install dd-opentracing-cpp library.
wget https://github.com/DataDog/dd-opentracing-cpp/archive/${DD_OPENTRACING_CPP_VERSION}.tar.gz -O dd-opentracing-cpp.tar.gz
mkdir -p dd-opentracing-cpp/.build
tar zxvf dd-opentracing-cpp.tar.gz -C ./dd-opentracing-cpp/ --strip-components=1
cd dd-opentracing-cpp/.build
# Download and install the correct version of opentracing-cpp, & other deps.
../scripts/install_dependencies.sh
cmake ..
make
make install
```

在 Cpp 代码中 include <datadog/opentracing.h> 并启动 tracer：

```cpp
// tracer_example.cpp
#include <datadog/opentracing.h>
#include <iostream>
#include <string>

int main(int argc, char* argv[]) {
  datadog::opentracing::TracerOptions tracer_options{"localhost", 8126, "compiled-in example"};
  auto tracer = datadog::opentracing::makeTracer(tracer_options);

  // Create some spans.
  {
    auto span_a = tracer->StartSpan("A");
    span_a->SetTag("tag", 123);
    auto span_b = tracer->StartSpan("B", {opentracing::ChildOf(&span_a->context())});
    span_b->SetTag("tag", "value");
  }

  tracer->Close();
  return 0;
}
```

**Note:** 链接 libdd_opentracing 和 libopentracing，保证这两个 lib 在 LD_LIBRARY_PATH 中。

```shell
g++ -std=c++14 -o tracer_example tracer_example.cpp -ldd_opentracing -lopentracing
./tracer_example
```

### 动态加载 {#dynamic-loading}

```shell
get_latest_release() {
  wget -qO- "https://api.github.com/repos/$1/releases/latest" |
    grep '"tag_name":' |
    sed -E 's/.*"([^"]+)".*/\1/';
}
DD_OPENTRACING_CPP_VERSION="$(get_latest_release DataDog/dd-opentracing-cpp)"
OPENTRACING_VERSION="$(get_latest_release opentracing/opentracing-cpp)"
# Download and install OpenTracing-cpp
wget https://github.com/opentracing/opentracing-cpp/archive/${OPENTRACING_VERSION}.tar.gz -O opentracing-cpp.tar.gz
mkdir -p opentracing-cpp/.build
tar zxvf opentracing-cpp.tar.gz -C ./opentracing-cpp/ --strip-components=1
cd opentracing-cpp/.build
cmake ..
make
make install
# Install dd-opentracing-cpp shared plugin.
wget https://github.com/DataDog/dd-opentracing-cpp/releases/download/${DD_OPENTRACING_CPP_VERSION}/linux-amd64-libdd_opentracing_plugin.so.gz
gunzip linux-amd64-libdd_opentracing_plugin.so.gz -c > /usr/local/lib/libdd_opentracing_plugin.so
```

在 Cpp 代码中 include <opentracing/dynamic_load.h> 并在 libdd_opentracing_plugin.so 中加载 tracer：

```cpp
// tracer_example.cpp
#include <opentracing/dynamic_load.h>
#include <iostream>
#include <string>

int main(int argc, char* argv[]) {
  // Load the tracer library.
  std::string error_message;
  auto handle_maybe = opentracing::DynamicallyLoadTracingLibrary(
      "/usr/local/lib/libdd_opentracing_plugin.so", error_message);
  if (!handle_maybe) {
    std::cerr << "Failed to load tracer library " << error_message << "\n";
    return 1;
  }

  // Read in the tracer's configuration.
  std::string tracer_config = R"({
      "service": "dynamic-load example",
      "agent_host": "localhost",
      "agent_port": 8126
    })";

  // Construct a tracer.
  auto& tracer_factory = handle_maybe->tracer_factory();
  auto tracer_maybe = tracer_factory.MakeTracer(tracer_config.c_str(), error_message);
  if (!tracer_maybe) {
    std::cerr << "Failed to create tracer " << error_message << "\n";
    return 1;
  }
  auto& tracer = *tracer_maybe;

  // Create some spans.
  {
    auto span_a = tracer->StartSpan("A");
    span_a->SetTag("tag", 123);
    auto span_b = tracer->StartSpan("B", {opentracing::ChildOf(&span_a->context())});
    span_b->SetTag("tag", "value");
  }

  tracer->Close();
  return 0;
}
```

**Note:** 链接 libopentracing, 确保 libopentracing.so 在 LD_LIBRARY_PATH 中:

```shell
g++ -std=c++11 -o tracer_example tracer_example.cpp -lopentracing
./tracer_example
```

## 运行 {#run}

配置 dd agent 主机地址和端口号环境变量然后启动编译好的二进制。

```shell
DD_AGENT_HOST=localhost \
DD_TRACE_AGENT_PORT=9529 \
./run_your_cpp_binary_with_parameters
```

## 环境变量支持 {#envs}

- DD_ENV: 为服务设置环境变量。
- DD_VERSION: APP 版本号。
- DD_SERVICE: 用于设置应用程序的服务名称，也可以通过 TracerOptions 或者 JSON 配置。
- DD_SERVICE_MAPPING: 定义服务名映射用于在 Tracing 里重命名服务。
- DD_TAGS: 为每个 Span 添加默认 Tags。
- DD_AGENT_HOST: Datakit 监听的地址名，默认 localhost。
- DD_AGENT_PORT: Datakit 监听的端口号，默认 9529。
- DD_TRACE_SAMPLE_RATE: 设置采样率从 0.0(0%) ~ 1.0(100%)。
