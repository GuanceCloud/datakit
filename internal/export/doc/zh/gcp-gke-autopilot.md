# GCP GKE Autopilot 集成
---

GKE Autopilot 不允许 DataKit 以传统 DaemonSet 方式挂载宿主机 `hostPath`、容器运行时 socket、宿主机日志目录或使用特权容器。因此，DataKit 在 GKE Autopilot 中使用独立发布的 `datakit-gke-autopilot` Helm chart，默认部署为单副本 Deployment，并通过 GCP Cloud API 采集容器相关数据。

普通 `datakit` chart 面向常规 Kubernetes 环境，默认安装为 DaemonSet；`datakit-gke-autopilot` chart 面向 GKE Autopilot，默认安装为 Deployment。两个 chart 使用同一套模板，但 `datakit-gke-autopilot` 带有专用默认 values，会关闭宿主机访问并打开 Cloud API 采集能力。

## 采集能力 {#capabilities}

<!-- markdownlint-disable MD046 -->
???+ note "版本要求"

    Cloud API 采集容器指标、对象和日志的功能需要 DataKit 2.3.0 及以上版本。早期版本不支持 GKE Autopilot 模式。
<!-- markdownlint-enable MD046 -->

`datakit-gke-autopilot` 默认启用 `dk,container`，并开启 election。单个 leader DataKit 实例会采集集群级数据，避免多个副本重复调用同一批 Cloud API。

- 容器指标：通过 Cloud Monitoring 采集并写入 `docker_containers`。
- 容器 stdout/stderr 日志：通过 Cloud Logging 采集。
- Kubernetes 资源指标和对象：通过 Kubernetes API 采集 Pod、Deployment、Service、Node 等资源。
- GCP 认证：使用 Workload Identity，不需要 Service Account key 文件。

该模式不支持宿主机指标、容器运行时 socket、本地容器文件日志和 eBPF。Cloud Monitoring 指标通常会有数分钟延迟。Cloud Logging 使用重叠时间窗口和 `timestamp + insertId` 去重，正常轮询不会重复；Pod 被替换或 leader 切换后，可能重新读取重叠窗口内日志，因此语义为至少一次。

## 前置条件 {#prerequisites}

- 已有 GKE Autopilot 集群，并且本地 `kubectl` 能访问该集群。
- 已获取 DataWay 地址，例如 `https://openway.<<<custom_key.brand_main_domain>>>?token=<YOUR-TOKEN>`。
- 已启用 GCP Cloud Monitoring 和 Cloud Logging API。
- 准备一个 GCP Service Account，并授予 `roles/monitoring.viewer` 和 `roles/logging.viewer`。
- 为 Kubernetes ServiceAccount 配置 Workload Identity 绑定。

下面示例使用 `my-project`、`datakit` namespace、`my-datakit` release 名称和 `datakit-cloud-monitor` GCP Service Account。请按实际环境替换。

```shell
PROJECT_ID="my-project"
NAMESPACE="datakit"
RELEASE_NAME="my-datakit"
GSA_NAME="datakit-cloud-monitor"
KSA_NAME="datakit"
```

创建 GCP Service Account 并授予权限：

```shell
gcloud iam service-accounts create "${GSA_NAME}" \
  --project "${PROJECT_ID}"

gcloud projects add-iam-policy-binding "${PROJECT_ID}" \
  --member "serviceAccount:${GSA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com" \
  --role roles/monitoring.viewer

gcloud projects add-iam-policy-binding "${PROJECT_ID}" \
  --member "serviceAccount:${GSA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com" \
  --role roles/logging.viewer
```

授权 Helm chart 创建的 Kubernetes ServiceAccount 使用该 GCP Service Account：

```shell
gcloud iam service-accounts add-iam-policy-binding \
  "${GSA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com" \
  --project "${PROJECT_ID}" \
  --role roles/iam.workloadIdentityUser \
  --member "serviceAccount:${PROJECT_ID}.svc.id.goog[${NAMESPACE}/${KSA_NAME}]"
```

