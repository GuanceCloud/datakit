---
title     : 'Dianping CAT'
summary   : 'The performance, capacity, and business indicator monitoring system of Meituan Dianping'
__int_icon      : 'icon/cat'
tags:
  - 'TRACING'
  - 'APM'
dashboard :
  - desc  : 'Cat dashboard'
    path  : 'dashboard/en/cat'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---

[:octicons-tag-24: Version-1.9.0](../datakit/changelog.md#cl-1.9.0) ·
[:octicons-beaker-24: Experimental](../datakit/index.md#experimental)

---

{{.AvailableArchs}}


---

[Dianping-cat](https://github.com/dianping/cat){:target="_blank"}  Cat is an open-source distributed real-time monitoring system mainly used to monitor the performance, capacity, and business indicators of the system. It is a monitoring system developed by Meituan Dianping Company and is currently open source and widely used.

Cat collects various indicator data of the system, such as CPU, memory, network, disk, etc., for real-time monitoring and analysis, helping developers quickly locate and solve system problems.
At the same time, it also provides some commonly used monitoring functions, such as alarms, statistics, log analysis, etc., to facilitate system monitoring and analysis by developers.


## Data Type {#data}

Data transmission protocol:

- Plaintext: Plain text mode, currently not supported by DataKit.

- Native: Text form separated by specific symbols, currently supported by DataKit.


Data Classification：

| type | long type         | doc               | DataKit support | Corresponding data type |
|------|-------------------|:------------------|:---------------:|:------------------------|
| t    | transaction start | transaction start |      true       | trace                   |
| T    | transaction end   | transaction end   |      true       | trace                   |
| E    | event             | event             |      false      | -                       |
| M    | metric            | metric            |      false      | -                       |
| L    | trace             | trace             |      false      | -                       |
| H    | heartbeat         | heartbeat         |      true       | metric                      |




## CAT start mode {#cat-start}

The data is all in the DataKit, and the web page of cat no longer has data, so the significance of starting is not significant.

Moreover, the cat server will also send transaction data to the dk, causing a large amount of garbage data on the <<<custom_key.brand_name>>> page. It is not recommended to start a cat_ Home (cat server) service.

The corresponding configuration can be configured in client.xml, please refer to the following text.



## Configuration {#config}

client config：

```xml
<?xml version="1.0" encoding="utf-8"?>
<config mode="client">
    <servers>
        <!-- datakit ip, cat port , http port -->
        <server ip="10.200.6.16" port="2280" http-port="9529"/>
    </servers>
</config>
```

> Note: The 9529 port in the configuration is the HTTP port of the DataKit. 2280 is the 2280 port opened by the cat input.

<!-- markdownlint-disable MD046 -->

=== "Host Installation"

    Go to the `conf.d/cat` directory under the DataKit installation directory, copy `cat.conf.sample` and name it `cat.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

=== "Kubernetes"

    The collector can now be turned on by [ConfigMap Injection Collector Configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting).

<!-- markdownlint-enable -->

---

Notes on configuration files:

1. `startTransactionTypes` `MatchTransactionTypes` `block` `routers` `sample`  is the data returned to the client end
1. `routers` is DataKit IP or Domain
1. `tcp_port`  client config `servers ip` address

---

## Tracing {#tracing}

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "tracing"}}

### `{{$m.Name}}`

{{$m.Desc}}

- Tags

{{$m.TagsMarkdownTable}}

- Metrics

{{$m.FieldsMarkdownTable}}
{{end}}

{{ end }}

## Metric {#metric}

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "metric"}}

### Metric `{{$m.Name}}`

{{$m.Desc}}

- Tags

{{$m.TagsMarkdownTable}}

- Metrics

{{$m.FieldsMarkdownTable}}
{{end}}

{{ end }}
<!-- markdownlint-enable -->
