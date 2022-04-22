{{.CSS}}

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
- 操作系统支持：{{.AvailableArchs}}

# {{.InputName}}

采集 container 和 Kubernetes 的指标数据、对象数据和容器日志，上报到观测云。

## 前置条件

- 目前 container 会默认连接 Docker 服务，需安装 Docker v17.04 及以上版本。
- 采集 Kubernetes 数据需要 DataKit 以 [DaemonSet 方式部署](datakit-daemonset-deploy)。
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


各个字段说明：

| 字段名            | 必填 | 取值             | 默认值 | 说明                                                                                                                                                       |
| -----             | ---- | ----             | ----   | ----                                                                                                                                                       |
| `disable`         | N    | true/false       | false  | 是否禁用该 pod/容器的日志采集                                                                                                                              |
| `source`          | N    | 字符串           | 无     | 日志来源，参见[容器日志采集的 source 设置](container#6de978c3)                                                                                                               |
| `service`         | N    | 字符串           | 无     | 日志隶属的服务，默认值为日志来源（source）                                                                                                                 |
| `pipeline`        | N    | 字符串           | 无     | 适用该日志的 Pipeline 脚本，默认值为与日志来源匹配的脚本名（`<source>.p`）                                                                                 |
| `only_images`     | N    | 字符串数组       | 无     | 针对 Pod 内部多容器情景，如果填写了任何 image 通配，则只采集能匹配这些 image 的容器的日志，类似白名单功能；如果字段为空，即认为采集该 Pod 中所有容器的日志 |
| `multiline_match` | N    | 正则表达式字符串 | 无     | 用于多行日志匹配时的首行识别，例如 `"multiline_match":"^\\d{4}"` 表示行首是4个数字，在正则表达式规则中`\d` 是数字，前面的 `\` 是用来转义                   |

如果是在终端命令行添加 Annotations，注意添加转义字符（以下示例两边是单引号，所以无需对双引号做转义）：

```shell
kubectl annotate pods my-pod datakit/logs='[{\"disable\":false,\"source\":\"testing-source\",\"service\":\"testing-service\",\"pipeline\":\"test.p\",\"only_images\":[\"image:<your_image_regexp>\"],\"multiline_match\":\"^\\d{4}-\\d{2}\"}]'
```

> 关于 Docker 容器添加 Label 的方法，参见[这里](https://docs.docker.com/engine/reference/commandline/run/#set-metadata-on-container--l---label---label-file)。

在 Kubernetes 可以在创建 Deployment 时，以 `template` 模式添加 Pod Annotations，例如：

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
              "multiline_match": "^\d{4}-\d{2}",
							"only_images": ["image:.*nginx.*", "image:.*my_app.*"]
            }
          ]
```

### 通过 Sidecar 形式采集 Pod 内部日志

参见 [logfwd](logfwd)

### 环境变量配置

支持以环境变量的方式修改配置参数：

> 只有 DataKit 以 K8s DaemonSet 方式运行时生效，==主机部署时，以下环境变量不生效==。

| 环境变量名                                             | 对应的配置参数项                    | 参数示例                                                       |
| :----------------------------------------------------- | ----------------------------------- | ------------------------------------------------------------   |
| `ENV_INPUT_CONTAINER_DOCKER_ENDPOINT`                  | `docker_endpoint`                   | `unix:///var/run/docker.sock`                                  |
| `ENV_INPUT_CONTAINER_CONTAINERD_ADDRESS`               | `containerd_address`                | `/var/run/containerd/containerd.sock`                          |
| `ENV_INPUT_CONTIANER_EXCLUDE_PAUSE_CONTAINER`          | `exclude_pause_container`           | `true`/`false`                                                 |
| `ENV_INPUT_CONTAINER_LOGGING_REMOVE_ANSI_ESCAPE_CODES` | `logging_remove_ansi_escape_codes ` | `true`/`false`                                                 |
| `ENV_INPUT_CONTAINER_TAGS`                             | `tags`                              | `tag1=value1,tag2=value2` 如果配置文件中有同名 tag，会覆盖它   |
| `ENV_INPUT_CONTAINER_CONTAINER_INCLUDE_METRIC`         | `container_include_metric`          | `"image:*"` 以英文逗号隔开                                     |
| `ENV_INPUT_CONTAINER_CONTAINER_EXCLUDE_METRIC`         | `container_exclude_metric`          | `"image:*"` 以英文逗号隔开                                     |
| `ENV_INPUT_CONTAINER_CONTAINER_INCLUDE_LOG`            | `container_include_log`             | `"image:pubrepo.jiagouyun.com/datakit/logfwd*"` 以英文逗号隔开 |
| `ENV_INPUT_CONTAINER_CONTAINER_EXCLUDE_LOG`            | `container_exclude_log`             | `"image:pubrepo.jiagouyun.com/datakit/logfwd*"` 以英文逗号隔开 |
| `ENV_INPUT_CONTAINER_MAX_LOGGING_LENGTH`               | `max_logging_length`                | `32766`                                                        |
| `ENV_INPUT_CONTAINER_KUBERNETES_URL`                   | `kubernetes_url`                    | `https://kubernetes.default:443`                               |
| `ENV_INPUT_CONTAINER_BEARER_TOKEN`                     | `bearer_token`                      | `/run/secrets/kubernetes.io/serviceaccount/token`              |
| `ENV_INPUT_CONTAINER_BEARER_TOKEN_STRING`              | `bearer_token_string`               | `<your-token-string>`                                          |
| `ENV_K8S_CLUSTER_NAME`                                 | k8s `cluster_name` 字段的缺省值     | `kube`                                                         |

补充，k8s 数据的 `cluster_name` 字段可能会为空，为此提供注入环境变量的方式，取值优先级依次为：
1. k8s 集群返回的 ClusterName 值（不为空）
2. 环境变量 `ENV_K8S_CLUSTER_NAME` 指定的值
3. 默认值 `kubernetes`

### 支持 Kubernetes 自定义 Export

详见[Kubernetes-prom](kubernetes-prom)

### 支持 containerd

- 容器指标和对象：适配 docker container 指标集，详见下面文档
- 容器/Pod 日志：推荐使用 [logfwd](logfwd) 进行采集。
- Kubernetes 其它采集均不受影响

如果 containerd.sock 路径不是默认的 `/var/run/containerd/containerd.sock`，需要指定新的 `containerd.sock` 路径：

- 主机部署：修改 container.conf 的 `containerd_address` 配置项
- 以 Kubernetes daemonset 运行 DataKit：更改 datakit.yaml 的 volumes `containerd-socket`，将新路径 mount 到 DataKit daemonset 中，同时配置环境变量 `ENV_INPUT_CONTAINER_CONTAINERD_ADDRESS`，值为新路径。例如新的路径是 `/var/containerd/containerd.sock`，datakit.yaml 片段如下：

```
      # 添加 env
      - env:
        - name: ENV_INPUT_CONTAINER_CONTAINERD_ADDRESS
          value: /var/containerd/containerd.sock
```
```
      # 修改 mountPath
        - mountPath: /var/containerd/containerd.sock
          name: containerd-socket
          readOnly: true
```
```
      # 修改 volumes
      volumes:
      - hostPath:
          path: /var/containerd/containerd.sock
        name: containerd-socket
```



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

- 指标列表

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

- 指标列表

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

### 容器日志采集的 source 设置

在容器环境下，日志来源（`source`）设置是一个很重要的配置项，它直接影响在页面上的展示效果。但如果挨个给每个容器的日志配置一个 source 未免残暴。如果不手动配置容器日志来源，DataKit 有如下规则（优先级递减）用于自动推断容器日志的来源：

> 所谓不手动指定容器日志来源，就是指在 Pod Annotation 中不指定，在 container.conf 中也不指定（目前 container.conf 中无指定容器日志来源的配置项）

- 容器名：一般从容器的 `io.kubernetes.container.name` 这个 label 上取值。如果不是 Kubernetes 创建的容器（比如只是单纯的 Docker 环境，那么此 label 没有，故不以容器名作为日志来源）
- short-image-name: 镜像名，如 `nginx.org/nginx:1.21.0` 则取 `nginx`。在非 Kubernetes 容器环境下，一般首先就是取（精简后的）镜像名
- `unknown`: 如果镜像名无效（如 `sha256:b733d4a32c...`），则取该未知值

## 延伸阅读

- [eBPF 采集器：支持容器环境下的流量采集](ebpf)
- [Pipeline：文本数据处理](pipeline)
- [正确使用正则表达式来配置](datakit-input-conf#9da8bc26) 
- [Kubernetes 下 DataKit 的几种配置方式](k8s-config-how-to)
- [DataKit 日志采集综述](datakit-logging)
