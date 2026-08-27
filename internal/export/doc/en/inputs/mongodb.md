---
title     : 'MongoDB'
summary   : 'Collect MongoDB metrics, objects, queries, and logs'
tags:
  - 'DATABASE'
__int_icon      : 'icon/mongodb'
dashboard :
  - desc  : 'Mongodb'
    path  : 'dashboard/en/mongodb'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---


{{.AvailableArchs}}

---

MongoDb database, Collection, MongoDb database cluster running status data Collection.

## Config {#config}

### Preconditions {#requirements}

- Already tested version:
    - [x] 6.0
    - [x] 5.0
    - [x] 4.0
    - [x] 3.0
    - [x] 2.8.0

- Developed and used MongoDB version `4.4.5`;
- Write the configuration file in the corresponding directory and then start DataKit to complete the configuration;
- For secure connections using TLS, please configure the response certificate file path and configuration under `## TLS connection config` in the configuration file;
- If MongoDB has access control enabled, you need to configure the necessary user rights to establish an authorized connection:

```sh
# Run MongoDB shell.
$ mongo

# Authenticate as the admin/root user.
> use admin
> db.auth("<admin OR root>", "<YOUR_MONGODB_ADMIN_PASSWORD>")

# Create the user for the DataKit.
> db.createUser({
  "user": "datakit",
  "pwd": "<YOUR_COLLECT_PASSWORD>",
  "roles": [
    { role: "read", db: "admin" },
    { role: "clusterMonitor", db: "admin" },
    { role: "backup", db: "admin" },
    { role: "read", db: "local" }
  ]
})
```

