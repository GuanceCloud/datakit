# GCP GKE Autopilot 集成
---

选择一种采集模式，再选择 Helm 或 YAML 安装，避免重复采集。

## 选择采集模式 {#choose-mode}

| 对比项 | Partner | Cloud API |
| --- | --- | --- |
| 工作负载 | DaemonSet，每个节点运行 DataKit | Deployment，默认单副本 |
| Helm Chart | `datakit`，加载包内 Partner values | `datakit-gke-autopilot` |
| 容器指标来源 | 节点 containerd | Cloud Monitoring |
| 容器日志来源 | 节点上的 Kubernetes Pod stdout/stderr 日志 | Cloud Logging |
| 接入前提 | 同步 Google 正式 V2 allowlist | GCP API、IAM 权限和 Workload Identity |
| YAML 部署 | 本地从 Chart 渲染完整清单 | 下载专用 Deployment YAML |

两种模式均通过 Kubernetes API 采集资源数据，默认启用 `dk,container` 和 election。Cloud API 模式要求 DataKit 2.3.0 及以上版本。

## 公共准备 {#prerequisites}

- 已有 GKE Autopilot 集群，本地 `kubectl` 的 context 指向该集群，账号有部署工作负载和 RBAC 资源的权限。
- DataKit 能访问包含有效工作空间 token 的 DataWay 地址，节点能拉取对应镜像。
- 示例使用 `datakit` namespace 和资源名，确认没有其它安装占用同名工作负载、Service、ClusterRole 或 ClusterRoleBinding。
- Helm 安装和 Partner YAML 渲染需要本地 Helm 3；Cloud API 的 IAM 配置需要 `gcloud`。

