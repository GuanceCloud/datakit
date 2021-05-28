{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# {{.InputName}}

采集 Docker 服务数据和容器数据，分别以指标、对象和日志的方式上报到 DataFlux 中。

## 前置条件

- Docker 版本 >= 1.24

## 配置

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}} 
```

## 指标集

以下所有指标集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

{{$m.Desc}}

-  标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }} 

### Docker 容器日志采集

对于 Docker 容器日志采集，有着更细致的配置，主要解决了区分日志 `source` 和使用 pipeline 的问题。

日志采集配置项为 `[[inputs.docker.logfilter]]`，该项是数组配置，意即可以有多个 logfilter 来处理采集到的容器日志，比如某个容器中既有 MySQL 日志，也有 Redis 日志，那么此时可能需要两个 logfilter 来分别处理它们。

- `filter_message` 为匹配日志文本的正则表达式，该参数类型是字符串数组，只要任意一个正则匹配成功即可。未匹配的日志内容将被丢弃。[正则表达式参见这里](https://golang.org/pkg/regexp/syntax/#hdr-Syntax)
>Tips：为保证此处正则表达式的正确书写，请务必将正则表达式用 `'''这里是一个正则表达式'''` 这种形式来配置（即两边用三个单引号来包围正则文本），否则可能导致正则转义问题。
- `source` 指定数据来源，如果为空值，则默认使用容器名
- `service` 指定该条日志的服务名，如果为空值，则使用 `source` 字段值
- `pipeline` 只需写文件名即可，不需要写全路径，使用方式见 [Pipeline 文档](pipeline)。当此值为空值或该文件不存在时，将不使用 pipeline 功能

#### 日志切割注意事项

使用 pipeline 功能时，如果切割成功，则：

- 取其中的 `time` 字段作为此条数据的产生时间。如果没有 `time` 字段或解析此字段失败，默认使用当前 DataKit 所在机器的系统时间
- 所切割出来日志结果中，必定有一个 `status` 字段。如果切割出来的原始数据中没有该字段，则默认将 `status` 置为 `info`

当前有效的 `status` 字段值如下（三列均不区分大小写）：

| 简写 | 可能的全称                  | 对应值     |
| :--- | ---                         | -------    |
| `f`  | `emerg`                     | `emerg`    |
| `a`  | `alert`                     | `alert`    |
| `c`  | `critical`                  | `critical` |
| `e`  | `error`                     | `error`    |
| `w`  | `warning`                   | `warning`  |
| `n`  | `notice`                    | `notice`   |
| `i`  | `info`                      | `info`     |
| `d`  | `debug`, `trace`, `verbose` | `debug`    |
| `o`  | `s`, `OK`                   | `OK`       |

### 远程采集 Docker 配置

如果需要远程采集 Docker 容器信息，则需要 Docker 开启相应的监听端口。以 Ubuntu 为例，需要在该远程主机的 `/etc/docker` 径路下打开或创建 `daemon.json` 文件，添加内容如下后重启 Docker 服务：

```json
{
  "hosts":[
    "tcp://0.0.0.0:2375",
    "unix:///var/run/docker.sock"
  ]
}
```

重启该 Docker 服务后，便可以监听 `2375` 端口。详情见[官方配置文档](https://docs.docker.com/config/daemon/#configure-the-docker-daemon)。

远程采集 Docker 需要修改 `inputs.docker.endpoint` 配置，示例如下：

```
endpoint = "tcp://remote-docker-ip:port"
```

此外，建议在 `[inputs.docker.tag]` 中添加目标 Docker 的信息，用以辨识远程 Docker 服务：

```toml
[inputs.docker.tags]
	host = "<your-real-docker-server-hostname>"
```

否则采集到的容器数据中会[带上 DataKit 所在主机的 hostname](datakit-how-to#cdcbfcc9)：

### 关联 Kubernetes 服务

如果主机上装有 Kubernetes，则 DataKit 会尝试连接 Kubernetes 服务，进行容器和 Kubernetes 关联，可以得到该容器在 Kubernetes 服务中的 pod 相关信息。

例如该容器由 Kubernetes 创建，可以得到 `pod_name` 和 `pod_namespace` 两个对象数据。

注意，DataKit 只会关联本机的 Kubernetes 服务，不会关键远程 Kubernetes 服务。即尝试连接 Kubernetes 服务的地址是 `127.0.0.1:10255`。

在连接 Kubernetes 时，可能会因为 Kubernetes 认证问题报错：

- 报错一：`/run/secrets/kubernetes.io/serviceaccount/token: no such file or directory`

执行如下两个命令准备对应文件：

```shell
# mkdir -p /run/secrets/kubernetes.io/serviceaccount
# touch /run/secrets/kubernetes.io/serviceaccount/token
```

- 报错二： `error making HTTP request to http://<k8s-host>/stats/summary: dial tcp <k8s-hosst>:10255: connect: connect refused`

按如下步骤调整 k8s 配置：

  1. 编辑所有节点的 `/var/lib/kubelet/config.yaml` 文件，加入`readOnlyPort` 这个参数：`readOnlyPort: 10255`
  1. 重启 kubelet 服务：`systemctl restart kubelet.service`
