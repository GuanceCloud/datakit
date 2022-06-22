{{.CSS}}
# DataKit 自身指标
---

- DataKit 版本：{{.Version}}
- 操作系统支持：`{{.AvailableArchs}}`

self 采集器用于 DataKit 自身基本信息的采集，包括运行环境信息、CPU、内存占用情况等。

## 视图预览
Datakit 性能指标展示，包括 CPU 使用率，内存信息，运行时间，日志记录等

![image.png](../imgs/self-1.png)


## 前置条件

暂无

## 安装配置

### 指标采集

#### 配置

self 采集器会自动运行，无需配置，且无法关闭。

#### 指标

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "metric"}}

#### `{{$m.Name}}`

{{$m.Desc}}

- 标签

{{$m.TagsMarkdownTable}}

- 字段列表

{{$m.FieldsMarkdownTable}}
{{end}}

{{ end }}


### 日志采集 (默认)
Datakit 日志采集默认开启，主配置文件 /usr/local/datakit/conf.d/datakit.conf 默认路径
```
log = "/var/log/datakit/log"
gin_log = "/var/log/datakit/gin.log"
```
日志预览<br />![image.png](../imgs/self-2.png)


## 场景视图
<场景 - 新建仪表板 - 内置模板库 - Datakit>

## 常见问题排查
<[无数据上报排查](why-no-data.md)>

## 延申阅读

- [主机采集器](hostobject.md)
