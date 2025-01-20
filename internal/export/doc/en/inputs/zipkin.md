---
title     : 'Zipkin'
summary   : 'Zipkin Tracing Data Ingestion'
tags      :
  - 'ZIPKIN'
  - 'APM'
  - 'TRACING'
__int_icon      : 'icon/zipkin'
dashboard :
  - desc  : 'N/A'
    path  : '-'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---

{{.AvailableArchs}}

---

The Zipkin Agent embedded in Datakit is used to receive, calculate and analyze the data of Zipkin Tracing protocol.

## Configuration {#config}

### Collector Config {#input-config}

<!-- markdownlint-disable MD046 -->
=== "Host Installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    After configuration, restart DataKit.

=== "Kubernetes"

    At present, the collector can be turned on by [injecting the collector configuration in ConfigMap mode](../datakit/datakit-daemonset-deploy.md#configmap-setting).

    Multiple environment variables supported that can be used in Kubernetes showing below:

    | Environment Variable Name             | Type        | Example                                                                          |
    | ------------------------------------- | ----------- | -------------------------------------------------------------------------------- |
    | `ENV_INPUT_ZIPKIN_PATH_V1`            | string      | "/api/v1/spans"                                                                  |
    | `ENV_INPUT_ZIPKIN_PATH_V2`            | string      | "/api/v2/spans"                                                                  |
    | `ENV_INPUT_ZIPKIN_IGNORE_TAGS`        | JSON string | `["block1", "block2"]`                                                           |
    | `ENV_INPUT_ZIPKIN_KEEP_RARE_RESOURCE` | bool        | true                                                                             |
    | `ENV_INPUT_ZIPKIN_DEL_MESSAGE`        | bool        | true                                                                             |
    | `ENV_INPUT_ZIPKIN_CLOSE_RESOURCE`     | JSON string | `{"service1":["resource1"], "service2":["resource2"], "service3":["resource3"]}` |
    | `ENV_INPUT_ZIPKIN_SAMPLER`            | float       | 0.3                                                                              |
    | `ENV_INPUT_ZIPKIN_TAGS`               | JSON string | `{"k1":"v1", "k2":"v2", "k3":"v3"}`                                              |
    | `ENV_INPUT_ZIPKIN_THREADS`            | JSON string | `{"buffer":1000, "threads":100}`                                                 |
    | `ENV_INPUT_ZIPKIN_STORAGE`            | JSON string | `{"storage":"./zipkin_storage", "capacity": 5120}`                               |
<!-- markdownlint-enable -->

## Tracing {#tracing}

{{range $i, $m := .Measurements}}

{{if eq $m.Type "tracing"}}

### `{{$m.Name}}`

{{$m.Desc}}

- Tags

{{$m.TagsMarkdownTable}}

- Metrics

{{$m.FieldsMarkdownTable}}
{{end}}

{{end}}

## Zipkin Docs {#docs}

- [Quick Start](https://zipkin.io/pages/quickstart.html){:target="_blank"}
- [Docs](https://zipkin.io/pages/instrumenting.html){:target="_blank"}
- [Source Code](https://github.com/openzipkin/zipkin){:target="_blank"}
