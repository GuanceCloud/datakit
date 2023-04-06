{{.CSS}}
# Kubernetes
---

本文档介绍如何在 K8s 中通过 DaemonSet 方式安装 DataKit。

## 安装 {#install}

=== "Daemonset"

    先下载 [datakit.yaml](https://static.guance.com/datakit/datakit.yaml){:target="_blank"}，其中开启了很多[默认采集器](datakit-input-conf.md#default-enabled-inputs)，无需配置。
    
    ???+ attention
    
        如果要修改这些采集器的默认配置，可通过 [Configmap 方式挂载单独的 conf](k8s-config-how-to.md#via-configmap-conf) 来配置。部分采集器可以直接通过环境变量的方式来调整，具体参见具体采集器的文档。总而言之，不管是默认开启的采集器，还是其它采集器，在 DaemonSet 方式部署 DataKit 时，通过 [Configmap](https://kubernetes.io/docs/tasks/configure-pod-container/configure-pod-configmap/){:target="_blank"} 来配置采集器总是生效的。
    
    修改 `datakit.yaml` 中的 dataway 配置
    
    ```yaml
    	- name: ENV_DATAWAY
    		value: https://openway.guance.com?token=<your-token> # 此处填上 DataWay 真实地址
    ```
    
    如果选择的是其它节点，此处更改对应的 DataWay 地址即可，如 AWS 节点：
    
    ```yaml
    	- name: ENV_DATAWAY
    		value: https://aws-openway.guance.com?token=<your-token> 
    ```
    
    安装 yaml
    
    ```shell
    $ kubectl apply -f datakit.yaml
    ```
    
    安装完后，会创建一个 datakit 的 DaemonSet 部署：
    
    ```shell
    $ kubectl get pod -n datakit
    ```

=== "Helm"

    前提条件
    
    * Kubernetes >= 1.14
    * Helm >= 3.0+
    
    添加 DataKit Helm 仓库：
    
    ```shell 
    $ helm repo add datakit  https://pubrepo.guance.com/chartrepo/datakit
    $ helm repo update 
    ```
    
    Helm 安装 Datakit（注意修改 `datakit.dataway_url` 参数）
    
    ```shell
    $ helm install datakit datakit/datakit -n datakit --set datakit.dataway_url="https://openway.guance.com?token=<your-token>" --create-namespace 
    ```
    
    查看部署状态：
    
    ```shell
    $ helm -n datakit list
    ```
    
    可以通过如下命令来升级：
    
    ```shell
    $ helm repo update 
    $ helm upgrade datakit datakit/datakit -n datakit --set datakit.dataway_url="https://openway.guance.com?token=<your-token>" 
    ```
    
    可以通过如下命令来卸载：
    
    ```shell
    $ helm uninstall datakit -n datakit
    ```

## 资源限制 {#requests-limits}

DataKit 默认设置了 Requests 和 Limits，如果 DataKit 容器状态变为 OOMKilled ，可自定义修改配置。

=== "Yaml"

    datakit.yaml 中其大概格式为
    
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
 
具体配置，参见[官方文档](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#requests-and-limits){:target="_blank"}。


## Kubernetes 污点容忍度配置 {#toleration}

DataKit 默认会在 Kubernetes 集群的所有 node 上部署（即忽略所有污点），如果 Kubernetes 中某些 node 节点添加了污点调度，且不希望在其上部署 DataKit，可修改 datakit.yaml，调整其中的污点容忍度：

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

## DataKit 中其它环境变量设置 {#using-k8-env}

> 注意： ENV_LOG 如果配置成 `stdout`，则不要将 ENV_LOG_LEVEL 设置成 `debug`，否则可能循环产生日志，产生大量日志数据。

在 DaemonSet 模式中，DataKit 支持多个环境变量配置

- datakit.yaml 中其大概格式为

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
- json：一些较为复杂的配置，需要以 json 字符串形式来设置环境变量
- bool：开关类型，给定**任何非空字符串**即表示开启该功能，建议均以 `"on"` 作为其开启时的取值。如果不开启，必须将其删除或注释掉。
- string-list：以英文逗号分割的字符串，一般用于表示列表
- duration：一种字符串形式的时间长度表示，比如 `10s` 表示 10 秒，这里的单位支持 h/m/s/ms/us/ns。**不要给负数**。
- int：整数类型
- float：浮点类型

对于 string/bool/string-list/duration，建议都用双引号修饰一下，避免 k8s 解析 yaml 可能导致的问题。

### 最常用环境变量 {#env-common}

| 环境变量名称                              | 类型        | 默认值 | 必须   | 说明                                                                                                  |
| ---------:                                | ----:       | ---:   | ------ | ----                                                                                                  |
| `ENV_DATAWAY`                             | string      | 无     | 是     | 配置 DataWay 地址，如 `https://openway.guance.com?token=xxx`                                          |
| `ENV_DEFAULT_ENABLED_INPUTS`              | string-list | 无     | 否     | 默认开启[采集器列表](datakit-input-conf.md#default-enabled-inputs)，以英文逗号分割，如 `cpu,mem,disk` |
| `ENV_ENABLE_INPUTS` :fontawesome-solid-x: | string-list | 无     | 否     | 同 ENV_DEFAULT_ENABLED_INPUTS，将废弃                                                                 |
| `ENV_GLOBAL_HOST_TAGS`                    | string-list | 无     | 否     | 全局 tag，多个 tag 之间以英文逗号分割，如 `tag1=val,tag2=val2`                                        |
| `ENV_GLOBAL_TAGS` :fontawesome-solid-x:   | string-list | 无     | 否     | 同 ENV_GLOBAL_HOST_TAGS，将废弃                                                                       |

???+ note "区分*全局主机 tag* 和*全局选举 tag*"

    `ENV_GLOBAL_HOST_TAGS` 用来指定主机类全局 tag，这些 tag 的值一般跟随主机变迁，比如主机名、主机 IP 等。当然，其它不跟随主机变迁的 tag 也能加进来。所有非选举类采集器，会默认带上 `ENV_GLOBAL_HOST_TAGS` 中指定的 tag。

    而 `ENV_GLOBAL_ELECTION_TAGS` 建议只添加不随主机切换而变迁的 tags，如集群名、项目名等。对于[参与选举的采集器](election.md#inputs)，只会添加 `ENV_GLOBAL_ELECTION_TAGS` 中指定的 tag，不会增加 `ENV_GLOBAL_HOST_TAGS` 中指定的 tag。

    不管是主机类全局 tag 还是环境类全局 tag，如果原始数据中已经有对应 tag，则不会追加已存在的 tag，我们认为应该沿用原始数据中的 tag。

### 日志配置相关环境变量 {#env-log}

| 环境变量名称            | 类型   | 默认值                     | 必须   | 说明                                                             |
| ---------:              | ----:  | ---:                       | ------ | ----                                                             |
| `ENV_GIN_LOG`           | string | */var/log/datakit/gin.log* | 否     | 如果改成 `stdout`，DataKit 自身 gin 日志将不写文件，而是终端输出 |
| `ENV_LOG`               | string | */var/log/datakit/log*     | 否     | 如果改成 `stdout`，DatakIt 自身日志将不写文件，而是终端输出      |
| `ENV_LOG_LEVEL`         | string | info                       | 否     | 设置 DataKit 自身日志等级，可选 `info/debug`                     |
| `ENV_DISABLE_LOG_COLOR` | bool   | -                          | 否     | 关闭日志颜色                                                     |

###  DataKit pprof 相关 {#env-pprof}

| 环境变量名称       | 类型   | 默认值 | 必须   | 说明                |
| ---------:         | ----:  | ---:   | ------ | ----                |
| `ENV_ENABLE_PPROF` | bool   | -      | 否     | 是否开启 `pprof`    |
| `ENV_PPROF_LISTEN` | string | 无     | 否     | `pprof`服务监听地址 |

### 选举相关环境变量 {#env-elect}

| 环境变量名称                        | 类型        | 默认值    | 必须   | 说明                                                                                                                                                                                       |
| ---------:                          | ----:       | ---:      | ------ | ----                                                                                                                                                                                       |
| `ENV_ENABLE_ELECTION`               | bool        | -         | 否     | 开启[选举](election.md)，默认不开启，如需开启，给该环境变量任意一个非空字符串值即可                                                                                                        |
| `ENV_NAMESPACE`                     | string      | `default` | 否     | DataKit 所在的命名空间，默认为空表示不区分命名空间，接收任意非空字符串，如 `dk-namespace-example`。如果开启了选举，可以通过此环境变量指定工作空间。                                        |
| `ENV_ENABLE_ELECTION_NAMESPACE_TAG` | bool        | -         | 否     | 开启该选项后，所有选举类的采集均会带上 `election_namespace=<your-election-namespace>` 的额外 tag，这可能会导致一些时间线的增长（[:octicons-tag-24: Version-1.4.7](changelog.md#cl-1.4.7)） |
| `ENV_GLOBAL_ELECTION_TAGS`          | string-list | 无        | 否     | 全局选举 tag，多个 tag 之间以英文逗号分割，如 `tag1=val,tag2=val2`。ENV_GLOBAL_ENV_TAGS 将被弃用                                                                                           |
| `ENV_CLUSTER_NAME_K8S`              | string      | -         | 否     | DataKit 所在的 cluster，如果非空，会在 `global_election_tags` 添加一个指定 tag，key 是 `cluster_name_k8s`，value 是环境变量的值。（[:octicons-tag-24: Version-1.5.8](changelog.md#cl-1.5.8)）|
### HTTP/API 相关环境变量 {#env-http-api}

| 环境变量名称                     | 类型        | 默认值            | 必须   | 说明                                                                                                                                                                                                        |
| ---------:                       | ----:       | ---:              | ------ | ----                                                                                                                                                                                                        |
| `ENV_DISABLE_404PAGE`            | bool        | -                 | 否     | 禁用 DataKit 404 页面（公网部署 DataKit RUM 时常用）                                                                                                                                                        |
| `ENV_HTTP_LISTEN`                | string      | localhost:9529    | 否     | 可修改地址，使得外部可以调用 [DataKit 接口](apis.md)                                                                                                                                                           |
| `ENV_HTTP_PUBLIC_APIS`           | string-list | 无                | 否     | 允许外部访问的 DataKit [API 列表](apis.md)，多个 API 之间以英文逗号分割。当 DataKit 部署在公网时，用来禁用部分 API                                                                                             |
| `ENV_HTTP_TIMEOUT`               | duration    | 30s               | 否     | 设置 9529 HTTP API 服务端超时时间 [:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6) · [:octicons-beaker-24: Experimental](index.md#experimental)                                                     |
| `ENV_HTTP_CLOSE_IDLE_CONNECTION` | bool        | -                 | 否     | 如果开启，则 9529 HTTP server 会主动关闭闲置连接（闲置时间等同于 `ENV_HTTP_TIMEOUT`） [:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6) · [:octicons-beaker-24: Experimental](index.md#experimental) |
| `ENV_REQUEST_RATE_LIMIT`         | float       | 无                | 否     | 限制 9529 [API 每秒请求数](datakit-conf.md#set-http-api-limit)                                                                                                                                              |
| `ENV_RUM_ORIGIN_IP_HEADER`       | string      | `X-Forwarded-For` | 否     | RUM 专用                                                                                                                                                                                                    |
| `ENV_RUM_APP_ID_WHITE_LIST`      | string      | 无                | 否     | RUM app-id 白名单列表，以 `,` 分割，如 `appid-1,appid-2`                                                                                                                                                    |

### Confd 配置相关环境变量 {#env-confd}

| 环境变量名                 | 类型   | 适用场景            | 说明     | 样例值 |
| ----                     | ----   | ----               | ----     | ---- |
| ENV_CONFD_BACKEND        | string |  全部              | 后端源类型  | `etcdv3`或`zookeeper`或`redis`或`consul` |
| ENV_CONFD_BASIC_AUTH     | string | `etcdv3`或`consul` | 可选      | |
| ENV_CONFD_CLIENT_CA_KEYS | string | `etcdv3`或`consul` | 可选      | |
| ENV_CONFD_CLIENT_CERT    | string | `etcdv3`或`consul` | 可选      | |
| ENV_CONFD_CLIENT_KEY     | string | `etcdv3`或`consul`或`redis` | 可选      | |
| ENV_CONFD_BACKEND_NODES  | string |  全部              | 后端源地址 | `["IP地址:2379","IP地址2:2379"]` (nacos 加 http://) |
| ENV_CONFD_PASSWORD       | string | `etcdv3`或`consul`或`nacos` | 可选      |  |
| ENV_CONFD_SCHEME         | string | `etcdv3`或`consul` | 可选      |  |
| ENV_CONFD_SEPARATOR      | string | `redis`            | 可选默认0 |  |
| ENV_CONFD_USERNAME       | string | `etcdv3`或`consul`或`nacos` | 可选      |  |
| ENV_CONFD_ACCESS_KEY         | string | `nacos`或`aws` | 可选                    |          |
| ENV_CONFD_SECRET_KEY         | string | `nacos`或`aws` | 可选                    |          |
| ENV_CONFD_CIRCLE_INTERVAL    | string | `nacos`或`aws` | 可选                    | 默认 60   |
| ENV_CONFD_CONFD_NAMESPACE    | string | `nacos` | 配置信息空间ID        | `6aa36e0e-bd57-4483-9937-e7c0ccf59599` |
| ENV_CONFD_PIPELINE_NAMESPACE | string | `nacos` | `pipeline`信息空间ID | `d10757e6-aa0a-416f-9abf-e1e1e8423497` |
| ENV_CONFD_REGION             | string | `aws`   | AWS 服务区           | `cn-north-1` |

### Git 配置相关环境变量 {#env-git}

| 环境变量名称       | 类型     | 默认值 | 必须   | 说明                                                                                                   |
| ---------:         | ----:    | ---:   | ------ | ----                                                                                                   |
| `ENV_GIT_BRANCH`   | string   | 无     | 否     | 指定拉取的分支。<stong>为空则是默认</strong>，默认是远程指定的主分支，一般是 `master`。                |
| `ENV_GIT_INTERVAL` | duration | 无     | 否     | 定时拉取的间隔。（如 `1m`）                                                                            |
| `ENV_GIT_KEY_PATH` | string   | 无     | 否     | 本地 PrivateKey 的全路径。（如 `/Users/username/.ssh/id_rsa`）                                         |
| `ENV_GIT_KEY_PW`   | string   | 无     | 否     | 本地 PrivateKey 的使用密码。（如 `passwd`）                                                            |
| `ENV_GIT_URL`      | string   | 无     | 否     | 管理配置文件的远程 git repo 地址。（如 `http://username:password@github.com/username/repository.git`） |

### Sinker 配置相关环境变量 {#env-sinker}

| 环境变量名称 | 类型         | 默认值 | 必须   | 说明                             |
| ---------:   | ----:        | ---:   | ------ | ----                             |
| `ENV_SINKER` | string(JSON) | 无     | 否     | 安装时指定 Dataway Sinker 的配置 |

Sinker 用来指定 [dataway 的 sinker 配置](datakit-sink-dataway.md)，它是一个形如下面的 JSON 格式：

```json
[
	{
		"categories": ["L", "M"],
		"filters": [
			"{measurement='cpu' and tag='some-host'}"
		],
		"proxy": "",
		"url": "http://openway.guance.com?token=<YOUR-TOKEN>"
	}
]
```

在配置 ENV 时，我们需要将其变成一行：

```json
[ { "categories": ["L", "M"], "filters": [ "{measurement='cpu' and tag='some-host'}" ], "url": "http://openway.guance.com?token=<YOUR-TOKEN>" } ]
```

如果将这一行 JSON 应用在命令行中，需要对其中的 `"` 进行转义，如：

```shell
DK_SINKER="[ { \"categories\": [\"L\", \"M\"], \"filters\": [ \"{measurement='cpu' and tag='some-host'}\" ], \"url\": \"http://openway.guance.com?token=<YOUR-TOKEN>\" } ]"
```

### IO 模块配置相关环境变量 {#env-io}

| 环境变量名称                  | 类型     | 默认值             | 必须   | 说明                                                                         |
| ---------:                    | ---:     | ---:               | ------ | ----                                                                         |
| `ENV_IO_FILTERS`              | json     | 无                 | 否     | 添加[行协议过滤器](datakit-filter.md)                                        |
| `ENV_IO_FLUSH_INTERVAL`       | duration | 10s                | 否     | IO 发送时间频率                                                              |
| `ENV_IO_FLUSH_WORKERS`        | int      | `cpu_core * 2 + 1` | 否     | IO 发送 worker 数（:octicons-tag-24: Version-1.5.9](changelog.md#cl-1.5.9)） |
| `ENV_IO_MAX_CACHE_COUNT`      | int      | 1000               | 否     | 发送 buffer（点数）大小                                                      |
| `ENV_IO_ENABLE_CACHE`         | bool     | false              | 否     | 是否开启发送失败的磁盘缓存                                                   |
| `ENV_IO_CACHE_ALL`            | bool     | false              | 否     | 是否 cache 所有发送失败的数据                                                |
| `ENV_IO_CACHE_MAX_SIZE_GB`    | int      | 10                 | 否     | 发送失败缓存的磁盘大小（单位 GB）                                            |
| `ENV_IO_CACHE_CLEAN_INTERVAL` | duration | 5s                 | 否     | 定期发送缓存在磁盘内的失败任务                                               |

???+ note "关于 buffer 和 queue 的说明"

    `ENV_IO_MAX_CACHE_COUNT` 用来控制数据的发送策略，即当内存中 cache 的（行协议）点数超过该数值的时候，就会尝试将内存中当前 cache 的点数发送到中心。如果该 cache 的阈值调的太大，数据就都堆积在内存，导致内存飙升，但会提高 GZip 的压缩效果。如果太小，可能影响发送吞吐率。

`ENV_IO_FILTERS` 是一个 json 字符串，示例如下:

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

| 环境变量名称         | 类型   | 默认值         | 必须   | 说明                                                                                                 |
| ---------:           | ----:  | ---:           | ------ | ----                                                                                                 |
| `ENV_DCA_LISTEN`     | string | localhost:9531 | 否     | 可修改改地址，使得 [DCA](dca.md) 客户端能管理该 DataKit，一旦开启 ENV_DCA_LISTEN 即默认启用 DCA 功能 |
| `ENV_DCA_WHITE_LIST` | string | 无             | 否     | 配置 DCA 白名单，以英文逗号分隔                                                                      |

### Refer Table 有关环境变量 {#env-reftab}

| 环境变量名称                      | 类型   | 默认值 | 必须   | 说明                                                    |
| ---------:                        | ----:  | ---:   | ------ | ----                                                    |
| `ENV_REFER_TABLE_URL`             | string | 无     | 否     | 设置数据源 URL                                          |
| `ENV_REFER_TABLE_PULL_INTERVAL`   | string | 5m     | 否     | 设置数据源 URL 的请求时间间隔                           |
| `ENV_REFER_TABLE_USE_SQLITE`      | bool   | false  | 否     | 设置是否使用 SQLite 保存数据                            |
| `ENV_REFER_TABLE_SQLITE_MEM_MODE` | bool   | false  | 否     | 当使用 SQLite 保存数据时，使用 SQLite 内存模式/磁盘模式 |

### 其它杂项 {#env-others}

| 环境变量名称                    | 类型     | 默认值 | 必须   | 说明                                                       |
| ---------:                      | ----:    | ---:   | ------ | ----                                                       |
| `ENV_CLOUD_PROVIDER`            | string   | 无     | 否     | 支持安装阶段填写云厂商(`aliyun/aws/tencent/hwcloud/azure`) |
| `ENV_HOSTNAME`                  | string   | 无     | 否     | 默认为本地主机名，可安装时指定，如， `dk-your-hostname`    |
| `ENV_IPDB`                      | string   | 无     | 否     | 指定 IP 信息库类型，目前只支持 `iploc/geolite2` 两种       |
| `ENV_ULIMIT`                    | int      | 无     | 否     | 指定 Datakit 最大的可打开文件数                            |
| `ENV_DATAWAY_TIMEOUT`           | duration | 30s    | 否     | 设置 DataKit 请求 DataWay 的超时时间                       |
| `ENV_DATAWAY_ENABLE_HTTPTRACE`  | bool     | false  | 否     | 在 debug 日志中输出 dataway HTTP 请求的网络日志            |
| `ENV_DATAWAY_HTTP_PROXY`        | string   | 无     | 否     | 设置 DataWay HTTP 代理                                     |

### 特殊环境变量 {#env-special}

#### ENV_K8S_NODE_NAME {#env_k8s_node_name}

当 k8s node 名称跟其对应的主机名不同时，可将 k8s 的 node 名称顶替默认采集到的主机名，在 *datakit.yaml* 中增加环境变量：

> [1.2.19](changelog.md#cl-1.2.19) 版本的 datakit.yaml 中默认就带了这个配置，如果是从老版本的 yaml 直接升级而来，需要对 *datakit.yaml* 做如下手动改动。

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

- [DataKit 选举](election.md)
- [DataKit 的几种配置方式](k8s-config-how-to.md)
