---
title     : 'Doris'
summary   : '采集 Doris 的指标数据'
tags:
  - '数据库'
__int_icon      : 'icon/doris'
dashboard :
  - desc  : 'Doris'
    path  : 'dashboard/zh/doris'
monitor   :
  - desc  : 'Doris'
    path  : 'monitor/zh/doris'
---

{{.AvailableArchs}}

---

Doris 采集器通过 FE Query Port 的 MySQL 协议采集 Doris 数据库对象和自定义 SQL 指标。
如需采集完整 FE/BE Prometheus 指标，可继续使用配置样例中的 Prometheus 采集器配置。

## 配置 {#config}

已测试的版本：

- [x] 2.0.0

### 前置条件 {#requirements}

使用 Doris 管理账号连接 FE Query Port，创建 `datakit` 采集账号：

```sql
CREATE USER 'datakit'@'%' IDENTIFIED BY '123456';
GRANT NODE_PRIV ON *.*.* TO 'datakit'@'%';
```

使用 `datakit` 账号验证 FE SQL 连接：

```shell
mysql -h <FE_HOST> -P 9030 -u datakit -p
```

连接后可执行以下 SQL 验证对象采集所需信息：

```sql
SHOW FRONTENDS;
SHOW BACKENDS;
```

如配置自定义查询，需确保 `datakit` 账号具备对应表或视图的查询权限。

### 采集器配置 {#input-config}

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/samples` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    可通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting) 或 [配置 ENV_DATAKIT_INPUTS](../datakit/datakit-daemonset-deploy.md#env-setting) 开启采集器。

<!-- markdownlint-enable -->

### 自定义查询 {#custom-query}

可通过 `[[inputs.doris.custom_queries]]` 配置自定义 SQL，将查询结果上报为指标。

```toml
[[inputs.doris.custom_queries]]
  sql = '''
    SELECT
      TABLE_SCHEMA AS table_schema,
      COUNT(*) AS table_count
    FROM information_schema.tables
    GROUP BY TABLE_SCHEMA
  '''
  metric = "doris_custom"
  tags = ["table_schema"]
  fields = ["table_count"]
  interval = "10s"
```

- `metric`：上报的指标集名称。
- `tags`：查询结果中作为标签的列名。
- `fields`：查询结果中作为字段的列名，字段值需要为数值类型。
- `interval`：该查询的采集周期；未配置时使用 `inputs.doris.interval`。

## 指标 {#metric}

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "metric"}}

### `{{$m.Name}}`

{{$m.MarkdownTable}}

{{ end }}
{{ end }}

## 对象 {#object}

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "object"}}

### `{{$m.Name}}`

{{$m.MarkdownTable}}

{{ end }}
{{ end }}
