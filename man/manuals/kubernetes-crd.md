# Kubernetes DataKit CRD 扩展采集
---

[:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6) · [:octicons-beaker-24: Experimental](index.md#experimental) 

## 介绍

本文档介绍如何在 Kubernetes 集群中创建 DataKit resouce 并配置扩展采集器。

### 添加鉴权

如果是升级版 DataKit 需要在 `datakit.yaml` 的 `apiVersion: rbac.authorization.k8s.io/v1` 项添加鉴权，即复制以下几行添加到末尾：

```
- apiGroups:
  - guance.com
  resources:
  - datakits
  verbs:
  - get
  - list
```

### 创建 v1beta1 DataKit 实例，创建 DataKit 实例对象

将以下内容写入 yaml 配置，例如 `datakit-crd.yaml`，修改配置项 `input-conf` `k8s-namespace` 和 `k8s-deployment`，并执行 apply 命令。

Datakit 会发现 DataKit 实例并按照 namespace 和 deployment 找寻对应的 pod，根据 input-conf 启动采集器。

配置项字段含义如下：

- `input-conf`：主配置，内容和 Datakit 采集器相同，其中支持如下几个通配符：
  - `$IP`：通配 Pod 的内网 IP
  - `$NAMESPACE`：Pod Namespace
  - `$PODNAME`：Pod Name
- `k8s-namespace`：指定 namespace
- `k8s-deployment`：指定 deployment，配合 namespace 定位 pod 组

```yaml
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: datakits.guance.com
spec:
  group: guance.com
  versions:
    - name: v1beta1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
              properties:
                input-conf:
                  type: string
                k8s-namespace:
                  type: string
                k8s-deployment:
                  type: string
                tags:
                  type: string
  scope: Namespaced
  names:
    plural: datakits
    singular: datakit
    kind: DataKit
    shortNames:
    - dk
---
apiVersion: v1
kind: Namespace
metadata:
  name: datakit-crd
---
apiVersion: "guance.com/v1beta1"
kind: DataKit
metadata:
  name: my-test-crd-object
  namespace: datakit-crd
spec:
  input-conf: |
    [[inputs.prom]]
      url = "http://$IP:8080/metrics"
      source = "hello-prom-testing]"
      metric_types = ["counter", "gauge"]
      interval = "10s"
      [inputs.prom.tags]
      namespace = "$NAMESPACE"
  k8s-namespace: "default"
  k8s-deployment: "prom-testing"
  tags: "key1=value1"
```

## Ngxin Ingress 配置示例

### 使用 DataKit CRD 扩展采集 Nginx Ingress 指标

#### 前提条件

- 已部署 [DaemonSet DataKit](datakit-daemonset-deploy.md)
- Deployment 的 *`Pod labels`* 必须存在 `app=你的deploy名称`
  
  如果 `Deployment` 名称为 `ingress-nginx-controller`，那边 yaml 配置如下：
  ```
  ...
  spec:
    selector:
      matchLabels:
        app.kubernetes.io/component: controller
    template:
      metadata:
        creationTimestamp: null
        labels:
          app: ingress-nginx-controller
  ...
  ```
  > !!! `ingress-nginx-controller` 只是一个示例名称

#### 采集基础信息

|  资源   | 名称  |
|  ----  | ----  |
| Namespace  | ingress-nginx |
| Deployment  | ingress-nginx-controller |
| Metrics Port | 10254 |
| source | prom-ingress |

#### 配置步骤

##### 创建 Datakit CustomResourceDefinition

执行如下创建命令：

```bash
cat <<EOF | kubectl apply -f -
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: datakits.guance.com
spec:
  group: guance.com
  versions:
    - name: v1beta1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
              properties:
                input-conf:
                  type: string
                k8s-namespace:
                  type: string
                k8s-deployment:
                  type: string
                tags:
                  type: string
  scope: Namespaced
  names:
    plural: datakits
    singular: datakit
    kind: DataKit
    shortNames:
    - dk
EOF
```

查看部署情况：

```bash
$ kubectl  get crd | grep datakit
datakits.guance.com
```

##### 创建 Datakit 资源

Prometheus 详细配置可参考[链接](../integrations/kubernetes-prom.md)

执行如下 `yaml` ：

```
apiVersion: guance.com/v1beta1
kind: DataKit
metadata:
  name: prom-ingress
  namespace: datakit
spec:
  input-conf: |-
    [[inputs.prom]]
      url = "http://$IP:10254/metrics"
      source = "prom-ingress"
      metric_types = ["counter", "gauge", "histogram"]
      measurement_name = "prom_ingress"
      interval = "60s"
      tags_ignore = ["build","le","method","release","repository"]
      metric_name_filter = ["nginx_process_cpu_seconds_total","nginx_process_resident_memory_bytes","request_size_sum","response_size_sum","requests","success","config_last_reload_successful"]
    [[inputs.prom.measurements]]
      prefix = "nginx_ingress_controller_"
      name = "prom_ingress"
    [inputs.prom.tags]
      namespace = "$NAMESPACE"
  k8s-deployment: ingress-nginx-controller
  k8s-namespace: ingress-nginx
  tags: key1=value1
```

> !!! 注意`namespace` 可自定义，`k8s-deployment` 和 `k8s-namespace` 则必须准确

查看部署情况：

```bash
$ kubectl get dk -n datakit
NAME           AGE
prom-ingress   18m
```

##### 查看指标采集情况

登录 `Datakit pod` ，执行以下命令：

```bash
$ datakit monitor
```

![](imgs/datakit-crd-ingress.png)

也可以登录 [观测云平台](https://www.guance.com/){:target="_blank"} ,【指标】-【查看器】查看指标数据


