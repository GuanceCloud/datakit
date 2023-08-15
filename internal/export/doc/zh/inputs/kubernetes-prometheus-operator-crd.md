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

<!-- markdownlint-disable MD014 -->
```shell
$ git clone https://github.com/nacos-group/nacos-k8s.git
$ cd nacos-k8s
$ chmod +x quick-startup.sh
$ ./quick-startup.sh
```
<!-- markdownlint-enable -->

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

<!-- markdownlint-disable MD014 -->
```shell
$ wget https://github.com/prometheus-operator/prometheus-operator/releases/download/v0.62.0/bundle.yaml
$ kubectl apply -f bundle.yaml
$ kubectl get crd

NAME                                        CREATED AT
alertmanagerconfigs.monitoring.coreos.com   2022-08-11T03:15:57Z
alertmanagers.monitoring.coreos.com         2022-08-11T03:15:57Z
podmonitors.monitoring.coreos.com           2022-08-11T03:15:57Z
probes.monitoring.coreos.com                2022-08-11T03:15:57Z
prometheuses.monitoring.coreos.com          2022-08-11T03:15:57Z
servicemonitors.monitoring.coreos.com       2022-08-11T03:15:57Z
thanosrulers.monitoring.coreos.com          2022-08-11T03:15:57Z
```
<!-- markdownlint-enable -->

- 创建 PodMonitor

<!-- markdownlint-disable MD014 -->
``` shell
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
<!-- markdownlint-enable -->

几个重要的配置项要和 Nacos 一致：

- namespace: default
- app: `nacos`
- port: client
- path: `/nacos/actuator/prometheus`

配置参数[文档](https://doc.crds.dev/github.com/prometheus-operator/kube-prometheus/monitoring.coreos.com/PodMonitor/v1@v0.7.0){:target="_blank"}，目前 Datakit 只支持 require 部分，暂不支持诸如 `baseAuth` `bearerToeknSecret` 和 `tlsConfig` 等认证配置。

### 指标集和 tags {#measurement-and-tags}

详见参考[此处](kubernetes-prom.md#measurement-and-tags)。

### 验证 {#check}

启动 Datakit，使用 `datakit monitor -V` 或在观测云页面上查看，能找到以 `nacos_` 开头的指标集说明采集成功。
