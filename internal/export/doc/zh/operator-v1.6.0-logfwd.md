# Operator < 1.6.0 logfwd 用法

此配置方式在 DataKit-Operator v1.6.0 及之前版本使用。v1.7.0 版本采用了新的 CRD 配置方式，v1.6.0 版本中引入的 CRD + Annotation 混合方案已被废弃。

1. 在目标 Kubernetes 集群，[下载和安装 DataKit-Operator](datakit-operator.md#install)
1. 在 deployment 添加指定 Annotation，表示需要挂载 logfwd sidecar。注意 Annotation 要添加在 template 中
    - key 统一是 `admission.datakit/logfwd.instances`
    - value 是一个 JSON 字符串，是具体的 logfwd 配置，示例如下：

```json
[
    {
        "datakit_addr": "datakit-service.datakit.svc:9533",
        "loggings": [
            {
                "logfiles":      ["<your-logfile-path>"],
                "ignore":        [],
                "storage_index": "<your-storage-index>",
                "source":        "<your-source>",
                "service":       "<your-service>",
                "pipeline":      "<your-pipeline.p>",
                "character_encoding": "",
                "multiline_match": "<your-match>",
                "tags": {}
            },
            {
                "logfiles": ["<your-logfile-path-2>"],
                "source": "<your-source-2>"
            }
        ]
    }
]
```

参数说明，可参考 [logfwd 配置](../integrations/logfwd.md#config)：

- `datakit_addr` 是 DataKit logfwdserver 地址
- `loggings` 为主要配置，是一个数组，可参考 [DataKit logging 采集器](../integrations/logging.md)
    - `logfiles` 日志文件列表，可以指定绝对路径，支持使用 glob 规则进行批量指定，推荐使用绝对路径
    - `ignore` 文件路径过滤，使用 glob 规则，符合任意一条过滤条件将不会对该文件进行采集
    - `storage_index` 指定日志存储索引
    - `source` 数据来源，如果为空，则默认使用 'default'
    - `service` 新增标记 tag，如果为空，则默认使用 $source
    - `pipeline` Pipeline 脚本路径，如果为空将使用 $source.p，如果 $source.p 不存在将不使用 Pipeline（此脚本文件存在于 DataKit 端）
    - `character_encoding` 选择编码，如果编码有误会导致数据无法查看，默认为空即可。支持 `utf-8/utf-16le/utf-16le/gbk/gb18030`
    - `multiline_match` 多行匹配，详见 [DataKit 日志多行配置](../integrations/logging.md#multiline)，注意因为是 JSON 格式所以不支持 3 个单引号的"不转义写法"，正则 `^\d{4}` 需要添加转义写成 `^\\d{4}`
    - `tags` 添加额外 `tag`，书写格式是 JSON map，例如 `{ "key1":"value1", "key2":"value2" }`

<!-- markdownlint-disable MD046 -->
???+ note

    注入 logfwd 时，DataKit Operator 默认复用相同路径的 volume，避免因为存在同样路径的 volume 而注入报错。

    路径末尾有斜线和无斜线的意义不同，例如 `/var/log` 和 `/var/log/` 是不同路径，不能复用。
<!-- markdownlint-enable -->

## 用例 {#datakit-operator-inject-logfwd-example}

下面是一个 Deployment 示例，使用 shell 持续向文件写入数据，且配置该文件的采集：

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
    name: logging-deployment
    labels:
    app: logging
spec:
    replicas: 1
    selector:
    matchLabels:
        app: logging
    template:
    metadata:
        labels:
        app: logging
        annotations:
        admission.datakit/logfwd.instances: '[{"datakit_addr":"datakit-service.datakit.svc:9533","loggings":[{"logfiles":["/var/log/log-test/*.log"],"source":"deployment-logging","tags":{"key01":"value01"}}]}]'
    spec:
        containers:
        - name: log-container
        image: busybox
        args: [/bin/sh, -c, 'mkdir -p /var/log/log-test; i=0; while true; do printf "$(date "+%F %H:%M:%S") [%-8d] Bash For Loop Examples.\\n" $i >> /var/log/log-test/1.log; i=$((i+1)); sleep 1; done']
```

使用 yaml 文件创建资源：

```shell
$ kubectl apply -f logging.yaml
...
```

验证如下：

```shell
$ kubectl get pod

NAME                                   READY   STATUS    RESTARTS      AGE
logging-deployment-5d48bf9995-vt6bb    1/1     Running   0             4s

$ kubectl get pod logging-deployment-5d48bf9995-vt6bb -o=jsonpath={.spec.containers\[\*\].name}
log-container datakit-logfwd
```

最终可以在<<<custom_key.brand_name>>>日志平台查看日志是否采集。
