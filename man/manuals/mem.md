{{.CSS}}
# 内存
---

- DataKit 版本：{{.Version}}
- 操作系统支持：`{{.AvailableArchs}}`

mem 采集器用于收集系统内存信息，一些通用的指标如主机总内存、用的内存、已使用的内存等  

## 前置条件

暂无

## 配置

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}}
```

配置好后，重启 DataKit 即可。

支持以环境变量的方式修改配置参数（只在 DataKit 以 K8s daemonset 方式运行时生效，主机部署的 DataKit 不支持此功能）：

| 环境变量名           | 对应的配置参数项 | 参数示例                                                     |
| :---                 | ---              | ---                                                          |
| `ENV_INPUT_MEM_TAGS` | `tags`           | `tag1=value1,tag2=value2` 如果配置文件中有同名 tag，会覆盖它 |
| `ENV_INPUT_MEM_INTERVAL` | `interval` | `10s` |

## 视图预览
内存性能指标展示，包括内存使用率，内存大小，缓存，缓冲等

![image](imgs/input-mem-1.png)

## 版本支持

操作系统支持：Linux / Windows / Mac

## 前置条件

- 服务器 <[安装 Datakit](../datakit/datakit-install.md)>

## 安装配置

说明：示例 Linux 版本为：CentOS Linux release 7.8.2003 (Core)，Windows 版本请修改对应的配置文件

### 部署实施

(Linux / Windows 环境相同)

#### 指标采集 (默认)

1、Mem 数据采集默认开启，对应配置文件 /usr/local/datakit/conf.d/host/mem.conf

参数说明

- interval：数据采集频率

```
[[inputs.mem]]
  interval = '10s'
```

2、Mem 指标采集验证  /usr/local/datakit/datakit -M |egrep "最近采集|mem"

![image](imgs/input-mem-2.png)

指标预览

![image](imgs/input-mem-3.png)

#### 插件标签 (非必选)

参数说明

- 该配置为自定义标签，可以填写任意 key-value 值
- 以下示例配置完成后，所有 mem 指标都会带有 app = oa 的标签，可以进行快速查询
- 相关文档 <[DataFlux Tag 应用最佳实践](../best-practices/guance-skill/tag.md)>

```
# 示例
[inputs.mem.tags]
   app = "oa"
```

重启 Datakit

```
systemctl restart datakit
```

## 场景视图

<场景 - 新建仪表板 - 内置模板库 - Memory>

## 异常检测

<监控 - 模板新建 - 主机检测库>

## 指标详解

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

```bash
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


## 常见问题排查

<[无数据上报排查](why-no-data.md)>

## 进一步阅读

<[主机可观测最佳实践](../best-practices/integrations/host.md)>

