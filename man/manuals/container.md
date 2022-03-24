{{.CSS}}

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
- 操作系统支持：{{.AvailableArchs}}

# {{.InputName}}

采集 container 和 Kubernetes 的指标数据、对象数据和容器日志，上报到观测云。

## 前置条件

- 目前 container 会默认连接 Docker 服务，需安装 Docker v17.04 及以上版本。
- 采集 Kubernetes 数据需要 DataKit 以 [DaemonSet 方式部署](datakit-daemonset)。
- 采集 Kubernetes Pod 指标数据，[需要 Kubernetes 安装 Metrics-Server 组件](https://github.com/kubernetes-sigs/metrics-server#installation)。

## 配置

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}}
```

> 对象数据采集间隔是 5 分钟，指标数据采集间隔是 20 秒，暂不支持配置

### 根据容器 image 配置指标和日志采集

配置文件中的 `container_include_metric / container_exclude_metric` 是针对指标数据，`container_include_log / container_exclude_log` 是针对日志数据。

- `container_include` 和 `container_exclude` 必须以 `image` 开头，格式为 `"image:<glob规则>"`，表示 glob 规则是针对容器 image 生效
- [Glob 规则](https://en.wikipedia.org/wiki/Glob_(programming))是一种轻量级的正则表达式，支持 `*` `?` 等基本匹配单元

例如，配置如下：

```
  ## 当容器的 image 能够匹配 `hello*` 时，会采集此容器的指标
  container_include_metric = ["image:hello*"]

  ## 忽略所有容器
  container_exclude_metric = ["image:*"]
```

> ==[Daemonset 方式部署](datakit-daemonset-deploy)时，可通过 [Configmap 方式挂载单独的 conf](k8s-config-how-to#ebf019c2) 来配置这些镜像的开关==

假设有 3 个容器，image 分别是：

```
容器A：hello/hello-http:latest
容器B：world/world-http:latest
容器C：registry.jiagouyun.com/datakit/datakit:1.2.0
```

使用以上 `include / exclude` 配置，将会只采集 `容器A` 指标数据，因为它的 image 能够匹配 `hello*`。另外 2 个容器不会采集指标，因为它们的 image 匹配 `*`。

补充，如何查看容器 image。

- docker 模式（容器由 docker 启动和管理）：

```
docker inspect --format "{{`{{.Config.Image}}`}}" $CONTAINER_ID
```

- Kubernetes 模式（容器由 Kubernetes 创建，有自己的所属 Pod）：

```
echo `kubectl get pod -o=jsonpath="{.items[0].spec.containers[0].image}"`
```

### 通过 Annotation/Label 调整容器日志采集

可以通过配置容器的 Labels，或容器所属 Pod 的 Annotations，为容器指定日志配置。

以 Kubernetes 为例，创建 Pod 添加 Annotations 如下：

- Key 为固定的 `datakit/logs`
- Value 是一个 JSON 字符串，支持 `source` `service` 和 `pipeline` 等配置项

```json
[
  {
    "disable"        : false,
    "source"         : "testing-source",
    "service"        : "testing-service",
    "pipeline"       : "test.p",
    "only_images"    : ["image:<your_image_regexp>"], # 用法和上文的 `根据 image 过滤容器` 完全相同，`image:` 后面填写正则表达式
    "multiline_match": "^\d{4}-\d{2}"
  }
]
```

如果是在终端命令行添加 Annotations，注意添加转义字符（以下示例两边是单引号，所以无需对双引号做转义）：

```
## foo 是 Pod name
kubectl annotate pods foo datakit/logs='[{\"disable\":false,\"source\":\"testing-source\",\"service\":\"testing-service\",\"pipeline\":\"test.p\",\"only_images\":[\"image:<your_image_regexp>\"],\"multiline_match\":\"^\\d{4}-\\d{2}\"}]'
```

注意：

- 如果该 JSON 配置的 `disable` 字段为 `true`，则不采集此 Pod 的所有容器日志。
- `only_images` 针对 Pod 内部多容器情景，如果填写了任何 image 通配，则只采集能匹配这些 image 的容器的日志，类似白名单功能；如果字段为空，即认为采集该 Pod 中所有容器的日志
- `multiline_match` 配置要做转义，例如 `"multiline_match":"^\\d{4}"` 表示行首是4个数字，在正则表达式规则中`\d` 是数字，前面的 `\` 是用来转义。
- 容器添加 Labels 的方法[文档](https://docs.docker.com/engine/reference/commandline/run/#set-metadata-on-container--l---label---label-file)
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
              "disable": false,
              "source": "testing-source",
              "service": "testing-service",
              "pipeline": "test.p",
              "multiline_match": "^\d{4}-\d{2}"
            }
          ]
```

