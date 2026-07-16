---
title     : 'Kubernetes Prometheus CRD'
summary   : 'Collecting on Prometheus-Operator CRD'
tags      :
  - 'PROMETHEUS'
  - 'KUBERNETES'
__int_icon: 'icon/kubernetes'
---

## Introduction {#intro}

This document describes how to enable DataKit to support Prometheus-Operator CRD and collecting corresponding metrics.

## Description {#description}

Prometheus has a complete Kubernetes application metrics collection scheme, and the process is briefly described as follows:

1. Create Prometheus-Operator in the Kubernetes cluster
2. Create a corresponding CRD instance according to the requirements, which must carry the necessary configuration for collecting target metrics, such as `matchLabels`, `port` and `path` and so on
3. Prometheus-Operator listens for CRD instances and starts metric collection based on their configuration items

<!-- markdownlint-disable MD046 -->
???+ note

    Prometheus-Operator [official link](https://github.com/prometheus-operator/prometheus-operator) and [application example](https://alexandrev.medium.com/prometheus-concepts-servicemonitor-and-podmonitor-8110ce904908){:target="_blank"}。
<!-- markdownlint-enable MD046 -->

Here, DataKit plays the role of step 3. DataKit watches and discovers Prometheus-Operator CRDs, starts metric collection based on their configuration, and uploads the metrics to <<<custom_key.brand_name>>>. When a PodMonitor or ServiceMonitor is created, updated, or deleted, the corresponding collection tasks are dynamically added, rebuilt, or stopped without restarting DataKit.

DataKit supports `PodMonitor` and `ServiceMonitor` in `monitoring.coreos.com/v1`. The main supported fields are:

```markdown
- PodMonitor
    - selector
    - podTargetLabels
    - podMetricsEndpoints:
        - scheme
          port
          path
          params
          tlsConfig.insecureSkipVerify
    - namespaceSelector:
        any
        matchNames
- ServiceMonitor
    - selector
    - targetLabels
    - endpoints:
        - scheme
          port
          path
          params
          tlsConfig.insecureSkipVerify
    - namespaceSelector:
        any
        matchNames
```

Notes:

- The `interval` field in a Monitor does not control the actual scrape interval. Use the KubernetesPrometheus collector's `scrape_interval` setting instead.
- For `tlsConfig`, only `insecureSkipVerify` is supported. Certificates cannot be loaded from Kubernetes Secrets or ConfigMaps.
- Monitor authentication fields such as `basicAuth`, `bearerTokenSecret`, and `authorization` are not supported.
- `podTargetLabels` in ServiceMonitor is not supported.

Use `params` to specify `measurement`, for example:

```yaml
params:
    measurement:
    - new-measurement
```

## Enablement and RBAC {#enable-and-rbac}

Enable PodMonitor and ServiceMonitor discovery as needed in the KubernetesPrometheus collector configuration:

```toml
[inputs.kubernetesprometheus]
  enable_discovery_of_prometheus_pod_monitors     = true
  enable_discovery_of_prometheus_service_monitors = true
```

The following permissions are recommended for the DataKit ServiceAccount:

```yaml
- apiGroups: ["monitoring.coreos.com"]
  resources: ["podmonitors", "servicemonitors"]
  verbs: ["get", "list", "watch"]
```

DataKit uses `list` to load existing Monitors and `watch` to receive creation, update, and deletion events. When DataKit is upgraded with an older RBAC configuration that grants only `get` and `list`, collection does not fail. DataKit emits one WARN log and automatically falls back to running `list` every 20 seconds. In fallback mode, configuration changes may take up to about 20 seconds to take effect. Update RBAC to receive changes in real time.

## Examples {#example}

Take the Nacos cluster as an example.

Installing Nacos

```bash
git clone https://github.com/nacos-group/nacos-k8s.git
cd nacos-k8s
chmod +x quick-startup.sh
./quick-startup.sh
```

*nacos/nacos-quick-start.yaml* container port configuration:

```yaml
      containers:
        - name: k8snacos
          imagePullPolicy: Always
          image: nacos/nacos-server:latest
          ports:
            - containerPort: 8848
              name: client
            - containerPort: 9848
              name: client-rpc
            - containerPort: 9849
              name: raft-rpc
            - containerPort: 7848
              name: old-raft-rpc
```

- metrics access: `$IP:8848/nacos/actuator/prometheus`

- metrics port: 8848

There is now a Nacos metrics service in the Kubernetes cluster that collects metrics.

### Create Prometheus-Operator CRD {#create-crd}

- Install Prometheus-Operator



```bash
$ wget https://github.com/prometheus-operator/prometheus-operator/releases/download/v0.62.0/bundle.yaml
$ kubectl apply -f bundle.yaml
$ kubectl get crd
NAME                                        CREATED AT
alertmanagerconfigs.monitoring.coreos.com   2023-08-11T16:31:33Z
alertmanagers.monitoring.coreos.com         2023-08-11T16:31:33Z
podmonitors.monitoring.coreos.com           2023-08-11T16:31:33Z
probes.monitoring.coreos.com                2023-08-11T16:31:33Z
prometheuses.monitoring.coreos.com          2023-08-11T16:31:33Z
servicemonitors.monitoring.coreos.com       2023-08-11T16:31:34Z
thanosrulers.monitoring.coreos.com          2023-08-11T16:31:34Z
```

- Create PodMonitor

```bash
$ cat pod-monitor.yaml
apiVersion: monitoring.coreos.com/v1
kind: PodMonitor
metadata:
  name: nacos
  labels:
    app: nacos
spec:
  podMetricsEndpoints:
  - port: client
    path: /nacos/actuator/prometheus
  namespaceSelector:
    matchNames:
    - default
  selector:
    matchLabels:
      app: nacos

$ kubectl apply -f pod-monitor.yaml
```

Several important configuration items should be consistent with Nacos:

- namespace: default
- app: `nacos`
- port: client
- path: `/nacos/actuator/prometheus`

See the [Prometheus Operator API reference](https://prometheus-operator.dev/docs/api-reference/api/){:target="_blank"} for the complete PodMonitor schema. The fields supported by DataKit are listed above.

### Measurements and Tags {#measurement-and-tags}

See [Measurement Naming Rules](kubernetesprometheus.md#measurement-naming-rules) and [Automatically Added Tags](kubernetesprometheus.md#auto-added-tags).

### Check {#check}

Start DataKit, use `datakit monitor -V` or view it on the <<<custom_key.brand_name>>> page, and you can find a metric set beginning with `nacos_` to indicate that the collection was successful.
