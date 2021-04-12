{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：{{.AvailableArchs}}

# 简介

该采集器是网络拨测结果数据采集，所有拨测产生的数据，都以行协议方式，通过 `/v1/write/logging` 接口,上报DataFlux平台

## 前置条件

暂无

## 配置

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```python
{{.InputSample}}
```

配置好后，重启 DataKit 即可。

##  http 任务说明 

|  参数名        |   type  | 必选  |          说明          |
|---------------|----------|----|------------------------|
| type          |  string  |  Y | 云拨测类型，可选选项`http`,`tcp`,`dns`  |
| name       |  string   |  Y | 任务名称|
| url       |  string   |  Y | url|
| method     |  string   |  Y | url 请求方法|
| status       |  string   |  Y | 任务状态，可选值`ok`,`stop`|
| frequency       |  string   |  Y | 任务频率|
| advance_options  |  struct   |   | |
| success_when       |  array   | Y  | |


##  HTTP 拨测结果指标集

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

-  标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}
