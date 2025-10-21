
# Kubernetes
---

本文档介绍如何在 K8s 中通过 DaemonSet 方式安装 DataKit。

## 安装 {#install}

<!-- markdownlint-disable MD046 -->
=== "DaemonSet"

    先下载 [*datakit.yaml*](https://static.<<<custom_key.brand_main_domain>>>/datakit/datakit.yaml){:target="_blank"}，其中开启了很多[默认采集器](datakit-input-conf.md#default-enabled-inputs)，无需配置。
    
    ???+ note
    
        如果要修改这些采集器的默认配置，可通过 [ConfigMap 方式挂载单独的配置文件](k8s-config-how-to.md#via-configmap-conf) 来配置。部分采集器可以直接通过环境变量的方式来调整，具体参见具体采集器的文档。总而言之，不管是默认开启的采集器，还是其它采集器，在 DaemonSet 方式部署 DataKit 时，通过 [ConfigMap](https://kubernetes.io/docs/tasks/configure-pod-container/configure-pod-configmap/){:target="_blank"} 来配置采集器总是生效的。
    
    修改 `datakit.yaml` 中的 Dataway 配置
    
    ```yaml
    - name: ENV_DATAWAY
      value: https://openway.<<<custom_key.brand_main_domain>>>?token=<your-token> # 此处填上 DataWay 真实地址
    ```
    
    如果选择的是其它节点，此处更改对应的 Dataway 地址即可，如 AWS 节点：
    
    ```yaml
    - name: ENV_DATAWAY
      value: https://aws-openway.<<<custom_key.brand_main_domain>>>?token=<your-token> 
    ```
    
    安装 yaml
    
    ```shell
    $ kubectl apply -f datakit.yaml
    ```
    
    安装完后，会创建一个 DataKit 的 DaemonSet 部署：
    
    ```shell
    $ kubectl get pod -n datakit
    ```

=== "Helm"

    前提条件
    
    * Kubernetes >= 1.14
    * Helm >= 3.0+
    
    Helm 安装 DataKit（注意修改 `datakit.dataway_url` 参数）, 其中开启了很多[默认采集器](datakit-input-conf.md#default-enabled-inputs)，无需配置。更多 Helm 相关可参考 [Helm 管理配置](datakit-helm.md)
    
    
    ```shell
    helm install datakit datakit \
        <<<% if custom_key.brand_key == 'guance' -%>>>
        --repo  https://pubrepo.<<<custom_key.brand_main_domain>>>/chartrepo/datakit \
        <<<% else -%>>>
        --repo  https://pubrepo.<<<custom_key.brand_main_domain>>>/chartrepo/truewatch \
        <<<% endif -%>>>
        -n datakit --create-namespace \
        --set datakit.dataway_url="https://openway.<<<custom_key.brand_main_domain>>>?token=<YOUR-TOKEN>" 
    ```
    
    查看部署状态：
    
    ```shell
    helm -n datakit list
    ```
    
    可以通过如下命令来升级：
    
    ```shell
    helm -n datakit get  values datakit -o yaml > values.yaml
    helm upgrade datakit datakit \
        <<<% if custom_key.brand_key == 'guance' -%>>>
        --repo  https://pubrepo.<<<custom_key.brand_main_domain>>>/chartrepo/datakit \
        <<<% else -%>>>
        --repo  https://pubrepo.<<<custom_key.brand_main_domain>>>/chartrepo/truewatch \
        <<<% endif -%>>>
        -n datakit \
        -f values.yaml
    ```
    
    可以通过如下命令来卸载：
    
    ```shell
    helm uninstall datakit -n datakit
    ```

    ### 更多 Helm 示例 {#helm-examples}

    除了手动编辑上面的 *values.yaml* 来调整 DataKit 配置（如果遇到复杂转义操作，还是建议直接用 *values.yaml* 来操作），还可以在 Helm 安装阶段就指定这些参数。需要注意的是，这些参数的设置需符合 Helm 的命令行语法。

    **设置默认的采集器列表**

    ```shell
    helm install datakit datakit \
        <<<% if custom_key.brand_key == 'guance' -%>>>
        --repo  https://pubrepo.<<<custom_key.brand_main_domain>>>/chartrepo/datakit \
        <<<% else -%>>>
        --repo  https://pubrepo.<<<custom_key.brand_main_domain>>>/chartrepo/truewatch \
        <<<% endif -%>>>
        -n datakit --create-namespace \
        --set datakit.dataway_url="https://openway.<<<custom_key.brand_main_domain>>>?token=<your-token>" \
        --set datakit.default_enabled_inputs="statsd\,dk\,cpu\,mem"
    ```

    注意，此处需要将 `,` 转义一下，不然 Helm 会报错。

    **设置环境变量**

    DataKit 支持非常多的[环境变量设置](datakit-daemonset-install.md#env-setting)，我们可以用如下的方式来追加一组环境变量设置：

    ```shell
    helm install datakit datakit \
        <<<% if custom_key.brand_key == 'guance' -%>>>
        --repo  https://pubrepo.<<<custom_key.brand_main_domain>>>/chartrepo/datakit \
        <<<% else -%>>>
        --repo  https://pubrepo.<<<custom_key.brand_main_domain>>>/chartrepo/truewatch \
        <<<% endif -%>>>
        -n datakit --create-namespace \
        --set datakit.dataway_url="https://openway.<<<custom_key.brand_main_domain>>>?token=tkn_xxx" \
        --set "extraEnvs[0].name=ENV_INPUT_OTEL_GRPC" \
        --set 'extraEnvs[0].value=\{"trace_enable":true\,"metric_enable":true\,"addr":"0.0.0.0:4317"\}' \
        --set "extraEnvs[1].name=ENV_INPUT_CPU_PERCPU" \
        --set 'extraEnvs[1].value=true'
    ```

    此处 `extraEnvs` 是 DataKit Helm 包中定义的设置环境变量的入口，由于环境变量是数组结构，故此处我们用数组下标（从 0 开始）的方式来追加多个环境变量。其中 `name` 即环境变量名，`value` 即对应的值。值得注意的是，某些环境变量的值是 JSON 字符串，此处我们也要注意对一些字符（比 `{},` 等字符）做转义。

    **安装指定版本**

    可以通过 `image.tag` 来指定 DataKit 镜像版本号：

    ```shell
    helm install datakit datakit \
        <<<% if custom_key.brand_key == 'guance' -%>>>
        --repo  https://pubrepo.<<<custom_key.brand_main_domain>>>/chartrepo/datakit \
        <<<% else -%>>>
        --repo  https://pubrepo.<<<custom_key.brand_main_domain>>>/chartrepo/truewatch \
        <<<% endif -%>>>
        -n datakit --create-namespace \
        --set image.tag="1.70.0" \
        ...
    ```
<!-- markdownlint-enable -->

### 资源限制 {#requests-limits}

DataKit 默认设置了 Requests 和 Limits，如果 DataKit 容器状态变为 OOMKilled ，可自定义修改配置。

<!-- markdownlint-disable MD046 -->
=== "Yaml"

    *datakit.yaml* 中其大概格式为
    
    ```yaml
    ...
            resources:
              requests:
                cpu: "200m"
                memory: "128Mi"
              limits:
                cpu: "2000m"
                memory: "4Gi"
    ...
    ```

=== "Helm"

    Helm values.yaml 中其大概格式为
    
    ```yaml
    ...
    resources:
      requests:
        cpu: "200m"
        memory: "128Mi"
      limits:
        cpu: "2000m"
        memory: "4Gi"
    ...
    ```
<!-- markdownlint-enable -->

具体配置，参见[官方文档](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#requests-and-limits){:target="_blank"}。

### 安全限制 {#security-context}

DataKit 推荐以 root 用户和特权模式运行，同时也提供非 root 用户运行和权限控制，但这会影响部分数据采集。通过以下步骤，可以降低 DataKit 容器的权限级别，同时确保能采集网络数据、容器数据等。

前提条件：

- Kubernetes 集群
- 节点访问权限（用于执行主机级配置）
- DataKit 镜像已包含 UID 为 10001 的 `datakit` 用户（DataKit 1.83.0 及以后的镜像已创建 `datakit` 用户，默认用户仍是 `root`）

配置步骤如下：

1. 在宿主机上创建用户组并设置权限，在每个 Kubernetes 节点上执行以下命令：

    ```bash
    # 创建专用用户组
    groupadd datakit-reader
    
    # 设置日志目录权限
    chgrp -R datakit-reader /var/log/pods
    chmod -R g+rx /var/log/pods

    # 设置 Docker 套接字权限（如使用）
    chgrp datakit-reader /var/run/docker.sock
    chmod g+r /var/run/docker.sock
    
    # 设置 Containerd 套接字权限
    chgrp datakit-reader /var/run/containerd/containerd.sock
    chmod g+r /var/run/containerd/containerd.sock
    
    # 设置 CRI-O 套接字权限（如使用）
    chgrp datakit-reader /var/run/crio/crio.sock
    chmod g+r /var/run/crio/crio.sock
    
    # 设置 Kubelet 目录权限
    chgrp -R datakit-reader /var/lib/kubelet/pods
    chmod -R g+rx /var/lib/kubelet/pods
    ```

1. 获取用户组 GID，在每个节点上执行以下命令获取 `datakit-reader` 组的 GID：

    ```bash
    getent group datakit-reader | cut -d: -f3
    ```

记下输出的 GID 值（例如 `12345`），后续步骤中需要使用。

1. 配置 Kubernetes Deployment/DaemonSet，更新您的 DataKit Kubernetes 配置文件：

    ```yaml
    apiVersion: apps/v1
    kind: DaemonSet  # 或 Deployment
    metadata:
      name: datakit
      namespace: monitoring
    spec:
      template:
        spec:
          # 安全上下文配置
          securityContext:
            runAsUser: 10001  # datakit 用户的 UID
            runAsGroup: 10001 # datakit 用户的 GID
            fsGroup: 10001    # 文件系统组
            supplementalGroups: [12345]  # 上一步获取的 datakit-reader 组 GID
          containers:
          - name: datakit
            image: your-datakit-image:tag
            # 容器安全上下文
            securityContext:
              privileged: false             # 关闭特权模式
              allowPrivilegeEscalation: false
              readOnlyRootFilesystem: true  # 可选：设置根文件系统为只读
              capabilities:
                drop: ["ALL"]  # 丢弃所有特权能力
                add: ["SYS_ADMIN", "SYS_PTRACE", "DAC_READ_SEARCH", "NET_RAW"]
          # 其他内容
    ```

注意：

- 在以非 root 用户运行时，DataKit 的以下功能可能会受到限制：
    - 部分系统指标无法采集：某些需要 root 权限的系统文件和目录可能无法访问
    - 容器运行时数据受限：如果容器 sock 和容器日志目录不是默认路径，需要重新挂载和授权
    - 内核级指标缺失：部分需要特权能力的系统调用无法执行
- 此配置需要在每个 Kubernetes 节点上执行主机级权限设置
- 当节点扩容时，需要在新节点上重复权限设置步骤
- 考虑使用配置管理工具（Ansible、Chef、Puppet）自动化节点配置


如果非 root 模式无法满足监控需求，可以随时回退到 root 模式：

1. 从 yaml 中移除 `runAsUser`、`runAsGroup` 和 `supplementalGroups` 等配置
1. 将 `privileged` 设置为 `true`
1. 重新部署 DataKit

### Kubernetes 污点容忍度配置 {#toleration}

DataKit 默认会在 Kubernetes 集群的所有 Node 上部署（即忽略所有污点），如果 Kubernetes 中某些 Node 节点添加了污点调度，且不希望在其上部署 DataKit，可修改 *datakit.yaml*，调整其中的污点容忍度：

```yaml
      tolerations:
      - operator: Exists    <--- 修改这里的污点容忍度
```

具体绕过策略，参见[官方文档](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration){:target="_blank"}。

## 采集器配置 {#input-config}

DataKit 在 Kubernetes 中采集器配置方式有两种：

1. ConfigMap：通过注入 ConfigMap 来追加采集器配置
1. 环境变量：通过一个特定的环境变量，注入完整的 toml 采集配置

### ConfigMap 设置 {#configmap-setting}

部分采集器的开启，需通过 ConfigMap 来注入。以下是 MySQL 和 Redis 采集器的注入示例：

```yaml
# datakit.yaml

volumeMounts: # datakit.yaml 中已有该配置，直接搜索即可定位到
- mountPath: /usr/local/datakit/conf.d/db/mysql.conf
  name: datakit-conf
  subPath: mysql.conf
    readOnly: true
- mountPath: /usr/local/datakit/conf.d/db/redis.conf
  name: datakit-conf
  subPath: redis.conf
    readOnly: true

# 直接在 datakit.yaml 底部追加
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: datakit-conf
  namespace: datakit
data:
    mysql.conf: |-
      [[inputs.mysql]]
         ...
    redis.conf: |-
      [[inputs.redis]]
         ...
```

### ENV 设置采集器 {#env-setting}

采集器的开启，也可以通过 ENV_DATAKIT_INPUTS 这个环境变量来注入。以下是 MySQL 和 Redis 采集器的注入示例：

- *datakit.yaml* 中其大概格式为

```yaml
spec:
  containers:
    - env
    - name: ENV_XXX
      value: YYY
    - name: ENV_DATAKIT_INPUTS
      value: |
        [[inputs.mysql]]
          interval = "10s"
          ...
        [inputs.mysql.tags]
          some_tag = "some_value"

        [[inputs.redis]]
          interval = "10s"
          ...
        [inputs.redis.tags]
          some_tag = "some_value"
```

- Helm values.yaml 中其大概格式为

```yaml
  extraEnvs: 
    - name: "ENV_XXX"
      value: "YYY"
    - name: "ENV_DATAKIT_INPUTS"
      value: |
        [[inputs.mysql]]
          interval = "10s"
          ...
        [inputs.mysql.tags]
          some_tag = "some_value"

        [[inputs.redis]]
          interval = "10s"
          ...
        [inputs.redis.tags]
          some_tag = "some_value"
```

注入的内容，将存入容器的 *conf.d/env_datakit_inputs.conf* 文件中。

## DataKit 主配置 {#using-k8-env}

DataKit 在 Kubernetes 中不再使用 *datkait.conf* 来配置，只能使用环境变量。在 DaemonSet 模式中，DataKit 支持多个环境变量配置

- *datakit.yaml* 中其大概格式为

```yaml
spec:
  containers:
    - env
    - name: ENV_XXX
      value: YYY
    - name: ENV_OTHER_XXX
      value: YYY
```

- Helm values.yaml 中其大概格式为

```yaml
  extraEnvs: 
    - name: "ENV_XXX"
      value: "YYY"
    - name: "ENV_OTHER_XXX"
      value: "YYY"    
```

### 环境变量类型说明 {#env-types}

以下环境变量的取值分为如下几种数据类型：

- string：字符串类型
- JSON：一些较为复杂的配置，需要以 JSON 字符串形式来设置环境变量
- bool：开关类型，给定**任何非空字符串**即表示开启该功能，建议均以 `"on"` 作为其开启时的取值。如果不开启，必须将其删除或注释掉。
- string-list：以英文逗号分割的字符串，一般用于表示列表
- duration：一种字符串形式的时间长度表示，比如 `10s` 表示 10 秒，这里的单位支持 h/m/s/ms/us/ns。**不要给负数**。
- int：整数类型
- float：浮点类型

对于 string/bool/string-list/duration，建议都用双引号修饰一下，避免 k8s 解析 yaml 可能导致的问题。

### 最常用环境变量 {#env-common}

<!-- markdownlint-disable MD046 -->
{{ CodeBlock .NonInputENVSampleZh.envCommon 0}}
<!-- markdownlint-enable -->

<!-- markdownlint-disable MD046 -->
???+ note "区分*全局主机 tag* 和*全局选举 tag*"

    `ENV_GLOBAL_HOST_TAGS` 用来指定主机类全局 tag，这些 tag 的值一般跟随主机变迁，比如主机名、主机 IP 等。当然，其它不跟随主机变迁的 tag 也能加进来。所有非选举类采集器，会默认带上 `ENV_GLOBAL_HOST_TAGS` 中指定的 tag。

    而 `ENV_GLOBAL_ELECTION_TAGS` 建议只添加不随主机切换而变迁的 tags，如集群名、项目名等。对于[参与选举的采集器](election.md#inputs)，只会添加 `ENV_GLOBAL_ELECTION_TAGS` 中指定的 tag，不会增加 `ENV_GLOBAL_HOST_TAGS` 中指定的 tag。

    不管是主机类全局 tag 还是环境类全局 tag，如果原始数据中已经有对应 tag，则不会追加已存在的 tag，我们认为应该沿用原始数据中的 tag。

???+ note "关于禁用保护模式（ENV_DISABLE_PROTECT_MODE）"

    保护模式一旦被禁用，即可以设置一些危险的配置参数，DataKit 将接受任何配置参数。这些参数可能会导致 DataKit 一些功能异常，或者影响采集器的采集功能。比如 HTTP 发送 Body 设置太小，会影响数据上传功能；某些采集器的采集频率过高，可能影响被采集的实体。
<!-- markdownlint-enable -->

<!--
### Point Pool {#env-pointpool}

[:octicons-tag-24: Version-1.28.0](changelog.md#cl-1.28.0) ·
[:octicons-beaker-24: Experimental](index.md#experimental)
-->

### Dataway {#env-dataway}

<!-- markdownlint-disable MD046 -->
{{ CodeBlock .NonInputENVSampleZh.envDataway 0}}
<!-- markdownlint-enable -->

### 日志配置 {#env-log}

<!-- markdownlint-disable MD046 -->
{{ CodeBlock .NonInputENVSampleZh.envLog 0}}
<!-- markdownlint-enable -->

### Pprof 相关 {#env-pprof}

<!-- markdownlint-disable MD046 -->
{{ CodeBlock .NonInputENVSampleZh.envPprof 0}}
<!-- markdownlint-enable -->

> `ENV_ENABLE_PPROF`：[:octicons-tag-24: Version-1.9.2](changelog.md#cl-1.9.2) 已默认开启 pprof。

### 选举 {#env-elect}

<!-- markdownlint-disable MD046 -->
{{ CodeBlock .NonInputENVSampleZh.envElect 0}}
<!-- markdownlint-enable -->

### HTTP/API {#env-http-api}

<!-- markdownlint-disable MD046 -->
{{ CodeBlock .NonInputENVSampleZh.envHTTPAPI 0}}
<!-- markdownlint-enable -->

### Confd {#env-confd}

<!-- markdownlint-disable MD046 -->
{{ CodeBlock .NonInputENVSampleZh.envConfd 0}}
<!-- markdownlint-enable -->

### Git {#env-git}

<!-- markdownlint-disable MD046 -->
{{ CodeBlock .NonInputENVSampleZh.envGit 0}}
<!-- markdownlint-enable -->

### Sinker {#env-sinker}

<!-- markdownlint-disable MD046 -->
{{ CodeBlock .NonInputENVSampleZh.envSinker 0}}
<!-- markdownlint-enable -->

### IO 模块 {#env-io}

<!-- markdownlint-disable MD046 -->
{{ CodeBlock .NonInputENVSampleZh.envIO 0}}
<!-- markdownlint-enable -->

<!-- markdownlint-disable MD046 -->
???+ note "关于 buffer 和 queue 的说明"

    `ENV_IO_MAX_CACHE_COUNT` 用来控制数据的发送策略，即当内存中 cache 的（行协议）点数超过该数值的时候，就会尝试将内存中当前 cache 的点数发送到中心。如果该 cache 的阈值调的太大，数据就都堆积在内存，导致内存飙升，但会提高 GZip 的压缩效果。如果太小，可能影响发送吞吐率。
<!-- markdownlint-enable -->

`ENV_IO_FILTERS` 是一个 JSON 字符串，示例如下：

```json
{
  "logging":[
    "{ source = 'datakit' and ( host in ['ubt-dev-01', 'tanb-ubt-dev-test'] )}",
    "{ source = 'abc' and ( host in ['ubt-dev-02', 'tanb-ubt-dev-test-1'] )}"
  ],

  "metric":[
    "{ measurement in in ['datakit', 'redis_client'] )}"
  ],
}
```

### DCA {#env-dca}

<!-- markdownlint-disable MD046 -->
{{ CodeBlock .NonInputENVSampleZh.envDca 0}}
<!-- markdownlint-enable -->

### Refer Table {#env-reftab}

<!-- markdownlint-disable MD046 -->
{{ CodeBlock .NonInputENVSampleZh.envRefta 0}}
<!-- markdownlint-enable -->

### 数据录制 {#env-recorder}

[:octicons-tag-24: Version-1.22.0](changelog.md#1.22.0)

数据录制相关的功能，参见[这里的文档](datakit-tools-how-to.md#record-and-replay)。

<!-- markdownlint-disable MD046 -->
{{ CodeBlock .NonInputENVSampleZh.envRecorder 0}}
<!-- markdownlint-enable -->

### Remote Job 远程任务 {#remote_job}

[:octicons-tag-24: Version-1.63.0](changelog.md#cl-1.63.0)

<!-- markdownlint-disable MD046 -->
{{ CodeBlock .NonInputENVSampleZh.remote_job 0}}
<!-- markdownlint-enable -->

### 其它杂项 {#env-others}

<!-- markdownlint-disable MD046 -->
{{ CodeBlock .NonInputENVSampleZh.envOthers 0}}
<!-- markdownlint-enable -->

### 特殊环境变量 {#env-special}

#### ENV_K8S_NODE_NAME {#env_k8s_node_name}

当 k8s node 名称跟其对应的主机名不同时，可将 k8s 的 node 名称顶替默认采集到的主机名，在 *datakit.yaml* 中增加环境变量：

> [1.2.19](changelog.md#cl-1.2.19) 版本的 *datakit.yaml* 中默认就带了这个配置，如果是从老版本的 yaml 直接升级而来，需要对 *datakit.yaml* 做如下手动改动。

```yaml
- env:
    - name: ENV_K8S_NODE_NAME
        valueFrom:
            fieldRef:
                apiVersion: v1
                fieldPath: spec.nodeName
```

#### ENV_K8S_CLUSTER_NODE_NAME {#env-rename-node}

[:octicons-tag-24: Version-1.36.0](changelog.md#1.36.0)

如果不同集群存在同名 Node，且这些集群的数据都打到**同一个工作空间**，可以通过 `ENV_K8S_CLUSTER_NODE_NAME` 来手动修改**采集到的 Node 名称**。在部署时，*datakit.yaml* 中位于 `ENV_K8S_NODE_NAME` **后面**新增一个配置段：

```yaml
- name: ENV_K8S_CLUSTER_NODE_NAME
  value: cluster_a_$(ENV_K8S_NODE_NAME) # 注意，此处引用的 ENV_K8S_NODE_NAME 必须在前面已有定义
```

这样之后，该集群获取到的主机名（主机对象列表）会多一个 `cluster_a_` 的前缀，除此之外，主机日志/进程/CPU/Mem 等指标集上，`host` 这个 tag 的值也都多了这个前缀。

### 采集器专用环境变量 {#inputs-envs}

部分采集器支持外部注入环境变量，以调整采集器自身的默认配置。具体参见各个具体的采集器文档。

## 延伸阅读 {#more-readings}

<font size=3>
<div class="grid cards" markdown>
- [<font color="coral"> :fontawesome-solid-arrow-right-long: &nbsp; <u>DataKit 选举</u></font>](election.md)
</div>
<div class="grid cards" markdown>
- [<font color="coral"> :fontawesome-solid-arrow-right-long: &nbsp; <u>DataKit 的几种配置方式</u></font>](k8s-config-how-to.md)
</div>
</font>