### 通过 Sidecar 形式采集 Pod 内部日志

参见 [logfwd](logfwd)

### 环境变量配置

支持以环境变量的方式修改配置参数：

> 只有 DataKit 以 K8s DaemonSet 方式运行时生效，==主机部署时，以下环境变量不生效==。

| 环境变量名                                             | 对应的配置参数项                    | 参数示例                                                     |
| :----------------------------------------------------- | ----------------------------------- | ------------------------------------------------------------ |
| `ENV_INPUT_CONTIANER_EXCLUDE_PAUSE_CONTAINER`          | `exclude_pause_container`           | `true`/`false`                                               |
| `ENV_INPUT_CONTAINER_LOGGING_REMOVE_ANSI_ESCAPE_CODES` | `logging_remove_ansi_escape_codes ` | `true`/`false`                                               |
| `ENV_INPUT_CONTAINER_TAGS`                             | `tags`                              | `tag1=value1,tag2=value2` 如果配置文件中有同名 tag，会覆盖它 |

### 支持 Kubernetes 自定义 Export

详见[Kubernetes-prom](kubernetes-prom)

### 支持 containerd（试用版）

containerd 默认开启，通过连接 `/var/run/containerd/containerd.sock` 实现对 containerd 容器的对象采集（默认忽略 `pause` 容器），支持全部 namespace，详细对象字段见后文。

containerd 容器日志，推荐使用 [logfwd](logfwd) 进行采集。

## 指标集

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

```toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

### 指标

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "metric"}}

#### `{{$m.Name}}`

{{$m.Desc}}

- 标签

{{$m.TagsMarkdownTable}}

- 字段列表

{{$m.FieldsMarkdownTable}}
{{end}}

{{ end }}

### 对象

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "object"}}

#### `{{$m.Name}}`

{{$m.Desc}}

- 标签

{{$m.TagsMarkdownTable}}

- 字段列表

{{$m.FieldsMarkdownTable}}
{{end}}

{{ end }}

### 日志

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "logging"}}

#### `{{$m.Name}}`

{{$m.Desc}}

- 标签

{{$m.TagsMarkdownTable}}

- 字段列表

{{$m.FieldsMarkdownTable}}
{{end}}

{{ end }}

## FAQ

### 容器日志的特殊字节码过滤

容器日志可能会包含一些不可读的字节码（比如终端输出的颜色等），可以

- 将 `logging_remove_ansi_escape_codes` 设置为 `true` 
- DataKit DaemonSet 部署时，将 `ENV_INPUT_CONTAINER_LOGGING_REMOVE_ANSI_ESCAPE_CODES` 置为 `true`

此配置会影响日志的处理性能，基准测试结果如下：

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

每一条文本的处理耗时将==额外增加== `1616 ns` 不等。如果日志中不带有颜色等修饰，不要开启该功能。

## 延伸阅读

- [eBPF 采集器：支持容器环境下的流量采集](ebpf)
- [Pipeline：文本数据处理](pipeline)
- [正确使用正则表达式来配置](datakit-input-conf#9da8bc26) 
- [Kubernetes 下 DataKit 的几种配置方式](k8s-config-how-to)
- [DataKit 日志采集综述](datakit-logging)
