---
title     : 'MongoDB'
summary   : '采集 MongoDB 的指标数据'
tags:
  - '数据库'
__int_icon      : 'icon/mongodb'
dashboard :
  - desc  : 'MongoDB 监控视图'
    path  : 'dashboard/zh/mongodb'
monitor   :
  - desc  : 'MongoDB 监控器'
    path  : 'monitor/zh/mongodb'
---


{{.AvailableArchs}}

---

MongoDB 数据库，Collection， MongoDB 数据库集群运行状态数据采集。

## 配置 {#config}

### 前置条件 {#requirements}

- 已测试的版本：
    - [x] 6.0
    - [x] 5.0
    - [x] 4.0
    - [x] 3.0
    - [x] 2.8.0

- 开发使用 MongoDB 版本 `4.4.5`;
- 编写配置文件在对应目录下然后启动 DataKit 即可完成配置；
- 使用 TLS 进行安全连接请在配置文件中配置 `## TLS connection config` 下响应证书文件路径与配置；
- 如果 MongoDB 启动了访问控制那么需要配置必须的用户权限用于建立授权连接：

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

>更多权限说明可参见官方文档 [Built-In Roles](https://www.mongodb.com/docs/manual/reference/built-in-roles/){:target="_blank"}。

执行完上述命令后将创建的「用户名」和「密码」填入 Datakit 的配置文件 `conf.d/db/mongodb.conf` 中。

### 采集器配置 {#input-config}

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    目前可以通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting)来开启采集器。
<!-- markdownlint-enable -->

### TLS 配置 (self-signed) {#tls}

使用 `openssl` 生成证书文件用于 MongoDB TLS 配置，用于开启服务端加密和客户端认证。

- 配置 TLS 证书

安装 `openssl` 运行以下命令：

```shell
sudo apt install openssl -y
```

- 配置 MongoDB 服务端加密

使用 `openssl` 生成证书级密钥文件，运行以下命令并按照命令提示符输入相应验证块信息：

```shell
sudo openssl req -x509 -newkey rsa:<bits> -days <days> -keyout <mongod.key.pem> -out <mongod.cert.pem> -nodes
```

- `bits`: rsa 密钥位数，例如 2048
- `days`: expired 日期
- `mongod.key.pem`: 密钥文件
- `mongod.cert.pem`: CA 证书文件

运行上面的命令后生成 `cert.pem` 文件和 `key.pem` 文件，我们需要合并两个文件内的 `block` 运行以下命令：

```shell
sudo bash -c "cat mongod.cert.pem mongod.key.pem >>mongod.pem"
```

合并后配置 `/etc/mongod.config` 文件中的 TLS 子项

```yaml
# TLS config
net:
  tls:
    mode: requireTLS
    certificateKeyFile: </etc/ssl/mongod.pem>
```

使用配置文件启动 MongoDB 运行以下命令：

```shell
mongod --config /etc/mongod.conf
```

使用命令行启动 MongoDB 运行一下命令：

```shell
mongod --tlsMode requireTLS --tlsCertificateKeyFile </etc/ssl/mongod.pem> --dbpath <.db/mongodb>
```

复制 `mongod.cert.pem` 为 `mongo.cert.pem` 到 MongoDB 客户端并启用 TLS：

```shell
mongo --tls --host <mongod_url> --tlsCAFile </etc/ssl/mongo.cert.pem>
```

- 配置 MongoDB 客户端认证

使用 `openssl` 生成证书级密钥文件，运行以下命令：

```shell
sudo openssl req -x509 -newkey rsa:<bits> -days <days> -keyout <mongod.key.pem> -out <mongod.cert.pem> -nodes
```

- `bits`: rsa 密钥位数，例如 2048
- `days`: expired 日期
- `mongo.key.pem`: 密钥文件
- `mongo.cert.pem`: CA 证书文件

合并 `mongod.cert.pem` 和 `mongod.key.pem` 文件中的 block 运行以下命令：

```shell
sudo bash -c "cat mongod.cert.pem mongod.key.pem >>mongod.pem"
```

复制 `mongod.cert.pem` 文件到 MongoDB 服务端然后配置 `/etc/mongod.config` 文件中的 TLS 项：

```yaml
# Tls config
net:
  tls:
    mode: requireTLS
    certificateKeyFile: </etc/ssl/mongod.pem>
    CAFile: </etc/ssl/mongod.cert.pem>
```

启动 MongoDB 运行以下命令：

```shell
mongod --config /etc/mongod.conf
```

复制 `mongod.cert.pem` 为 `mongo.cert.pem` ，复制 `mongod.pem` 为 `mongo.pem` 到 MongoDB 客户端并启用 TLS：

```shell
mongo --tls --host <mongod_url> --tlsCAFile </etc/ssl/mongo.cert.pem> --tlsCertificateKeyFile </etc/ssl/mongo.pem>
```

> 注意：使用自签名证书时，`mongodb.conf` 配置中 `insecure_skip_verify` 必须是 `true`

## 指标 {#metric}

以下所有数据采集，默认会追加全局选举 tag，也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

```toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "metric"}}

### `{{$m.Name}}`

- 说明

{{$m.Desc}}

- 标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}
{{ end }}

## 自定义对象 {#object}

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "custom_object"}}

### `{{$m.Name}}`

{{$m.Desc}}

- 标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}
{{end}}

{{ end }}

## 日志采集 {#logging}

去注释配置文件中 `# enable_mongod_log = false` 然后将 `false` 改为 `true`，其他关于 mongod log 配置选项在 `[inputs.mongodb.log]` 中，注释掉的配置极为默认配置，如果路径对应正确将无需任何配置启动 Datakit 后将会看到指标名为 `mongod_log` 的采集指标集。

日志原始数据 sample

```log
{"t":{"$date":"2021-06-03T09:12:19.977+00:00"},"s":"I",  "c":"STORAGE",  "id":22430,   "ctx":"WTCheckpointThread","msg":"WiredTiger message","attr":{"message":"[1622711539:977142][1:0x7f1b9f159700], WT_SESSION.checkpoint: [WT_VERB_CHECKPOINT_PROGRESS] saving checkpoint snapshot min: 653, snapshot max: 653 snapshot count: 0, oldest timestamp: (0, 0) , meta checkpoint timestamp: (0, 0)"}}
```

日志切割字段

| 字段名    | 字段值                        | 说明                                                           |
| --------- | ----------------------------- | -------------------------------------------------------------- |
| message   |                               | Log raw data                                                   |
| component | STORAGE                       | The full component string of the log message                   |
| context   | WTCheckpointThread            | The name of the thread issuing the log statement               |
| msg       | WiredTiger message            | The raw log output message as passed from the server or driver |
| status    | I                             | The short severity code of the log message                     |
| time      | 2021-06-03T09:12:19.977+00:00 | Timestamp                                                      |
