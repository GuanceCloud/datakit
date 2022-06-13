{{.CSS}}
# DaemonSet 安装 DataKit 
---

- DataKit 版本：{{.Version}}
- 操作系统支持：Linux

本文档介绍如何在 K8s 中通过 DaemonSet 方式安装 DataKit。

## 安装步骤 

- Helm 安装
- 普通 yaml 安装

### Helm 安装

#### 前提条件

* Kubernetes >= 1.14
* Helm >= 3.0+

#### 添加 DataKit Helm 仓库

```shell 
$ helm repo add datakit  https://pubrepo.guance.com/chartrepo/datakit
$ helm repo update 
```

#### Helm 安装 Datakit

```shell
$ helm install datakit datakit/datakit -n datakit --set datakit.dataway_url="https://openway.guance.com?token=<your-token>" --create-namespace 
```

> 注意修改 `datakit.dataway_url` 参数。

具体执行如下：

```
$ helm install datakit datakit/datakit -n datakit --set datakit.dataway_url="https://openway.guance.com?token=xxxxxxxxx" --create-namespace 
```

#### 查看部署状态

```shell
$ helm -n datakit list
```

#### 升级

```shell
$ helm repo update 
$ helm upgrade datakit datakit/datakit -n datakit --set datakit.dataway_url="https://openway.guance.com?token=<your-token>" 
```

#### 卸载

```shell
$ helm uninstall datakit -n datakit
```

### 普通 yaml 安装