在新工作目录准备以下 Helm 用户配置，替换 DataWay 地址和集群名。文件含 token，须妥善保管。**Cloud API YAML 安装跳过下面的 Helm 配置，直接进入 [Cloud API 部署](gcp-gke-autopilot.md#cloud-api)。**

```shell
umask 077
cat > datakit-user-values.yaml <<'VALUES'
datakit:
  dataway_url: "https://openway.<<<custom_key.brand_main_domain>>>?token=<YOUR-TOKEN>"
  cluster_name_k8s: "my-autopilot-cluster"
VALUES
```

添加仓库并查看可用版本，将 `CHART_VERSION` 替换为所选 Chart 版本。Partner 要求包内包含下文的两个配置文件。

```shell
helm repo add truewatch https://pubrepo.truewatch.com/chartrepo/truewatch
helm repo update truewatch
helm search repo truewatch/datakit --versions
CHART_VERSION="<本次正式发布的 Chart 版本>"
```

## Partner 部署 {#partner}

**Partner V2 仅支持 TrueWatch，最低 GKE 版本为 `1.35.6-gke.1258000`。** 使用 TrueWatch 的 `datakit` Chart 和 `pubrepo.truewatch.com/truewatch/datakit` 镜像；部署前确认两者均已发布。

保持 `image.tag` 为空，镜像自动跟随 Chart 的 `appVersion`。完成下载和 allowlist 同步后，选择 Helm 或 YAML 安装。

### 获取 Partner 配置 {#partner-download}

下载普通 `datakit` Chart，检查 Partner 文件：

```shell
helm pull truewatch/datakit --version "$CHART_VERSION" --untar
test -f datakit/values-gke-autopilot-partner.yaml
test -f datakit/gke-autopilot-partner-allowlist.yaml
```

### 同步 V2 allowlist {#allowlist}

账号需要管理 AllowlistSynchronizer 的权限。集群须允许 `gke://TrueWatch/datakit/truewatch-datakit-autopilot-v2.yaml`；GKE 默认允许 `gke://*`。如果管理员限制过来源，需保留已有路径并加入此路径，参见 [Google allowlist 安装说明](https://docs.cloud.google.com/kubernetes-engine/docs/how-to/run-autopilot-partner-workloads){:target="_blank"}。

每个集群安装一次同步器。以下命令用于新建同步器；已有管理同一路径的同步器时，跳过 `apply`，并将检查命令中的 `truewatch-datakit` 替换为已有同步器名称。**Partner 的 Helm 和 YAML 安装均须先完成同步。**

```shell
kubectl apply -f datakit/gke-autopilot-partner-allowlist.yaml
kubectl wait --for=condition=Ready allowlistsynchronizer/truewatch-datakit --timeout=10m
kubectl get allowlistsynchronizer truewatch-datakit -o yaml
kubectl get workloadallowlist truewatch-datakit-autopilot-v2
```

确认 `Ready=True`、V2 文件在 `status.managedAllowlistStatus` 中为 `Installed`，且对应 WorkloadAllowlist 存在，再安装 DataKit。

### Partner Helm 安装 {#partner-helm}

```shell
helm upgrade --install datakit ./datakit \
  --namespace datakit --create-namespace \
  -f datakit/values-gke-autopilot-partner.yaml \
  -f datakit-user-values.yaml --wait --timeout 15m
```

Chart 创建 `datakit-dataway` Secret，通过 `secretKeyRef` 将 DataWay 地址传给 DataKit。

### Partner YAML 安装 {#partner-yaml}

与 Helm 安装二选一。本方式使用 Helm 3 在本地渲染，之后由 `kubectl` 管理资源，不创建 Helm release。`--no-hooks` 排除 Chart 的 Helm 测试 Pod。

```shell
umask 077
helm template datakit ./datakit --namespace datakit --no-hooks \
  -f datakit/values-gke-autopilot-partner.yaml \
  -f datakit-user-values.yaml > datakit-partner.yaml
kubectl create namespace datakit --dry-run=client -o yaml | kubectl apply -f -
kubectl apply --namespace datakit -f datakit-partner.yaml
```

渲染清单包含完整工作负载和 Secret，须妥善保管。静态 `datakit.yaml` 用于普通节点，`datakit-gke-autopilot.yaml` 用于 Cloud API；Partner 使用此处渲染的清单。

### Partner 配置范围 {#partner-configuration}

Partner 预设将 `fullnameOverride` 固定为 `datakit`，仅修改 Helm release 名不会改变资源名。需要与集群中已有同名资源区分时，在用户 values 中设置其它 `fullnameOverride`（例如 `datakit-autopilot`），并使用独立 namespace；同步替换本文命令中的 namespace 和 DaemonSet 名。每个 namespace 内的兼容 Service 名仍为 `datakit-service`。

V2 仅只读挂载 `/var/run/containerd`、`/proc` 和 `/var/log/pods`，不支持 eBPF、任意宿主机文件采集或完整宿主机监控。缓存使用 `emptyDir`，Pod 重建会丢失本地缓存。

可在 `datakit-user-values.yaml` 中添加 `extraEnvs`，例如将运行日志输出到 stdout，方便使用 `kubectl logs` 排查：

```yaml
extraEnvs:
  - name: ENV_LOG
    value: "stdout"
```

变量名须匹配 `^ENV_[A-Z0-9_]+$`，可通过 `valueFrom` 引用同 namespace 中已有的 Secret/ConfigMap。已有 `extraEnvs` 时追加条目；DataWay、集群名等使用 `datakit.*` 设置。保持预设的镜像来源、allowlist、挂载和安全配置，不启用 `iploc`、`dkconfig`、Git SSH key 挂载或附加采集器 Chart。

## Cloud API 部署 {#cloud-api}

Cloud API 通过 Cloud Monitoring/Logging 采集，无宿主机访问；不支持本地文件日志、宿主机监控或 eBPF。指标可能延迟数分钟，Pod 或 leader 切换可能导致近期日志重复。

### 配置 GCP 权限 {#cloud-permissions}

使用有权启用 API、创建服务账号和配置 IAM 的账号执行。设置集群所属项目，启用 API 并通过 Workload Identity 绑定服务账号；已有服务账号时复用其名称并跳过创建命令。

```shell
PROJECT_ID="my-project"
GSA_NAME="datakit-cloud-monitor"
gcloud services enable monitoring.googleapis.com logging.googleapis.com \
  iam.googleapis.com iamcredentials.googleapis.com --project "$PROJECT_ID"
gcloud iam service-accounts create "$GSA_NAME" --project "$PROJECT_ID"
gcloud projects add-iam-policy-binding "$PROJECT_ID" \
  --member "serviceAccount:${GSA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com" \
  --role roles/monitoring.viewer
gcloud projects add-iam-policy-binding "$PROJECT_ID" \
  --member "serviceAccount:${GSA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com" \
  --role roles/logging.viewer
gcloud iam service-accounts add-iam-policy-binding \
  "${GSA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com" --project "$PROJECT_ID" \
  --role roles/iam.workloadIdentityUser \
  --member "serviceAccount:${PROJECT_ID}.svc.id.goog[datakit/datakit]"
```

示例绑定 `datakit` namespace 中的 `datakit` Kubernetes ServiceAccount；覆盖 namespace 或 `fullnameOverride` 时，须同步修改绑定。

### Cloud API Helm 安装 {#cloud-helm}

在 `datakit-user-values.yaml` 中追加实际 GCP Service Account 邮箱：

```yaml
serviceAccountAnnotations:
  iam.gke.io/gcp-service-account: "datakit-cloud-monitor@my-project.iam.gserviceaccount.com"
```

```shell
helm upgrade --install datakit truewatch/datakit-gke-autopilot \
  --version "$CHART_VERSION" --namespace datakit --create-namespace \
  -f datakit-user-values.yaml --wait --timeout 15m
```

### Cloud API YAML 安装 {#cloud-yaml}

与 Cloud API Helm 安装二选一。下载专用 Deployment YAML：

```shell
umask 077
curl -f -o datakit-gke-autopilot.yaml \
  https://static.<<<custom_key.brand_main_domain>>>/datakit-v2/datakit-gke-autopilot.yaml
```

下载后设置 `ENV_DATAWAY`（含 token）、`ENV_CLUSTER_NAME_K8S`，并将 **ServiceAccount** 的现有 annotation 替换为实际邮箱；此路径不读取 Helm 用户 values：

```yaml
metadata:
  annotations:
    iam.gke.io/gcp-service-account: "datakit-cloud-monitor@my-project.iam.gserviceaccount.com"
```

```shell
kubectl apply -f datakit-gke-autopilot.yaml
```

### Cloud API 配置 {#cloud-configuration}

GCP project、cluster 和 location 默认从 GKE metadata server 发现。需要覆盖时，在 Helm 用户 values 中加入以下完整列表。**`extraEnvs` 会替换 Chart 默认列表，必须保留前三项，否则 Cloud API 采集将被关闭。**

```yaml
extraEnvs:
  - name: ENV_NAMESPACE
    value: "datakit"
  - name: ENV_INPUT_CONTAINER_GCP_CLOUD_API_ENABLED
    value: "true"
  - name: ENV_INPUT_CONTAINER_ENABLE_K8S_NODE_LOCAL
    value: "false"
  - name: ENV_INPUT_CONTAINER_GCP_PROJECT_ID
    value: "my-project"
  - name: ENV_INPUT_CONTAINER_GCP_CLUSTER_NAME
    value: "my-cluster"
  - name: ENV_INPUT_CONTAINER_GCP_CLUSTER_LOCATION
    value: "asia-southeast1"
```

YAML 安装直接修改容器 env。默认单副本；扩容时保持 election 开启，并让所有副本使用同一套选举配置。可选非 root 运行：Helm 设置 `gkeAutopilot.runAsNonRoot=true`；YAML 按文件中的 `securityContext` 注释配置，确保挂载目录对 UID/GID `10001` 可写。

## 验证运行与数据 {#verify}

按所选模式检查 rollout：

```shell
# Partner
kubectl -n datakit rollout status daemonset/datakit --timeout=15m
# Cloud API
kubectl -n datakit rollout status deployment/datakit --timeout=15m
```

随后检查 Pod 和健康接口：

```shell
kubectl -n datakit get daemonset,deployment,pods -o wide
kubectl get --raw '/api/v1/namespaces/datakit/services/http:datakit-service:9529/proxy/v1/health'
```

确认工作负载 desired 数大于 0、ready 数等于 desired 数，实际 Pod 持续正常运行，健康接口返回 `live=true`。Partner 在没有节点的集群中可能显示 DaemonSet `0/0`，此时 Helm 安装成功不代表 DataKit 已运行；先部署业务工作负载，等待 Autopilot 创建节点，再检查 DataKit Pod 和健康接口。

在平台按 `cluster_name_k8s` 和部署时间检查容器指标、Kubernetes 对象及已启用采集的业务 Pod 日志，确认数据属于目标集群且时间持续更新。DataWay 地址须包含有效 token；仅提供域名无法完成数据上报验收。Pod Ready 只表示运行就绪。

Partner V2 禁止 `exec`、`attach` 和 `port-forward`；排查使用 Pod 状态、Events、`kubectl logs` 和 Service proxy。分享日志前移除 token 和完整 DataWay 地址。

## 升级、回退和卸载 {#lifecycle}

**Helm 安装**：保存当前 Chart 版本和用户配置，选择目标版本，重复对应模式的安装命令并重新验收。Partner 先在另一个目录下载目标 Chart，每次显式传入 Partner 和用户两个 values 文件，保持 `image.tag` 为空，使镜像跟随目标 `appVersion`。Cloud API 的用户配置须包含 Workload Identity annotation。

Partner Chart 管理的 DataWay Secret 更新会触发 DaemonSet 滚动更新。更新 `extraEnvs.valueFrom` 引用的外部 Secret/ConfigMap 后，执行 `kubectl -n datakit rollout restart daemonset/datakit` 重新加载环境变量。

按需执行下列命令，将 `<REVISION>` 替换为历史版本编号；回退后复核镜像和外部配置：

```shell
helm history datakit --namespace datakit
helm rollback datakit <REVISION> --namespace datakit --wait --timeout 15m
helm uninstall datakit --namespace datakit
```

**YAML 安装**：保留旧清单，使用新版本重新渲染 Partner 清单或下载 Cloud API 清单，补齐配置后 apply；回退时 apply 旧清单。新版移除的资源需单独核对并删除。卸载使用 `kubectl delete -n datakit -f <安装时使用的清单>`。Cloud API 清单包含 Namespace；如果 namespace 中还有其它资源，先从卸载清单中移除该对象。YAML 中的 Secret 或 DataWay 地址需按敏感信息保管。

仅在集群中已没有使用此 allowlist 的工作负载时，按需删除 Partner 同步器；保留其它部署共享的 namespace 和配置：

```shell
kubectl delete -f datakit/gke-autopilot-partner-allowlist.yaml
```

## 故障排查 {#troubleshooting}

- Partner 包内缺少配置：确认下载了包含本功能的正式 Chart，更新仓库索引后重新下载。
- allowlist 未同步：检查同步器的 `status.managedAllowlistStatus` 错误、完整 GKE 版本及集群允许的路径。以 `Ready=True` 和文件 `Installed` 为准。
- Partner 被拒绝：检查 DaemonSet Events、Pod 的 `cloud.google.com/matching-allowlist=truewatch-datakit-autopilot-v2`，以及是否注入 sidecar、增加挂载或替换镜像。参见 [Google 准入故障排查](https://docs.cloud.google.com/kubernetes-engine/docs/troubleshooting/autopilot-privileged-workloads){:target="_blank"}。
- Partner 报 `Image Mismatch`：将用户 values 的 `image.repository` 恢复为 `pubrepo.truewatch.com/truewatch/datakit`，保持 `image.tag` 为空；YAML 安装从同版本 Chart 重新渲染并 apply。报 `ErrImagePull` 或 `ImagePullBackOff` 时，检查对应 TrueWatch 镜像是否已发布，以及节点到仓库的网络和拉取权限。
- 日志出现 `dataway.emptyToken` 或 `token missing`：补齐 DataWay 地址中的有效工作空间 token，再按原安装方式更新。空 token 会导致上报被拒绝、数据丢弃和选举失败，即使 Pod Ready、健康检查正常也不能视为采集验收通过。
- Cloud API 被拒绝或无数据：检查是否引入宿主机权限，以及 GCP IAM 权限、ServiceAccount annotation 和 Workload Identity 的 namespace/name 是否一致。
- Pod 正常但数据缺失或重复：检查 DataWay 连通性、配置、Kubernetes API 权限、业务 Pod 日志采集设置及 election；非 root 写入失败时检查目录权限。
- `kubectl logs` 只有启动日志：在 Partner 用户配置中设置 `ENV_LOG=stdout`（见[配置示例](gcp-gke-autopilot.md#partner-configuration)），再按原方式升级；YAML 需重新渲染后 apply。
