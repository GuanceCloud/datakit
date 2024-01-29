
# C++ Examples
---

Applying DDTrace in C + + code requires modifying business code and manually embedding points in existing business code. This document demonstrates how to bury points in C + + code with a simple read file demo.

## Install Libraries and Dependencies {#dependence}
<!-- markdownlint-disable MD046 -->
=== "Linux/macOS"

    - Download DDTrace-C++ SDK
    
    ```shell
    git clone https://github.com/DataDog/dd-opentracing-cpp
    ```
    
    - Compile and install the SDK
    
    ```shell
    # Install dependencies
    cd dd-opentracing-cpp && sudo scripts/install_dependencies.sh
    
    # Compile and install
    mkdir .build && cd .build && cmake .. && make && make install
    ```

    If you have problems compiling the SDK, you can temporarily test it with the prepared header file [5] and dynamic library [6] of Guance Cloud.

=== "Windows"

    Coming Soon...

???+ attention "cmake installation"

    Cmake may not be installed to a higher version through yum or apt-get. It is recommended to download the latest version directly from its [official website][3]{:target="_blank"}. You can also use the observation cloud hosted [source] [1] or [Windows binary] [2] directly.
    
    ```shell
    # Install cmake from source
    wget https://static.guance.com/gfw/cmake-3.24.2.tar.gz
    tar -zxvf cmake-3.24.2.tar.gz
    ./bootstrap --prefix=/usr/local
    make
    make install

    # Verify the current version
    cmake --version
    cmake version 3.24.2
    ```

    Gcc install version 7. x:

    ```shell
    yum install centos-release-scl
    yum install devtoolset-7-gcc*
    scl enable devtoolset-7 bash
    which gcc
    gcc --version
    ```
<!-- markdownlint-enable -->
## C++ Code Sample {#simple-example}

The following C + + code demonstrates the basic trace burial operation, which simulates a business that reads a local disk file.

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

### Compile and Run {#build-run}

```shell
LD_LIBRARY_PATH=/usr/local/lib64 g++ -std=c++14 -o demo demo.cc -ldd_opentracing -I ./dd-opentracing-cpp/deps/include
LD_LIBRARY_PATH=/usr/local/lib64  DD_AGENT_HOST=localhost DD_TRACE_AGENT_PORT=9529 ./demo
```

Here you can put *libdd_opentracing.so* and the corresponding header file into any directory and adjust the `LD_LIBRARY_PATH` and `-I` parameters.


After running the program for a period of time, you can see trace data similar to the following in Guance Cloud:

<figure markdown>
  ![](https://static.guance.com/images/datakit/cpp-ddtrace-example.png){ width="800"}
  <figcaption>C++ trace data display</figcaption>
</figure>

## Environment Variable Support {#envs}

## Supported Environment Variables {#start-options}

The following environment variables support specifying some configuration parameters of ddtrace when starting the program, and their basic form is:

```shell
DD_XXX=<env-value> DD_YYY=<env-value> ./demo
```

Several commonly used ENVs are as follows. For more ENV support, see [DDTrace Original][7]{:target="_blank"}.

|                       Key | Default Value | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| ------------------------: | ------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
|              `DD_VERSION` | -             | Set the application version, such as *1.2.3*、*2022.02.13*                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
|           `DD_AGENT_HOST` | `localhost`   | Set DataKit address                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
|     `DD_TRACE_AGENT_PORT` | -             | Set the receiving port for DataKit trace data. You need to manually specify [HTTP port for DataKit] [4] (typically 9529)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
|                  `DD_ENV` | -             | Set the current application environment, such as prod, pre-prod, etc.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
|              `DD_SERVICE` | -             | Set the application service name                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| `DD_TRACE_SAMPLING_RULES` | -             | Here, a JSON array is used to represent the sampling setting (the sampling rate is applied in the order of the array), where `sample_rate` is the sampling rate and the value range is `[0.0, 1.0]`。<br> **Example 1**: Set the global sampling rate to 20%: `DD_TRACE_SAMPLE_RATE='[{"sample_rate": 0.2}]' ./my-app` <br>**Example 2**: If the service name is generic `app1.*`, and the span name is `abc`, the sampling rate is set to 10%, except that the sampling rate is set to 20%: `DD_TRACE_SAMPLE_RATE='[{"service": "app1.*", "name": "b", "sample_rate": 0.1}, {"sample_rate": 0.2}]' ./my-app` <br> |
|                 `DD_TAGS` | -             | Here you can inject a set of global tags that will appear in each span and profile data. Multiple tags can be separated by spaces and English commas, such as `layer:api,team:intake`、`layer:api team:intake`                                                                                                                                                                                                                                                                                                                                                                                                     |

<!-- markdownlint-disable MD053 -->
[1]: https://static.guance.com/gfw/cmake-3.24.2.tar.gz
[2]: https://static.guance.com/gfw/cmake-3.24.2-windows-x86_64.msi
[3]: https://cmake.org/download/
[4]: ../datakit/datakit-conf.md#config-http-server
[5]: https://static.guance.com/gfw/dd-cpp-include.tar.gz
[6]: https://static.guance.com/gfw/libdd_opentracing.so
[7]: https://docs.datadoghq.com/tracing/trace_collection/library_config/cpp/
<!-- markdownlint-enable -->