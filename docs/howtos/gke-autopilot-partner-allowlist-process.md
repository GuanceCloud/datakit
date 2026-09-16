# GKE Autopilot Partner 操作指南

本指南汇总 TrueWatch DataKit 的 Google 审核、发布、复测和扩展流程。客户安装命令统一维护在 [中文文档](../../internal/export/doc/zh/gcp-gke-autopilot.md#partner) / [英文文档](../../internal/export/doc/en/gcp-gke-autopilot.md#partner)。记录更新于 2026-09-16。

## 1. 方案与入口

复用普通 `datakit` Chart，加载包内 `values-gke-autopilot-partner.yaml`，部署访问节点 containerd、procfs 和 Pod 日志的 DaemonSet。Google V2 是权限修订号；DataKit 镜像版本跟随 CI 写入的 `Chart.appVersion`，`image.tag` 留空。

| 入口 | 用途 |
| --- | --- |
| [Google Gerrit：V2 Change 10680](https://gke-ap-allowlist-review.googlesource.com/c/TrueWatch/+/10680) | 查看本次提交、审核意见和合并记录；后续修订沿用该项目 |
| [GoogleSource：TrueWatch 仓库](https://gke-ap-allowlist.googlesource.com/TrueWatch/) | 维护 `datakit/` 下的版本化 allowlist；本机检出位于 `~/workspace/gke-ap-allowlist-TrueWatch` |
| [Google Partner 目录](https://docs.cloud.google.com/kubernetes-engine/docs/resources/autopilot-partners) / [安装说明](https://docs.cloud.google.com/kubernetes-engine/docs/how-to/run-autopilot-partner-workloads) | Partner 接入、允许来源和正式同步流程 |
| [WorkloadAllowlist 字段](https://docs.cloud.google.com/kubernetes-engine/docs/reference/crds/workloadallowlist) / [准入排错](https://docs.cloud.google.com/kubernetes-engine/docs/troubleshooting/autopilot-privileged-workloads) | 扩展规则和定位 mismatch |
| [GKE 控制台](https://console.cloud.google.com/kubernetes/list?project=keen-alignment-466309-q1) | 本次测试项目的集群入口 |
| [Issue #3198](https://gitlab.jiagouyun.com/cloudcare-tools/datakit/-/issues/3198) / [MR !4176](https://gitlab.jiagouyun.com/cloudcare-tools/datakit/-/merge_requests/4176) | 需求与 DataKit 实现 |

当前 V2 的关键约定：

- 审核源文件：[truewatch-datakit-autopilot-v2.yaml](gke-autopilot/truewatch-datakit-autopilot-v2.yaml)；最低 GKE `1.35.6-gke.1258000`。
- Synchronizer 路径：`TrueWatch/datakit/truewatch-datakit-autopilot-v2.yaml`；集群允许来源使用 `gke://TrueWatch/datakit/truewatch-datakit-autopilot-v2.yaml`。
- WorkloadAllowlist 名及 Pod 的 `cloud.google.com/matching-allowlist`：`truewatch-datakit-autopilot-v2`。
- **本交付仅支持 TrueWatch**：`pubrepo.truewatch.com/truewatch/datakit`。
- 仅三个只读 hostPath：`/var/run/containerd`、`/proc`、`/var/log/pods`；缓存为 `emptyDir`。无 privileged、hostNetwork/PID/IPC，drop ALL capabilities，不包含 eBPF 或完整宿主机监控。
- V2 在 V1 基础上将 DataKit 配置变量名扩展为 `^ENV_[A-Z0-9_]+$`，支持字面值和逐键 `valueFrom`；保留 `POD_NAME`、`HOST_IP`、`HOST_ROOT`，未扩大宿主机权限。

## 2. Google 审核与后续扩展

**普通 DataKit 发版、已允许范围内的 ENV 配置变化可以复用 V2。** 更换镜像仓库、增加容器、挂载、capability 或其它超出匹配范围的字段，需要重新审核。V2 中的 `containerImageDigests` 用于已审核的其它仓库镜像匹配，并非官方仓库每次发版都要追加的版本表。

需要新权限修订时：

1. 从目标版本 Chart 渲染工作负载，列出“新增能力 → 所需权限 → 验证方法”。只保留必要权限，另建版本化文件及 `metadata.name`（例如 V3），保留已发布 V2。
2. 在团队已获 Google 授权的调试环境验证候选规则：正常 workload 允许，未批准的镜像、hostPath、capability 等仍被拒绝。准备候选 YAML、渲染清单、权限差异、实际镜像 digest 和脱敏的运行/数据证据。[候选生成说明](https://docs.cloud.google.com/kubernetes-engine/docs/how-to/autopilot-privileged-allowlists)中的 customer-owned 调试能力需要单独授权。
3. 用有 TrueWatch 项目权限的账号访问 Gerrit。新机器从项目页面获取认证及 clone 方法，安装 Gerrit `commit-msg` hook；`gcloud` 已登录不等于拥有 Gerrit 提交权限。以下在 Google allowlist 仓库执行，候选文件准备好后提交：

    ```shell
    git fetch origin
    git switch -c datakit-allowlist-v3 origin/main
    # 将审核候选保存为 datakit/truewatch-datakit-autopilot-v3.yaml
    git add datakit/truewatch-datakit-autopilot-v3.yaml
    git commit -m "Add DataKit Autopilot WorkloadAllowlist v3"
    git push origin HEAD:refs/for/main
    ```

4. Push 返回 Gerrit Change 地址；按 Google reviewer 意见修改，`git commit --amend` 保留 `Change-Id`，再次推送到 `refs/for/main` 更新同一审核。沿用团队 Partner 联系渠道推进审核与 Submit；审核材料按 reviewer 要求提供，不把整个报告包提交进 allowlist 仓库。
5. **Merged 后继续验证正式分发**：在没有手工安装候选 WLA 的客户形态环境中，使用新路径创建/更新 Synchronizer，确认 `Ready=True`、文件 `Installed` 和对应 WLA 存在。若路径仍 `not found`，携带 Change、完整路径、集群版本及同步错误联系 Google；不能用直接 apply WLA 代替正式同步验收。
6. 同步修改 DataKit 的审核源文件、包内 Synchronizer 路径、Partner values 的 `allowlistName` 和中英文文档，再完成下述发布与验收。升级先同步新规则；保留旧规则供旧工作负载及回退使用。

## 3. 复用原有 CI 发布

| 分支 | TrueWatch job | 版本与 Helm 仓库 |
| --- | --- | --- |
| `dev` | 无 Helm 发布 job | 合入 dev 不会上传 Chart |
| `testing` / `testing-*` | `release-testing-tw` | `CI_TESTING_VERSION` 写入 appVersion；`https://registry.jiagouyun.com/chartrepo/truewatch` |
| `master` | `release-prod-tw` | `CI_VERSION` 写入 appVersion；`https://pubrepo.truewatch.com/chartrepo/truewatch` |

以 [gitlab-ci.yml](../../gitlab-ci.yml) 为准，使用现有 Runner 凭据和发布流程，无需增加 Chart、job 或改 Makefile。Chart 会先上传，镜像随后构建推送，必须等待**整个 TrueWatch job 成功**并确认 `pubrepo.truewatch.com/truewatch/datakit:<appVersion>` 可拉取。

测试 Chart 的包版本不含 testing 后缀，相同 `CI_VERSION` 的不同测试分支可能覆盖同一包版本；下载后必须核对 `helm show chart` 的 `appVersion`，记录 job、包 SHA-256 和实际镜像 digest。digest 用于记录证据，部署仍跟随 CI 版本。

维护入口：

- [Partner values](../../charts/datakit/values-gke-autopilot-partner.yaml) 与 [Synchronizer](../../charts/datakit/gke-autopilot-partner-allowlist.yaml)：随普通 Chart 打包。同步器在 `templates/` 外，安装时独立 apply，Helm 不自动创建或接管它。
- [DaemonSet](../../charts/datakit/templates/daemonset.yaml) / [RBAC](../../charts/datakit/templates/clusterrole.yaml)：直接维护的资源模板；普通 values、README 则修改 `templates/charts-*.template.*` 源文件。
- 修改后本地检查：`python3 charts/tests/gke-autopilot-partner.py`（Helm 3、Python 3、PyYAML），再按客户文档验收实际发布包。

## 4. 按客户路径复测

1. 选择 GKE 版本满足要求的 Autopilot 集群，记录版本、节点架构和 container runtime。需要复用本次集群时，先确认它仍存在且允许测试，再获取凭据：

    ```shell
    gcloud container clusters get-credentials dk-autopilot-certification \
      --project keen-alignment-466309-q1 --region asia-southeast1
    ```

2. 在新目录按[客户文档](../../internal/export/doc/zh/gcp-gke-autopilot.md#partner-download)下载普通 `datakit` Chart；测试时替换为上表的测试仓库。核对 appVersion 及包内两个 Partner YAML，准备含有效 DataWay token 的私有用户 values。
3. 集群允许正式 `gke://` 来源后，先按文档同步 V2。已有同步器管理该路径时复用；本次测试使用 `truewatch-datakit-v2-verification`，包内新建名称为 `truewatch-datakit`。检查实际名称：

    ```shell
    SYNC_NAME="truewatch-datakit-v2-verification"
    kubectl wait --for=condition=Ready "allowlistsynchronizer/$SYNC_NAME" --timeout=10m
    kubectl get allowlistsynchronizer "$SYNC_NAME" -o yaml
    kubectl get workloadallowlist truewatch-datakit-autopilot-v2
    ```

4. 分别验证两种安装方式：Helm 加载 Partner values 和用户 values；YAML 用 `helm template --no-hooks` 渲染后 `kubectl apply`。YAML 路径需要本地 Helm 3，但不创建 Helm release。静态 `datakit.yaml` 是普通节点部署，`datakit-gke-autopilot.yaml` 是 Cloud API Deployment。
5. 检查 DaemonSet `desired > 0`、`ready = desired`、实际 Pod 的镜像和重启数；通过 Service proxy `/v1/health` 确认 `live=true`。创建输出唯一日志标记的业务 Pod，用集群标签和时间窗口在后端核对日志、指标、Kubernetes 对象及选举结果。
6. 验证 ENV 字面值/ConfigMap 引用更新、DataWay Secret 更新触发滚动、回退后实际配置恢复；Helm 用 upgrade/rollback，YAML 重新渲染/apply 并回放旧清单。有升级前版本时增加跨版本升级验证。
7. 卸载并检查 namespace 内资源及本次专用集群 RBAC；共享的 Synchronizer 保留。若要退役专用 allowlist，先清理依赖它的工作负载，再移除同步路径。保存版本、命令、脱敏状态及后端查询结果。

复测中的已知注意事项：

- 隔离已有安装必须同时使用独立 namespace 和 `fullnameOverride`；Partner 预设固定名称 `datakit`，只改 release 名无法隔离集群 RBAC。
- 空集群的 DaemonSet 可能是 `0/0`，Helm 成功也没有实际运行；先部署业务 Pod 触发节点，再验收。
- V2 禁止 exec/attach/port-forward；健康检查使用文档中的 Service proxy。默认日志写文件，可通过允许的 `ENV_LOG=stdout` 配置后用 `kubectl logs` 查看。
- 无 token 的 DataWay 地址会导致 `dataway.emptyToken`、上报失败和选举失败；Ready 不代表数据入库。
- `helm template` 使用 `--no-hooks` 排除历史 Helm test Pod；普通 Helm install/upgrade 不执行该测试 Pod。
- TrueWatch 基础镜像拉取出现 `TLS handshake timeout` 时，检查 Runner 及其 BuildKit 的 DNS、代理、CA 和出口连通性，修复后重试原 job；该错误发生在镜像构建拉取阶段。

## 5. 本次结果与后续验收

历史验证记录：V2 Change 10680 于 2026-09-06 merged，2026-09-15 已通过 Google 正式路径同步；9 月 16 日复测仍为 Ready/Installed。Google 合并、正式分发、DataKit 制品发布是三个独立完成点。

2026-09-16 使用 [TrueWatch job 475400](https://gitlab.jiagouyun.com/cloudcare-tools/datakit/-/jobs/475400) 发布的普通 Chart `2.12.0`，`appVersion=2.12.0-testing_testing-gke-autopilot-partner-v2`，在 GKE `1.35.7-gke.1222000` / COS / Linux amd64 / containerd 2.1.9 上完成 Helm、YAML 安装、配置更新、实际回退和卸载，两种部署均达到 **3/3 Ready、零重启、health live=true**。确认 containerd 连接及业务日志 tailer 启动，测试资源已清理。

本轮未提供有效 DataWay token，因此**数据入库、后端查询、选举成功仍待验收**；跨 DataKit 版本升级、arm64 运行及生产负载也未覆盖。该结果属于测试仓库制品，正式发布后仍需重新下载核对并复测。

脱敏 JSON/JSONL、发布包及校验信息保留在本机 `~/aa-plans/datakit-truewatch-gke-validation-2026-09-16-fzbuwwyw/`：`chart-summary.json`、`image-summary.json`、`commands.jsonl`、各阶段 status/health、`collection-summary.json` 和 `cleanup-check.json`。这些是历史证据；以后每次验收记录新的 CI job、版本和结果。
