
# DataKit Election
---

:fontawesome-brands-linux: :fontawesome-brands-windows: :fontawesome-brands-apple: :material-kubernetes: :material-docker:

---

When there is only one collected object (such as Kubernetes) in the cluster, but in the case of batch deployment, the configuration of multiple DataKits is exactly the same, and the collection of the central object is turned on. In order to avoid repeated collection, we can turn on the election function of DataKits.

DataKit has two election modes:

- DataKit self-election: In the same election namespace, the elected DataKit is responsible for all data collection, while other DataKit are in a pending state. The advantage is that the configuration is simple and there is no need to deploy additional applications. However, the disadvantage is that the elected DataKit has a higher resource utilization as all collectors run on this DataKit, which increases system resource usage.
- Collector task election[:octicons-tag-24: Version-1.7.0](changelog.md#cl-1.7.0): This mode is only applicable to Kubernetes environment. By deploying the [DataKit Operator](datakit-operator.md#datakit-operator-overview-and-install) program, task distribution can be achieved among various collectors of DataKit. The advantage is that the resource utilization of each DataKit is more balanced. However, the disadvantage is that an additional program needs to be deployed in the cluster.

## Collector Task Election Mode {#plugins-election}

### Deploy DataKit Operator {#install-operator}

The collector election mode requires the use of the DataKit Operator v1.0.5+ program. Refer to [here](datakit-operator.md#datakit-operator-install) for the deployment document.

### Election Configuration {#plugins-election-config}

Add an environment variable `ENV_DATAKIT_OPERATOR` with the value of the DataKit Operator address in DataKit's installation YAML file, for example:

```yaml
      containers:
      - env:
        - name: ENV_DATAKIT_OPERATOR
          value: https://datakit-operator.datakit.svc:443
```

The default service address of DataKit Operator is `datakit-operator.datakit.svc:443`.

<!-- markdownlint-disable MD046 -->
???+ info

    The priority of collector task election is higher than that of DataKit self-election. If a usable DataKit Operator address is configured, task election will be used first, otherwise DataKit self-election will be used.

## DataKit Self Election {#self-election}

### Election Configuration {#config}

=== "`datakit.conf`"

    Edit `conf.d/datakit.conf`, and the election-related configuration is as follows:
    
    ```toml
    [election]
      # Open the election
      enable = false
    
      # Set the namespace of the election (default)
      namespace = "default"
    
      # tag that allows election space to be appended to data
      enable_namespace_tag = false
    
      ## election.tags: Election-related global tags
      [election.tags]
        #  project = "my-project"
        #  cluster = "my-cluster"
    ```
    
    You can configure the DataKit namespace if you want to separate elections for multiple DataKits, such as these 10 DataKits and the other 8 DataKits, without interfering with each other. DataKits in the same namespace participate in the same election.
    
    After the election is opened, if `enable_election_tag = true`（[:octicons-tag-24: Version-1.4.7](changelog.md#cl-1.4.7)） is opened at the same time, tag: `election_namespace = <your-namespace-name>` is automatically added to the data collected by the election class.
    
    After the election is opened in `conf.d/datakit.conf`, configure `election = true` in the collectors that need to participate in the election. (Currently, all collectors that support the election have `election` entries in their configuration files.)
    
    Note: Collectors that support elections but are configured as `election = false` do not participate in elections, and their collection behavior and tag settings are not affected by elections; If `datakit.conf`` closes the election, but the collector opens the election, its collection behavior and tag setting are the same as those of closing the election.

=== "Kubernetes"

    See [here](datakit-daemonset-deploy.md#env-elect)

### Viewing Election Status {#status}

After the election is configured, you can check the current election status of Datakit by [viewing the monitor](datakit-monitor.md#view). In the `Basic Info` section, there will be a line like this:

```not-set
Elected default::success|MacBook-Pro.local(elected: 4m40.554909s)
```

Here's what each part means:

- `default` indicates the election-namespace in which the current Datakit participates in the election. A workspace can have multiple election-namespaces dedicated to elections.
- `success` indicates that the current Datakit has election enabled and has been chosen as the leader.
- `MacBook-Pro.local` shows the hostname of the Datakit that was elected in the current namespace. If this hostname is the same as the current Datakit, the duration for which it has been the leader will be displayed afterward (`elected: 4m40.554909s`) [:octicons-tag-24: Version-1.5.8](changelog.md#cl-1.5.8)

If it is displayed as follows, it means that the current Datakit was not elected, but it will show which one was elected:

```not-set
Elected default::defeat|host-abc
```

Here's the breakdown:

- `default` indicates the namespace in which the current Datakit is participating in the election, as explained above.
- `defeat` indicates that the current Datakit has election enabled but was not successful. In addition to this, there are several other possible statuses:

    - **disabled**: The election feature is not enabled.
    - **success**: The election was successfully completed.
    - **banned**: The election feature is enabled, but it is not on the whitelist allowed for election [:octicons-tag-24: Version-1.35.0](../datakit/changelog.md#cl-1.35.0)

- `host-abc` shows the hostname of the Datakit that was elected in the current namespace.

### Election Principle {#how}

Take MySQL as an example. In the same cluster (such as k8s cluster), suppose there are 10 DataKits, 2 MySQL instances, and all DataKits have elections turned on (in DaemonSet mode, the configuration of each DataKit is the same) and MySQL collector:

- Once a DataKit is elected, all MySQL data collection (the same is true for other election types) will be collected by the DataKit, regardless of whether the collected objects are one or more, and the winner takes all. Other DataKits that are not selected are on standby.
- Guance Cloud will test whether the currently leader DataKit is alive(via heartbeat). If it is abnormal, the DataKit will be kicked off forcibly, and one of other DataKits in standby state will replace it.
- DataKit that does not open the election (it may not be in the current cluster). If MySQL collection is also configured, it will still collect MySQL data without election constraints.
- The scope of the election is at the level of `workspace + election-namespace` . In a single `workspace + election-namespace`, only one DataKit can be selected as the leader at a time.
    - With regard to workspaces, in *datakit.conf*, it is represented by the `token` URL parameter in the DataWay address string, and each workspace has its corresponding token.
    - The namespace for the election, in *datakit.conf*, is represented by the `namespace` configuration item. Multiple namespaces from different Datakits can be configured within one workspace.

<!-- markdownlint-disable MD013 -->
### Election Class Collector's Global Tag Settings {#global-tags}
<!-- markdownlint-enable -->

<!-- markdownlint-disable MD046 -->
=== "`datakit.conf`"

    Under the condition of `conf.d/datakit.conf` opening the election, all the data collected by the collector that opened the election will try to append the global-env-tag in `datakit.conf`:
    
    ```toml
    [election]
      [election.tags]
        # project = "my-project"
        # cluster = "my-cluster"
    ```
    
    If the original data has the corresponding tags, the tag in the original data will prevail and will not be overwritten here.
    
    If the election is not turned on, the data collected by the election collector will be accompanied by the `global_host_tags` configured in `datakit.conf` (same as the non-election collector): [:octicons-tag-24: Version-1.4.8](changelog.md#cl-1.4.8).


    ```toml
    [global_host_tags]
      ip         = "__datakit_ip"
      host       = "__datakit_hostname"
    ```

=== "Kubernetes"

    See [here](datakit-daemonset-deploy.md#env-elect) for the configuration of elections in Kubernetes and [here](datakit-daemonset-deploy.md#env-common) for the setting of global tags.
<!-- markdownlint-enable -->

## Election Whitelist {#election-whitelist}

[:octicons-tag-24: Version-1.35.0](../datakit/changelog.md#cl-1.35.0)

<!-- markdownlint-disable MD046 -->
=== "*datakit.conf*"

    For host installations, the election whitelist is configured through the `datakit.conf` file:

    ```toml
    [election]

      # election whitelist. If list empty, all host/node are allowed for election.
      node_whitelist = ["host-name-1", "host-name-2", "..."]
    ```

=== "Kubernetes"

    See [here](datakit-daemonset-deploy.md#env-elect)
<!-- markdownlint-enable -->

## Collection List Supporting Election {#inputs}

The list of collectors currently supporting elections is as follows:

- [Apache](../integrations/apache.md)
- [ElasticSearch](../integrations/elasticsearch.md)
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

> In fact, there are more collectors that support elections, and this information may not be up-to-date. Please refer to the specific documentation of the collector for the most accurate information.

## FAQ {#faq}

### :material-chat-question: `host` Field Problem {#host}

For objects collected by collectors participating in elections, such as MySQL, because the DataKit collecting their data may change (election rotation occurs), by default, the data collected by such collectors will not take the tag `host` to avoid timeline growth. We recommend adding an additional `tags` field to the MySQL collector configuration:

```toml
[inputs..tags]
  host = "real-mysql-instance-name"
```

This way, the `host` field configured in tags will continue to be used when the DataKit has an election rotation.
