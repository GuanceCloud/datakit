{{.CSS}}

# C++

---

在 C++ 代码中应用 DDTrace，需要修改业务代码，需手动在现有业务代码中进行埋点。本文档以一个简单的读取文件 demo 来演示如何在 C++ 代码中进行埋点。

## 安装库和依赖 {#dependence}

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

    如果编译 SDK 过程中有问题，可临时用观测云已经准备好的[头文件][5]以及 [动态库][6] 来测试。

=== "Windows"

    Comming Soon...

???+ attention "cmake 安装"

    cmake 可能无法通过 yum 或 apt-get 安装到较高的版本，建议直接去其[官网][3]{:target="_blank"}下载最新版本。也可以直接使用观测云托管的[源码][1]或[Windows 二进制][2]。
    
    ```shell
    # 从源码安装 cmake
    wget https://static.guance.com/gfw/cmake-3.24.2.tar.gz
    tar -zxvf cmake-3.24.2.tar.gz
    ./bootstrap --prefix=/usr/local
    make
    make install

    # 验证一下当前版本
    cmake --version
    cmake version 3.24.2
    ```

    gcc 安装 7.x 的版本：

    ```shell
    yum install centos-release-scl
    yum install devtoolset-7-gcc*
    scl enable devtoolset-7 bash
    which gcc
    gcc --version
    ```

## C++ 代码示例 {#simple-example}

以下 C++ 代码演示了基本的 trace 埋点操作，其模拟的业务是一个读取本地磁盘文件的操作。

```cpp
#include <datadog/opentracing.h>
#include <datadog/tags.h>

#include <stdarg.h>
#include <memory> 
#include <chrono>
#include <fstream>
#include <string>
#include <chrono>
#include <thread>
#include <sys/stat.h>

datadog::opentracing::TracerOptions tracer_options{"", 0, "compiled-in-example"};
auto tracer = datadog::opentracing::makeTracer(tracer_options);

std::ifstream::pos_type filesize(const char* filename) {
	std::ifstream in(filename, std::ifstream::ate | std::ifstream::binary);
	return in.tellg(); 
}

std::string string_format(const std::string fmt_str, ...) {
	int final_n, n = ((int)fmt_str.size()) * 2;
	std::unique_ptr<char[]> formatted;
	va_list ap;
	while(1) {
		formatted.reset(new char[n]);
		strcpy(&formatted[0], fmt_str.c_str());
		va_start(ap, fmt_str);
		final_n = vsnprintf(&formatted[0], n, fmt_str.c_str(), ap);
		va_end(ap);
		if (final_n < 0 || final_n >= n)
			n += abs(final_n - n + 1);
		else
			break;
	}
	return std::string(formatted.get());
}

int runApp(const char* f) {
	auto span_a = tracer->StartSpan(f);
	span_a->SetTag(datadog::tags::environment, "production");
	span_a->SetTag("tag", "app-ok");
	span_a->SetTag("file_len", string_format("%d", filesize(f)));
}

int main(int argc, char* argv[]) {
	for (;;) {
		runApp(argv[0]);
		runApp("file-not-exists");
		std::this_thread::sleep_for(std::chrono::milliseconds(1000));
	}

	tracer->Close();
	return 0;
} 
```

### 编译运行 {#build-run}

```shell
LD_LIBRARY_PATH=/usr/local/lib64 g++ -std=c++14 -o demo demo.cc -ldd_opentracing -I ./dd-opentracing-cpp/deps/include
LD_LIBRARY_PATH=/usr/local/lib64  DD_AGENT_HOST=localhost DD_TRACE_AGENT_PORT=9529 ./demo
```

此处可以将 *libdd_opentracing.so* 以及对应的头文件放到任意目录，然后调整 `LD_LIBRARY_PATH` 以及 `-I` 参数即可。


程序运行一段时间后，即可在观测云看到类似如下 trace 数据：

<figure markdown>
  ![](https://zhuyun-static-files-production.oss-cn-hangzhou.aliyuncs.com/images/datakit/cpp-ddtrace-example.png){ width="800"}
  <figcaption>C++ trace 数据展示</figcaption>
</figure>

## 环境变量支持 {#envs}

## 支持的环境变量 {#start-options}

以下环境变量支持在启动程序的时候指定 ddtrace 的一些配置参数，其基本形式为：

```shell
DD_XXX=<env-value> DD_YYY=<env-value> ./demo
```

常用的几个 ENV 如下。更多 ENV 支持，可参见 [DDTrace 原始文档][7]{:target="_blank"}。

| Key                       | 默认值      | 说明                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| ---:                      | ---         | ---                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| `DD_VERSION`              | -           | 设置应用程序版本，如 *1.2.3*、*2022.02.13*                                                                                                                                                                                                                                                                                                                                                                                                              |
| `DD_AGENT_HOST`           | `localhost` | 设置 DataKit 地址                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| `DD_TRACE_AGENT_PORT`     | -           | 设置 DataKit trace 数据的接收端口。这里需手动指定 [DataKit 的 HTTP 端口][4]（一般为 9529）                                                                                                                                                                                                                                                                                                                                                              |
| `DD_ENV`                  | -           | 设置应用当前的环境，如 prod、pre-prod 等                                                                                                                                                                                                                                                                                                                                                                                                                |
| `DD_SERVICE`              | -           | 设置应用服务名                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| `DD_TRACE_SAMPLING_RULES` | -           | 这里用 JSON 数组来表示采样设置（采样率应用以数组顺序为准），其中 `sample_rate` 为采样率，取值范围为 `[0.0, 1.0]`。<br> **示例一**：设置全局采样率为 20%：`DD_TRACE_SAMPLE_RATE='[{"sample_rate": 0.2}]' ./my-app` <br>**示例二**：服务名通配 `app1.*`、且 span 名称为 `abc`的，将采样率设置为 10%，除此之外，采样率设置为 20%：`DD_TRACE_SAMPLE_RATE='[{"service": "app1.*", "name": "b", "sample_rate": 0.1}, {"sample_rate": 0.2}]' ./my-app` <br> |
| `DD_TAGS`                 | -           | 这里可注入一组全局 tag，这些 tag 会出现在每个 span 和 profile 数据中。多个 tag 之间可以用空格和英文逗号分割，例如 `layer:api,team:intake`、`layer:api team:intake`                                                                                                                                                                                                                                                                                   |

[1]: https://static.guance.com/gfw/cmake-3.24.2.tar.gz
[2]: https://static.guance.com/gfw/cmake-3.24.2-windows-x86_64.msi
[3]: https://cmake.org/download/
[4]: /datakit/datakit-conf/#config-http-server
[5]: https://static.guance.com/gfw/dd-cpp-include.tar.gz
[6]: https://static.guance.com/gfw/libdd_opentracing.so
[7]: https://docs.datadoghq.com/tracing/trace_collection/library_config/cpp/

