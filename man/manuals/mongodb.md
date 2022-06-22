{{.CSS}}
# MongoDB
---

- DataKit 版本：{{.Version}}
- 操作系统支持：`{{.AvailableArchs}}`

MongoDb 数据库，Collection， MongoDb 数据库集群运行状态数据采集。

## 前置条件

- 开发使用 MongoDB 版本 4.4.5
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

使用 openssl 生成证书文件用于 MongoDB TLS 配置，用于开启服务端加密和客户端认证。

### 预配置

安装 openssl 运行以下命令:

```command
sudo apt install openssl -y
```

### 配置 MongoDB 服务端加密

使用 openssl 生成证书级密钥文件，运行以下命令:

```command
sudo openssl req -x509 -newkey rsa:<bits> -days <days> -keyout <mongod.key.pem> -out <mongod.cert.pem> -nodes -subj '/CN=<mongod_url>'
```

- `bits`: rsa 密钥位数，例如 2048
- `days`: expired 日期
- `mongod.key.pem`: 密钥文件
- `mongod.cert.pem`: CA 证书文件
- `mongod_url`: MongoDB server url

运行上面的命令后生成 `cert.pem` 文件和 `key.pem` 文件，我们需要合并两个文件内的 `block` 运行以下命令:

```command
sudo bash -c "cat mongod.cert.pem mongod.key.pem >>mongod.pem"
```

合并后配置 /etc/mongod.config 文件中的 TLS 子项

```yaml
# TLS config
net:
  tls:
    mode: requireTLS
    certificateKeyFile: /etc/ssl/mongod.pem
```

使用新的配置启动启动 MongoDB 运行以下命令:

```command
sudo mongod --config /etc/mongod.conf
```

复制 mongod.cert.pem 文件到 MongoDB 客户端测试使用 TLS 连接服务端 运行以下命令:

```command
mongo --tls --host <mongod_url> --tlsCAFile /etc/ssl/certs/mongod.cert.pem
```

### 配置 MongoDB 客户端认证

使用 openssl 生成证书级密钥文件，运行以下命令:

```command
sudo openssl req -x509 -newkey rsa:<bits> -days <days> -keyout <mongo.key.pem> -out <mongo.cert.pem> -nodes
```

- bits: rsa 密钥位数，例如 2048
- days: expired 日期
- mongo.key.pem: 密钥文件
- mongo.cert.pem: CA 证书文件

复制 mongo.cert.pem 文件到 MongoDB 服务端然后配置 /etc/mongod.config 文件中的 TLS 子项

```yaml
# Tls config
net:
  tls:
    mode: requireTLS
    certificateKeyFile: /etc/ssl/mongod.pem
    CAFile: /etc/ssl/mongo.cert.pem
```

启动 MongoDB 运行以下命令:

```command
sudo mongod --config /etc/mongod.conf
```

合并 mongo.cert.pem 和 mongo.key.pem 文件中的 block 运行以下命令:

```command
sudo bash -c "cat mongo.cert.pem mongo.key.pem >>mongo.pem"
```

启动 MongoDB 客户端并使用 TLS 客户端认证 运行以下命令:

```command
mongo --tls --host <mongod_url> --tlsCAFile /etc/ssl/certs/mongod.cert.pem --tlsCertificateKeyFile /etc/ssl/certs/mongo.pem
```

> 使用自签名证书时 mongodb.conf 中的配置项 `[inputs.mongodb.tlsconf]` 中 `insecure_skip_verify` 必须是 `true`

## 视图预览
MongoDB 性能指标展示：包括每秒查询操作，文档操作，TTL索引，游标，队列信息等

![image](imgs/input-mongodb-1.png)

![image](imgs/input-mongodb-2.png)

## 版本支持

操作系统支持：Linux / Windows / Mac

## 前置条件

- 服务器 <[安装 Datakit](../datakit/datakit-install.md)>
- MongoDB 用户授权 (使用超级账户执行)

