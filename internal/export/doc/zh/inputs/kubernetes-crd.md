---
title     : 'Kubernetes CRD'
summary   : 'Create Datakit CRD to collect'
tags      :
  - 'PROMETHEUS'
  - 'KUBERNETES'
__int_icon: 'icon/kubernetes'
---

:material-kubernetes:

---

[:octicons-beaker-24: Experimental](../datakit/index.md#experimental)

## 介绍 {#intro}

**此功能在 Datakit 1.63.0 弃用。**

本文档介绍如何在 Kubernetes 集群中创建 Datakit resource 并配置扩展采集器。

### 添加鉴权 {#authorization}

如果是升级版 DataKit 需要在 `datakit.yaml` 的 `apiVersion: rbac.authorization.k8s.io/v1` 项添加鉴权，即复制以下几行添加到末尾：

```yaml
- apiGroups:
  - <<<custom_key.brand_main_domain>>>
  resources:
  - datakits
  verbs:
  - get
  - list
```

### 创建 v1beta1 DataKit 实例和 DataKit 实例对象 {#create}

将以下内容写入 yaml 配置，例如 *datakit-crd.yaml*，其中各个字段的含义如下：

- `k8sNamespace`：指定 namespace，配合 deployment 定位一个集合的 Pod，必填项
- `k8sDaemonSet`：指定 DaemonSet 名称，配合 namespace 定位一个集合的 Pod
- `k8sDeployment`：指定 deployment 名称，配合 namespace 定位一个集合的 Pod
- `inputConf`：采集器配置文件，依据 namespace 和 deployment 找到对应的 Pod，替换 Pod 的通配符信息，再根据 inputConf 内容运行采集器。支持以下通配符
    - `$IP`：Pod 的内网 IP
    - `$NAMESPACE`：Pod Namespace
    - `$PODNAME`：Pod Name
    - `$NODENAME`：当前所在 node 的 name

执行 `kubectl apply -f datakit-crd.yaml` 命令。

<!-- markdownlint-disable MD046 -->
???+ attention

    - DaemonSet 和 Deployment 是两种不同的 Kubernetes resource，但在此处，`k8sDaemonSet` 和 `k8sDeployment` 是可以同时存在的。即在同一个 Namespace 下，DaemonSet 创建的 Pod 和 Deployment 创建的 Pod 共用同一份 CRD 配置。但是不推荐这样做，因为在具体配置中会有类似 `source` 这种字段用来标识数据源，混用会导致数据界线不够清晰。建议在同一份 CRD 配置中 `k8sDaemonSet` 和 `k8sDeployment` 只存在一个。

    - Datakit 只采集和它处于同一个 node 的 Pod，属于就近采集，不会跨 node 采集。
<!-- markdownlint-enable -->

## 示例 {#example}

完整示例如下，包括：

- 创建 CRD Datakit
- 测试所用的 namespace 和 Datakit 实例对象
- 配置 Prom 采集器（`inputConf`）

```yaml
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: datakits.<<<custom_key.brand_main_domain>>>
spec:
  group: <<<custom_key.brand_main_domain>>>
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
              instances:
                type: array
                items:
                  type: object
                  properties:
                    k8sNamespace:
                      type: string
                    k8sDaemonSet:
                      type: string
                    k8sDeployment:
                      type: string
                    datakit/logs:
                      type: string
                    inputConf:
                      type: string
  scope: Namespaced
  names:
    plural: datakits
    singular: datakit
    kind: Datakit
    shortNames:
    - dk
---
apiVersion: v1
kind: Namespace
metadata:
  name: datakit-crd
---
apiVersion: "<<<custom_key.brand_main_domain>>>/v1beta1"
kind: Datakit
metadata:
  name: my-test-crd-object
  namespace: datakit-crd
spec:
  instances:
    - k8sNamespace: "testing-namespace"
      k8sDaemonSet: "testing-daemonset"
      inputConf: |
        [inputs.prom]
          url="http://prom"
```

### NGINX Ingress 配置示例 {#example-nginx}

这里使用 DataKit CRD 扩展采集 Ingress 指标，即通过 prom 采集器来收集 Ingress 的指标。

#### 前提条件 {#nginx-requirements}

- 已部署 [DaemonSet DataKit](../datakit/datakit-daemonset-deploy.md)
- 如果 `Deployment` 名称为 `ingress-nginx-controller`，那边 yaml 配置如下：

  ``` yaml
  ...
  spec:
    selector:
      matchLabels:
        app.kubernetes.io/component: controller
    template:
      metadata:
        creationTimestamp: null
        labels:
          app: ingress-nginx-controller  # 这里只是一个示例名称
  ...
  ```

#### 配置步骤 {#nginx-steps}

- 先创建 Datakit CustomResourceDefinition

执行如下创建命令：

```bash
cat <<EOF | kubectl apply -f -
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: datakits.<<<custom_key.brand_main_domain>>>
spec:
  group: <<<custom_key.brand_main_domain>>>
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
                instances:
                  type: array
                  items:
                    type: object
                    properties:
                      k8sNamespace:
                        type: string
                      k8sDeployment:
                        type: string
                      datakit/logs:
                        type: string
                      inputConf:
                        type: string
  scope: Namespaced
  names:
    plural: datakits
    singular: datakit
    kind: Datakit
    shortNames:
    - dk
EOF
```

查看部署情况：

```bash
kubectl get crds | grep <<<custom_key.brand_main_domain>>>

datakits.<<<custom_key.brand_main_domain>>>   2022-08-18T10:44:09Z
```

- 创建 Datakit 资源

Prometheus 详细配置可参考[链接](kubernetes-prom.md)

执行如下 `yaml` ：

```yaml
apiVersion: <<<custom_key.brand_main_domain>>>/v1beta1
kind: DataKit
metadata:
  name: prom-ingress
  namespace: datakit
spec:
  instances:
    - k8sNamespace: ingress-nginx
      k8sDeployment: ingress-nginx-controller
      inputConf: |-
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
```

> !!! 注意 `namespace` 可自定义，`k8sDeployment` 和 `k8sNamespace` 则必须准确

查看部署情况：

```bash
$ kubectl get dk -n datakit
NAME           AGE
prom-ingress   18m
```

- 查看指标采集情况

登录 `Datakit pod` ，执行以下命令：

```bash
datakit monitor
```

<figure markdown>
  ![](https://static.<<<custom_key.brand_main_domain>>>/images/datakit/datakit-crd-ingress.png){ width="800" }
  <figcaption> Ingress 数据采集 </figcaption>
</figure>

也可以登录 [<<<custom_key.brand_name>>>平台](https://www.<<<custom_key.brand_main_domain>>>/){:target="_blank"} ,【指标】-【查看器】查看指标数据
