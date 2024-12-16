---
icon: zy/datakit
---

# DataKit
---

## Overview {#intro}

DataKit is an open source and integrated data collection Agent, which provides full-platform operating system (Linux/Windows/macOS) support and has comprehensive data collection capabilities, covering various scenarios such as host, container, middleware, Tracing, log and security inspection.

## Main Features {#features}

- Support data collection of metrics, logs and Tracing in host, middleware, log, APM and other fields
- Complete support for Kubernetes cloud native ecology
- [Pipeline](../pipeline/index.md): Easy structured data extraction
- Support access to other third-party data collection
    - [Telegraf](../integrations/telegraf.md)
    - [Prometheus](../integrations/prom.md)
    - [Statsd](../integrations/statsd.md)
    - [Fluentd](../integrations/logstreaming.md)
    - [Filebeat](../integrations/beats_output.md)
    - [Function](https://func.guance.com/doc/practice-write-data-via-datakit/){:target="_blank"}
    - Tracing
        - [OpenTelemetry](../integrations/opentelemetry.md)
        - [DDTrace](../integrations/ddtrace.md)
        - [Zipkin](../integrations/zipkin.md)
        - [Jaeger](../integrations/jaeger.md)
        - [SkyWalking](../integrations/skywalking.md)

## Description {#spec}

### Experimental Functionality {#experimental}

When DataKit is released, it will bring some experimental functions. These functions are often new functions released for the first time. The implementation of these functions may be lacking in consideration or imprecise. Therefore, when using experimental functions, the following possible situations should be considered:

- The function is unstable.
- For some functional configurations, compatibility is not guaranteed during subsequent iterations.
- The functionality may be removed due to its limitations, but there will be other corresponding measures to meet the corresponding requirements.

For this part of the function, please use it carefully.

In the process of using the experimental function, related questions can be submitted to issue

- [GitLab](https://gitlab.jiagouyun.com/cloudcare-tools/datakit/-/issues/new?issue%5Bmilestone_id%5D=){:target="_blank"}
- [Github](https://github.com/GuanceCloud/datakit/issues/new){:target="_blank"}
- [JihuLab](https://jihulab.com/guance-cloud/datakit/-/issues/new){:target="_blank"}

### Legend Description {#legends}

| Legend                                                                                                                       | Description                                                          |
| ---                                                                                                                        | ---                                                           |
| :fontawesome-solid-flag-checkered:                                                                                         | Indicates that the collector supports election                                          |
| :fontawesome-brands-linux: :fontawesome-brands-windows: :fontawesome-brands-apple: :material-kubernetes: :material-docker: | Examples are used to represent Linux, Windows, macOS, Kubernetes, and Docker respectively |

## Precautions {#disclaimer}

When using Datakit, there may be some impacts on existing systems as follows:

1. Log collection can lead to high-speed disk reads; the larger the volume of logs, the higher the IOPS.
1. If the RUM SDK is integrated into Web/App applications, there will be continuous uploads of RUM-related data. If there are bandwidth limitations on the uploads, it may cause Web/App pages to become unresponsive.
1. After enabling eBPF, due to the large amount of data collected, it will consume a certain amount of memory and CPU resources. In particular, when bpf-netlog is enabled, it generates a large number of logs based on all TCP packets from the host and container network interfaces.
1. During periods of high Datakit activity (when configured with large number of log/tracing collection, etc.), it will occupy a significant amount of CPU and memory resources. It is recommended to set reasonable cgroup limits on Datakit.
1. When Datakit is deployed in Kubernetes, it will exert a certain amount of request pressure on the API server.
