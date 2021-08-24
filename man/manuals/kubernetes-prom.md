{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：Linux

# Kubernetes 集群中自定义 Exporter 指标采集

## 介绍

该方案可以在 Kubernetes 集群中通过配置，采集集群中的自定义的 Pod 的 Exporter 数据。

目前只支持 Promtheus 格式的数据。

通过在 Kubernetes Pod 添加指定的 Annotation，实现 Exporter 功能。Annotation 格式内容如下：

- Key 固定为 `datakit/prom.exporter`
- Value 为 JSON 格式，示例如下：

```
{
  "disable": false,
  "url": "http://$IP:9100/metrics",
  "source": "prom",
  "interval": "10s",
  "measurement_name": "",
  "measurement_prefix": "",
  "metric_name_filter": [],
  "metric_types": [
    "counter",
    "gauge"
  ],
  "tags_ignore": [],
  "tls_open": false,
  "tls_ca": "/tmp/ca.crt",
  "tls_cert": "/tmp/peer.crt",
  "tls_key": "/tmp/peer.key",
  "measurements": [
    {
      "name": "cpu",
      "prefix": "cpu_"
    },
    {
      "name": "mem",
      "prefix": "mem_"
    }
  ],
  "tags": {
    "namespace": "$NAMESPACE",
    "pod_name": "$PODNAME"
  }
}
```

字段说明：

- `url` 是必填项，建议使用 `$IP` 变量自动替换为 Pod IP
- `source` 表示数据来源，建议填写
- `interval` 表示采集间隔时长
- `tags` 为自定义 tags，默认添加 `namespace` 和 `pod_name` 两项
- 其余字段详情可对照 [Prom 指标采集](prom)

变量说明:

- `$IP`：通配 Pod 的内网 IP，形如 `172.16.0.3`，无需额外配置
- `$NAMESPACE`：Pod Namespace
- `$PODNAME`：Pod Name

## 操作过程

假设 Pod 名称为 `dummy-abc`

0. 登录到 Kubernetes 所在主机
1. 复制上述 JSON 配置示例，将其写入文件（以 `/tmp/annotation.json` 文件名为例），按需修改配置参数
2. 添加 Annotation
  ```shell
  conf=`echo $(cat /tmp/annotation.json)`;kubectl annotate --overwrite pods dummy-abc datakit/prom.exporter="$conf"
  ```
  终端打印 `pod/dummy-abc annotated` 表示添加成功，可以使用以下命令查看 Annotation 详情
  ```shell
  kubectl get pod dummy-abc -o jsonpath='{.metadata.annotations}'
  ``` 
3. 导出 Pod yaml
  ```shell
  kubectl get pod dummy-abc -o yaml >> dummy-abc.yaml
  ``` 
4. 使用新的 yaml 创建资源
  ```shell
  kubectl apply -f dummy-abc.yaml
  ```

至此，Annotation 已经添加完成。

DataKit Kubernetes 会自动忽略相同配置（变量替换之后的配置），避免重复采集。