<!-- markdownlint-disable MD046 -->
???+ note

    `datakit-gke-autopilot` chart 默认设置 `fullnameOverride=datakit`，因此示例命令会创建名为 `datakit` 的 Kubernetes ServiceAccount。如果覆盖了 `fullnameOverride`，需要按实际 ServiceAccount 名称调整 Workload Identity 绑定。
<!-- markdownlint-enable MD046 -->

## Helm 安装 {#helm-install}

添加 Helm 仓库：

```shell
helm repo add datakit-gke https://pubrepo.truewatch.com/chartrepo/truewatch
```

安装 DataKit：

```shell
helm install my-datakit datakit-gke/datakit-gke-autopilot \
         --namespace datakit \
         --create-namespace \
         --set datakit.dataway_url="https://openway.<<<custom_key.brand_main_domain>>>?token=<YOUR-TOKEN>" \
         --set serviceAccountAnnotations."iam\\.gke\\.io/gcp-service-account"="datakit-cloud-monitor@my-project.iam.gserviceaccount.com"
```

查看状态：

```shell
helm -n datakit list
kubectl -n datakit get pod -l app.kubernetes.io/instance=my-datakit
kubectl -n datakit logs deploy/datakit
```

升级前建议先备份当前 values：

```shell
helm -n datakit get values my-datakit -o yaml > values-current.yaml
```

升级：

```shell
helm upgrade my-datakit datakit-gke/datakit-gke-autopilot \
         --namespace datakit \
         -f values-current.yaml
```

<!-- markdownlint-disable MD046 -->
???+ note "从旧 GKE Autopilot chart 升级"

    旧 `helm-gke-autopilot` 分支发布的 `datakit-gke-autopilot` chart 会生成类似 `my-datakit-datakit-gke-autopilot` 的 DaemonSet 和 ServiceAccount。当前 chart 默认生成名为 `datakit` 的 Deployment 和 ServiceAccount。升级前需要将 Workload Identity 绑定调整到 `datakit/datakit`，并确认目标 namespace 中没有其它同名 `datakit` 资源。
<!-- markdownlint-enable MD046 -->

## YAML 安装 {#yaml-install}

如果不使用 Helm，可以下载 GKE Autopilot 专用的 Deployment YAML：

```shell
curl -o datakit-gke-autopilot.yaml \
  https://static.<<<custom_key.brand_main_domain>>>/datakit/datakit-gke-autopilot.yaml
```

使用前需要修改 `ENV_DATAWAY`，并在 ServiceAccount 上补充 Workload Identity GCP Service Account annotation：

```yaml
metadata:
  annotations:
    iam.gke.io/gcp-service-account: "datakit-cloud-monitor@my-project.iam.gserviceaccount.com"
```

应用 YAML：

```shell
kubectl apply -f datakit-gke-autopilot.yaml
```

## 关键配置 {#configuration}

默认配置如下：

- `workload.kind=Deployment`
- `workload.replicas=1`
- `gkeAutopilot.enabled=true`
- `datakit.default_enabled_inputs=dk,container`
- `datakit.enabled_election=true`
- `ENV_INPUT_CONTAINER_GCP_CLOUD_API_ENABLED=true`
- `ENV_INPUT_CONTAINER_ENABLE_K8S_NODE_LOCAL=false`
- CPU/内存 request 为 `500m/500Mi`
- 默认只挂载 `/usr/local/datakit/cache`

默认副本数为 `1`。如果需要高可用，可以调大 `workload.replicas`，但需要保持 election 开启，否则多个副本会重复采集 Cloud API 数据。

GCP project ID、GKE 集群名称和集群位置默认从 GKE metadata server 自动发现。跨项目采集或自动发现不符合预期时，可按需覆盖：

```yaml
extraEnvs:
  - name: ENV_INPUT_CONTAINER_GCP_PROJECT_ID
    value: "my-project"
  - name: ENV_INPUT_CONTAINER_GCP_CLUSTER_NAME
    value: "my-cluster"
  - name: ENV_INPUT_CONTAINER_GCP_CLUSTER_LOCATION
    value: "asia-southeast1"
```

