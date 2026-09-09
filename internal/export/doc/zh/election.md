
# DataKit 选举
---

:fontawesome-brands-linux: :fontawesome-brands-windows: :fontawesome-brands-apple: :material-kubernetes: :material-docker:

---

多个 DataKit 使用相同配置采集同一组目标（例如 Kubernetes 集群或 MySQL 实例）时，可以通过中心选举选出一个采集 Leader，避免选举类采集器重复采集。其他 DataKit 保持待命，在 Leader 失效后参与接替。

中心选举支持两种后端：[DataWay/Kodo](election.md#dataway-election) 和 [DataKit Operator](election.md#operator-election)。两者共用 DataKit 的选举配置、[选举白名单](election.md#election-whitelist)和采集器启停逻辑；后端负责裁决，当选的 DataKit 负责采集。

从 DataKit [1.85.0](changelog-2025.md#cl-1.85.0) 起，采集器任务选举已被移除。本文说明的 Operator 中心选举只裁决采集 Leader，不分配采集器任务，也不在 Operator 中运行采集器。

## DataKit 中心选举 {#self-election}

### 选举原理 {#how}

同一选举后端中，选举作用域由工作空间和选举命名空间共同确定：

- 工作空间由 DataWay 地址中的 token 确定；Operator 模式复用首个 DataWay token，无需额外配置 token。
- `namespace`（环境变量 `ENV_NAMESPACE`）是 DataKit 的选举命名空间，默认 `default`，与 Kubernetes namespace 无关。不同选举命名空间分别选举。
- 同一作用域中只选出一个采集 Leader。它负责本机配置的所有选举类采集器，其他候选 DataKit 待命；Leader 通过心跳维持租约，失效后由其他候选者参与接替。

例如，10 个 DataKit 都配置了相同的两个 MySQL 实例，并为 MySQL 采集器开启选举，则当选的 DataKit 负责采集这两个实例。名单内多个节点不会按采集器或目标分摊任务。

DataWay/Kodo 与 Operator 是两个独立的选举域。采集同一组目标的 DataKit 应统一使用同一后端；切换时按[迁移与回滚步骤](election.md#operator-migration)执行，避免两个域各产生一个 Leader。

### 选举配置 {#config}

两种后端都需要开启 DataKit 的全局选举开关，并在需要参与选举的采集器中设置 `election = true`。支持选举的采集器会在配置文件中提供该选项，参见[采集器列表](election.md#inputs)。

主机部署通过 `[election].enable = true` 开启；Kubernetes 部署通过非空的 `ENV_ENABLE_ELECTION` 开启。该环境变量按是否非空判断，关闭时应移除或设为空，不要填写字符串 `false`。

全局选举关闭时，采集器按未启用选举的方式运行；采集器自身设置 `election = false` 时，其采集行为和标签不受选举影响。以下配置示例均合并到已有配置中，主机配置保存后需[重启 DataKit](datakit-service-how-to.md#manage-service)，Kubernetes 配置变更后需更新 DataKit Pod。

#### DataWay/Kodo 选举 {#dataway-election}

这是默认选举后端。开启选举并保持 `operator_url` / `ENV_ELECTION_OPERATOR_URL` 为空，即通过已有 DataWay/Kodo 链路参与选举。

<!-- markdownlint-disable MD046 -->
=== "主机部署"

    编辑 `conf.d/datakit.conf`：

    ```toml
    [election]
      enable = true
      namespace = "default"
      operator_url = ""
    ```

=== "Kubernetes"

    在 DataKit 容器的 `env` 列表中配置：

    ```yaml
    - name: ENV_ENABLE_ELECTION
      value: "on"
    - name: ENV_NAMESPACE
      value: "default"
    ```

    保持 `ENV_ELECTION_OPERATOR_URL` 未设置或为空。
<!-- markdownlint-enable MD046 -->

#### DataKit Operator 选举 {#operator-election}

使用 Operator 中心选举须同时满足以下条件：

1. **DataKit Operator v1.9.1 及以上版本**：从 v1.9.1 起支持中心选举。
2. **DataKit 2.12.0 及以上版本**：支持选择 Operator 作为中心选举后端。
3. **手动配置 DataKit 环境变量**：Kubernetes 部署需显式设置 `ENV_ELECTION_OPERATOR_URL`，同时开启 `ENV_ENABLE_ELECTION`。仅升级或部署 Operator 不会自动启用此后端。
4. **为 Operator 补充选举 RBAC 权限**：Operator 的 ServiceAccount 需要在 Operator 所在 Kubernetes namespace 中读写 `coordination.k8s.io` 的 `leases` 资源。Role/RoleBinding 配置和已有部署的增量升级步骤，详见 [Operator 文档](datakit-operator.md#central-election-upgrade)。仅升级镜像不能补齐旧部署缺少的权限。

Operator 的选举能力说明参见 [DataKit 中心选举协调](datakit-operator.md#central-election)。

<!-- markdownlint-disable MD046 -->
=== "Kubernetes"

    手动在 DataKit 容器的 `env` 列表中配置以下变量，并更新 Pod：

    ```yaml
    - name: ENV_ENABLE_ELECTION
      value: "on"
    - name: ENV_ELECTION_OPERATOR_URL
      value: "https://datakit-operator.datakit.svc:443"
    - name: ENV_NAMESPACE
      value: "default"
    ```

    所有参与同一次选举的 DataKit 应使用相同的选举命名空间，并能访问同一个 Operator 服务。

=== "Helm"

    在 DataKit 的 values 中显式配置以下参数，由 Chart 生成对应的选举环境变量，再执行 Helm 更新：

    ```yaml
    datakit:
      enabled_election: true
      election_operator_url: "https://datakit-operator.datakit.svc:443"
    ```

    其他参数参见 [DataKit Helm 部署](datakit-helm.md)。

=== "主机部署"

    对于能够访问 Operator 服务的主机，在 `conf.d/datakit.conf` 中配置等价的 TOML 参数：

    ```toml
    [election]
      enable = true
      namespace = "default"
      operator_url = "https://<operator-host>:443"
    ```

    将 `<operator-host>` 替换为该主机可访问且 TLS 证书匹配的 Operator 地址。
<!-- markdownlint-enable MD046 -->

集群内常用地址为 `https://datakit-operator.datakit.svc:443`，自定义 Service 名称或 namespace 时应相应调整。省略协议会补充 `https://`；明文 HTTP 仅允许用于回环地址。URL 不能包含用户名、密码、查询参数或 fragment，非法配置需要修正。

DataKit 仅在进程启动时探测一次 `GET /v1/dk-election/status`，超时为 2 秒，不携带 token，也不发起竞选。只有返回 `200` 且 `content.status` 为 `ready` 才选择 Operator：

| 选举开关 | Operator 地址与启动检查 | 本次进程使用的后端 |
| --- | --- | --- |
| 关闭 | 任意 | 不参与选举，不探测 Operator |
| 开启 | 地址为空 | DataWay/Kodo |
| 开启 | 地址非空，检查通过 | DataKit Operator |
| 开启 | 地址非空，检查失败 | DataWay/Kodo |

旧 Operator 返回 `404`、缺少 Lease RBAC、cache 未就绪、连接失败、超时或响应不符合协议时，启动检查失败，日志会记录回退原因。`/v1/ping` 或 `/v1/ready` 正常不能证明中心选举可用。竞选与心跳分别使用 `POST /v1/dk-election` 和 `POST /v1/dk-election/heartbeat`。

**后端在本次进程生命周期内固定**：Operator 运行期故障不会切回 DataWay/Kodo；补齐权限或恢复 Operator 后，已选择 DataWay/Kodo 的 DataKit 也不会自动切回。只有重启 DataKit 才会重新探测并选择。

Operator 短暂不可用时，当前 Leader 可以在本地安全租约内继续采集。超过服务端租约减去一个心跳间隔后，DataKit 会暂停选举类采集；收到新的有效 `success` 响应后才恢复。

#### 选举白名单 {#election-whitelist}

选举白名单从 DataKit [1.35.0](changelog.md#cl-1.35.0) 起支持，**对 DataWay/Kodo 和 Operator 两种后端都有效**。它限制哪些 DataKit 可以成为候选者，不决定选举后端。

- 列表为空：所有节点均可参与选举。
- 列表非空：只有名称匹配的节点可参与，其余节点状态为 `banned`。
- 多个匹配节点之间正常选举，列表顺序不代表优先级。若只配置一个节点，该节点不可用时，名单外节点不会自动接替。

匹配使用 DataKit 实际选举身份对应的主机名，区分大小写，要求完整一致，不支持通配符或正则。标准 Kubernetes 部署通常使用 `ENV_K8S_NODE_NAME` 对应的 Node 名称；如配置了 `ENV_K8S_CLUSTER_NODE_NAME` 或 `ENV_HOSTNAME` 覆盖主机名，应使用覆盖后的名称。可通过 `datakit_election_status` 指标的 `id` 标签或选举启动日志的 `id` 字段确认。

<!-- markdownlint-disable MD046 -->
=== "主机部署"

    在两种后端的配置中都可以追加 `node_whitelist`：

    ```toml
    [election]
      enable = true
      node_whitelist = ["node-a", "node-b"]
    ```

=== "Kubernetes"

    保持 `ENV_ENABLE_ELECTION` 开启，并在 DataKit 容器的 `env` 列表中配置 JSON 数组：

    ```yaml
    - name: ENV_ENABLE_ELECTION
      value: "on"
    - name: ENV_ELECTION_NODE_WHITELIST
      value: '["node-a", "node-b"]'
    ```

    也支持逗号分隔的 `node-a,node-b`；名称前后不要加空格。完整参数参见[选举环境变量](datakit-daemonset-deploy.md#env-elect)。
<!-- markdownlint-enable MD046 -->

白名单在 DataKit 启动时检查。应将一致的名单配置到参与同一次选举的所有 DataKit，并重启相关实例使其生效。修改名单不会通过 Operator 自动分发。

### 选举状态查看 {#status}

配置完选举后，通过[查看 monitor](datakit-monitor.md#view) 即可知道当前 DataKit 的选举状态，在 `Basic Info` 栏中，有如下行：

```not-set
Elected default::success|MacBook-Pro.local(elected: 4m40.554909s)
```

其中：

- `default` 表示当前 DataKit 参与选举的命名空间。一个工作空间可以有多个选举专用的命名空间
- `success` 表示当前 DataKit 开启了选举且选举成功
- `MacBook-Pro.local` 表示当前命名空间被选上的 DataKit 所在主机名。如果该主机名就是当前这个 DataKit，则后面会显示其当选 leader 的时长（`elected: 4m40.554909s`）[:octicons-tag-24: Version-1.5.8](changelog.md#cl-1.5.8)

如果是如下显示，则表示当前 DataKit 未被选上，但会显示当前是哪个主机被选上：

```not-set
Elected default::defeat|host-abc
```

其中：

- `default` 表示当前 DataKit 参与选举的命名空间，同上
- `defeat` 表示当前 DataKit 开启了，但选举失败。除此之外，还有如下几种可能的状态：

    - **disabled**：未开启选举功能
    - **success**：选举成功完成
    - **banned**：选举功能已开启，但自己未列在选举允许的白名单中 [:octicons-tag-24: Version-1.35.0](../datakit/changelog.md#cl-1.35.0)

- `host-abc` 表示当前命名空间被选上的 DataKit 所在主机名

自监控指标还提供以下信息：

- `datakit_election_provider_info`：进程启动时选定的选举后端，`provider` 标签为 `dataway` 或 `operator`；用于确认实际选择，不能仅根据配置的 URL 判断
- `datakit_election_last_success_timestamp_seconds`：最近一次 Leader 成功响应时间
- `datakit_election_lease_remaining_seconds`：Operator 模式本地安全租约的剩余秒数
- `datakit_election_epoch`：最近一次 Operator 任期
- `datakit_election_request_errors_total`：按有限错误原因分类的请求失败数
- `datakit_election_transitions_total`：按原因记录的 Leader 生命周期迁移

现有 `datakit_election_status` 指标继续提供 status 与当前 holder。以上指标和日志均不会包含 workspace token 或带认证参数的完整请求 URL。

### Operator 切换与回滚 {#operator-migration}

切换前，确认已满足 [Operator 版本、DataKit 版本、手动环境变量配置和额外 RBAC 要求](election.md#operator-election)。两种后端各自裁决 Leader，不能通过只更新部分 DataKit 完成迁移。

1. 部署 Operator v1.9.1 及以上版本，并按 [Operator 文档](datakit-operator.md#central-election-upgrade)补齐权限，确认 `GET /v1/dk-election/status` 返回 `200` 且 `content.status` 为 `ready`。
2. 将参与同一次选举的所有 DataKit 升级到 2.12.0 及以上版本，暂时保持 `operator_url` / `ENV_ELECTION_OPERATOR_URL` 为空，继续使用 DataWay/Kodo。
3. 记录当前 Leader 的选举身份，将所有 DataKit 的[白名单](election.md#election-whitelist)临时固定为该名称；完成重启，确认其他节点均为 `banned`。
4. 为所有 DataKit 设置相同的 Operator 地址，完成受控更新。从最后一个允许竞选的 DataWay/Kodo 模式进程退出起，等待旧后端租约失效；确认所有实例选择 Operator，且只有一个采集 Leader。
5. 移除临时白名单，再次完成更新，恢复其他节点的候选资格，并验证选举状态。

回滚时执行对称操作：先将所有实例的白名单固定为当前 Operator Leader 并完成更新，再清空所有实例的 Operator 地址并重启。等待旧 Operator 租约失效、确认 DataWay/Kodo Leader 稳定后，移除临时白名单并再次更新。

如果无法固定候选节点，应先暂停选举类采集并等待旧后端租约失效，再启用新后端，避免两个选举域重叠采集。

### 选举类采集器的全局 tag 设置 {#global-tags}

<!-- markdownlint-disable MD046 -->
=== "*datakit.conf*"

    在 `conf.d/datakit.conf` 开启选举的条件下，选举类采集器采集的数据会尝试追加 `[election.tags]` 中配置的全局标签：
    
    ```toml
    [election]
      enable_namespace_tag = false # 设为 true 时追加 election_namespace 标签
      [election.tags]
        # project = "my-project"
        # cluster = "my-cluster"
    ```

    如果原始数据上就带有了这里的 tag，则以原始数据中带有的 tag 为准，此处不会覆盖。

    如果没有开启选举，则选举采集器采集到的数据中，均会带上 *datakit.conf* 中配置的 `global_host_tags`（跟非选举类采集器一样）：[:octicons-tag-24: Version-1.4.8](changelog.md#cl-1.4.8)

    ```toml
    [global_host_tags]
      ip         = "__datakit_ip"
      host       = "__datakit_hostname"
    ```

=== "Kubernetes"

    使用 `ENV_GLOBAL_ELECTION_TAGS` 配置选举类采集器的全局标签；设置非空的 `ENV_ENABLE_ELECTION_NAMESPACE_TAG` 可追加 `election_namespace` 标签。参数详情参见[选举环境变量](datakit-daemonset-deploy.md#env-elect)。
<!-- markdownlint-enable MD046 -->

## 支持选举的采集列表 {#inputs}

目前支持选举的采集器列表如下：

- [Apache](../integrations/apache.md)
- [Elasticsearch](../integrations/elasticsearch.md)
- [GitLab](../integrations/gitlab.md)
- [InfluxDB](../integrations/influxdb.md)
- [Container](../integrations/container.md)
- [MongoDB](../integrations/mongodb.md)
- [MySQL](../integrations/mysql.md)
- [NSQ](../integrations/nsq.md)
- [Nginx](../integrations/nginx.md)
- [PostgreSQL](../integrations/postgresql.md)
- [Prom](../integrations/prom.md)
- [RabbitMQ](../integrations/rabbitmq.md)
- [Redis](../integrations/redis.md)
- [Solr](../integrations/solr.md)
- [TDengine](../integrations/tdengine.md)

> 事实上，支持选举的采集器会更多，此处可能更新不及时，以具体采集器的文档为准。

## FAQ {#faq}

### `host` 字段问题 {#host}

对于由参与选举的采集器采集的对象，比如 MySQL，由于采集其数据的 DataKit 可能会变迁（发生了选举轮换），故默认情况下，这类采集器采集的数据不会带上 `host` 这个 tag，以避免时间线增长。我们建议在 MySQL 采集器配置上，增加额外的 `tags` 字段：

```toml
[inputs.mysql.tags]
  host = "real-mysql-instance-name"
```

这样，当 DataKit 发生选举轮换时，会继续沿用 tags 中配置的 `host` 字段。
