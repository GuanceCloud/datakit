{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# {{.InputName}}

采集 NSQ 运行数据并以指标的方式上报到 DataFlux 中。

## 前置条件

- 已安装 NSQ（[NSQ 官方网址](https://nsq.io/)）

- NSQ 版本 >= 1.0.0

## 配置

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}} 
```

NSQ 采集器提供两种配置方式，分别为 `lookupd` 和 `nsqd`，具体说明如下：

- `lookupd`：配置 NSQ 集群的 `lookupd` 地址，采集器会自动发现 NSQ Server 并采集数据，扩展性更佳
- `nsqd`：配置固定的 NSQD 地址列表，采集器只会采集该列表的 NSQ Server 数据

以上两种配置方式是互斥的，`lookupd` 优先级更高，推荐使用 `lookupd` 配置方式。

配置好后，重启 DataKit 即可。

此 input 支持选举功能，[关于选举](election)。

## 指标集

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

{{$m.Desc}}

-  标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }} 
