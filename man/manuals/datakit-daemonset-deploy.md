{{.CSS}}

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
- 操作系统支持：Linux

# DaemonSet 安装 DataKit 

本文档介绍如何在 K8s 中通过 DaemonSet 方式安装 DataKit。

## 安装步骤 

先下载[datakit.yaml](https://static.guance.com/datakit/datakit.yaml) ,在该配置中，需做如下配置：

- kubernetes：用来采集 Kubernetes 中心指标，需要填写 kubernetes 中心采集地址

其它主机相关的采集器都是[默认开启的](datakit-conf-how-to#764ffbc2)，无需额外配置。

### 修改配置

修改 `datakit.yaml` 中的 dataway 配置

```yaml
	- name: ENV_DATAWAY
		value: https://openway.guance.com?token=<your-token> # 此处填上 dataway 真实地址
```

### 安装 yaml

```shell
kubectl apply -f datakit.yaml
```

### 查看运行状态：

安装完后，会创建一个 datakit 的 DaemonSet 部署：

```shell
kubectl get pod -n datakit
```

### DataKit 中其它环境变量设置

在 DaemonSet 模式中，DataKit 支持多个环境变量配置，如下表所示：

| 环境变量名称               | 默认值                     | 是否必须 | 说明                                                                                                               |
| ---------                  | ---                        | ------   | ----                                                                                                               |
| ENV_DATAWAY                | 无                         | 是       | 可配置多个 dataway，以英文逗号分割，如 `https://openway.guance.com?token=xxx,https://openway.guance.com?token=yyy` |
| ENV_GLOBAL_TAGS            | 无                         | 否       | 全局 tag，多个 tag 之间以英文逗号分割，如 `tag1=val,tag2=val2`                                                     |
| ENV_LOG_LEVEL              | `info`                     | 否       | 可选值 `info/debug`                                                                                                |
| ENV_LOG                    | `/var/log/datakit/log`     | 否       | 如果改成 `stdout`，日志将不写文件，而是终端输出                                                                    |
| ENV_GIN_LOG                | `/var/log/datakit/gin.log` | 否       | 如果改成 `stdout`，日志将不写文件，而是终端输出                                                                    |
| ENV_HTTP_LISTEN            | `localhost:9529`           | 否       | 可修改改地址，使得外部可以调用 [DataKit 接口](apis)                                                                |
| ENV_DCA_LISTEN             | `localhost:9531`           | 否       | 可修改改地址，使得 [DCA](dca) 客户端能管理该 DataKit                                                               |
| ENV_DCA_WHITE_LIST         | 无                         | 否       | 配置 DCA 白名单，以英文逗号分隔                                                                                    |
| ENV_RUM_ORIGIN_IP_HEADER   | `X-Forwarded-For`          | 否       | RUM 专用                                                                                                           |
| ENV_DISABLE_404PAGE        | 无                         | 否       | 禁用 DataKit 404 页面（公网部署 DataKit RUM 时常用）                                                               |
| ENV_DEFAULT_ENABLED_INPUTS | 无                         | 否       | 默认开启[采集器列表](datakit-conf-how-to#764ffbc2)，以英文逗号分割，如 `cpu,mem,disk`。                                                            |
| ENV_ENABLE_ELECTION        | 默认不开启                 | 否       | 开启[选举](election)，默认不开启，如需开启，给该环境变量任意一个非空字符串值即可                                   |
| ENV_NAMESPACE              | 无                         | 否       | DataKit 所在的命名空间，默认为空表示不区分命名空间，接收任意非空字符串，如 `dk-namespace-example`。如果开启了选举，可以通过此环境变量指定工作空间。                  |
| ENV_HOSTNAME               | 无                         | 否       | 默认为本地主机名，可安装时指定，如， `dk-your-hostname`                                                            |
| ENV_GIT_URL                 | 无                         | 否       | 管理配置文件的远程 git repo 地址。（如 `http://username:password@github.com/username/repository.git`）  |
| ENV_GIT_KEY_PATH            | 无                         | 否       | 本地 PrivateKey 的全路径。（如 `/Users/username/.ssh/id_rsa`）                                        |
| ENV_GIT_KEY_PW              | 无                         | 否       | 本地 PrivateKey 的使用密码。（如 `passwd`）                                                           |
| ENV_GIT_BRANCH              | 无                         | 否       | 指定拉取的分支。<stong>为空则是默认</strong>，默认是远程指定的主分支，一般是 `master`。                      |
| ENV_GIT_INTERVAL            | 无                         | 否       | 定时拉取的间隔。（如 `1m`）                                                                           |
| ENV_CLOUD_PROVIDER          | 无                         | 否       | 支持安装阶段填写云厂商(`aliyun/aws/tencent/hwcloud/azure`)                                            |

> 注意：
>  `ENV_ENABLE_INPUTS` 已被弃用（但仍有效），建议使用 `ENV_DEFAULT_ENABLED_INPUTS`。如果俩个环境变量同时指定，则**只有后者生效**。
>  `ENV_LOG` 如果配置成 `stdout`，则不要将 `ENV_LOG_LEVEL` 设置成 `debug`，否则可能循环产生日志，产生大量日志数据。
