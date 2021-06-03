{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# {{.InputName}}

Kubernetes 集群指标采集，主要用于收集各种资源指标

## 前置条件

注意：以下使用是 datakit 部署运行在 k8s 集群外

### 创建监控的 ServiceAccount 账号（该账户拥有只读权限）

- 创建 `account.yaml` 编排文件, 文件内容如下:

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
- apiGroups:
  - ""
  resources:
  - nodes
  - namespaces
  - pods
  - services
  - endpoints
  - persistentvolumes
  - persistentvolumeclaims
  verbs:
  - get
  - list
- apiGroups:
  - apps
  resources:
  - deployments
  - daemonsets
  - statefulsets
  - replicasets
  verbs:
  - get
  - list
- apiGroups:
  - "extensions"
  resources:
  - ingresses
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - batch
  resources:
  - jobs
  verbs:
  - get
  - list
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

- 由集群管理员执行以下命令创建监控只读权限的账户

```sh
## 执行编排yaml
$kubectl apply -f account.yaml

## 确认创建成功
$kubectl get sa -n datakit-monitor
NAME              SECRETS   AGE
datakit-monitor   1         3d13h
default           1         3d13h
```

- 获取服务地址

```sh
$kubectl config view -o jsonpath='{"Cluster name\tServer\n"}{range .clusters[*]}{.name}{"\t"}{.cluster.server}{"\n"}{end}'
```

注意：以上得到的集群服务地址，将用于 kubernetes 采集器的 `url` 配置项中

- 获取认证 token 和证书

```sh
## 获取token
$kubectl get secrets -n datakit-monitor -o jsonpath="{.items[?(@.metadata.annotations['kubernetes\.io/service-account\.name']=='datakit-monitor')].data.token}"| base64 --decode > token

## 获取CA证书
$kubectl get secrets -n datakit-monitor -o jsonpath="{.items[?(@.metadata.annotations['kubernetes\.io/service-account\.name']=='datakit-monitor')].data.ca\\.crt}" | base64 --decode > ca_crt.pem

## 确认结果
$ls -l 
-rw-r--r--  1 liushaobo  staff   1066  6  1 15:48 ca.crt
-rw-r--r--  1 liushaobo  staff    953  6  1 15:48 token
```

注意：以上得到的结果文件将分别用于 kubernetes 采集器 `bearer_token` 和 `tls_ca`配置项中

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

