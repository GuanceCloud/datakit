---
title     : 'Graphite Exporter'
summary   : '采集 Graphite Exporter 暴露的指标数据'
__int_icon      : 'icon/graphite'
dashboard :
  - desc  : '暂无'
    path  : '-'
monitor   :
  - desc  : '暂无'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# Graphite
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

Graphite 采集器可以接收以 Graphite plaintext protocol 格式的指标数据，转换并供例如 Prometheus 等使用，只要配置相应的 Exporter 地址，就可以将指标数据接入

## 配置 {#config}

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置好后，[重启 Datakit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    可通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting) 或 [配置 ENV_DATAKIT_INPUTS](../datakit/datakit-daemonset-deploy.md#env-setting) 开启采集器。
<!-- markdownlint-enable -->

## 指标映射配置 {#metric-mapping-configuration}

Graphite 采集器可以通过在配置文件里配置映射格式将 **点格式**(例如 `testA.testB.testC`)的 Graphite plaintext protocol 转为带标记的指标。这个指标的转换规则类似于 `statsd_exporter` 的转换规则，但是在这里是 TOML 格式的配置。

没有配置映射规则的指标，会把除 `_`, `:` 之外的符号的非字母数字符号都替换为 `_`。

一个示例的映射规则如下：

```toml
[inputs.graphite.metric_mapper]
name = "test"
[[inputs.graphite.metric_mapper.mappings]]
match = "test.dispatcher.*.*.*"
name = "dispatcher_events_total"

[inputs.graphite.metric_mapper.mappings.labels]
action = "$2"
job = "test_dispatcher"
outcome = "$3"
processor = "$1"

[[inputs.graphite.metric_mapper.mappings]]
match = "*.signup.*.*"
name = "signup_events_total"

[inputs.graphite.metric_mapper.mappings.labels]
job = "${1}_server"
outcome = "$3"
provider = "$2"

[[inputs.graphite_metric_mapper.mappings]]
match = "servers\\.(.*)\\.networking\\.subnetworks\\.transmissions\\.([a-z0-9-]+)\\.(.*)"
match_type = "regex"
name = "servers_networking_transmissions_${3}"

[inputs.graphite.metric_mapper.mappings.labels]
hostname = "${1}"
device = "${2}"
```

以上规则会把 Graphite 指标转为以下的格式：

```txt
test.dispatcher.FooProcessor.send.success
  => dispatcher_events_total{processor="FooProcessor", action="send", outcome="success", job="test_dispatcher"}

foo_product.signup.facebook.failure
  => signup_events_total{provider="facebook", outcome="failure", job="foo_product_server"}

test.web-server.foo.bar
  => test_web__server_foo_bar{}

servers.rack-003-server-c4de.networking.subnetworks.transmissions.eth0.failure.mean_rate
  => servers_networking_transmissions_failure_mean_rate{device="eth0",hostname="rack-003-server-c4de"}
```

### 支持的映射规则说明 {#support-mapping}

#### 全局映射(Glob mapping) {#glob-mapping}

默认的全局映射规则使用 `*` 去代表指标中动态的部分。

> 注意：此时使用的是 **点格式** 指标，例如 `test.a.b.c.d`。

类似的配置如下：

```toml
[inputs.graphite.metric_mapper]
name = "test"
[[inputs.graphite.metric_mapper.mappings]]
match = "test.dispatcher.*.*.*"
name = "dispatcher_events_total"

[inputs.graphite.metric_mapper.mappings.labels]
action = "$2"
job = "test_dispatcher"
outcome = "$3"
processor = "$1"

[[inputs.graphite.metric_mapper.mappings]]
match = "*.signup.*.*"
name = "signup_events_total"

[inputs.graphite.metric_mapper.mappings.labels]
job = "${1}_server"
outcome = "$3"
provider = "$2"
```

转换得到的内容如下：

```txt
test.dispatcher.FooProcessor.send.success
 => dispatcher_events_total{processor="FooProcessor", action="send", outcome="success", job="test_dispatcher"}

foo_product.signup.facebook.failure
 => signup_events_total{provider="facebook", outcome="failure", job="foo_product_server"}

test.web-server.foo.bar
 => test_web_server_foo_bar{}
```

> 注意： 每个映射规则都必须有 `name` 字段，用 `$n` 来匹配行中的第 `n` 个替换。

```txt
[[inputs.graphite.metric_mapper.mappings]]
match = "test.*.*.counter"
name = "${2}_total"

[inputs.graphite.metric_mapper.mappings.labels]
provider = "$1"
```

例如 `test.a.b.counter`，对应的 `$1$` 则为 `a`，对应的 `$2` 则为 `b`，以此类推。

#### 正则匹配规则 {#regular-regex-mapping}

正则匹配规则使用常规的正则匹配来匹配指标名。需指定 `match_type = regex`

> 注意： 正则匹配相较于全局规则较慢

示例如下：

```toml
[[inputs.graphite_metric_mapper.mappings]]
match = "servers\.(.*)\.networking\.subnetworks\.transmissions\.([a-z0-9-]+)\.(.*)"
match_type = "regex"
name = "servers_networking_transmissions_${3}"

[inputs.graphite.metric_mapper.mappings.labels]
hostname = "${1}"
device = "${2}"
```

> 注意： 在 TOML 里，在字符串中反斜杠 (`\`) 需要被转义才能使用，因此需要使用 `\\`

#### 更多规则 {#more-details}

请参照 [statsd_exporter](https://github.com/prometheus/statsd_exporter){:target="_blank"}


### 严格匹配 {#strict-match}

如果在 **配置了映射规则的前提下**，只想要配置了映射规则的指标而忽略掉所以未配置规则的指标，可以通过设置 `strict_match` 来实现。

```toml
[inputs.graphite.metric_mapper]
strict_match = true
```

