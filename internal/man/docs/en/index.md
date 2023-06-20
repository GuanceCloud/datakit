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
- [Pipeline](../developers/pipeline/index.md): Easy structured data extraction
- Support access to other third-party data collection
    - [Telegraf](telegraf.md)
    - [Prometheus](prom.md)
    - [Statsd](statsd.md)
    - [Fluentd](logstreaming.md)
    - [Filebeats](beats_output.md)
    - [Function](../dataflux-func/write-data-via-datakit.md)
    - Tracing
        - [OpenTelemetry](opentelemetry.md)
        - [DDTrace](ddtrace.md)
        - [Zipkin](zipkin.md)
        - [Jaeger](jaeger.md)
        - [Skywalking](skywalking.md)

## Description {#spec}

### Experimental Functionality {#experimental}

When DataKit is released, it will bring some experimental functions. These functions are often new functions released for the first time. The implementation of these functions may be lacking in consideration or imprecise. Therefore, when using experimental functions, the following possible situations should be considered:

- The function is unstable.
- For some functional configurations, compatibility is not guaranteed during subsequent iterations.
- The functionality may be removed due to its limitations, but there will be other corresponding measures to meet the corresponding requirements.

For this part of the function, please use it carefully.

In the process of using the experimental function, related questions can be submitted to issue

- [Gitlab](https://gitlab.jiagouyun.com/cloudcare-tools/datakit/-/issues/new?issue%5Bmilestone_id%5D=){:target="_blank"} 
- [Github](https://github.com/GuanceCloud/datakit/issues/new){:target="_blank"}
- [Jihu](https://jihulab.com/guance-cloud/datakit/-/issues/new){:target="_blank"}

### Legend Description {#legends}

| Legend                                                                                                                       | Description                                                          |
| ---                                                                                                                        | ---                                                           |
| :fontawesome-solid-flag-checkered:                                                                                         | Indicates that the collector supports election                                          |
| :fontawesome-brands-linux: :fontawesome-brands-windows: :fontawesome-brands-apple: :material-kubernetes: :material-docker: | Examples are used to represent Linux, Windows, macOS, Kubernetes, and Docker respectively |
