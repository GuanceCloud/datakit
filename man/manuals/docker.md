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

如果需要远程采集 Docker 容器信息，需要 Docker 开启相应的监听端口。

以 ubuntu 为例，需要在 `/etc/docker` 径路下打开或创建 `daemon.json` 文件，添加内容如下：

```json
{
    "hosts":["tcp://0.0.0.0:2375","unix:///var/run/docker.sock"]
}
```

重启服务后，Docker 便可以监听 `2375` 端口。详情见[官方配置文档](https://docs.docker.com/config/daemon/#configure-the-docker-daemon)。

### 采集器配置

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}} 
```

- 通过对 `collect_metric` 等三个配置项的开关，选择是否要开启此类数据的采集
- datkait 在连接 kubernetes 时，可能会因为 kubernetes 配置问题报错。以下是这两种报错的解决办法
    - `/run/secrets/kubernetes.io/serviceaccount/token: no such file or directory`。执行如下两个命令：
        - `mkdir -p /run/secrets/kubernetes.io/serviceaccount`
        - `touch /run/secrets/kubernetes.io/serviceaccount/token`
    - `error making HTTP request to http://<k8s-host>/stats/summary: dial tcp <k8s-hosst>:10255: connect: connect refused`，按如下方式调整 k8s 配置：
        - 编辑所有节点的 `/var/lib/kubelet/config.yaml` 文件，加入`readOnlyPort` 这个参数：`readOnlyPort: 10255`
        - 重启kubelet 服务：`systemctl restart kubelet.service`

- 当 `include_exited` 为 `true` 会采集非运行状态的容器

### Docker容器日志采集说明

- `log_option` 为数组配置，可以有多个。`container_name_match` 为正则表达式（[文档链接](https://golang.org/pkg/regexp/syntax/#hdr-Syntax)），如果容器名能匹配该正则，会使用 `source` 和 `service` 以及 `pipeline` 配置参数
- `pipeline` 为空值或该文件不存在时，将不使用 pipeline 功能

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

指标集为配置文件 `inputs.docker.log_option` 字段 `source`，例如配置如下：

```toml
[[inputs.docker.log_option]]
    container_name_match = "nginx-version-*"
    source = "nginx"
    service = "nginx"
    pipeline = "nginx.p"
```

如果日志来源的容器，容器名能够匹配 `container_name_match` 正则，该容器日志的指标集为 `source` 字段。

如果没有匹配到或 `source` 为空，指标集为容器名。

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
