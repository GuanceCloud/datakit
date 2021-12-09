{{.CSS}}

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
- 操作系统支持：Linux

# Kubernetes Pod 日志采集

## 介绍

本文档介绍如何采集 Kubernetes Pod 日志数据。

需要在 Kubernetes deployment 上添加特定的 template annotations，来采集由其创建的 Pod stdout 日志。Annotation 要求如下：

- Key 为固定的 `datakit/pod.logging`
- Value 如下：

```toml
## your logging source, if it's empty, use 'default'
source = ""

## add service tag, if it's empty, use $source.
service = ""

## grok pipeline script path
pipeline = ""

## optional status:
##   "emerg","alert","critical","error","warning","info","debug","OK"
ignore_status = []

## removes ANSI escape codes from text strings
remove_ansi_escape_codes = false

[tags]
# some_tag = "some_value"
# more_tag = "some_other_value"
```

## 操作过程

- 登录到 Kubernetes 所在主机

- 打开 `deployment.yaml`，添加 template annotations 示例如下：

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pod-logging-deployment
  labels:
    app: pod-logging-testing
spec:
  template:
    metadata:
      labels:
        app: pod-logging-testing
      annotations:
        datakit/pod.logging: |
          ## your logging source, if it's empty, use 'default'
          source = "pod-logging-testing-source"

          ## add service tag, if it's empty, use $source.
          service = ""

          ## grok pipeline script path
          pipeline = ""

          ## optional status:
          ##   "emerg","alert","critical","error","warning","info","debug","OK"
          ignore_status = []

          ## removes ANSI escape codes from text strings
          remove_ansi_escape_codes = false

          [tags]
          # some_tag = "some_value"
          # more_tag = "some_other_value"
```

> 注意， `annotations` 一定添加在 `template` 字段下，这样 deployment.yaml 创建的 Pod 才会携带 `datakit/pod.logging`。

- 使用新的 yaml 创建资源

```shell
kubectl apply -f deployment.yaml
```

至此，Annotation 已经添加完成。DataKit 稍后会读取到 Pod 的 Annotation，并采集 `url` 上暴露出来的指标。
