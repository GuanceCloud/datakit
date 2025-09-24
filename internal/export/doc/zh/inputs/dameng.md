---
title     : '达梦数据库'
summary   : '采集达梦数据库的指标数据'
tags:
  - '数据库'
__int_icon      : 'icon/dameng'
dashboard :
  - desc  : 'dameng'
    path  : 'dashboard/zh/dameng'
monitor   :
  - desc  : '暂无'
    path  : '-'
---

{{.AvailableArchs}}

---

达梦数据库采集器可以从达梦数据库实例中采集实例运行状态指标。

## 配置 {#config}

### 采集器配置 {#input-config}

<!-- markdownlint-disable MD046 -->
=== "主机部署"

    进入 DataKit 安装目录下的 `conf.d/samples` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    目前可以通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting)来开启采集器。
<!-- markdownlint-enable -->

## 指标 {#metric}

以下所有数据采集，默认会追加全局选举 tag，也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "metric"}}

### `{{$m.Name}}`

{{$m.MarkdownTable}}

{{ end }}
{{ end }}

## 日志 {#logging}

- 在达梦数据库运行过程中，会将一些关键信息记录到安装目录下一级 log 目录下的名称为 `dm_<instance-name>_YYYYMM.log`
  的日志文件中。比如：`dm_DMSERVER_202507.log`

- 达梦数据库采集器默认是未开启日志采集功能，可在 *conf.d/samples/{{.InputName}}.conf* 中 将 `files` 打开，并写入达梦日志文件的绝对路径。比如：

```toml
[[inputs.dameng]]

  ...

  [inputs.dameng.log]
    files = ["/home/dmdba/dmdbms/log/dm_DMSERVER_202507.log"]
```

开启日志采集后，默认会产生日志来源（`source`）为 `dameng` 的日志。

> 注意：日志采集仅支持已安装 DataKit 主机上的日志。

### 日志 Pipeline 切割 {#pipeline}

原始日志为

``` log
2025-07-03 10:16:20.659 [INFO] database P0000001485 T0000000000000001485  INI parameter ROLLSEG_POOLS changed, the original value 19, new value 1
```

切割后的字段说明：

| 字段名             | 字段值                                            | 说明                                                        |
| ---                | ---                                              | ---                                                         |
| `msg` | `database P0000001485 T0000000000000001485  …… new value 1`   | 日志内容                                  |
| `status`           | `INFO`                                           | 当前日志的级别（ERROR,FATAL,WARNING,INFO）|
| `time`             | `1751537780`                                     | 日志产生时间                                                                       |
