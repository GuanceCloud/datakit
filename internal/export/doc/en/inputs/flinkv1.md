---
title     : 'Flink'
summary   : 'Collect Flink metrics'
tags:
  - 'MIDDLEWARE'
__int_icon      : 'icon/flink'
dashboard :
  - desc  : 'Flink'
    path  : 'dashboard/en/flink'
monitor   :
  - desc  : 'Flink'
    path  : 'monitor/en/flink'
---


{{.AvailableArchs}}

---

Flink collector can take many metrics from Flink instances, such as Flink server status and network status, and collect the metrics to DataFlux to help you monitor and analyze various abnormal situations of Flink.

## Configuration  {#config}

### Preconditions {#requirements}

> Explanation: Example Flink version is: Flink 1.14. 2 (CentOS), each version of the indicator may be different.

At present, Flink officially provides two methods for reporting metrics: [Prometheus](https://nightlies.apache.org/flink/flink-docs-release-1.14/docs/deployment/metric_reporters/#prometheus){:target="_blank"} and [PrometheusPushGateway](https://nightlies.apache.org/flink/flink-docs-release-1.14/docs/deployment/metric_reporters/#prometheuspushgateway){:target="_blank"}. Their main differences are:

- Prometheus PushGateway is to report all metrics of the cluster to PushGateway in a unified way, so you need to install PushGateway additionally.
- Prometheus mode requires each node of the cluster to expose a unique port, and does not need to install other software, but it requires N available ports, which is slightly complicated to configure.

### PrometheusPushGateway Way (recommended) {#push-gateway}

- Download and Install: PushGateway can be downloaded at [Prometheuse official page](https://prometheus.io/download/#pushgateway){:target="_blank"}.

Start PushGateway: (This command is for reference only, and the specific command may vary according to the actual environment)

```shell
nohup ./pushgateway &
```

- Configure `flink-conf.yaml` to report metrics uniformly to PushGateway

Configure the configuration file for Flink `conf/flink-conf.yaml` sample:

```bash
metrics.reporter.promgateway.class: org.apache.flink.metrics.prometheus.PrometheusPushGatewayReporter # Fixed this value and cannot be changed
metrics.reporter.promgateway.host: localhost # IP address of promgateway
metrics.reporter.promgateway.port: 9091 # promgateway listening port
metrics.reporter.promgateway.interval: 15 SECONDS # collection interval
metrics.reporter.promgateway.groupingKey: k1=v1;k2=v2

# The following are optional parameters
# metrics.reporter.promgateway.jobName: myJob
# metrics.reporter.promgateway.randomJobNameSuffix: true
# metrics.reporter.promgateway.deleteOnShutdown: false
```

Start Flink: `./bin/start-cluster.sh` (This command is for reference only, and the specific command may vary depending on the actual environment)

### Prometheus Mode {#prometheus}

- Configure `flink-conf.yaml` to expose metrics for each node. Configure the configuration file for Flink `conf/flink-conf.yaml` sample:

```bash
metrics.reporter.prom.class: org.apache.flink.metrics.prometheus.PrometheusReporter
metrics.reporter.prom.port: 9250-9260 # The port range of each node is different according to the number of nodes, and one port corresponds to one node
```

- Start Flink: `./bin/start-cluster.sh` (This command is for reference only, and the specific command may vary depending on the actual environment)
- Change the Flink configuration and add the following to turn on Prometheus collection.

```bash
metrics.reporter.prom.class: org.apache.flink.metrics.prometheus.PrometheusReporter
metrics.reporter.prom.port: 9250-9260
```

> Note: The `metrics.reporter.prom.port` setting is based on the number of clustered `jobmanagers` and `taskmanager`

- Restart the Flink cluster application configuration
- curl http://{Flink iP}:9250-9260 to start collecting

## Metric {#metric}

Flink collects multiple metrics by default, and these [metrics](https://nightlies.apache.org/flink/flink-docs-release-1.14/docs/ops/metrics/#system-metrics){:target="_blank"} provide insight into the current state.

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

- Tags

{{$m.TagsMarkdownTable}}

- Metrics

{{ end }}
