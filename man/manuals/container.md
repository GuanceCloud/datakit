{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# {{.InputName}}

采集 container 指标数据、对象数据和容器日志，以及当前主机上的 kubelet Pod 指标和对象，上报到 DataFlux 中。

## 前置条件

- 目前 container 会默认连接 Docker 服务，需安装 Docker(v1.24+)

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

### 标签定制和删除

- `drop_tags`：对于某些 Tag，收集的意义不大，但会导致时间线暴涨。目前将 `contaienr_id` 这个 tag 移除掉了。
- `pod_name_rewrite`：在 kubernetes 中，Pod 名称具有相似性，但一般会默认追加一串随机串，如 `kube-proxy-678b9fd78b`，一旦将 "kube-proxy" 配置进去之后，会剪掉尾部的 `-678b9fd78b`

### 容器日志采集

对于容器日志采集，有着更细致的配置，主要解决了区分日志 `source` 和使用 pipeline 的问题。

日志采集配置项为 `[[inputs.container.logfilter]]`，该项是数组配置，意即可以有多个 logfilter 来处理采集到的容器日志，比如某个容器中既有 MySQL 日志，也有 Redis 日志，那么此时可能需要两个 logfilter 来分别处理它们。

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

### kubelet 相关采集

在配置文件中打开 `inputs.container.kubelet` 项，填写对应的 `kubelet_url`（默认为 `127.0.0.1:10255`）可以采集 kubelet Pod 相关数据。

kubelet 该端口默认关闭，开启方式请查看[官方文档](https://kubernetes.io/zh/docs/reference/command-line-tools-reference/kubelet/)搜索 `--read-only-port`。

如果 `kubelet_url` 配置的主机端口未监听，或尝试连接失败，则不再采集 kubelet 相关数据。DataKit 只有在启动或重启时才会对该端口进行连接验证，一旦验证失败，直到下次重启前都不会再次连接 kubelet。如果 kubelet 经过修复已经开启对应端口，需重启 DataKit 才能采集。

在连接 kubelet 时，可能会因为 kubelet 认证问题报错：

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
