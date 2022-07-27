{{.CSS}}
# 主机目录
---

- 操作系统支持：{{.AvailableArchs}}

hostdir 采集器用于目录文件的采集，例如文件个数，所有文件大小等。

## 前置条件

暂无

## 配置

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}}
```

配置好后，重启 DataKit 即可。

## 视图预览

Directory 性能指标展示，包括目录数，文件数，目录文件大小等

![image](imgs/input-directory-1.png)

## 版本支持

操作系统支持：Linux / Windows / Mac

## 前置条件

- 服务器 <[安装 Datakit](../datakit/datakit-install.md)>

## 安装配置

说明：示例 Linux 版本为：CentOS Linux release 7.8.2003 (Core)
### 部署实施

(Linux / Windows 环境相同)

#### 指标采集 (必选)

1、开启 Datakit Hostdir 插件，复制 sample 文件
```
cd /usr/local/datakit/conf.d/host/
cp hostdir.conf.sample hostdir.conf
```

2、修改配置文件 hostdir.conf

参数说明

- interval：采集频率 (默认10s)
- dir：目录路径
- exclude_patterns：不包含的文件名后缀 (例如 *.exe、*.so等)
```
[[inputs.hostdir]]
  interval = "10s"
  dir = "/usr/local/datakit/"
  # exclude_patterns = []
```

3、Directory 指标采集验证  /usr/local/datakit/datakit -M |egrep "最近采集|hostdir"

![image](imgs/input-directory-2.png)

指标预览

![image](imgs/input-directory-3.png)

#### 插件标签 (非必选)

参数说明

- 该配置为自定义标签，可以填写任意 key-value 值
- 以下示例配置完成后，所有 hostdir 指标都会带有 app = oa 的标签，可以进行快速查询
- 相关文档 <[DataFlux Tag 应用最佳实践](../best-practices/guance-skill/tag.md)>

```
# 示例
[inputs.hostdir.tags]
   app = "oa"
```

重启 Datakit

```
systemctl restart datakit
```

## 场景视图

<场景 - 新建仪表板 - 内置模板库 - Directory>

## 异常检测

<监控 - 模板新建 - 主机检测库>

## 指标详解

以下所有指标集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

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

## 常见问题排查

<[无数据上报排查](../datakit/why-no-data.md)>

Q：如果想监控多个目录，怎么配置？<br />A：需要填写多个 input 配置。
```
[[inputs.hostdir]]
  interval = "10s"
  dir = "/usr/local/datakit/"
[[inputs.hostdir]]
  interval = "10s"
  dir = "/usr/local/dataflux-func/"

```

## 进一步阅读

<[主机可观测最佳实践](../best-practices/integrations/host.md)>