3.4 + 版本
```
> use admin
> db.createUser(
  {
    user: "user",
    pwd: "password",
    roles: [ { role: "clusterMonitor", db: "admin" } ]
  }
)
```
3.4 - 版本
```
> use admin
> db.createUser(
  {
    user: "user",
    pwd: "password"
  }
> db.grantRolesToUser("user", [{role: "read", actions: "find", db: "local"}])
)
```
## 安装配置
说明：示例 MongoDB 版本为 Linux 环境 db version v4.0.22，Windows 版本请修改对应的配置文件
### 部署实施
#### 指标采集 (必选)

1、开启 Datakit MongoDB 插件，复制 sample 文件
```
cd /usr/local/datakit/conf.d/db/mongodb
cp mongodb.conf.sample mongodb.conf
```

2、修改 mongodb 配置文件

```
vi mongodb.conf
```
参数说明

- interval：数据采集频率
- servers：服务连接地址 (user:passwd)
- gather_replica_set_stats：是否开启副本集状态采集
- gather_cluster_stats：是否开启集群状态采集
- gather_per_db_stats：是否开启每个数据库状态采集
- gather_per_col_stats：是否开启数据库表状态采集
- col_stats_dbs：数据库表筛选 (如果为空，代表都选择)
- gather_top_stat：命令行状态采集
- enable_tls：是否开启 tls 验证
```
[[inputs.mongodb]]
  interval = "10s"
  servers = ["mongodb://user:passwd@127.0.0.1:27017"]
  # gather_replica_set_stats = false
  # gather_cluster_stats = false
  # gather_per_db_stats = true
  # gather_per_col_stats = true
  # col_stats_dbs = []
  # gather_top_stat = true
  # enable_tls = false
```

3、MongoDB 指标采集验证  /usr/local/datakit/datakit -M |egrep "最近采集|mongodb"

![image](imgs/input-mongodb-3.png)

4、重启 Datakit (如果需要开启日志，请配置日志采集再重启)

```
systemctl restart datakit
```

指标预览

![image](imgs/input-mongodb-4.png)

#### 安全认证 (非必选)

参数说明

- ca_certs：ca 证书
- cert：openssl 证书
- cert_key：openssl 私钥
- insecure_skip_verify：是否忽略 tls 认证
- server_name：服务名称 (自定义)

```
    [inputs.mongodb.tlsconf]
    # ca_certs = ["/etc/ssl/certs/mongod.cert.pem"]
    # cert = "/etc/ssl/certs/mongo.cert.pem"
    # cert_key = "/etc/ssl/certs/mongo.key.pem"
    # insecure_skip_verify = true
    # server_name = ""
```

#### 日志插件 (非必选)

参数说明

- files：日志文件路径
- pipeline：日志切割文件(内置)，实际文件路径 /usr/local/datakit/pipeline/mongod.p
- 相关文档 <[DataFlux pipeline 文本数据处理](/datakit/pipeline.md)>

```
[inputs.mongodb.log]
files = ["/var/log/mongodb/mongod.log"]
pipeline = "mongod.p"
```

重启 Datakit (如果需要开启自定义标签，请配置插件标签再重启)

```
systemctl restart datakit
```

MongoDB 日志采集验证  /usr/local/datakit/datakit -M |egrep "最近采集|mongodb_log"

![image](imgs/input-mongodb-5.png)

#### 插件标签 (非必选)

参数说明

- 该配置为自定义标签，可以填写任意 key-value 值
- 以下示例配置完成后，所有 mongodb 指标都会带有 app = oa 的标签，可以进行快速查询
- 相关文档 <[DataFlux Tag 应用最佳实践](../best-practices/guance-skill/tag.md)>

```
# 示例
[inputs.mongodb.tags]
   app = "oa"
```

重启 Datakit

```
systemctl restart datakit
```

## 场景视图

<场景 - 新建仪表板 - 内置模板库 - MongoDB 监控视图>

## 异常检测

<监控 - 模板新建 - MongoDB 检测库>

## 指标详解

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

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


## 常见问题排查

<[无数据上报排查](why-no-data.md)>

## 进一步阅读

<[全面认识 MongoDB](https://baijiahao.baidu.com/s?id=1709426775094926497&wfr=spider&for=pc)>
