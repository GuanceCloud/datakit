{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：Linux

# DaemonSet 安装 DataKit 

本文档介绍如何在 在 K8s 中通过 DaemonSet 方式安装 DataKit

## 安装步骤 

### 下载专用 yaml 配置

```shell
wget -O datakit-default.yaml http://path/to/yaml
```

### 修改配置

#### kubernetes 配置

修改 `datakit-default.yaml` 中的 dataway 配置

```yaml
        - name: ENV_DATAWAY
          value: <dataway_url> # 此处填上 dataway 真实地址
```

修改 Kubernetes API 地址：

通过如下命令，获取 k8s API 地址：

```shell
$kubectl config view -o jsonpath='{"Cluster name\tServer\n"}{range .clusters[*]}{.name}{"\t"}{.cluster.server}{"\n"}{end}'
```

将地址填到 yaml 如下配置中：

```yaml
      [[inputs.kubernetes]]
          url = "<https://your-k8s-api-server>"

```

#### container 配置

默认情况下，container 采集器没有开启指标采集，如需开启指标采集，修改 `datakit-default.yaml` 中如下配置：

```yaml
      [inputs.container]
        endpoint = "unix:///var/run/docker.sock"

        enable_metric = true # 将此处设置成 true
        enable_object = true
```

### 安装 yaml

```shell
kubectl apply -f datakit-default.yaml
```

### 查看运行状态：

安装完后，会创建一个 datakit-monitor 的 DaemonSet 部署：

```shell
kubectl get pod -n datakit-monitor
```

### DataKit 中其它环境变量设置

在 DaemonSet 模式中，DataKit 支持多个环境变量配置，如下表所示：


| 环境变量名称                 | 默认值           | 是否必须 | 说明                                                                                         |
| ---------                    | ---              | ------   | ----                                                                                         |
| `ENV_GLOBAL_TAGS`            | 无               | 否       | 全局 tag，多个 tag 之间以英文逗号分割，如 `tag1=val,tag2=val2`                               |
| `ENV_LOG_LEVEL`              | `info`           | 否       | 可选值 `info/debug`                                                                          |
| `ENV_DATAWAY`                | 无               | 否       | 可配置多个 dataway，以英文逗号分割，如 `https://dataway?token=xxx,https://dataway?token=yyy` |
| `ENV_HTTP_LISTEN`            | `localhost:9529` | 否       | 可修改改地址，使得外部可以调用 [DataKit 接口](apis)                                          |
| `ENV_RUM_ORIGIN_IP_HEADER`   | `X-Forward-For`  | 否       | RUM 专用                                                                                     |
| `ENV_DEFAULT_ENABLED_INPUTS` | 无               | 否       | 默认开启采集器列表，以英文逗号分割，如 `cpu,mem,disk`                                        |
| `ENV_ENABLE_ELECTION`        | `false`          | 否       | 开启[选举](election)，默认不开启                                                             |
