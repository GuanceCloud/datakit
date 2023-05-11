
# DataKit Election
---

:fontawesome-brands-linux: :fontawesome-brands-windows: :fontawesome-brands-apple: :material-kubernetes: :material-docker:

---

When there is only one collected object (such as Kubernetes) in the cluster, but in the case of batch deployment, the configuration of multiple DataKits is exactly the same, and the collection of the central object is turned on. In order to avoid repeated collection, we can turn on the election function of DataKits.

Datakit has two election modes:

- Datakit self-election: In the same election namespace, the elected Datakit is responsible for all data collection, while other Datakits are in a pending state. The advantage is that the configuration is simple and there is no need to deploy additional applications. However, the disadvantage is that the elected Datakit has a higher resource utilization as all collectors run on this Datakit, which increases system resource usage.
- Collector task election[:octicons-tag-24: Version-1.7.0](changelog.md#cl-1.7.0): This mode is only applicable to Kubernetes environment. By deploying the [Datakit Operator](datakit-operator.md#datakit-operator-overview-and-install) program, task distribution can be achieved among various collectors of Datakit. The advantage is that the resource utilization of each Datakit is more balanced. However, the disadvantage is that an additional program needs to be deployed in the cluster.

## Collector Task Election Mode {#plugins-election}

### Deploy Datakit Operator {#install-operator}

The collector election mode requires the use of the Datakit Operator v1.0.5+ program. Refer to [here](datakit-operator.md#datakit-operator-install) for the deployment document.

### Election Configuration {#plugins-election-config}

Add an environment variable `ENV_DATAKIT_OPERATOR` with the value of the Datakit Operator address in Datakit's installation YAML file, for example:

```yaml
      containers:
      - env:
        - name: ENV_DATAKIT_OPERATOR
          value: https://datakit-operator.datakit.svc:443
```

The default service address of Datakit Operator is `datakit-operator.datakit.svc:443`.

<!-- markdownlint-disable MD046 -->
???+ info

    The priority of collector task election is higher than that of Datakit self-election. If a usable Datakit Operator address is configured, task election will be used first, otherwise Datakit self-election will be used.

## Datakit Self Election {#self-election}

### Election Configuration {#config}

=== "datakit.conf"

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
    
    Note: Collectors that support elections but are configured as `election = false` do not participate in elections, and their collection behavior and tag settings are not affected by elections; If datakit.conf closes the election, but the collector opens the election, its collection behavior and tag setting are the same as those of closing the election.

=== "Kubernetes"

    See [here](datakit-daemonset-deploy.md#env-elect)

### Election Principle {#how}

Take MySQL as an example. In the same cluster (such as k8s cluster), suppose there are 10 DataKits, 2 MySQL instances, and all DataKits have elections turned on (in Daemonset mode, the configuration of each DataKit is the same) and MySQL collector:

- Once a DataKit is elected, all MySQL data collection (the same is true for other election types) will be collected by the DataKit, regardless of whether the collected objects are one or more, and the winner takes all. Other DataKits that are not selected are on standby.
- Guance Cloud center will judge whether the currently selected DataKit is normal. If it is abnormal, the DataKit will be kicked off forcibly, and other DataKits in standby state will replace it.
- DataKit that does not open the election (it may not be in the current cluster). If MySQL collection is also configured, it will still collect MySQL data without election constraints.
- The scope of the election is at the level of `Workspace + Namespace` . In a single `Workspace + Namespace`, only one DataKit can be selected at a time.
    - With regard to workspaces, in datakit.conf, it is represented by the `token` URL parameter in the DataWay address string, and each workspace has its corresponding token.
    - The namespace for the election, in datakit.conf, is represented by the `namespace` configuration item. Multiple namespaces can be configured in one workspace.

### Election Class Collector's Global Tag Settings {#global-tags}

=== "datakit.conf"

    Under the condition of `conf.d/datakit.conf` opening the election, all the data collected by the collector that opened the election will try to append the global-env-tag in datakit.conf:
    
    ```toml
    [global_election_tags]
      # project = "my-project"
      # cluster = "my-cluster"
    ```
    
    If the original data has the corresponding tag in `global_election_tags`, the tag in the original data will prevail and will not be overwritten here.
    
    If the election is not turned on, the data collected by the election collector will be accompanied by the `global_host_tags` configured in datakit.conf (same as the non-election collector): [:octicons-tag-24: Version-1.4.8](changelog.md#cl-1.4.8).


    ```toml
    [global_host_tags]
      ip         = "__datakit_ip"
      host       = "__datakit_hostname"
    ```

=== "Kubernetes"

    See [here](datakit-daemonset-deploy.md#env-elect) for the configuration of elections in Kubernetes and [here](datakit-daemonset-deploy.md#env-common) for the setting of global tags.

## Collection List Supporting Election {#inputs}

The list of collectors currently supporting elections is as follows:

- [Apache](apache.md)
- [ElasticSearch](elasticsearch.md)
- [Gitlab](gitlab.md)
- [InfluxDB](influxdb.md)
- [Container](container.md)
- [MongoDB](mongodb.md)
- [MySQL](mysql.md)
- [NSQ](nsq.md)
- [Nginx](nginx.md)
- [PostgreSQL](postgresql.md)
- [Prom](prom.md)
- [RabbitMQ](rabbitmq.md)
- [Redis](redis.md)
- [Solr](solr.md)
- [TDengine](tdengine.md)

## FAQ {#faq}

### :material-chat-question: `host` Field Problem {#host}

For objects collected by collectors participating in elections, such as MySQL, because the DataKit collecting their data may change (election rotation occurs), by default, the data collected by such collectors will not take the tag `host` to avoid timeline growth. We recommend adding an additional `tags` field to the MySQL collector configuration:

```toml
[inputs..tags]
  host = "real-mysql-instance-name"
```

This way, the `host` field configured in tags will continue to be used when the DataKit has an election rotation.
