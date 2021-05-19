{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# {{.InputName}}

采集 Docker 服务数据和容器数据，分别以指标、对象和日志的方式上报到 DataFlux 中。

## 前置条件

- 已安装 docker 1.24（[docker 官方链接](https://www.docker.com/get-started)）

## 配置

### Docker服务监听端口配置

- 如果不需要开启远程采集 Docker，忽略此段即可

- 如果需要远程采集 Docker 容器信息，则需要 Docker 开启相应的监听端口

以 ubuntu 为例，需要在 `/etc/docker` 径路下打开或创建 `daemon.json` 文件，添加内容如下：

```json
{
    "hosts":["tcp://0.0.0.0:2375","unix:///var/run/docker.sock"]
}
```

重启该 Docker 服务后，便可以监听 `2375` 端口。详情见[官方配置文档](https://docs.docker.com/config/daemon/#configure-the-docker-daemon)。

此外，建议在 `inputs.docker.tag` 配置（详情见下）中添加 `host` 标签字段，用以辨识远程 Docker 服务。

### 采集器配置

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}} 
```

- 通过对 `collect_metric` 等三个配置项的开关，选择是否要开启此类数据的采集
- DataKit 在连接 kubernetes 时，可能会因为 kubernetes 配置问题报错。以下是这两种报错的解决办法
    - `/run/secrets/kubernetes.io/serviceaccount/token: no such file or directory`。执行如下两个命令：
        - `mkdir -p /run/secrets/kubernetes.io/serviceaccount`
        - `touch /run/secrets/kubernetes.io/serviceaccount/token`
    - `error making HTTP request to http://<k8s-host>/stats/summary: dial tcp <k8s-hosst>:10255: connect: connect refused`，按如下方式调整 k8s 配置：
        - 编辑所有节点的 `/var/lib/kubelet/config.yaml` 文件，加入`readOnlyPort` 这个参数：`readOnlyPort: 10255`
        - 重启kubelet 服务：`systemctl restart kubelet.service`

- 当 `include_exited` 为 `true` 会采集非运行状态的容器

### Docker容器日志采集说明

对于 Docker 容器日志采集，有着更细致的配置，主要解决了区分日志 `source` 和使用 pipeline 的问题。

日志采集配置项为 `inputs.docker.logfilter`，该项是数组配置，可以有多个。

- `filter_message` 为匹配日志文本的正则表达式，该参数类型是字符串数组，只要任意一个正则匹配成功，则使用后续的 `source`、`service` 和 `pipeline` 参数。[正则表达式文档](https://golang.org/pkg/regexp/syntax/#hdr-Syntax)
- `source` 指定数据来源，如果为空值，则默认使用容器名
- `service` 指定该条日志的服务名，如果为空值，则使用 `source` 字段值
- `pipeline` 只需写文件名即可，不需要写全路径，使用方式见[文档](pipeline)。当此值为空值或该文件不存在时，将不使用 pipeline 功能

**使用 pipeline 功能时，取其中的 `time` 字段作为此条数据的产生时间。如果没有 `time` 字段或解析此字段失败，默认使用数据采集到的时间**

**数据必须含有 `status` 字段。如果使用 pipeline 功能时且得到有效的 `status`，否则默认使用“info”**

有效的 `status` 字段值（不区分大小写）：

| status 有效字段值                | 对应值     |
| :---                             | ---        |
| `f`, `emerg`                     | `emerg`    |
| `a`, `alert`                     | `alert`    |
| `c`, `critical`                  | `critical` |
| `e`, `error`                     | `error`    |
| `w`, `warning`                   | `warning`  |
| `n`, `notice`                    | `notice`   |
| `i`, `info`                      | `info`     |
| `d`, `debug`, `trace`, `verbose` | `debug`    |
| `o`, `s`, `OK`                   | `OK`       |

指标集为配置文件 `inputs.docker.logfilter` 字段 `source`，例如配置如下：

```toml
[[inputs.docker.logfilter]]
    filter_message = [ '''^\[GIN.*''']
    source = "gin"
    service = "gin"
    pipeline = "ginlog.p"
```

采集器会对所有容器日志进行正则匹配，如果有任意一条成功匹配到上述配置文件中的 `filter_message`，则会将此条日志的 `source` 和 `service` 设置为 `gin`，并使用 `ginlog.p` 的 pipeline。

注意，此功能需要对所有容器的日志文本进行正则匹配，假设 N 为全部 `filter_message` 正则的数量，则每一条日志文本最坏情况下需要匹配 N 次，会影响性能。

## 指标集

以下所有指标集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[[inputs.{{.InputName}}.tags]]` 另择 host 来命名。

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

{{$m.Desc}}

-  标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }} 
