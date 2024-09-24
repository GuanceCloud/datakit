---
title     : 'Graphite Exporter'
summary   : 'Collect Graphite Exporter exposed by Graphite Exporter'
tags:
  - 'THIRD PARTY'
__int_icon      : 'icon/graphite'
dashboard :
  - desc  : 'N/A'
    path  : '-'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---


{{.AvailableArchs}}

---

The Graphite collector can receive metrics data in Graphite plaintext protocol format, transform it, and make it available for use by systems like Prometheus. By configuring the appropriate Exporter address, you can integrate the metrics data into these systems.

## Configuration {#config}

### Preconditions {#requirements}

### Collector Configuration {#input-config}
<!-- markdownlint-disable MD046 -->
=== "Host Installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    After configuration, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    Can be turned on by [ConfigMap Injection Collector Configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting) or [Config ENV_DATAKIT_INPUTS](../datakit/datakit-daemonset-deploy.md#env-setting) .
<!-- markdownlint-enable -->

## Metric Mapping Configuration {#metric-mapping-configuration}

Graphite collector can be configured to translate specific **dot-separated** graphite metrics into labeled metrics via configuration file. The conversion rules for these metrics are similar to the rules in statsd_exporter, but here they are configured in TOML format.

Metrics that don't match any mapping in the configuration file are translated into metrics without any labels and with names in which every non-alphanumeric character except `_` and `:` is replaced with `_`.

An example mapping configuration:

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

This would transform these example graphite metrics into metrics as follows:

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

### Support Mapping Configuration {#support-mapping}

#### Glob Mapping {#glob-mapping}

The default glob mapping style uses * to denote parts of the metric name that may vary.

> Noted: now we use `dot-separated`, like `test.a.b.c.d`

An example mapping configuration:

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

This would transform these example metrics into metrics as follows:

```txt
test.dispatcher.FooProcessor.send.success
 => dispatcher_events_total{processor="FooProcessor", action="send", outcome="success", job="test_dispatcher"}

foo_product.signup.facebook.failure
 => signup_events_total{provider="facebook", outcome="failure", job="foo_product_server"}

test.web-server.foo.bar
 => test_web_server_foo_bar{}
```

> Noted: Every mapping configuration must have `name` field, The metric's name can contain $n-style references to be replaced by the n-th wildcard match in the matching line. That allows for dynamic rewrites, such as:

```txt
[[inputs.graphite.metric_mapper.mappings]]
match = "test.*.*.counter"
name = "${2}_total"

[inputs.graphite.metric_mapper.mappings.labels]
provider = "$1"
```

Here use `test.a.b.c.counter` as an example, `$1` corresponds to `a`, `$2`corresponds to `b`, and so on.

#### Regular expression matching {#regular-regex-mapping}

The regex matching rules use standard regular expression matching to match metric names. You need to specify match_type = regex.

> Noted: regex matching is slower than glob matching

An example mapping configuration:

```toml
[[inputs.graphite_metric_mapper.mappings]]
match = "servers\.(.*)\.networking\.subnetworks\.transmissions\.([a-z0-9-]+)\.(.*)"
match_type = "regex"
name = "servers_networking_transmissions_${3}"

[inputs.graphite.metric_mapper.mappings.labels]
hostname = "${1}"
device = "${2}"
```

> Noted: In TOML, backslashes (`\`) need to be escaped when used in strings, so you need to double-escape the backslashes by `\\`

#### More details {#more-details}

please refer to [statsd_exporter](https://github.com/prometheus/statsd_exporter){:target="_blank"}

### StrictMatch {#strict-match}

If you have a very large set of metrics you may want to skip the ones that don't match the mapping configuration. If that is the case you can force this behavior using the `strict_match`, and it will only store those metrics you really want.

```toml
[inputs.graphite.metric_mapper]
strict_match = true
```
