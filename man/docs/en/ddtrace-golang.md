# Golang Example

---

## Install Dependency {#dependence}

Install the ddtrace golang library to run in the development directory.

```shell
go get -v github.com/DataDog/dd-trace-go
```

## Set DataKit {#set-datakit}

First [install][1], [start datakit][2], and open [ddtrace collector][3]

## Code Example {#code-example}

The following code demonstrates trace data collection for a file open operation.

In the `main()` entry code, set the basic trace parameters and start trace:

``` go
package main

import (
	"io/ioutil"
	"os"
	"time"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func main() {
	tracer.Start(
		tracer.WithEnv("prod"),
		tracer.WithService("test-file-read"),
		tracer.WithServiceVersion("1.2.3"),
		tracer.WithGlobalTag("project", "add-ddtrace-in-golang-project"),
	)

	// end of app exit, make sure tracer stopped
	defer tracer.Stop()

	tick := time.NewTicker(time.Second)
	defer tick.Stop()

	// your-app-main-entry...
	for {
		runApp()
		runAppWithError()

		select {
		case <-tick.C:
		}
	}
}

func runApp() {
	var err error
	// Start a root span.
	span := tracer.StartSpan("get.data")
	defer span.Finish(tracer.WithError(err))

	// Create a child of it, computing the time needed to read a file.
	child := tracer.StartSpan("read.file", tracer.ChildOf(span.Context()))
	child.SetTag(ext.ResourceName, os.Args[0])

	// Perform an operation.
	var bts []byte
	bts, err = ioutil.ReadFile(os.Args[0])
	span.SetTag("file_len", len(bts))
	child.Finish(tracer.WithError(err))
}

func runAppWithError() {
	var err error
	// Start a root span.
	span := tracer.StartSpan("get.data")

	// Create a child of it, computing the time needed to read a file.
	child := tracer.StartSpan("read.file", tracer.ChildOf(span.Context()))
	child.SetTag(ext.ResourceName, "somefile-not-found.go")

	defer func() {
		child.Finish(tracer.WithError(err))
		span.Finish(tracer.WithError(err))
	}()

	// Perform an error operation.
	if _, err = ioutil.ReadFile("somefile-not-found.go"); err != nil {
		// error handle
	}
}
```

### Compile and Run {#run}

=== "Linux/Mac"

    ```shell
    go build main.go -o my-app
    DD_AGENT_HOST=localhost DD_TRACE_AGENT_PORT=9529 ./my-app
    ```

=== "Windows"

    ```powershell
    go build main.go -o my-app.exe
    $env:DD_AGENT_HOST="localhost"; $env:DD_TRACE_AGENT_PORT="9529"; .\my-app.exe
    ```

After running the program for a period of time, you can see trace data similar to the following in Guance Cloud:

<figure markdown>
  ![](https://zhuyun-static-files-production.oss-cn-hangzhou.aliyuncs.com/images/datakit/golang-ddtrace-example.png){ width="800"}
  <figcaption>Golang Program trace Data Display</figcaption>
</figure>

## Supported Environment Variable {#start-options}

The following environment variables support specifying some configuration parameters of ddtrace when starting the program, and their basic form is:

```shell
DD_XXX=<env-value> DD_YYY=<env-value> ./my-app
```

???+ attention

    These environment variables will be overwritten by the corresponding fields injected with `WithXXX()` in the code, so the configuration of code injection has higher priority, and these ENVs will only take effect if the corresponding fields are not specified in the code.

| Key                       | Default Value      | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| :---                      | :--         | :--                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| `DD_VERSION`              | -           | Set the application version, such as *1.2.3*, *2022.02.13*                                                                                                                                                                                                                                                                                                                                                                                                           |
| `DD_SERVICE`              | -           | Set the application service name                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| `DD_ENV`                  | -           | Set the current application environment, such as prod, pre-prod, etc.                                                                                                                                                                                                                                                                                                                                                                                                             |
| `DD_AGENT_HOST`           | `localhost` | Set the IP address of the DataKit, and the trace data generated by the application will be sent to the DataKit                                                                                                                                                                                                                                                                                                                                                                                    |
| `DD_TRACE_AGENT_PORT`     | -           | Set the receiving port for DataKit trace data. You need to manually specify [HTTP port for DataKit][4](typically 9529）                                                                                                  |
| `DD_DOGSTATSD_PORT`       | -           | If you want to receive statsd data generated by ddtrace, you need to manually open [statsd collector][5]|
| `DD_TRACE_SAMPLING_RULES` | -           | Here, a JSON array is used to represent the sampling setting (the sampling rate is applied in the order of the array), where `sample_rate` is the sampling rate and the value range is `[0.0, 1.0]`。<br> **Example 1**: Set the global sampling rate to 20%: `DD_TRACE_SAMPLE_RATE='[{"sample_rate": 0.2}]' ./my-app` <br>**Example 2** If the service name is generic `app1.*`, and the span name is `abc` , the sampling rate is set to 10%, except that the sampling rate is set to 20%: `DD_TRACE_SAMPLE_RATE='[{"service": "app1.*", "name": "b", "sample_rate": 0.1}, {"sample_rate": 0.2}]' ./my-app` <br> |
| `DD_TRACE_SAMPLE_RATE`    | -           | Turn on the sample rate switch above                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| `DD_TRACE_RATE_LIMIT`     | -           | Set the number of span samples per second per golang process. If `DD_TRACE_SAMPLE_RATE` is already on, the default is 100.                                                                                                                                                                                                                                                                                                                                |
| `DD_TAGS`                 | -           | Here you can inject a set of global tags that will appear in each span and profile data. Multiple tags can be separated by spaces and English commas, for example `layer:api,team:intake`, `layer:api team:intake`.                                                                                                                                                                                                                                                                                   |
| `DD_TRACE_STARTUP_LOGS`   | `true`      | Open configuration and diagnostic logs related to ddtrace.                                                                                                                                                                                                                                                                                                                                                                                                                   |
| `DD_TRACE_DEBUG`          | `false`     | Open debug logs related to ddtrace.                                                                                                                                                                                                                                                                                                                                                                               |
| `DD_TRACE_ENABLED`        | `true`      | Turn on the trace switch. If you turn the switch off manually, no trace data will be generated.                                                                                                                                                                                                                                                                                                                                                                                    |
| `DD_SERVICE_MAPPING`      | -           | Rename service name dynamically, and each service name mapping can be separated by spaces and English commas, such as `mysql:mysql-service-name,postgres:postgres-service-name`, `mysql:mysql-service-name postgres:postgres-service-name`                                                                                                                                                                                                                                                                  |

[1]: /datakit/datakit-install/
[2]: /datakit/datakit-service-how-to/
[3]: /datakit/ddtrace/#config
[4]: /datakit/datakit-conf/#config-http-server
[5]: /datakit/statsd/
