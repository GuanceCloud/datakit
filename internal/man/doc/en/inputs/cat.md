# CAT
---


[:octicons-tag-24: Version-1.9.0](changelog.md#cl-1.9.0) ·
[:octicons-beaker-24: Experimental](index.md#experimental)

---

{{.AvailableArchs}}


---

[dianping-cat](https://github.com/dianping/cat){:target="_blank"}  Cat is an open-source distributed real-time monitoring system mainly used to monitor the performance, capacity, and business indicators of the system. It is a monitoring system developed by Meituan Dianping Company and is currently open source and widely used.

Cat collects various indicator data of the system, such as CPU, memory, network, disk, etc., for real-time monitoring and analysis, helping developers quickly locate and solve system problems. 
At the same time, it also provides some commonly used monitoring functions, such as alarms, statistics, log analysis, etc., to facilitate system monitoring and analysis by developers.


## Data Type {#data}

Data transmission protocol:

- Plaintext: Plain text mode, currently not supported by Datakit.

- Native: Text form separated by specific symbols, currently supported by Datakit.


数据分类：

| type | long type         | doc               | datakit support | Corresponding data type |
|------|-------------------|:------------------|:---------------:|:------------------------|
| t    | transaction start | transaction start |      true       | trace                   |
| T    | transaction end   | transaction end   |      true       | trace                   |
| E    | event             | event             |      false      | -                       |
| M    | metric            | metric            |      false      | -                       |
| L    | trace             | trace             |      false      | -                       |
| H    | heartbeat         | heartbeat         |      true       | 指标                      |




## CAT start mode {#cat-start}

The data is all in the datakit, and the web page of cat no longer has data, so the significance of starting is not significant. 

Moreover, the cat server will also send transaction data to the dk, causing a large amount of garbage data on the observation cloud page. It is not recommended to start a cat_ Home (cat server) service.

The corresponding configuration can be configured in client.xml, please refer to the following text.



## Config {#config}

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

> Note: The 9529 port in the configuration is the HTTP port of the datakit. 2280 is the 2280 port opened by the cat input.

Datakit config：

    Go to the `conf.d/cat` directory under the DataKit installation directory, copy `cat.conf.sample` and name it `cat.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```


=== "Kubernetes"

The collector can now be turned on by [ConfigMap Injection Collector Configuration](datakit-daemonset-deploy.md#configmap-setting).



Notes on configuration files:

1. `startTransactionTypes` `MatchTransactionTypes` `block` `routers` `sample`  is the data returned to the client end.
2. `routers` is Datakit IP or Domain.
3. `tcp_port`  client config `servers ip` address

---

## Guance Trace and Metric {#traces_mertics}

### trace {#traces}

login guance.com, and click on Application Performance.

<!-- markdownlint-disable MD033 -->
<figure>
  <img src="https://df-storage-dev.oss-cn-hangzhou.aliyuncs.com/songlongqi/cat/cat-gateway.png" style="height: 500px" alt=" trace details">
  <figcaption> trace details </figcaption>
</figure>


### Metric {#metrics}
To [download dashboard](https://df-storage-dev.oss-cn-hangzhou.aliyuncs.com/songlongqi/cat/DianPing-Cat%20%E7%9B%91%E6%8E%A7%E8%A7%86%E5%9B%BE.json){:target="_blank"}

At guance `Scenes` -> `dashboard` to `Create Dashboard`.

Effect display:

<!-- markdownlint-disable MD046 MD033 -->
<figure >
  <img src="https://df-storage-dev.oss-cn-hangzhou.aliyuncs.com/songlongqi/cat/metric.png" style="height: 500px" alt="cat metric">
  <figcaption> cat metric </figcaption>
</figure>


## Metrics Set {#cat-metrics}

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

{{$m.Desc}}

- tag

{{$m.TagsMarkdownTable}}

- fields

{{$m.FieldsMarkdownTable}}

{{ end }}
