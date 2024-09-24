---
title     : 'OceanBase'
summary   : '采集 OceanBase 的指标数据'
tags:
  - '数据库'
__int_icon      : 'icon/oceanbase'
dashboard :
  - desc  : 'OceanBase'
    path  : 'dashboard/zh/oceanbase'
monitor   :
  - desc  : '暂无'
    path  : '-'
---


{{.AvailableArchs}}

---

支持通过系统租户采集 OceanBase 监控指标。

已测试的版本：

- [x] OceanBase 3.2.4 企业版

## 配置 {#config}

### 前置条件 {#reqirement}

- 创建监控账号

使用 OceanBase 系统租户账号创建监控账号，并授予以下权限：

```sql
CREATE USER 'datakit'@'localhost' IDENTIFIED BY '<UNIQUEPASSWORD>';

-- MySQL 8.0+ create the datakit user with the caching_sha2_password method
CREATE USER 'datakit'@'localhost' IDENTIFIED WITH caching_sha2_password by '<UNIQUEPASSWORD>';

-- 授权
GRANT SELECT ON *.* TO 'datakit'@'localhost';
```

<!-- markdownlint-disable MD046 -->
???+ attention

    - 如用 `localhost` 时发现采集器有如下报错，需要将上述步骤的 `localhost` 换成 `::1` <br/>
    `Error 1045: Access denied for user 'datakit'@'localhost' (using password: YES)`

    - 以上创建、授权操作，均限定了 `datakit` 这个用户，只能在 OceanBase 主机上（`localhost`）访问。如果需要远程采集，建议将 `localhost` 替换成 `%`（表示 DataKit 可以在任意机器上访问），也可用特定的 DataKit 安装机器地址。
<!-- markdownlint-enable -->

### 采集器配置 {#input-config}

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    目前可以通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting)来开启采集器。

<!-- markdownlint-enable -->

## 慢查询支持 {#slow}

Datakit 可以将执行超过用户自定义时间的 SQL 语句报告给观测云，在日志中显示，来源名是 `oceanbase_log`。

该功能默认情况下是关闭的，用户可以在 OceanBase 的配置文件中将其打开，方法如下：

将 `--slow-query-time` 后面的值从 `0s` 改成用户心中的阈值，最小值 1 毫秒。一般推荐 10 秒。

```conf
  args = [
    ...
    '--slow-query-time' , '10s'                        ,
  ]
```

???+ info "字段说明"
    - `failed_obfuscate`：SQL 脱敏失败的原因。只有在 SQL 脱敏失败才会出现。SQL 脱敏失败后原 SQL 会被上报。
    更多字段解释可以查看[这里](https://www.oceanbase.com/docs/enterprise-oceanbase-database-cn-10000000000376688){:target="_blank"}。

???+ attention "重要信息"
    - 如果值是 `0s` 或空或小于 1 毫秒，则不会开启 OceanBase 采集器的慢查询功能，即默认状态。
    - 没有执行完成的 SQL 语句不会被查询到。

## 指标 {#metric}

以下所有数据采集，默认会追加全局选举 tag，也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "metric" }}

### `{{$m.Name}}`

{{$m.Desc}}

- 标签

{{$m.TagsMarkdownTable}}

- 字段列表

{{$m.FieldsMarkdownTable}}
{{end}}

{{ end }}

## 日志 {#logging}

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "logging"}}

### `{{$m.Name}}`

{{$m.Desc}}

- 标签

{{$m.TagsMarkdownTable}}

- 字段列表

{{$m.FieldsMarkdownTable}}
{{end}}

{{ end }}
