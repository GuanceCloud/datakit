{{.CSS}}

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# {{.InputName}}

采集 container 和 Kubernetes 的指标数据、对象数据和容器日志，上报到观测云。

## 前置条件

- 目前 container 会默认连接 Docker 服务，需安装 Docker v17.04 及以上版本。
- 采集 Kubernetes 数据需要 DataKit 以 Kubernetes daemonset 方式运行。
- 采集 Kubernetes Pod 指标数据，需要 Kubernetes 安装 Metrics-Server 组件，[链接](https://github.com/kubernetes-sigs/metrics-server#installation)。

## 配置

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}} 
```

- `container_include` 和 `container_exclude` 必须以 `image` 开头，格式为 `"image:<glob规则>"`，表示 glob 规则是针对 image 生效。
- glob 规则是一种轻量级的正则表达式，支持 `*` `?` 等基本匹配单元，[glob wiki](https://en.wikipedia.org/wiki/Glob_(programming))

*对象数据采集间隔是5分钟，指标数据采集间隔是15秒，暂不支持配置*

### 环境变量配置

支持以环境变量的方式修改配置参数（只在 DataKit 以 K8s daemonset 方式运行时生效，主机部署的 DataKit 不支持此功能）：

| 环境变量名                                             | 对应的配置参数项                    | 参数示例                                                     |
| :---                                                   | ---                                 | ---                                                          |
| `ENV_INPUT_CONTIANER_EXCLUDE_PAUSE_CONTAINER`          | `exclude_pause_container`           | `true`/`false`                                               |
| `ENV_INPUT_CONTAINER_LOGGING_REMOVE_ANSI_ESCAPE_CODES` | `logging_remove_ansi_escape_codes ` | `true`/`false`                                               |
| `ENV_INPUT_CONTAINER_TAGS`                             | `tags`                              | `tag1=value1,tag2=value2` 如果配置文件中有同名 tag，会覆盖它 |

### 指定容器日志配置

可以通过配置容器的 Labels，或容器所属 Pod 的 Annotations，为容器指定日志配置。

以 Kubernetes 为例，创建 Pod 添加 Annotations 如下：

- Key 为固定的 `datakit/logs`
- Value 是一个 JSON 字符串，支持 `source` `service` 和 `pipeline` 三个字段值

```json
[
  {
    "disable"  : false,
    "source"   : "testing-source",
    "service"  : "testing-service",
    "pipeline" : "test.p"
  }
]
```

拼接成一行并加上转义字符，最终结果是 `[{\"disable\": false, \"source\": \"testing-source\", \"service\": \"testing-service\", \"pipeline\": \"test.p\"}]`

注意：

- 如果该 JSON 配置的 `disable` 字段为 `true`，则不采集此 Pod 的所有容器日志。
- 容器不支持动态添加 Labels，容器的 Labels 跟其镜像绑定在一起，在生成镜像时已经固定。给容器添加 Labels 需要重新 build 一份镜像再添加 Labels，[官方示例文档](https://docs.docker.com/engine/reference/builder/#label)
- Kubernetes 一般不会直接创建 Pod 也不添加 Annotations，可以在创建 Deployment 时以 `template` 模式添加 Annotations，由此 Deployment 生成的所有 Pod 都会携带 Annotations，例如：

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: testing-log-deployment
  labels:
    app: testing-log
spec:
  template:
    metadata:
      labels:
        app: testing-log
      annotations:
        datakit/logs: |
          [
            {
              "source": "testing-source",
              "service": "testing-service",
              "pipeline": "test.p"
            }
          ]
```

### 容器日志的特殊字节码过滤

容器日志可能会包含一些不可读的字节码（比如终端输出的颜色等），可以将 `logging_remove_ansi_escape_codes` 设置为 `true` 对其删除过滤。

此配置可能会影响日志的处理性能，基准测试结果如下：

```
goos: linux
goarch: amd64
pkg: gitlab.jiagouyun.com/cloudcare-tools/test
cpu: Intel(R) Core(TM) i7-4770HQ CPU @ 2.20GHz
BenchmarkRemoveAnsiCodes
BenchmarkRemoveAnsiCodes-8        636033              1616 ns/op
PASS
ok      gitlab.jiagouyun.com/cloudcare-tools/test       1.056s
```

每一条文本的处理耗时增加 `1616 ns` 不等。如果不开启此功能将无额外损耗。

### 支持 Kubernetes 自定义 Export

详见[Kubernetes-prom](kubernetes-prom)

## 指标集

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

## 指标

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "metric"}}

### `{{$m.Name}}`
{{$m.Desc}}

-  标签

{{$m.TagsMarkdownTable}}

- 字段列表

{{$m.FieldsMarkdownTable}}
{{end}}

{{ end }}

## 对象

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "object"}}

### `{{$m.Name}}`
{{$m.Desc}}

-  标签

{{$m.TagsMarkdownTable}}

- 字段列表

{{$m.FieldsMarkdownTable}}
{{end}}

{{ end }}

## 日志

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "logging"}}

### `{{$m.Name}}`
{{$m.Desc}}

-  标签

{{$m.TagsMarkdownTable}}

- 字段列表

{{$m.FieldsMarkdownTable}}
{{end}}

{{ end }}
