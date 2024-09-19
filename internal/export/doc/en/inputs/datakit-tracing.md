# Datakit Tracing Overview

The third-party Tracing data currently supported by Datakit includes:

- DDTrace
- Apache Jaeger
- OpenTelemetry
- SkyWalking
- Zipkin

---

## Datakit Tracing Frontend {#datakit-tracing-frontend}

Tracing Frontend is an API that receives data from a variety of different types of Trace, typically via HTTP or gRPC from a variety of Trace SDKs. When DataKit receives this data, it converts it into a [unified Span structure](datakit-tracing-struct.md). It is then sent to [Backend](datakit-tracing.md#datakit-tracing-backend) for processing.

In addition to transforming the Span structure, Tracing Frontend also completes the configuration of the filter unit and arithmetic unit in [Tracing Backend](datakit-tracing.md#datakit-tracing-backend).

## Tracing Data Collection Common Configuration {#tracing-common-config}

The tracer generation in the configuration file refers to the currently configured Tracing Agent, and all supported Tracing Agents can use the following configuration:

```toml
  ## customer_tags is a list of keys contains keys set by client code like span.SetTag(key, value)
  ## that want to send to data center. Those keys set by client code will take precedence over
  ## keys in [inputs.tracer.tags]. DOT(.) IN KEY WILL BE REPLACED BY DASH(_) WHEN SENDING.
  customer_tags = ["key1", "key2", ...]

  ## Keep rare tracing resources list switch.
  ## If some resources are rare enough(not presend in 1 hour), those resource will always send
  ## to data center and do not consider samplers and filters.
  keep_rare_resource = false

  ## By default every error presents in span will be send to data center and omit any filters or
  ## sampler. If you want to get rid of some error status, you can set the error status list here.
  omit_err_status = ["404"]

  ## Ignore tracing resources map like service:[resources...].
  ## The service name is the full service name in current application.
  ## The resource list is regular expressions uses to block resource names.
  ## If you want to block some resources universally under all services, you can set the
  ## service name as "*". Note: double quotes "" cannot be omitted.
  [inputs.tracer.close_resource]
    service1 = ["resource1", "resource2", ...]
    service2 = ["resource1", "resource2", ...]
    "*" = ["close_resource_under_all_services"]

  ## Sampler config uses to set global sampling strategy.
  ## sampling_rate used to set global sampling rate.
  [inputs.tracer.sampler]
    sampling_rate = 1.0

  [inputs.tracer.tags]
    key1 = "value1"
    key2 = "value2"

  ## Threads config controls how many goroutines an agent cloud start.
  ## buffer is the size of jobs' buffering of worker channel.
  ## threads is the total number fo goroutines at running time.
  ## timeout is the duration(ms) before a job can return a result.
  [inputs.tracer.threads]
    buffer = 100
    threads = 8
    timeout = 1000
```

- `customer_tags`: By default, Datakit only picks up the Tags it is interested in (that is, the fields other than message that can be seen in the observation cloud link details),
  If users are interested in other tags reported on the link, they can add a notification Datakit to this configuration to pick them up. This configuration takes precedence over  `[inputs.tracer.tags]`。
- `keep_rare_resource`: If a link from a Resource has not been present within the last hour, the system considers it a rare link and reports it directly to the Data Center.
- `omit_err_status`: By default, data is reported directly to the Data Center if there is a Span with Error status in the link, and Datakit can be told to ignore links with some HTTP Error Status (for example, 429 too many requests) if the user needs to ignore it.
- `[inputs.tracer.close_resource]`: Users can configure this to close a Resource link with [span_type](datakit-tracing-struct) as Entry.
- `[inputs.tracer.sampler]`: Configure the global sampling rate for the current Datakit, [configuration sample](datakit-tracing.md#samplers).
- `[inputs.tracer.tags]`: Configure Datakit Global Tags with a lower priority than `customer_tags` 。
- `[inputs.tracer.threads]`: Configure the thread queue of the current Tracing Agent to control the CPU and Memory resources available during data processing.
    - buffer: The cache of the work queue. The larger the configuration, the greater the memory consumption. At the same time, the request sent to the Agent has a greater probability of queuing successfully and returning quickly, otherwise it will be discarded and return a 429 error.
    - threads: The maximum number of threads in the work queue. The larger the configuration, the more threads started, the higher the CPU consumption. Generally, it is configured as the number of CPU cores.
    - timeout: The task timed out, the larger the configuration, the longer the buffer.

## Datakit Tracing Backend {#datakit-tracing-backend}

The Datakit backend is responsible for manipulating the link data as configured, and currently supported operations include Tracing Filters and Samplers.

### Datakit Filters {#filters}

- `user_rule_filter`: Datakit default filter, triggered by user behavior.
- `omit_status_code_filter`: When `omit_err_status = ["404"]` is configured, an error with a status code of 404 in a link under the HTTP service will not be reported to the Data Center.
- `penetrate_error_filter`: Datakit default filter, triggered by link error.
- `close_resource_filter`: Configured in `[inputs.tracer.close_resource]`, the service name is the full service name or `*`, and the resource name is the regular expression of the resource.
    - Example 1: Configuration such as `login_server = ["^auth\_.*\?id=[0-9]*"]`, then the `login_server` service name `resource` looks like `auth_name?id=123` will be closed
    - Example 2: If configured as `"*" = ["heart_beat"]`, the `heart_beat` resource on all services under the current Datakit will be closed.
- `keep_rare_resource_filter`: When `keep_rare_resource = true` is configured, links determined to be rare will be reported directly to the Data Center.

Filters (Sampler is also a Filter) in the current version of Datakit are executed in a fixed order:

> error status penetration --> close resource filter --> omit certain http status code list --> rare resource keeper --> sampler <br>
> Each Datakit Filter has the ability to terminate the execution link, meaning that filters that meet the termination conditions will not execute subsequent filters.

### Datakit Samplers {#samplers}

Currently, Datakit respects client sampling priority, [DDTrace Sampling Rules](https://docs.datadoghq.com/tracing/faq/trace_sampling_and_storage){:target="_blank"}。

- Case one

Take DDTrace as an example. If the sampling priority tags is configured in the DDTrace lib sdk or client and the client sampling rate is 0.3 through the environment variable (DD_TRACE_SAMPLE_RATE) or the startup parameter (dd.trace.sample.rate) and the Datakit sampling rate (inputs.tracer.sampler) is not specified, the amount of data reported to the Data Center is approximately 30% of the total.

- Case two

If the customer only configures the Datakit sampling rate (inputs.tracer.sampler), for example: sampling_rate = 0.3, then the Datakit reports about 30% of the total data to the Data Center.

**Note**: In the case of multi-service multi-Datakit distributed deployment, configuring Datakit sampling rate needs to be uniformly configured to the same sampling rate to achieve sampling effect.

- Case three

That is, the client sampling rate is configured as A and the Datakit sampling rate is configured as B, where A and B are greater than 0 and less than 1. In this case, the amount of data reported to Data Center is about A\* B% of the total amount.

**Note**: In the case of multi-service multi-Datakit distributed deployment, configuring Datakit sampling rate needs to be uniformly configured to the same sampling rate to achieve sampling effect.

## Span Structure Description {#about-span-structure}

Business explanation of how Datakit uses the [DatakitSpan](datakit-tracing-struct.md) data structure

- Refer to [Datakit Tracing Structure](datakit-tracing-struct.md) for a detailed description of the Datakit Tracing data structure.
- Multiple Datakit Span data are placed in a Datakit Trace to form a Tracing data uploaded to the Data Center and ensure that all Spans have only one TraceID.
- For DDTrace, DDTrace data with the same TraceID may be reported in batches.
- In a production environment (multi-service, multi-Datakit deployment), a complete piece of Trace data is uploaded to the Data Center in batches, not in the order of invocation.
- `parent_id = 0` is root span.
- `span_type = entry` is the caller of the first resource on the service, the first span on the current service.
