# Kubernetes Prometheus CRD Support

## Introduction {#intro}

This document describes how to enable Datakit to support Prometheus-Operator CRD and capture corresponding metrics.

## Description {#description}

Prometheus has a complete Kubernetes application metrics collection scheme, and the process is briefly described as follows:

1. Create Prometheus-Operator in the Kubernetes cluster
2. Create a corresponding CRD instance according to the requirements, which must carry the necessary configuration for collecting target metrics, such as `matchLabels`, `port` and `path` and so on
3. Prometheus-Operator listens for CRD instances and starts metric collection based on their configuration items

???+ attention

    Prometheus-Operator [official link](https://github.com/prometheus-operator/prometheus-operator) and [application example](https://alexandrev.medium.com/prometheus-concepts-servicemonitor-and-podmonitor-8110ce904908)。

Here, Datakit plays the role of step 3, in which Datakit monitors and discovers Prometheus-Operator CRD, starts metric collection according to configuration, and finally uploads it to Guance Cloud.

Currently, Datakit supports Prometheus-Operator CRD resources —— `PodMonitor` and `ServiceMonitor` —— and their required configuration.

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
$ wget https://github.com/prometheus-operator/prometheus-operator/blob/main/bundle.yaml
$ kubectl apply -f bundle.yaml
$ kubectl get crd
NAME                                        CREATED AT
alertmanagerconfigs.monitoring.coreos.com   2022-11-02T16:31:33Z
alertmanagers.monitoring.coreos.com         2022-11-02T16:31:33Z
podmonitors.monitoring.coreos.com           2022-11-02T16:31:33Z
probes.monitoring.coreos.com                2022-11-02T16:31:33Z
prometheuses.monitoring.coreos.com          2022-11-02T16:31:33Z
prometheusrules.monitoring.coreos.com       2022-11-02T16:31:34Z
servicemonitors.monitoring.coreos.com       2022-11-02T16:31:34Z
thanosrulers.monitoring.coreos.com          2022-11-02T16:31:34Z
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

Configuration parameters [document](https://doc.crds.dev/github.com/prometheus-operator/kube-prometheus/monitoring.coreos.com/PodMonitor/v1@v0.7.0). Currently, Datakit only supports the requirement part, and does not support authentication configurations such as `baseAuth`, `bearerToeknSecret` and `tlsConfig`.

### Turn on Datakit Collection {#config}

Add the environment variable `ENV_INPUT_CONTAINER_ENABLE_AUTO_DISCOVERY_OF_PROMETHEUS_POD_MONITORS` to datakit.yaml with a value of `"true"` to start PodMonitor metrics collection.

To work with metric data in more detail, Datakit provides the environment variable `ENV_INPUT_CONTAINER_PROMETHEUS_MONITORING_MATCHES_CONFIG` in JSON format, as follows:

```json
{
    "matches": [
        {
            "namespaceSelector": {
                "any": true,
                "matchNamespaces": []
            },
            "selector": {
                 "matchLabels": {
                     "app": "nacos"
                 },
                 "matchExpressions": [
                     {
                         "key": "environment",
                         "operator": "NotIn",
                         "values": [ "production" ]
                     }
                 ]
            },
            "promConfig": {
                "metric_types": ["counter", "gauge"],
                "measurement_prefix": "nacos_",
                "measurement_name": "prom",
                "tags": {
                    "key1": "value1"
                }
            }
        }
    ]
}
```

- matches: array format, allowing multiple matches, only the first successful match will take effect.
    - namespaceSelector: Specify the namespace of the object
        - any：booler type, whether to accept all namespaces
        - matchNamespaces: an array of strings that specifies a list of namespaces
    - selector: Select the Pod object
        - matchLabels: The map of the K/V key-value pair, equivalent to the IN of the matchExpressions operator with all conditions being AND
        - matchExpressions: List of matching expressions
            - key: A string value that represents the key of the label
            - operator: A string value that represents the relationship between key and values and can only be In, NotIn, Exists, and DoesNotExist
            - values: an array of strings that must be empty if the operator is In or NotIn; If it is another operator, it must not be empty.
    - promConfig: Corresponding configuration of prom collector
        - metric_types
        - metric_name_filter
        - measurement_prefix
        - measurement_name
        - measurements
            - prefix
            - name
        - tags_ignore
        - tags_rename
            - overwrite_exist_tags
            - mapping
        - ignore_tag_kv_match
        - ignore_req_err
        - http_headers
        - as_logging
            - enable
            - service
        - tags
        - auth

`promConfig` supports most of the conf fields of the prom collector, which are listed in the above field list, as shown in [doc](prom.md)。

???+ attention

    matchLabels and matchExpressions are Kubernetes' common match methods, as shown in [doc](https://kubernetes.io/zh-cn/docs/concepts/overview/working-with-objects/labels/#label-selectors)。

???+ attention

    The value of the environment variable `ENV_INPUT_CONTAINER_PROMETHEUS_MONITORING_MATCHES_CONFIG` is in JSON format and needs to be compressed into a line and escaped. You can store the configuration through ConfigMap, and then specify the replacement by ENV. For example:
    ```yaml
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: datakit-prom-crd
    data:
      prom-match-config: |
        {
            "matches":[
                {
                    "namespaceSelector":{
                        "any":true
                    },
                    "selector":{
                        "matchLabels":{
                            "app":"nacos"
                        }
                    },
                    "promConfig":{
                        "metric_types":[
                            "counter",
                            "gauge"
                        ],
                        "measurement_prefix":"nacos_"
                    }
                }
            ]
        }
    ```

    ENV uses ConfigMap content:
    ```
      - env:
        - name: ENV_INPUT_CONTAINER_PROMETHEUS_MONITORING_MATCHES_CONFIG
          valueFrom:
            configMapKeyRef:
              name: datakit-prom-crd  # name of configmap
              key: prom-match-config # The primary key name of configmap
              optional: false
    ```

### Check {#check}

Start Datakit, use `datakit monitor -V` or view it on the Guance Cloud page, and you can find a metric set beginning with `nacos_` to indicate that the collection was successful.
