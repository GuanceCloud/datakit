# Kubernetes Kubelet
---

## 视图预览

![1653442677(1).png](../imgs/kubernetes-kubelet-1.png)

## 版本支持

操作系统：Linux

Kubernetes 版本：1.18+

## 前置条件

- Kubernetes 集群 <[安装 Datakit](kube-metric-server.md)>

## 安装配置

说明：示例 Kubernetes 版本为 1.22.6，DataKit 版本为 1.2.20，各个不同版本指标可能存在差异。

### 部署实施

#### 指标采集 (必选)

1、 ConfigMap 增加 kubelet.conf 配置

在部署 DataKit 使用的 datakit.yaml 文件中，ConfigMap 资源中增加 kubelet.conf。

```
apiVersion: v1
kind: ConfigMap
metadata:
  name: datakit-conf
  namespace: datakit
data:    
    #### kubelet ##下面是新增部分
    kubelet.conf: |-    
        [[inputs.prom]]
          ## Exporter地址或者文件路径（Exporter地址要加上网络协议http或者https）
          ## 文件路径各个操作系统下不同
          ## Windows example: C:\\Users
          ## UNIX-like example: /usr/local/
          urls = ["https://172.16.0.229:10250/metrics"]

          ## 采集器别名
          source = "prom-kubelet"

          ## 指标类型过滤, 可选值为 counter, gauge, histogram, summary
          # 默认只采集 counter 和 gauge 类型的指标
          # 如果为空，则不进行过滤
          metric_types = ["counter", "gauge"]

          ## 指标名称过滤
          # 支持正则，可以配置多个，即满足其中之一即可
          # 如果为空，则不进行过滤
          # metric_name_filter = [""]

          ## 指标集名称前缀
          # 配置此项，可以给指标集名称添加前缀
          measurement_prefix = ""

          ## 指标集名称
          # 默认会将指标名称以下划线"_"进行切割，切割后的第一个字段作为指标集名称，剩下字段作为当前指标名称
          # 如果配置measurement_name, 则不进行指标名称的切割
          # 最终的指标集名称会添加上measurement_prefix前缀
          measurement_name = "prom_kubelet"

          ## 采集间隔 "ns", "us" (or "µs"), "ms", "s", "m", "h"
          interval = "60s"

          ## 过滤tags, 可配置多个tag
          # 匹配的tag将被忽略
          tags_ignore = ["build_date","compiler","container_state","container_type","git_commit","git_tree_state","git_version","go_version","long_running","major","method","minor","name","node","path","platform","plugin_name","pod","server_type","status","uid","version"] 
          metric_name_filter = ["kubelet_node_name","kubelet_running_pods","kubelet_running_containers","volume_manager_total_volumes","kubelet_runtime_operations_total","kubelet_runtime_operations_errors_total","storage_operation_errors_total","rest_client_requests_total","process_resident_memory_bytes","process_cpu_seconds_total","go_goroutines"]
          
          ## TLS 配置
          tls_open = true
          #tls_ca = ""
          #tls_cert = ""
          #tls_key = ""

          ## 自定义指标集名称
          # 可以将包含前缀prefix的指标归为一类指标集
          # 自定义指标集名称配置优先measurement_name配置项
          #[[inputs.prom.measurements]]
          #  prefix = "etcd_"
          #  name = "etcd"

          ## 自定义认证方式，目前仅支持 Bearer Token
          [inputs.prom.auth]
           type = "bearer_token"
          # token = "xxxxxxxx"
           token_file = "/var/run/secrets/kubernetes.io/serviceaccount/token"

          ## 自定义Tags
          [inputs.prom.tags]
            instance = "172.16.0.229:10250"      
```

参数说明：

- urls：metrics 地址

- source：采集器别名
- metric_types：指标类型过滤
- metric_name_filter：指标名称过滤
- measurement_prefix：指标集名称前缀
- measurement_name：指标集名称
- interval：采集间隔
- tags_ignore:  忽略的 tag
- metric_name_filter:  保留的指标名
- tls_open：是否忽略安全验证 (如果是 https，请设置为 true，并设置相应证书)，此处为 true
- tls_ca：ca 证书路径
- type：自定义认证方式，kubelet 使用 bearer_token 认证
-  token_file：认证文件路径
- [inputs.prom.tags]：请参考插件标签

2、 挂载 kubelet.conf

在 datakit.yaml 文件的 volumeMounts 下面增加下面内容。

```
- mountPath: /usr/local/datakit/conf.d/prom/kubelet.conf
  name: datakit-conf
  subPath: kubelet.conf  
```

3、 重启 DataKit

```
kubectl delete -f datakit.yaml
kubectl apply -f datakit.yaml
```

指标预览

![1653445079(1).png](../imgs/kubernetes-kubelet-2.png)

#### 插件标签 (必选）

参数说明

- 该配置为自定义标签，可以填写任意 key-value 值

- 以下示例配置完成后，kubelet 指标都会带有 app = oa 的标签，可以进行快速查询
- 采集 kubelet 指标，必填的 key 是 instance，值是 kubeletmetrics 的 ip + 端口

```

## 自定义Tags
[inputs.prom.tags]
  instance = "172.16.0.229:10250" 
```

如果增加了自定义 tag，重启 Datakit 

```
kubectl delete -f datakit.yaml
kubectl apply -f datakit.yaml
```

## 场景视图

<场景 - 新建仪表板 - 内置模板库 - Kubernetes Kubelet 监控视图>

## 指标集

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

-  标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}

## 常见问题排查

- [无数据上报排查](why-no-data.md)

## 进一步阅读

- [DataFlux Tag 应用最佳实践](/best-practices/guance-skill/tag.md)

