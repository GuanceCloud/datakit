
# DataKit Election
---

:fontawesome-brands-linux: :fontawesome-brands-windows: :fontawesome-brands-apple: :material-kubernetes: :material-docker:

---

When multiple DataKit instances use the same configuration to collect the same targets, such as a Kubernetes cluster or MySQL instances, central election selects one Collection Leader to prevent duplicate collection by election-enabled collectors. Other DataKit instances remain on standby and can take over when the Leader becomes unavailable.

Central election supports two providers: [DataWay/Kodo](election.md#dataway-election) and [DataKit Operator](election.md#operator-election). Both use the same DataKit election configuration, [election whitelist](election.md#election-whitelist), and collector lifecycle. The provider arbitrates the election, and the elected DataKit runs the collectors.

Collector task election was removed in DataKit [1.85.0](changelog-2025.md#cl-1.85.0). The Operator central election described here selects a Collection Leader; it does not distribute collector tasks or run collectors inside Operator.

## DataKit Central Election {#self-election}

### Election Principle {#how}

Within one election provider, the workspace and election namespace define the election scope:

- The workspace is identified by the token in the DataWay URL. Operator mode reuses the first DataWay token, so no additional token is required.
- `namespace` (environment variable `ENV_NAMESPACE`) is the DataKit election namespace. It defaults to `default` and is independent of Kubernetes namespaces. Different election namespaces elect their Leaders separately.
- One Collection Leader is elected per scope. It runs all election-enabled collectors configured on that DataKit while the other candidates remain on standby. The Leader renews its lease through heartbeats; other candidates can take over when it becomes unavailable.

For example, if 10 DataKit instances configure the same two MySQL targets and enable election for the MySQL collector, the elected DataKit collects both targets. Whitelisted nodes do not divide collectors or targets among themselves.

DataWay/Kodo and Operator are independent election domains. DataKit instances collecting the same targets should use the same provider. Follow the [migration and rollback procedure](election.md#operator-migration) when switching to prevent each domain from electing its own Leader.

### Election Configuration {#config}

Both providers require the global DataKit election switch to be enabled and `election = true` in each collector that should participate. Collectors supporting election expose this option in their configuration; see the [collector list](election.md#inputs).

For host deployments, set `[election].enable = true`. For Kubernetes, set `ENV_ENABLE_ELECTION` to a non-empty value. This variable is checked for a non-empty value, so remove it or set it to an empty string to disable election; the string `false` also enables it.

When global election is disabled, collectors run as if election were disabled. A collector with `election = false` keeps its collection and tag behavior independent of election. Merge the examples below into existing configuration. After changing host configuration, [restart DataKit](datakit-service-how-to.md#manage-service); after changing Kubernetes configuration, update the DataKit Pods.

#### DataWay/Kodo {#dataway-election}

This is the default election provider. Enable election and leave `operator_url` / `ENV_ELECTION_OPERATOR_URL` empty to use the existing DataWay/Kodo path.

<!-- markdownlint-disable MD046 -->
=== "Host deployment"

    Edit `conf.d/datakit.conf`:

    ```toml
    [election]
      enable = true
      namespace = "default"
      operator_url = ""
    ```

=== "Kubernetes"

    Configure the DataKit container's `env` list:

    ```yaml
    - name: ENV_ENABLE_ELECTION
      value: "on"
    - name: ENV_NAMESPACE
      value: "default"
    ```

    Leave `ENV_ELECTION_OPERATOR_URL` unset or empty.
<!-- markdownlint-enable MD046 -->

#### DataKit Operator {#operator-election}

Operator central election requires all of the following:

1. **DataKit Operator v1.9.1 or later**: Central election is supported starting with v1.9.1.
2. **DataKit 2.12.0 or later**: This version supports selecting Operator as the central election provider.
3. **Manually configured DataKit environment variables**: For Kubernetes deployments, explicitly set `ENV_ELECTION_OPERATOR_URL` and enable `ENV_ENABLE_ELECTION`. Installing or upgrading Operator alone does not enable this provider.
4. **Additional election RBAC for Operator**: The Operator ServiceAccount needs access to read and write `leases` in the `coordination.k8s.io` API group within Operator's Kubernetes namespace. See the [Operator documentation](datakit-operator.md#central-election-upgrade) for the Role/RoleBinding and incremental upgrade procedure. Updating only the image does not add missing permissions to an existing deployment.

For the Operator's election capability, see [DataKit Central Election Coordination](datakit-operator.md#central-election).

<!-- markdownlint-disable MD046 -->
=== "Kubernetes"

    Manually configure the following variables in the DataKit container's `env` list, then update the Pods:

    ```yaml
    - name: ENV_ENABLE_ELECTION
      value: "on"
    - name: ENV_ELECTION_OPERATOR_URL
      value: "https://datakit-operator.datakit.svc:443"
    - name: ENV_NAMESPACE
      value: "default"
    ```

    All DataKit instances participating in the same election should use the same election namespace and be able to reach the same Operator service.

=== "Helm"

    Explicitly configure these DataKit values so that the Chart generates the election environment variables, then apply the Helm update:

    ```yaml
    datakit:
      enabled_election: true
      election_operator_url: "https://datakit-operator.datakit.svc:443"
    ```

    See [DataKit Helm Deployment](datakit-helm.md) for other parameters.

=== "Host deployment"

    For hosts that can reach the Operator service, configure the equivalent TOML settings in `conf.d/datakit.conf`:

    ```toml
    [election]
      enable = true
      namespace = "default"
      operator_url = "https://<operator-host>:443"
    ```

    Replace `<operator-host>` with an Operator address reachable from the host and covered by the TLS certificate.
<!-- markdownlint-enable MD046 -->

A typical in-cluster address is `https://datakit-operator.datakit.svc:443`; adjust it for a custom Service name or namespace. DataKit adds `https://` when the scheme is omitted. Plain HTTP is accepted only for a loopback address. URLs must not include user information, credentials, query parameters, or a fragment. Invalid configuration must be corrected.

At process startup, DataKit probes `GET /v1/dk-election/status` once with a 2-second timeout. The probe sends no token and does not campaign. Operator is selected only when the response is `200` with `content.status` set to `ready`:

| Election switch | Operator URL and startup check | Provider for this process |
| --- | --- | --- |
| Disabled | Any | No election and no Operator probe |
| Enabled | Empty URL | DataWay/Kodo |
| Enabled | Non-empty URL, check passes | DataKit Operator |
| Enabled | Non-empty URL, check fails | DataWay/Kodo |

An old Operator returning `404`, missing Lease RBAC, an unready cache, a connection failure, a timeout, or an invalid response fails the check. The startup log records the fallback reason. A healthy `/v1/ping` or `/v1/ready` response does not establish central election availability. Campaigns and heartbeats use `POST /v1/dk-election` and `POST /v1/dk-election/heartbeat`, respectively.

**The provider stays fixed for the process lifetime**. An Operator failure at runtime does not switch DataKit to DataWay/Kodo. Granting permissions or restoring Operator does not switch existing DataWay/Kodo processes back either. Only restarting DataKit triggers a new probe and selection.

During a short Operator outage, the current Leader can continue collection within its local safe lease. After the server lease minus one heartbeat interval has elapsed, DataKit pauses election-enabled collection. Collection resumes only after a new valid `success` response.

#### Election Whitelist {#election-whitelist}

The election whitelist has been supported since DataKit [1.35.0](changelog.md#cl-1.35.0) and **applies to both DataWay/Kodo and Operator**. It limits which DataKit instances can become candidates; it does not select the provider.

- An empty list allows all nodes to participate.
- A non-empty list allows only nodes whose names match. Other nodes report `banned`.
- Matching nodes participate in the normal election. List order does not indicate priority. If only one node is listed and it becomes unavailable, nodes outside the list do not automatically take over.

Matching uses the hostname that identifies DataKit in the election. It is exact and case-sensitive, with no wildcard or regular-expression support. Standard Kubernetes deployments usually use the Node name from `ENV_K8S_NODE_NAME`. If `ENV_K8S_CLUSTER_NODE_NAME` or `ENV_HOSTNAME` overrides the hostname, use the overridden name. Check the `id` label of `datakit_election_status` or the `id` field in the election startup log for this identity.

<!-- markdownlint-disable MD046 -->
=== "Host deployment"

    Add `node_whitelist` to the configuration for either provider:

    ```toml
    [election]
      enable = true
      node_whitelist = ["node-a", "node-b"]
    ```

=== "Kubernetes"

    Keep `ENV_ENABLE_ELECTION` enabled and configure a JSON array in the DataKit container's `env` list:

    ```yaml
    - name: ENV_ENABLE_ELECTION
      value: "on"
    - name: ENV_ELECTION_NODE_WHITELIST
      value: '["node-a", "node-b"]'
    ```

    A comma-separated value such as `node-a,node-b` is also supported; do not add spaces around names. See [election environment variables](datakit-daemonset-deploy.md#env-elect) for the full parameter list.
<!-- markdownlint-enable MD046 -->

DataKit checks the whitelist at startup. Apply the same list to all DataKit instances participating in the election and restart them for the change to take effect. Operator does not distribute whitelist changes.

### Viewing Election Status {#status}

After the election is configured, you can check the current election status of DataKit by [viewing the monitor](datakit-monitor.md#view). In the `Basic Info` section, there will be a line like this:

```not-set
Elected default::success|MacBook-Pro.local(elected: 4m40.554909s)
```

Here's what each part means:

- `default` indicates the election-namespace in which the current DataKit participates in the election. A workspace can have multiple election-namespaces dedicated to elections.
- `success` indicates that the current DataKit has election enabled and has been chosen as the leader.
- `MacBook-Pro.local` shows the hostname of the DataKit that was elected in the current namespace. If this hostname is the same as the current DataKit, the duration for which it has been the leader will be displayed afterward (`elected: 4m40.554909s`) [:octicons-tag-24: Version-1.5.8](changelog.md#cl-1.5.8)

If it is displayed as follows, it means that the current DataKit was not elected, but it will show which one was elected:

```not-set
Elected default::defeat|host-abc
```

Here's the breakdown:

- `default` indicates the namespace in which the current DataKit is participating in the election, as explained above.
- `defeat` indicates that the current DataKit has election enabled but was not successful. In addition to this, there are several other possible statuses:

    - **disabled**: The election feature is not enabled.
    - **success**: The election was successfully completed.
    - **banned**: The election feature is enabled, but it is not on the whitelist allowed for election [:octicons-tag-24: Version-1.35.0](../datakit/changelog.md#cl-1.35.0)

- `host-abc` shows the hostname of the DataKit that was elected in the current namespace.

Self-monitoring metrics also expose:

- `datakit_election_provider_info`: the provider selected at process startup, with `provider` set to `dataway` or `operator`; use this to confirm the actual selection instead of relying on the configured URL
- `datakit_election_last_success_timestamp_seconds`: the most recent successful Leader response
- `datakit_election_lease_remaining_seconds`: seconds remaining before the local safe lease deadline in Operator mode
- `datakit_election_epoch`: the most recently observed Operator epoch
- `datakit_election_request_errors_total`: request failures grouped by bounded error reasons
- `datakit_election_transitions_total`: Leader lifecycle transitions grouped by reason

The existing `datakit_election_status` metric continues to expose status and the current holder. These metrics and election logs never include a workspace token or a complete credential-bearing request URL.

### Migration and Rollback {#operator-migration}

Before switching, meet the [Operator version, DataKit version, manual environment configuration, and additional RBAC requirements](election.md#operator-election). Each provider elects its own Leader, so updating only some DataKit instances does not complete a migration.

1. Deploy Operator v1.9.1 or later, grant permissions as described in the [Operator documentation](datakit-operator.md#central-election-upgrade), and confirm that `GET /v1/dk-election/status` returns `200` with `content.status` set to `ready`.
2. Upgrade every DataKit participating in the election to 2.12.0 or later. Keep `operator_url` / `ENV_ELECTION_OPERATOR_URL` empty for now so that they continue using DataWay/Kodo.
3. Record the current Leader's election identity and temporarily pin the [whitelist](election.md#election-whitelist) on every DataKit to that name. Complete the restarts and confirm that other nodes report `banned`.
4. Configure the same Operator URL on every DataKit and complete a controlled update. Starting when the last DataWay/Kodo process allowed to campaign exits, wait for the old provider lease to expire. Verify that every instance selects Operator and that there is only one Collection Leader.
5. Remove the temporary whitelist, complete another update to restore other nodes' candidacy, and verify election status.

Rollback is symmetric: pin the whitelist on every instance to the current Operator Leader and complete the update, then clear the Operator URL on every instance and restart. Wait for the old Operator lease to expire and verify that the DataWay/Kodo Leader is stable before removing the temporary whitelist and updating again.

If candidacy cannot be pinned, pause election-enabled collection and wait for the old provider lease to expire before enabling the new provider, preventing overlapping collection across election domains.

<!-- markdownlint-disable MD013 -->
### Election Class Collector's Global Tag Settings {#global-tags}
<!-- markdownlint-enable MD013 -->

<!-- markdownlint-disable MD046 -->
=== "`datakit.conf`"

    When election is enabled in `conf.d/datakit.conf`, election-enabled collectors try to append the global tags configured in `[election.tags]`:
    
    ```toml
    [election]
      enable_namespace_tag = false # Set to true to add the election_namespace tag
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

    Use `ENV_GLOBAL_ELECTION_TAGS` for global tags on election-enabled collectors. Set `ENV_ENABLE_ELECTION_NAMESPACE_TAG` to a non-empty value to add the `election_namespace` tag. See [election environment variables](datakit-daemonset-deploy.md#env-elect) for details.
<!-- markdownlint-enable MD046 -->

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

### `host` Field Problem {#host}

For objects collected by collectors participating in elections, such as MySQL, because the DataKit collecting their data may change (election rotation occurs), by default, the data collected by such collectors will not take the tag `host` to avoid timeline growth. We recommend adding an additional `tags` field to the MySQL collector configuration:

```toml
[inputs.mysql.tags]
  host = "real-mysql-instance-name"
```

This way, the `host` field configured in tags will continue to be used when the DataKit has an election rotation.