>More authorization information can refer to official documentation [Built-In Roles](https://www.mongodb.com/docs/manual/reference/built-in-roles/){:target="_blank"}。

After done with commands above, filling the `user` and `pwd` to DataKit configuration file `conf.d/db/mongodb.conf`.

### Collector Configuration {#input-config}

<!-- markdownlint-disable MD046 -->
=== "Host Installation"

    Go to the `conf.d/samples` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    Once configured, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    The collector can now be turned on by [ConfigMap Injection Collector Configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting).
<!-- markdownlint-enable MD046 -->

### Database Monitoring {#dbm}

MongoDB Database Monitoring (DBM) contains Query Metrics, Activity, and Slow Operations collection. `dbm.enabled` is the parent switch, and `databases` is the database filter shared by all three collectors. An empty list collects all databases except `admin`.

```toml
[[inputs.mongodb]]
  # ... MongoDB connection settings

  [inputs.mongodb.dbm]
    enabled = true
    databases = []

  [inputs.mongodb.dbm.query_metrics]
    enabled = true
    interval = "60s"
    limit = 10000
    query_text_max_bytes = 512

  [inputs.mongodb.dbm.activity]
    enabled = true
    interval = "10s"

  [inputs.mongodb.dbm.slow_operations]
    enabled = false
    interval = "10s"
    max_operations = 1000
```

Query Metrics and Activity are enabled by default when DBM is enabled. Slow Operations is disabled by default and must be enabled explicitly. All three collectors use the first MongoDB server that initializes successfully.

#### Query Metrics {#dbm-queries}

Query Metrics uses `$queryStats` to collect execution counts, durations, and scan statistics grouped by query shape.

##### Prerequisites {#dbm-queries-requirements}

- MongoDB 8.0 or later is required.
- On self-hosted MongoDB, set `internalQueryStatsRateLimit` to a positive integer. The default value `0` disables Query Stats recording. For example:

```yaml
setParameter:
  internalQueryStatsRateLimit: 100
```

The parameter can also be changed dynamically with `setParameter`. A value of `-1` removes the recording rate limit.

##### Configuration {#dbm-queries-config}

- `enabled`: Query Metrics is enabled by default when DBM is enabled. Set it to `false` to disable Query Metrics independently.
- `interval`: Collection interval. The default is `60s`.
- `limit`: Maximum number of Query Stats entries processed per collection interval. The default is `10000`.
- `query_text_max_bytes`: Maximum number of obfuscated query bytes stored in the `query_text` tag. The default is `512` and the maximum is `1024`. Truncation preserves complete UTF-8 characters and sets `query_truncated` to `true`.

The first collection establishes the delta baseline. Later collections report:

- `mongodb_dbm_metric`: Cumulative values and per-interval deltas prefixed with `delta_`.
- `db_query`: Obfuscated query commands, reported at most once every 24 hours for each `query_signature`.

#### Activity {#dbm-activity}

Activity uses `$currentOp` to collect currently running operations.

- `enabled`: Activity is enabled by default when DBM is enabled. Set it to `false` to disable Activity independently.
- `interval`: Collection interval. The default is `10s`.

Each visible operation produces an obfuscated `mongodb_dbm_activity` log. The same snapshot also produces `mongodb_dbm_operation` metrics: `active_operation_count` counts active operations and `waiting_for_lock_count` counts operations waiting for a lock. These metrics do not represent all MongoDB connections.

#### Slow Operations {#dbm-slow-operations}

Slow Operations incrementally collects completed slow operations from each database's `system.profile` collection. Enable the MongoDB Profiler in each monitored database, for example:

```javascript
db.setProfilingLevel(1, { slowms: 100 })
```

- `enabled`: Disabled by default. Enable it explicitly together with `dbm.enabled`.
- `interval`: Collection interval. The default is `10s`.
- `max_operations`: Maximum number of slow operations reported per collection interval. The default is `1000`.

Each slow operation produces one `mongodb_dbm_slow_query` log. The log timestamp comes from `system.profile.ts`, the complete obfuscated command is stored in the `message` field, and tags such as `operation`, `command_type`, and `plan_summary` can be used to aggregate operation counts and durations.

### TLS config (self-signed) {#tls}

Use OpenSSL to generate a certificate file for MongoDB TLS configuration to enable server-side encryption and client-side authentication.

- Configure TLS certificates

Install OpenSSL and run the following command:

```shell
sudo apt install openssl -y
```

- Configure MongoDB server-side encryption

Use OpenSSL to generate a certificate-level key file, run the following command and enter the corresponding authentication block information at the command prompt:

```shell
sudo openssl req -x509 -newkey rsa:<bits> -days <days> -keyout <mongod.key.pem> -out <mongod.cert.pem> -nodes
```

- `bits`: rsa key digits, for example, 2048
- `days`: expired date
- `mongod.key.pem`: key file
- `mongod.cert.pem`: CA certificate file

Running the above command generates the `cert.pem` file and the `key.pem` file, and we need to merge the `block` inside the two files to run the following command:

```shell
sudo bash -c "cat mongod.cert.pem mongod.key.pem >>mongod.pem"
```

Configure the TLS subentry in the /etc/mongod.config file after merging

```yaml
# TLS config
net:
  tls:
    mode: requireTLS
    certificateKeyFile: </etc/ssl/mongod.pem>
```

Start MongoDB with the configuration file and run the following command:

```shell
mongod --config /etc/mongod.conf
```

Start MongoDB from the command line and run the following command:

```shell
mongod --tlsMode requireTLS --tlsCertificateKeyFile </etc/ssl/mongod.pem> --dbpath <.db/mongodb>
```

Copy mongod.cert.pem as mongo.cert.pem to MongoDB client and enable TLS:

```shell
mongo --tls --host <mongod_url> --tlsCAFile </etc/ssl/mongo.cert.pem>
```

- Configuring MongoDB Client Authentication

Use OpenSSL to generate a certificate-level key file and run the following command:

```shell
sudo openssl req -x509 -newkey rsa:<bits> -days <days> -keyout <mongod.key.pem> -out <mongod.cert.pem> -nodes
```

- `bits`: rsa key digits, for example, 2048
- `days`: expired date
- `mongo.key.pem`: key file
- `mongo.cert.pem`: CA certificate file

Merging the block in the mongod.cert.pem and mongod.key.pem files runs the following command:

```shell
sudo bash -c "cat mongod.cert.pem mongod.key.pem >>mongod.pem"
```

Copy the mongod.cert.pem file to the MongoDB server and configure the TLS entry in the /etc/mongod.config file.

```yaml
# Tls config
net:
  tls:
    mode: requireTLS
    certificateKeyFile: </etc/ssl/mongod.pem>
    CAFile: </etc/ssl/mongod.cert.pem>
```

Start MongoDB and run the following command:

```shell
mongod --config /etc/mongod.conf
```

Copy mongod.cert.pem for mongo.cert.pem; Copy mongod.pem for mongo.pem to MongoDB client and enable TLS:

```shell
mongo --tls --host <mongod_url> --tlsCAFile </etc/ssl/mongo.cert.pem> --tlsCertificateKeyFile </etc/ssl/mongo.pem>
```

**Note:**`insecure_skip_verify` must be `true` in mongodb.conf configuration when using self-signed certificates.

## Metric {#metric}

For all of the following data collections, the global election tags will added automatically, we can add extra tags in `[inputs.{{.InputName}}.tags]` if needed:

```toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}
{{if eq $m.Type "metric"}}

### `{{$m.Name}}`

{{$m.Desc}}

{{$m.MarkdownTable}}

{{ end }}
{{ end }}

## Object {#object}

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "object"}}

### `{{$m.Name}}`

{{$m.Desc}}

{{$m.MarkdownTable}}

{{if eq $m.Name "database"}}

#### `message` Metric Field Structure {#message-struct}

The `message` field contains MongoDB startup/configuration settings collected from the `parsed` section of the `getCmdLineOpts` admin command. It does not include command-line `argv`, connection counters, or database statistics. The exact keys depend on the MongoDB startup options and configuration file. Its basic structure is as follows:

```json
{
  "setting": {
    "net": {
      "port": 27017,
      "bindIp": "127.0.0.1"
    },
    "storage": {
      "dbPath": "/var/lib/mongodb",
      "engine": "wiredTiger"
    },
    "systemLog": {
      "destination": "file",
      "path": "/var/log/mongodb/mongod.log"
    }
  }
}
```

{{end}}
{{end}}

{{ end }}

## Logging {#logging}

The following DBM Activity measurement is collected as logs:

{{ range $i, $m := .Measurements }}
{{if eq $m.Type "logging"}}
### `{{$m.Name}}`

{{$m.Desc}}

{{$m.MarkdownTable}}

{{ end }}
{{ end }}

### Mongod Log Collection {#mongod-logging}

Annotate the configuration file `# enable_mongod_log = false` and change `false` to `true`. Other configuration options for mongod log are in `[inputs.mongodb.log]`, and the commented configuration is very default. If the path correspondence is correct, no configuration is needed. After starting DataKit, you will see a collection measurement named `mongod_log`.

Log raw data sample

```not-set
{"t":{"$date":"2021-06-03T09:12:19.977+00:00"},"s":"I",  "c":"STORAGE",  "id":22430,   "ctx":"WTCheckpointThread","msg":"WiredTiger message","attr":{"message":"[1622711539:977142][1:0x7f1b9f159700], WT_SESSION.checkpoint: [WT_VERB_CHECKPOINT_PROGRESS] saving checkpoint snapshot min: 653, snapshot max: 653 snapshot count: 0, oldest timestamp: (0, 0) , meta checkpoint timestamp: (0, 0)"}}
```

Log cut field

| Field Name | Field Value                   | Description                                                    |
| ---------- | ----------------------------- | -------------------------------------------------------------- |
| message    |                               | Log raw data                                                   |
| component  | STORAGE                       | The full component string of the log message                   |
| context    | WTCheckpointThread            | The name of the thread issuing the log statement               |
| msg        | WiredTiger message            | The raw log output message as passed from the server or driver |
| status     | I                             | The short severity code of the log message                     |
| time       | 2021-06-03T09:12:19.977+00:00 | Timestamp                                                      |
