{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# {{.InputName}}

Kubernetes 集群指标采集，收集以下数据：

- 各种资源对象指标
- 接入 kubernetes event 数据，通过 [kube-eventer](https://github.com/AliyunContainerService/kube-eventer) 配置
- 接入 kubernetes 的 Prometheus exporter数据源 (todo)

## 前置条件

- 创建监控的 serviceAccount 账号（该账户只读权限）

```yaml
# create namespace
apiVersion: v1
kind: Namespace
metadata:
  name: datakit-monitor
  labels:
    name: datakit
---
# create ServiceAccount
apiVersion: v1
kind: ServiceAccount
metadata:
  name: datakit-monitor
  namespace: datakit-monitor
  labels:
    name: datakit
---
# create ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: datakit-monitor
rules:
- apiGroups: [""]
  resources: ["nodes", "services", "endpoints", "pods", "daemonsets", "deployments", "statefulsets", "persistentvolumes", "persistentvolumeclaims"]
  verbs: ["get", "list", "watch"]
---
# ClusterRoleBinding
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: datakit-monitor
subjects:
- kind: ServiceAccount
  name: datakit-monitor
  namespace: datakit-monitor
roleRef:
  kind: ClusterRole
  name: datakit-monitor
  apiGroup: rbac.authorization.k8s.io
```

- 获取认证 token 和证书


## 配置

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}}
```

配置好后，重启 DataKit 即可。

## 指标集

以下所有指标集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[[inputs.{{.InputName}}.tags]]` 另择 host 来命名。

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

-  标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}