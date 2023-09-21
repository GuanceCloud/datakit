

# MongoDB

---

{{.AvailableArchs}}

---

MongoDb database, Collection, MongoDb database cluster running status data Collection.

## Preconditions {#requirements}

- Already tested version:
    - [x] 6.0
    - [x] 5.0
    - [x] 4.0
    - [x] 3.0

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

# Create the user for the Datakit.
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

After done with commands above, filling the `user` and `pwd` to Datakit configuration file `conf.d/db/mongodb.conf`.

## Configuration {#config}

=== "Host Installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    Once configured, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    The collector can now be turned on by [ConfigMap Injection Collector Configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting).

## TLS config (self-signed) {#tls}

Use openssl to generate a certificate file for MongoDB TLS configuration to enable server-side encryption and client-side authentication.

- Configure TLS certificates

Install openssl and run the following command:

```shell
sudo apt install openssl -y
```

- Configure MongoDB server-side encryption

Use openssl to generate a certificate-level key file, run the following command and enter the corresponding authentication block information at the command prompt:

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

Use openssl to generate a certificate-level key file and run the following command:

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

## Measurements {#measurements}

For all of the following data collections, a global tag named `host` is appended by default (the tag value is the host name of the DataKit), or other tags can be specified in the configuration by `[inputs.mongodb.tags]`:

```toml
 [inputs.mongodb.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

- tag

{{$m.TagsMarkdownTable}}

- metric list

{{$m.FieldsMarkdownTable}}

{{ end }}

## Mongod Log Collection {#logging}

Annotate the configuration file `# enable_mongod_log = false` and change `false` to `true`. Other configuration options for mongod log are in `[inputs.mongodb.log]`, and the commented configuration is very default. If the path correspondence is correct, no configuration is needed. After starting Datakit, you will see a collection measurement named `mongod_log`.

Log raw data sample

```
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
