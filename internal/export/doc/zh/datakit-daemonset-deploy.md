
# Kubernetes
---

本文档介绍如何在 K8s 中通过 DaemonSet 方式安装 DataKit。

## 安装 {#install}

<!-- markdownlint-disable MD046 -->
=== "DaemonSet"

    先下载 [*datakit.yaml*](https://static.guance.com/datakit/datakit.yaml){:target="_blank"}，其中开启了很多[默认采集器](datakit-input-conf.md#default-enabled-inputs)，无需配置。
    
    ???+ attention
    
        如果要修改这些采集器的默认配置，可通过 [ConfigMap 方式挂载单独的配置文件](k8s-config-how-to.md#via-configmap-conf) 来配置。部分采集器可以直接通过环境变量的方式来调整，具体参见具体采集器的文档。总而言之，不管是默认开启的采集器，还是其它采集器，在 DaemonSet 方式部署 Datakit 时，通过 [ConfigMap](https://kubernetes.io/docs/tasks/configure-pod-container/configure-pod-configmap/){:target="_blank"} 来配置采集器总是生效的。
    
    修改 `datakit.yaml` 中的 Dataway 配置
    
    ```yaml
    - name: ENV_DATAWAY
      value: https://openway.guance.com?token=<your-token> # 此处填上 DataWay 真实地址
    ```
    
    如果选择的是其它节点，此处更改对应的 Dataway 地址即可，如 AWS 节点：
    
    ```yaml
    - name: ENV_DATAWAY
      value: https://aws-openway.guance.com?token=<your-token> 
    ```
    
    安装 yaml
    
    ```shell
    $ kubectl apply -f datakit.yaml
    ```
    
    安装完后，会创建一个 Datakit 的 DaemonSet 部署：
    
    ```shell
    $ kubectl get pod -n datakit
    ```

=== "Helm"

    前提条件
    
    * Kubernetes >= 1.14
    * Helm >= 3.0+
    
    Helm 安装 Datakit（注意修改 `datakit.dataway_url` 参数）, 其中开启了很多[默认采集器](datakit-input-conf.md#default-enabled-inputs)，无需配置。更多 Helm 相关可参考 [Helm 管理配置](datakit-helm.md)
    
    
    ```shell
    $ helm install datakit datakit \
         --repo  https://pubrepo.guance.com/chartrepo/datakit \
         -n datakit --create-namespace \
         --set datakit.dataway_url="https://openway.guance.com?token=<your-token>" 
    ```
    
    查看部署状态：
    
    ```shell
    $ helm -n datakit list
    ```
    
    可以通过如下命令来升级：
    
    ```shell
    $ helm -n datakit get  values datakit -o yaml > values.yaml
    $ helm upgrade datakit datakit \
        --repo  https://pubrepo.guance.com/chartrepo/datakit \
        -n datakit \
        -f values.yaml
    ```
    
    可以通过如下命令来卸载：
    
    ```shell
    $ helm uninstall datakit -n datakit
    ```
<!-- markdownlint-enable -->

## 资源限制 {#requests-limits}

Datakit 默认设置了 Requests 和 Limits，如果 Datakit 容器状态变为 OOMKilled ，可自定义修改配置。

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

## Kubernetes 污点容忍度配置 {#toleration}

Datakit 默认会在 Kubernetes 集群的所有 Node 上部署（即忽略所有污点），如果 Kubernetes 中某些 Node 节点添加了污点调度，且不希望在其上部署 Datakit，可修改 *datakit.yaml*，调整其中的污点容忍度：

```yaml
      tolerations:
      - operator: Exists    <--- 修改这里的污点容忍度
```

具体绕过策略，参见[官方文档](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration){:target="_blank"}。

## ConfigMap 设置 {#configmap-setting}

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

## ENV 设置采集器 {#env-setting}

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

注入的内容，将存入容器的 conf.d/env_datakit_inputs.conf 文件中。

## Datakit 中其它环境变量设置 {#using-k8-env}

> 注意： ENV_LOG 如果配置成 `stdout`，则不要将 ENV_LOG_LEVEL 设置成 `debug`，否则可能循环产生日志，产生大量日志数据。

在 DaemonSet 模式中，Datakit 支持多个环境变量配置

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

???+ attention "关于禁用保护模式（ENV_DISABLE_PROTECT_MODE）"

    保护模式一旦被禁用，即可以设置一些危险的配置参数，Datakit 将接受任何配置参数。这些参数可能会导致 Datakit 一些功能异常，或者影响采集器的采集功能。比如 HTTP 发送 Body 设置太小，会影响数据上传功能；某些采集器的采集频率过高，可能影响被采集的实体。
<!-- markdownlint-enable -->

### Point Pool 配置相关环境变量 {#env-pointpool}

[:octicons-tag-24: Version-1.28.0](changelog.md#cl-1.28.0) ·
[:octicons-beaker-24: Experimental](index.md#experimental)

<!-- markdownlint-disable MD046 -->
{{ CodeBlock .NonInputENVSampleZh.envPointPool 0}}
<!-- markdownlint-enable -->

### Dataway 配置相关环境变量 {#env-dataway}

<!-- markdownlint-disable MD046 -->
{{ CodeBlock .NonInputENVSampleZh.envDataway 0}}
<!-- markdownlint-enable -->

### 日志配置相关环境变量 {#env-log}

<!-- markdownlint-disable MD046 -->
{{ CodeBlock .NonInputENVSampleZh.envLog 0}}
<!-- markdownlint-enable -->

### Pprof 相关 {#env-pprof}

<!-- markdownlint-disable MD046 -->
{{ CodeBlock .NonInputENVSampleZh.envPprof 0}}
<!-- markdownlint-enable -->

> `ENV_ENABLE_PPROF`：[:octicons-tag-24: Version-1.9.2](changelog.md#cl-1.9.2) 已默认开启 pprof。

### 选举相关环境变量 {#env-elect}

<!-- markdownlint-disable MD046 -->
{{ CodeBlock .NonInputENVSampleZh.envElect 0}}
<!-- markdownlint-enable -->

### HTTP/API 相关环境变量 {#env-http-api}

<!-- markdownlint-disable MD046 -->
{{ CodeBlock .NonInputENVSampleZh.envHTTPAPI 0}}
<!-- markdownlint-enable -->

### Confd 配置相关环境变量 {#env-confd}

<!-- markdownlint-disable MD046 -->
{{ CodeBlock .NonInputENVSampleZh.envConfd 0}}
<!-- markdownlint-enable -->

### Git 配置相关环境变量 {#env-git}

<!-- markdownlint-disable MD046 -->
{{ CodeBlock .NonInputENVSampleZh.envGit 0}}
<!-- markdownlint-enable -->

### Sinker 配置相关环境变量 {#env-sinker}

<!-- markdownlint-disable MD046 -->
{{ CodeBlock .NonInputENVSampleZh.envSinker 0}}
<!-- markdownlint-enable -->

### IO 模块配置相关环境变量 {#env-io}

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

### DCA 相关环境变量 {#env-dca}

<!-- markdownlint-disable MD046 -->
{{ CodeBlock .NonInputENVSampleZh.envDca 0}}
<!-- markdownlint-enable -->

### Refer Table 有关环境变量 {#env-reftab}

<!-- markdownlint-disable MD046 -->
{{ CodeBlock .NonInputENVSampleZh.envRefta 0}}
<!-- markdownlint-enable -->

### 数据录制有关环境变量 {#env-recorder}

[:octicons-tag-24: Version-1.22.0](changelog.md#1.22.0)

数据录制相关的功能，参见[这里的文档](datakit-tools-how-to.md#record-and-replay)。

<!-- markdownlint-disable MD046 -->
{{ CodeBlock .NonInputENVSampleZh.envRecorder 0}}
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

### 各个采集器专用环境变量 {#inputs-envs}

部分采集器支持外部注入环境变量，以调整采集器自身的默认配置。具体参见各个具体的采集器文档。

## 延伸阅读 {#more-readings}

- [Datakit 选举](election.md)
- [Datakit 的几种配置方式](k8s-config-how-to.md)
