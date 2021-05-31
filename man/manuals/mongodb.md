{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# 简介

MongoDb 数据库，Collection， MongoDb 数据库集群运行状态数据采集。

## 前置条件

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
