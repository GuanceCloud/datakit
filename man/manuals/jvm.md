- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}

# 简介

`jvm` 采集器 可以从 `JMX` 中采取很多指标，并将指标采集到 `dataflux` ，帮助你监控分析 `java` 运行情况


## 前置条件

- 已安装 DataKit（[DataKit 安装文档](../../../02-datakit采集器/index.md)）
- 已安装或下载 [jolokia](https://jolokia.org/download.html)，datakit 安装目录下 `data` 目录下已经下载好 jolokia 的 jar 包，并通过如下方式开启 Java 应用。 
```
 java -javaagent:/path/to/jolokia-jvm-<version>-agent.jar -jar your_app.jar
```

## 配置

进入 DataKit 安装目录下的 `conf.d/{{.InputName}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```
{{.InputSample}}
```

配置好后，重启 DataKit 即可。

## 指标集

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

-  标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}