先下载 [datakit.yaml](https://static.guance.com/datakit/datakit.yaml)，其中开启了很多[默认采集器](datakit-input-conf#default-enabled-inputs)，无需配置。

> 如果要修改这些采集器的默认配置，可通过 [Configmap 方式挂载单独的 conf](k8s-config-how-to.md#via-configmap-conf) 来配置。部分采集器可以直接通过环境变量的方式来调整，具体参见具体采集器的文档（[容器采集器示例](container#5cf8fecf)）。总而言之，不管是默认开启的采集器，还是其它采集器，在 DaemonSet 方式部署 DataKit 时，==通过 [Configmap](https://kubernetes.io/docs/tasks/configure-pod-container/configure-pod-configmap/) 来配置采集器总是生效的==

#### 修改配置

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

#### 安装 yaml

```shell
$ kubectl apply -f datakit.yaml
```

#### 查看运行状态

安装完后，会创建一个 datakit 的 DaemonSet 部署：

```shell
$ kubectl get pod -n datakit
```

#### Kubernetes 污点容忍度配置

DataKit 默认会在 Kubernetes 集群的所有 node 上部署（即忽略所有污点），如果 Kubernetes 中某些 node 节点添加了污点调度，且不希望在其上部署 DataKit，可修改 datakit.yaml，调整其中的污点容忍度：

```yaml
      tolerations:
      - operator: Exists    <--- 修改这里的污点容忍度
```

具体绕过策略，参见[官方文档](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration)。

#### ConfigMap 设置 {#configmap-setting}

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

DataKit 支持的环境变量如下各表所示。

### 最常用环境变量

| 环境变量名称               | 默认值 | 必须   | 说明                                                                                 |
| ---------:                 | ---:   | ------ | ----                                                                                 |
| ENV_DATAWAY                | 无     | 是     | 配置 DataWay 地址，如 `https://openway.guance.com?token=xxx`                         |
| ENV_DEFAULT_ENABLED_INPUTS | 无     | 否     | 默认开启[采集器列表](datakit-input-conf.md#default-enabled-inputs)，以英文逗号分割，如 `cpu,mem,disk` |
| ENV_ENABLE_INPUTS          | 无     | 否     | ==已弃用==，同 ENV_DEFAULT_ENABLED_INPUTS                                            |
| ENV_GLOBAL_TAGS            | 无     | 否     | 全局 tag，多个 tag 之间以英文逗号分割，如 `tag1=val,tag2=val2`                       |

### 日志配置相关环境变量

| 环境变量名称  | 默认值                     | 必须   | 说明                                                             |
| ---------:    | ---:                       | ------ | ----                                                             |
| ENV_GIN_LOG   | */var/log/datakit/gin.log* | 否     | 如果改成 `stdout`，DataKit 自身 gin 日志将不写文件，而是终端输出 |
| ENV_LOG       | */var/log/datakit/log*     | 否     | 如果改成 `stdout`，DatakIt 自身日志将不写文件，而是终端输出      |
| ENV_LOG_LEVEL | info                       | 否     | 设置 DataKit 自身日志等级，可选 `info/debug`                     |

###  DataKit pprof 相关

| 环境变量名称  | 默认值                     | 必须   | 说明                                            |
| ---------:    | ---:                       | ------ | ----                                            |
| ENV_ENABLE_PPROF   | false | 否     | 是否开启 `pprof` |
| ENV_PPROF_LISTEN       | 无     | 否     | `pprof`服务监听地址 |

### 选举相关环境变量

| 环境变量名称        | 默认值     | 必须   | 说明                                                                                                                                                |
| ---------:          | ---:       | ------ | ----                                                                                                                                                |
| ENV_ENABLE_ELECTION | 默认不开启 | 否     | 开启[选举](election.md)，默认不开启，如需开启，给该环境变量任意一个非空字符串值即可                                                                    |
| ENV_NAMESPACE       | 无         | 否     | DataKit 所在的命名空间，默认为空表示不区分命名空间，接收任意非空字符串，如 `dk-namespace-example`。如果开启了选举，可以通过此环境变量指定工作空间。 |

### HTTP/API 相关环境变量

| 环境变量名称             | 默认值            | 必须   | 说明                                                 |
| ---------:               | ---:              | ------ | ----                                                 |
| ENV_DISABLE_404PAGE      | 无                | 否     | 禁用 DataKit 404 页面（公网部署 DataKit RUM 时常用） |
| ENV_HTTP_LISTEN          | localhost:9529    | 否     | 可修改地址，使得外部可以调用 [DataKit 接口](apis)    |
| ENV_REQUEST_RATE_LIMIT   | 无(float)         | 否     | 限制 9529 [API 每秒请求数](datakit-conf#39e48d64)    |
| ENV_RUM_ORIGIN_IP_HEADER | `X-Forwarded-For` | 否     | RUM 专用                                             |

### Git 配置相关环境变量

| 环境变量名称     | 默认值 | 必须   | 说明                                                                                                   |
| ---------:       | ---:   | ------ | ----                                                                                                   |
| ENV_GIT_BRANCH   | 无     | 否     | 指定拉取的分支。<stong>为空则是默认</strong>，默认是远程指定的主分支，一般是 `master`。                |
| ENV_GIT_INTERVAL | 无     | 否     | 定时拉取的间隔。（如 `1m`）                                                                            |
| ENV_GIT_KEY_PATH | 无     | 否     | 本地 PrivateKey 的全路径。（如 `/Users/username/.ssh/id_rsa`）                                         |
| ENV_GIT_KEY_PW   | 无     | 否     | 本地 PrivateKey 的使用密码。（如 `passwd`）                                                            |
| ENV_GIT_URL      | 无     | 否     | 管理配置文件的远程 git repo 地址。（如 `http://username:password@github.com/username/repository.git`） |

### Sinker 配置相关环境变量

| 环境变量名称 | 默认值 | 必须   | 说明                              |
| ---------:   | ---:   | ------ | ----                              |
| ENV_SINK_M   | 无     | 否     | 安装时指定 Metric 的 sink。       |
| ENV_SINK_N   | 无     | 否     | 安装时指定 Network 的 sink。      |
| ENV_SINK_K   | 无     | 否     | 安装时指定 KeyEvent 的 sink。     |
| ENV_SINK_O   | 无     | 否     | 安装时指定 Object 的 sink。       |
| ENV_SINK_CO  | 无     | 否     | 安装时指定 CustomObject 的 sink。 |
| ENV_SINK_L   | 无     | 否     | 安装时指定 Logging 的 sink。      |
| ENV_SINK_T   | 无     | 否     | 安装时指定 Tracing 的 sink。      |
| ENV_SINK_R   | 无     | 否     | 安装时指定 RUM 的 sink。          |
| ENV_SINK_S   | 无     | 否     | 安装时指定 Security 的 sink。     |

### 其它杂项

| 环境变量名称                 | 默认值         | 必须 | 说明                                                       |
| -----------------:           | -------------: | ---- | ---------------------------------------------------------- |
| ENV_CLOUD_PROVIDER           | 无             | 否   | 支持安装阶段填写云厂商(`aliyun/aws/tencent/hwcloud/azure`) |
| ENV_DCA_LISTEN               | localhost:9531 | 否   | 可修改改地址，使得 [DCA](dca) 客户端能管理该 DataKit       |
| ENV_DCA_WHITE_LIST           | 无             | 否   | 配置 DCA 白名单，以英文逗号分隔                            |
| ENV_HOSTNAME                 | 无             | 否   | 默认为本地主机名，可安装时指定，如， `dk-your-hostname`    |
| ENV_IPDB                     | 无（string）   | 否   | 指定 IP 信息库类型，目前只支持 `iploc`                     |
| ENV_ULIMIT                   | 无             | 否   | 指定 Datakit 最大的可打开文件数                            |
| ENV_DATAWAY_TIMEOUT          | 30s            | 否   | 设置 DataKit 请求 DataWay 的超时时间                       |
| ENV_DATAWAY_ENABLE_HTTPTRACE | false          | 否   | 在 debug 日志中输出 dataway HTTP 请求的网络日志            |
| ENV_DATAWAY_HTTP_PROXY       | 无             | 否   | 设置 DataWay HTTP 代理                                     |

### 特殊环境变量

#### ENV_K8S_NODE_NAME

当 k8s node 名称跟其对应的主机名不同时，可将 k8s 的 node 名称顶替默认采集到的主机名，在 *datakit.yaml* 中增加环境变量：

> [1.2.19](changelog#9bec76a9) 版本的 datakit.yaml 中默认就带了这个配置，如果是从老版本的 yaml 直接升级而来，需要对 *datakit.yaml* 做如下手动改动。

```yaml
- env:
	- name: ENV_K8S_NODE_NAME
		valueFrom:
			fieldRef:
				apiVersion: v1
				fieldPath: spec.nodeName
```

### 各个采集器专用环境变量

部分采集器支持外部注入环境变量，以调整采集器自身的默认配置。具体参见各个具体的采集器文档。

## 延伸阅读

- [DataKit 选举](election)
- [DataKit 的几种配置方式](k8s-config-how-to)
- [DataKit DaemonSet 配置管理最佳实践](datakit-daemonset-bp)
