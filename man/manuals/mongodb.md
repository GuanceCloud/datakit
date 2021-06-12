{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# 简介

MongoDb 数据库，Collection， MongoDb 数据库集群运行状态数据采集。

## 前置条件

- 开发使用 _MongoDB_ 版本 4.4.5
- 编写配置文件在对应目录下然后启动 DataKit 即可完成配置。
- 使用 TLS 进行安全连接需要先将配置文件中`enable_tls = true`值置 true，然后配置`inputs.mongodb.tlsconf`中指定的证书文件路径。
- 如果 MongoDb 启动了访问控制那么需要配置必须的用户权限用于建立授权连接。例如：

```command
> db.grantRolesToUser("user", [{role: "read", actions: "find", db: "local"}])
```

## 配置

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}}
```

配置好后，重启 DataKit 即可。

## TLS config (self-signed)

使用 _openssl_ 生成证书文件用于 _MongoDB TLS_ 配置，用于开启服务端加密和客户端认证。

### 预配置

安装 _openssl_ 运行以下命令:

```command
sudo apt install openssl -y
```

### 配置 MongoDB 服务端加密

使用 _openssl_ 生成证书级密钥文件，运行以下命令:

```command
sudo openssl req -x509 -newkey rsa:<bits> -days <days> -keyout <mongod.key.pem> -out <mongod.cert.pem> -nodes -subj '/CN=<mongod_url>'
```

- bits: rsa 密钥位数，例如 2048
- days: expired 日期
- mongod.key.pem: 密钥文件
- mongod.cert.pem: CA 证书文件
- mongod_url: MongoDB server url

运行上面的命令后生成 _cert.pem_ 文件和 _key.pem_ 文件，我们需要合并两个文件内的 _block_ 运行以下命令:

```command
sudo bash -c "cat mongod.cert.pem mongod.key.pem >>mongod.pem"
```

合并后配置 _/etc/mongod.config_ 文件中的 TLS 子项

```config
# TLS config
net:
  tls:
    mode: requireTLS
    certificateKeyFile: /etc/ssl/mongod.pem
```

使用新的配置启动启动 _MongoDB_ 运行以下命令:

```command
sudo mongod --config /etc/mongod.conf
```

复制 _mongod.cert.pem_ 文件到 _MongoDB_ 客户端测试使用 TLS 连接服务端 运行以下命令:

```command
mongo --tls --host <mongod_url> --tlsCAFile /etc/ssl/certs/mongod.cert.pem
```

### 配置 MongoDB 客户端认证

使用 _openssl_ 生成证书级密钥文件，运行以下命令:

```command
sudo openssl req -x509 -newkey rsa:<bits> -days <days> -keyout <mongo.key.pem> -out <mongo.cert.pem> -nodes
```

- bits: rsa 密钥位数，例如 2048
- days: expired 日期
- mongo.key.pem: 密钥文件
- mongo.cert.pem: CA 证书文件

复制 _mongo.cert.pem_ 文件到 _MongoDB_ 服务端然后配置 _/etc/mongod.config_ 文件中的 TLS 子项

```config
# Tls config
net:
  tls:
    mode: requireTLS
    certificateKeyFile: /etc/ssl/mongod.pem
    CAFile: /etc/ssl/mongo.cert.pem
```

启动 _MongoDB_ 运行以下命令:

```command
sudo mongod --config /etc/mongod.conf
```

合并 _mongo.cert.pem_ 和 _mongo.key.pem_ 文件中的 _block_ 运行以下命令:

```command
sudo bash -c "cat mongo.cert.pem mongo.key.pem >>mongo.pem"
```

启动 _MongoDB_ 客户端并使用 TLS 客户端认证 运行以下命令:

```command
mongo --tls --host <mongod_url> --tlsCAFile /etc/ssl/certs/mongod.cert.pem --tlsCertificateKeyFile /etc/ssl/certs/mongo.pem
```

!!!important
使用自签名证书时 _/your/home/path/datakit/conf.d/mongodb_ 中的配置项 _\[inputs.mongodb.tlsconf\]_ 中 _insecure_skip_verify_ 必须是 _true_

## 指标集

以下所有指标集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

```toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

- 标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}

## mongod log 采集

### 基本配置

去注释配置文件中 `# enable_mongod_log = false` 然后将 `false` 改为 `true`，其他关于 mongod log 配置选项在 `[inputs.mongodb.log]` 中，注释掉的配置极为默认配置，如果路径对应正确将无需任何配置启动 Datakit 后将会看到指标名为 `mongod_log` 的采集指标集。

### 日志原始数据 sample

```
{"t":{"$date":"2021-06-03T09:12:19.977+00:00"},"s":"I",  "c":"STORAGE",  "id":22430,   "ctx":"WTCheckpointThread","msg":"WiredTiger message","attr":{"message":"[1622711539:977142][1:0x7f1b9f159700], WT_SESSION.checkpoint: [WT_VERB_CHECKPOINT_PROGRESS] saving checkpoint snapshot min: 653, snapshot max: 653 snapshot count: 0, oldest timestamp: (0, 0) , meta checkpoint timestamp: (0, 0)"}}
```

### 日志切割字段

| 字段名    | 字段值                        | 说明                                                           |
| --------- | ----------------------------- | -------------------------------------------------------------- |
| message   |                               | Log raw data                                                   |
| component | STORAGE                       | The full component string of the log message                   |
| context   | WTCheckpointThread            | The name of the thread issuing the log statement               |
| msg       | WiredTiger message            | The raw log output message as passed from the server or driver |
| status    | I                             | The short severity code of the log message                     |
| time      | 2021-06-03T09:12:19.977+00:00 | Timestamp                                                      |
