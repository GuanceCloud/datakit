# Kubernetes Prometheus CRD Support

## Introduction {#intro}

This document describes how to enable Datakit to support Prometheus-Operator CRD and capture corresponding metrics.

## Description {#description}

Prometheus has a complete Kubernetes application metrics collection scheme, and the process is briefly described as follows:

1. Create Prometheus-Operator in the Kubernetes cluster
2. Create a corresponding CRD instance according to the requirements, which must carry the necessary configuration for collecting target metrics, such as `matchLabels`, `port` and `path` and so on
3. Prometheus-Operator listens for CRD instances and starts metric collection based on their configuration items

???+ attention

    Prometheus-Operator [official link](https://github.com/prometheus-operator/prometheus-operator) and [application example](https://alexandrev.medium.com/prometheus-concepts-servicemonitor-and-podmonitor-8110ce904908){:target="_blank"}。

Here, Datakit plays the role of step 3, in which Datakit monitors and discovers Prometheus-Operator CRD, starts metric collection according to configuration, and finally uploads it to Guance Cloud.

Currently, Datakit supports Prometheus-Operator CRD resources —— `PodMonitor` and `ServiceMonitor` —— and their required configuration:

```markdown
- PodMonitor [monitoring.coreos.com/v1]
    - podTargetLabels
    - podMetricsEndpoints:
        - interval
          port
          path
	  params
    - namespaceSelector:
        any
        matchNames
- ServiceMonitor:
    - targetLabels
    - podTargetLabels
    - endpoints:
        - interval
          port
          path
	  params
    - namespaceSelector:
        any
        matchNames
```


## Examples {#example}

Take the nacos cluster as an example.

Installing nacos

```
$ git clone https://github.com/nacos-group/nacos-k8s.git
$ cd nacos-k8s
$ chmod +x quick-startup.sh
$ ./quick-startup.sh
```

nacos/nacos-quick-start.yaml container port configuration:
```
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
- metrics access: $IP:8848/nacos/actuator/prometheus
- metrics port: 8848

There is now a nacos metrics service in the Kubernetes cluster that collects metrics.

### Create Prometheus-Operator CRD {#create-crd}

1. Install Prometheus-Operator

```
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

2. Create PodMonitor

```
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
    interval: 15s
    path: /nacos/actuator/prometheus
  namespaceSelector:
    matchNames:
    - default
  selector:
    matchLabels:
      app: nacos

$ kubectl apply -f pod-monitor.yaml
```

Several important configuration items should be consistent with nacos:

- namespace: default
- app: nacos
- port: client
- path: /nacos/actuator/prometheus

Configuration parameters [document](https://doc.crds.dev/github.com/prometheus-operator/kube-prometheus/monitoring.coreos.com/PodMonitor/v1@v0.7.0){:target="_blank"}. Currently, Datakit only supports the requirement part, and does not support authentication configurations such as `baseAuth`, `bearerToeknSecret` and `tlsConfig`.

### Measurements and Tags {#measurement-and-tags}

Refer to [doc](kubernetes-prom.md#measurement-and-tags).

### Check {#check}

Start Datakit, use `datakit monitor -V` or view it on the Guance Cloud page, and you can find a metric set beginning with `nacos_` to indicate that the collection was successful.
