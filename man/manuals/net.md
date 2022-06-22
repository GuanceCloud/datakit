{{.CSS}}
# Net
---

- DataKit 版本：{{.Version}}
- 操作系统支持：`{{.AvailableArchs}}`

net 采集器用于采集主机网络信息，如各网络接口的流量信息等。对于 Linux 将采集系统范围 TCP 和 UDP 统计信息。

## 视图预览
网络性能指标展示，包括网络出入口流量，网络出入口数据包等

![image.png](../imgs/net-1.png)


## 前置条件

暂无

### 指标采集 (默认)

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}}
```

配置好后，重启 DataKit 即可。

支持以环境变量的方式修改配置参数（只在 DataKit 以 K8s daemonset 方式运行时生效，主机部署的 DataKit 不支持此功能）：

| 环境变量名                                | 对应的配置参数项            | 参数示例                                                     |
| :---                                      | ---                         | ---                                                          |
| `ENV_INPUT_NET_IGNORE_PROTOCOL_STATS`     | `ignore_protocol_stats`     | `true`/`false`                                               |
| `ENV_INPUT_NET_ENABLE_VIRTUAL_INTERFACES` | `enable_virtual_interfaces` | `true`/`false`                                               |
| `ENV_INPUT_NET_TAGS`                      | `tags`                      | `tag1=value1,tag2=value2` 如果配置文件中有同名 tag，会覆盖它 |
| `ENV_INPUT_NET_INTERVAL` | `interval` | `10s` |
| `ENV_INPUT_NET_INTERFACES` | `interfaces` | `'''eth[\w-]+''', '''lo'''` 以英文逗号隔开 |


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

-  标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}

## 场景视图
<场景 - 新建仪表板 - 内置模板库 - Net>
## 异常检测
<监控 - 模板新建 - 主机检测库>

## 常见问题排查
- [无数据上报排查](why-no-data.md)
## 进一步阅读
- [主机可观测最佳实践](/best-practices/integrations/host.md)
- [eBPF 数据采集](ebpf.md)
- [DataFlux Tag 应用最佳实践](/best-practices/guance-skill/tag.md)
