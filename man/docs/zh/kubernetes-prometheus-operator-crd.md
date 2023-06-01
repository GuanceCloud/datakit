# Prometheus CRD 支持

## 介绍 {#intro}

本文档介绍如何让 Datakit 支持 Prometheus-Operator CRD 并采集对应指标。

## 描述 {#description}

Prometheus 有一套完善的 Kubernetes 应用指标采集方案，流程简述如下：

1. 在 Kubernetes 集群中创建 Prometheus-Operator
2. 根据需求，创建对应的 CRD 实例，该实例必须携带采集目标指标的必要配置，例如 `matchLabels` `port` `path` 等配置
3. Prometheus-Operator 会监听 CRD 实例，并根据其配置项开启指标采集

<!-- markdownlint-disable MD046 -->
???+ attention

    Prometheus-Operator [官方链接](https://github.com/prometheus-operator/prometheus-operator){:target="_blank"} 和 [应用示例](https://alexandrev.medium.com/prometheus-concepts-servicemonitor-and-podmonitor-8110ce904908){:target="_blank"}。
<!-- markdownlint-enable -->

在此处，Datakit 扮演了第 3 步的角色，由 Datakit 来监听和发现 Prometheus-Operator CRD，并根据配置开启指标采集，最终上传到观测云。

目前 Datakit 支持 Prometheus-Operator 两种 CRD 资源 —— `PodMonitor` 和 `ServiceMonitor`，以及其必要（require）配置。

## 示例 {#example}

以 Nacos 集群为例。

安装 Nacos：

```shell
git clone https://github.com/nacos-group/nacos-k8s.git

cd nacos-k8s

chmod +x quick-startup.sh

./quick-startup.sh
```

*nacos/nacos-quick-start.yaml* 容器端口配置：

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

- metrics 接口：`$IP:8848/nacos/actuator/prometheus`
- metrics port：8848

现在在 Kubernetes 集群中存在一个 Nacos metrics 服务可以采集指标。

### 创建 Prometheus-Operator CRD {#create-crd}

- 安装 Prometheus-Operator

```shell
wget https://github.com/prometheus-operator/prometheus-operator/blob/main/bundle.yaml
kubectl apply -f bundle.yaml
kubectl get crd

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

- 创建 PodMonitor

``` shell
cat pod-monitor.yaml

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

几个重要的配置项要和 Nacos 一致：

- namespace: default
- app: `nacos`
- port: client
- path: `/nacos/actuator/prometheus`

配置参数[文档](https://doc.crds.dev/github.com/prometheus-operator/kube-prometheus/monitoring.coreos.com/PodMonitor/v1@v0.7.0){:target="_blank"}，目前 Datakit 只支持 require 部分，暂不支持诸如 `baseAuth` `bearerToeknSecret` 和 `tlsConfig` 等认证配置。

### 开启 Datakit 采集功能 {#config}

在 *datakit.yaml* 中添加环境变量 `ENV_INPUT_CONTAINER_ENABLE_AUTO_DISCOVERY_OF_PROMETHEUS_POD_MONITORS` 值为 `"true"`，开启 `PodMonitor` 指标采集。

为了更细致的处理指标数据，Datakit 提供环境变量 `ENV_INPUT_CONTAINER_PROMETHEUS_MONITORING_MATCHES_CONFIG`，内容是 JSON 格式，如下：

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

- matches：数组格式，允许多个匹配，只有最先匹配成功的会生效
    - namespaceSelector：指定对象的命名空间
        - any：bool 类型，是否接受所有命名空间
        - matchNamespaces：字符串数组，指定命名空间列表
    - selector：选择 Pod 对象
        - matchLabels：K/V 键值对的 map，等价于 matchExpressions operator 的 IN，所有条件都是 AND
        - matchExpressions：匹配表达式列表
            - key：字符串值， 表示 label 的 key
            - operator：字符串值，表示 key 和 values 的关系，只能是 In、NotIn、Exists 和 DoesNotExist
            - values：字符串数组，如果 operator 是 In 或者 NotIn，数组必须为空；如果是其他 operator 则必须不为空
    - promConfig：prom 采集器的对应配置
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

`promConfig` 支持 prom 采集器的大部分 conf 字段，已经列在上述字段列表，具体含义见[文档](prom.md)。

<!-- markdownlint-disable MD046 -->
???+ attention

    matchLabels 和 matchExpressions 是 Kubernetes 通用的 match 方式，详见[文档](https://kubernetes.io/zh-cn/docs/concepts/overview/working-with-objects/labels/#label-selectors){:target="_blank"}。

???+ attention

    环境变量 `ENV_INPUT_CONTAINER_PROMETHEUS_MONITORING_MATCHES_CONFIG` 的值是 JSON 格式，需要注意压缩成一行和转义。可以通过 ConfigMap 将配置存储，再由 ENV 指定替换即可。例如：
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

    ENV 使用 ConfigMap 内容：
    ```
      - env:
        - name: ENV_INPUT_CONTAINER_PROMETHEUS_MONITORING_MATCHES_CONFIG
          valueFrom:
            configMapKeyRef:
              name: datakit-prom-crd  # configmap 的名称
              key: prom-match-config # configmap 的主键名称
              optional: false
    ```
<!-- markdownlint-enable -->

### 验证 {#check}

启动 Datakit，使用 `datakit monitor -V` 或在观测云页面上查看，能找到以 `nacos_` 开头的指标集说明采集成功。
