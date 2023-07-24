
# DataKit 选举
---

:fontawesome-brands-linux: :fontawesome-brands-windows: :fontawesome-brands-apple: :material-kubernetes: :material-docker:

---

当集群中只有一个被采集对象（如 Kubernetes），但是在批量部署情况下，多个 DataKit 的配置完全相同，都开启了对该中心对象的采集，为了避免重复采集，我们可以开启 DataKit 的选举功能。

Datakit 有两种选举模式，即：

- Datakit 自选举：在同一个选举命名空间下，当选的 Datakit 负责全部采集，其他 Datakit 处于待定状态。优点是配置简单，不需要额外部署应用；缺点是对当选者的资源占用较大，所有的采集器都在这台 Datakit 上运行，系统资源占用增多。
- 采集器任务选举[:octicons-tag-24: Version-1.7.0](changelog.md#cl-1.7.0)：只适用于 Kubernetes 环境，通过部署 [Datakit Operator](datakit-operator.md#datakit-operator-overview-and-install) 程序，实现对 Datakit 各个采集器的任务分发。优点是各个 Datakit 的资源占用较为平均，缺点是需要在本集群额外部署一个程序。

## 采集器任务选举模式 {#plugins-election}

### 部署 Datakit Operator {#install-operator}

采集器选举模式需要用到 Datakit Operator v1.0.5 及以上版本，部署文档参见[此处](datakit-operator.md#datakit-operator-install)。

### 选举配置 {#plugins-election-config}

在 Datakit 安装 yaml 中添加一项环境变量 `ENV_DATAKIT_OPERATOR`，值为 Datakit Operator 地址，例如：

```yaml
      containers:
      - env:
        - name: ENV_DATAKIT_OPERATOR
          value: https://datakit-operator.datakit.svc:443
```

Datakit Operator 默认的 Service 地址是 `datakit-operator.datakit.svc:443`。

<!-- markdownlint-disable MD046 -->
???+ info

    采集器任务选举的优先级高于 Datakit 自选举。如果配置可用的 Datakit Operator 地址，会优先使用任务选举，否则会使用 Datakit 自选举。

## Datakit 自选举模式 {#self-election}

### 选举配置 {#config}

<!-- markdownlint-disable MD046 -->
=== "*datakit.conf*"

    编辑 `conf.d/datakit.conf`，选举有关的配置如下：
    
    ```toml
    [election]
      # 开启选举
      enable = false

      # 设置选举的命名空间(默认 default)
      namespace = "default"
    
      # 允许在数据上追加选举空间的 tag
      enable_namespace_tag = false
    
      ## election.tags: 选举相关全局标签
      [election.tags]
        #  project = "my-project"
        #  cluster = "my-cluster"
    ```
    
    如果要对多个 DataKit 区分选举，比如这 10 DataKit 和 另外 8 DataKit 分开选举，互相不干扰，可以配置 DataKit 命名空间。同一个命名空间下的 DataKit 参与同一起选举。
    
    开启选举后，如果同时开启 `enable_election_tag = true`（[:octicons-tag-24: Version-1.4.7](changelog.md#cl-1.4.7)），则在选举类采集的数据上，自动加上 tag: `election_namespace = <your-namespace-name>`。

    `conf.d/datakit.conf` 中开启选举后，在需要参加选举的采集器中配置 `election = true`（目前支持选举的采集器的配置文件中都带有 `election` 项）

    注意：支持选举但配置为 `election = false` 的采集器不参与选举，其采集行为、tag 设置均不受选举影响；如果 *datakit.conf* 关闭选举，但采集器开启选举，其采集行为、tag 设置均与关闭选举相同。

=== "Kubernetes"

    参见[这里](datakit-daemonset-deploy.md#env-elect)
<!-- markdownlint-enable -->

### 选举状态查看 {#status}

配置完选举后，通过[查看 monitor](datakit-monitor.md#view) 即可知道当前 Datakit 的选举状态，在 `Basic Info` 栏中，有如下行：

```not-set
Elected default::success|MacBook-Pro.local(elected: 4m40.554909s)
```

其中：

- `default` 表示当前 Datakit 参与选举的命名空间。一个工作空间可以有多个选举专用的命名空间
- `success` 表示当前 Datakit 开启了选举且选举成功
- `MacBook-Pro.local` 表示当前命名空间被选上的 Datakit 所在主机名。如果该主机名就是当前这个 Datakit，则后面会显示其当选 leader 的时长（`elected: 4m40.554909s`）[:octicons-tag-24: Version-1.5.8](changelog.md#cl-1.5.8)

如果是如下显示，则表示当前 Datakit 未被选上，但会显示当前是哪个主机被选上：

```not-set
Elected default::defeat|host-abc
```

其中：

- `default` 表示当前 Datakit 参与选举的命名空间，同上
- `defeat` 表示当前 Datakit 开启了，但选举失败
- `host-abc` 表示当前命名空间被选上的 Datakit 所在主机名

### 选举原理 {#how}

以 MySQL 为例，在同一个集群（如 K8s cluster）中，假定有 10 Datakit、2 个 MySQL 实例，且 Datakit 都开启了选举（DaemonSet 模式下，每个 Datakit 的配置都是一样的）以及 MySQL 采集器：

- 一旦某个 Datakit 被选举上，那么所有 MySQL （其它选举类的采集也一样）的数据采集，都将由该 Datakit 来采集，不管被采集对象是一个还是多个，赢者通吃。其它未选上的 Datakit 出于待命状态。
- 观测云中心会判断当前选上的 Datakit 是否正常，如果异常，则强行踢掉该 Datakit，其它待命状态的 Datakit 将替代它
- 未开启选举的 Datakit（可能它不在当前集群中），如果也配置了 MySQL 采集，不受选举约束，它仍然会去采集 MySQL 的数据
- 选举的范围是「工作空间+命名空间」级别的，单个「工作空间+命名空间」中，一次最多只能有一个 Datakit 被选上
    - 关于工作空间，在 *datakit.conf* 中，通过 Dataway 地址串中的 `token` URL 参数来表示，每个工作空间，都有其对应 token
    - 关于选举的命名空间，在 *datakit.conf* 中，通过 `namespace` 配置项来表示。一个工作空间可以配置多个命名空间

### 选举类采集器的全局 tag 设置 {#global-tags}

<!-- markdownlint-disable MD046 -->
=== "*datakit.conf*"

    在 `conf.d/datakit.conf` 开启选举的条件下，开启了选举的采集器采集到的数据，均会尝试追加 *datakit.conf* 中的 `global_election_tag`：
    
    ```toml
    [election]
      [election.tags]
        # project = "my-project"
        # cluster = "my-cluster"
    ```

    如果原始数据上就带有了这里的 tag，则以原始数据中带有的 tag 为准，此处不会覆盖。

    如果没有开启选举，则选举采集器采集到的数据中，均会带上 *datakit.conf* 中配置的 `global_host_tags`（跟非选举类采集器一样）：[:octicons-tag-24: Version-1.4.8](changelog.md#cl-1.4.8) ·

    ```toml
    [global_host_tags]
      ip         = "__datakit_ip"
      host       = "__datakit_hostname"
    ```

=== "Kubernetes"

    Kubernetes 中选举的配置参见[这里](datakit-daemonset-deploy.md#env-elect)，全局 tag 的设置参见[这里](datakit-daemonset-deploy.md#env-common)。
<!-- markdownlint-enable -->

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

### :material-chat-question: `host` 字段问题 {#host}

对于由参与选举的采集器采集的对象，比如 MySQL，由于采集其数据的 DataKit 可能会变迁（发生了选举轮换），故默认情况下，这类采集器采集的数据不会带上 `host` 这个 tag，以避免时间线增长。我们建议在 MySQL 采集器配置上，增加额外的 `tags` 字段：

```toml
[inputs.{{.InputName}}.tags]
  host = "real-mysql-instance-name"
```

这样，当 DataKit 发生选举轮换时，会继续沿用 tags 中配置的 `host` 字段。