`ENV_INPUT_CONTAINER_ENABLE_GCP_CLOUD_MONITORING` 和 `ENV_INPUT_CONTAINER_ENABLE_GCP_CLOUD_LOGGING` 默认均为 `true`，只有在需要关闭其中一项时才需要显式配置。

## 容器 host 标签 {#container-host-tag}

传统 DaemonSet 模式下，每个 DataKit 只采集本节点容器，容器指标、对象和日志的 `host` 可以由 DataKit 在 Feed 阶段统一追加。

GKE Autopilot Cloud API 模式下，单个 leader DataKit 会通过 Cloud API 采集集群级容器数据，因此不能使用 DataKit Pod 所在节点作为所有容器的 `host`。该模式会从 Kubernetes Pod 信息中提取容器所在的原始 GKE Node 名称，并写入容器指标、对象和日志的 `host` 标签。

该模式不支持 `ENV_K8S_CLUSTER_NODE_NAME` 对容器数据 `host` 的重命名。如果需要区分多个集群，建议使用 `cluster_name_k8s`、`gcp_project_id`、`gcp_location` 或自定义全局标签。

## 非 root 运行 {#run-as-non-root}

`datakit-gke-autopilot` 默认保持 root 用户运行，以兼容 DataKit 镜像默认行为。当前 Cloud API 模式不需要宿主机权限，如需以非 root 用户运行，可开启：

```shell
helm upgrade my-datakit datakit-gke/datakit-gke-autopilot \
         --namespace datakit \
         --reuse-values \
         --set gkeAutopilot.runAsNonRoot=true
```

开启后 Pod 会设置：

```yaml
securityContext:
  runAsNonRoot: true
  runAsUser: 10001
  runAsGroup: 10001
  fsGroup: 10001
```

该模式默认不需要 initContainer。chart 只挂载 `/usr/local/datakit/cache`，用于 WAL、Cloud Logging 状态和本地缓存，`fsGroup=10001` 可以保证该目录可写。如果额外挂载 `conf.d`、`data`、`pipeline`、`python.d` 等目录，需要确认这些目录对 UID/GID `10001` 可写；只有在额外 volume 不支持 `fsGroup` 或需要预置文件权限时，才需要自行增加 initContainer。

## 与普通 Helm 安装的区别 {#difference}

| chart | 默认工作负载 | 适用场景 | 宿主机访问 | 容器指标和日志来源 |
| --- | --- | --- | --- | --- |
| `datakit` | DaemonSet | 常规 Kubernetes | 默认使用 `hostPath`、容器运行时 socket 和宿主机日志目录 | 容器运行时和宿主机文件 |
| `datakit-gke-autopilot` | Deployment | GKE Autopilot | 默认不使用 `hostPath`、特权容器和宿主机网络 | Cloud Monitoring 和 Cloud Logging |

如果目标环境允许挂载宿主机目录和容器运行时 socket，优先使用普通 `datakit` chart。GKE Autopilot 或类似 Serverless Kubernetes 环境不允许这些宿主机能力时，使用 `datakit-gke-autopilot`。

## 故障排查 {#troubleshooting}

- Pod 被 GKE Autopilot 拒绝：检查是否额外启用了 `hostPath`、特权容器、`hostNetwork`、`hostPID` 或 `hostIPC`。
- 没有容器指标或日志：检查 GCP Service Account 是否有 `roles/monitoring.viewer` 和 `roles/logging.viewer`，Kubernetes ServiceAccount 是否正确设置 `iam.gke.io/gcp-service-account` annotation，并确认 Workload Identity 绑定的 namespace/name 与实际 ServiceAccount 一致。
- 没有 Kubernetes 对象或资源指标：检查 DataKit Pod 是否能访问 Kubernetes API，以及 `dk,container` 是否仍在 `datakit.default_enabled_inputs` 中。
- 多副本数据重复：检查 `datakit.enabled_election=true`，并确认所有副本使用同一个 namespace 内的 election 配置。
- 非 root 模式写入失败：检查额外挂载目录是否对 UID/GID `10001` 可写，必要时补充 volume 权限处理。
