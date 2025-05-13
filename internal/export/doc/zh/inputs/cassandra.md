---
title     : 'Cassandra'
summary   : '采集 Cassandra 的指标数据'
tags      :
  - '数据库'
__int_icon: 'icon/cassandra'
dashboard :
  - desc  : 'Cassandra'
    path  : 'dashboard/zh/cassandra'
monitor   :
  - desc  : 'Cassandra'
    path  : 'monitor/zh/cassandra'
---


{{.AvailableArchs}}

---

可以使用 [DDTrace](ddtrace.md) 采集 Cassandra 指标。采集数据流向如下：Cassandra -> DDTrace -> DataKit(StatsD)。

可以看到 DataKit 已经集成了 [StatsD](https://github.com/statsd/statsd){:target="_blank"} 的服务端，DDTrace 采集 Cassandra 的数据后使用 StatsD 的协议报告给了 DataKit。

## 配置 {#config}

### 前置条件 {#requrements}

- 已测试的版本：
    - [x] 5.0
    - [x] 4.1.3
    - [x] 3.11.15
    - [x] 3.0.24
    - [x] 2.1.22

- 下载 `dd-java-agent.jar` 包，参见 [这里](ddtrace.md){:target="_blank"};

- DataKit 侧：参见 [StatsD](statsd.md){:target="_blank"} 的配置。

- Cassandra 侧：

在 */usr/local/cassandra/bin* 下创建文件 *setenv.sh* 并赋予执行权限，再写入以下内容：

```shell
export CATALINA_OPTS="-javaagent:dd-java-agent.jar \
                      -Ddd.jmxfetch.enabled=true \
                      -Ddd.jmxfetch.statsd.host=${DATAKIT_HOST} \
                      -Ddd.jmxfetch.statsd.port=${DATAKIT_STATSD_HOST} \
                      -Ddd.jmxfetch.cassandra.enabled=true"
```

参数说明如下：

- `javaagent`: 这个填写 `dd-java-agent.jar` 的完整路径；
- `Ddd.jmxfetch.enabled`: 填 `true`, 表示开启 DDTrace 的采集功能；
- `Ddd.jmxfetch.statsd.host`: 填写 DataKit 监听的网络地址。不含端口号；
- `Ddd.jmxfetch.statsd.port`: 填写 DataKit 监听的端口号。一般为 `11002`，由 DataKit 侧的配置决定；
- `Ddd.jmxfetch.Cassandra.enabled`: 填 `true`, 表示开启 DDTrace 的 Cassandra 采集功能。开启后会多出名为 `cassandra` 的指标集；

重启 Cassandra 使配置生效。

### 采集器配置 {#input-config}

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    可通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting) 或 [配置 ENV_DATAKIT_INPUTS](../datakit/datakit-daemonset-deploy.md#env-setting) 开启采集器。

---

<!-- markdownlint-enable -->

## 指标 {#metric}

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

- 标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}
