{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：Linux

# DaemonSet 安装 DataKit 

本文档介绍如何在 在 K8s 中通过 DaemonSet 方式安装 DataKit

## 安装步骤 


- 下载专用 yaml 配置

```shell
wget -O datakit-default.yaml http://path/to/yaml
```

修改 `datakit-default.yaml`，修改其中的 `ENV_DATAWAY` 地址，将其设定成对应的地址即可。

### 其它环境变量设置

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

- 安装

```shell
kubectl apply -f datakit-default.yaml
```

- 查看运行状态：

安装完后，会创建一个 datakit-monitor 的 DaemonSet 部署：

```shell
kubectl get pod -n datakit-monitor
```
