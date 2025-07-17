---
title     : 'ElasticSearch'
summary   : '采集 ElasticSearch 的指标数据'
tags:
  - '数据库'
__int_icon      : 'icon/elasticsearch'
dashboard :
  - desc  : 'ElasticSearch'
    path  : 'dashboard/zh/elasticsearch'
monitor   :
  - desc  : 'ElasticSearch'
    path  : 'monitor/zh/elasticsearch'
---


{{.AvailableArchs}}

---

ElasticSearch 采集器主要采集节点运行情况、集群健康、JVM 性能状况、索引性能、检索性能等。

## 配置 {#config}

### 前置条件 {#requirements}

- ElasticSearch 版本 >= 6.0.0
- ElasticSearch 默认采集 `Node Stats` 指标，如果需要采集 `Cluster-Health` 相关指标，需要设置 `cluster_health = true`
- 设置 `cluster_health = true` 可产生如下指标集
    - `elasticsearch_cluster_health`
- 设置 `cluster_stats = true` 可产生如下指标集
    - `elasticsearch_cluster_stats`

### 用户权限配置 {#user-permission}

如果开启账号密码访问，需要配置相应的权限，否则会导致监控信息获取失败错误。

目前支持 [Elasticsearch](elasticsearch.md#perm-es)、[Open Distro for Elasticsearch](elasticsearch.md#perm-open-es) 和 [OpenSearch](elasticsearch.md#perm-opensearch)。

#### Elasticsearch {#perm-es}

- 创建角色 `monitor`，设置如下权限

```http
POST /_security/role/monitor
{
  "applications": [],
  "cluster": [
      "monitor"
  ],
  "indices": [
      {
          "allow_restricted_indices": false,
          "names": [
              "*"
          ],
          "privileges": [
              "manage_ilm",
              "monitor"
          ]
      }
  ],
  "run_as": []
}
```

- 创建自定义用户，并赋予新创建的 `monitor` 角色。
- 其他信息请参考配置文件说明

#### Open Distro for ElasticSearch {#perm-open-es}

- 创建用户
- 创建角色 `monitor`，设置如下权限：

``` http
PUT _opendistro/_security/api/roles/monitor
{
  "description": "monitor es cluster",
  "cluster_permissions": [
    "cluster:admin/opendistro/ism/managedindex/explain",
    "cluster_monitor",
    "cluster_composite_ops_ro"
  ],
  "index_permissions": [
    {
      "index_patterns": [
        "*"
      ],
      "fls": [],
      "masked_fields": [],
      "allowed_actions": [
        "read",
        "indices_monitor"
      ]
    }
  ],
  "tenant_permissions": []
}
```

- 设置角色与用户之间的映射关系

#### OpenSearch {#perm-opensearch}

- 创建用户
- 创建角色 `monitor`，设置如下权限：

``` http
PUT _plugins/_security/api/roles/monitor
{
  "description": "monitor es cluster",
  "cluster_permissions": [
    "cluster:admin/opendistro/ism/managedindex/explain",
    "cluster_monitor",
    "cluster_composite_ops_ro"
  ],
  "index_permissions": [
    {
      "index_patterns": [
        "*"
      ],
      "fls": [],
      "masked_fields": [],
      "allowed_actions": [
        "read",
        "indices_monitor"
      ]
    }
  ],
  "tenant_permissions": []
}
```

- 设置角色与用户之间的映射关系

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

## 指标 {#metric}

以下所有数据采集，默认会追加全局选举 tag，也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

``` toml
[inputs.{{.InputName}}.tags]
# some_tag = "some_value"
# more_tag = "some_other_value"
# ...
```

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "metric"}}

### `{{$m.Name}}`

{{$m.MarkdownTable}}

{{ end }}
{{ end }}

## 自定义对象 {#object}

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "custom_object"}}

### `{{$m.Name}}`

{{$m.Desc}}

{{$m.MarkdownTable}}
{{end}}

{{ end }}

## 日志 {#logging}

<!-- markdownlint-disable MD046 -->
???+ info

    需将 DataKit 安装到 ElasticSearch 主机上才能采集到对应日志。
<!-- markdownlint-enable -->

如需采集 ElasticSearch 的日志，可在 {{.InputName}}.conf 中 将 `files` 打开，并写入 ElasticSearch 日志文件的绝对路径。比如：

```toml
[[inputs.{{.InputName}}]]
  ...
[inputs.{{.InputName}}.log]
files = ["/path/to/your/file.log"]
```

开启日志采集以后，默认会产生日志来源（`source`）为 `elasticsearch` 的日志。

### 日志 Pipeline 功能切割字段说明 {#pipeline}

- ElasticSearch 通用日志切割
  
通用日志文本示例：

``` log
[2021-06-01T11:45:15,927][WARN ][o.e.c.r.a.DiskThresholdMonitor] [master] high disk watermark [90%] exceeded on [A2kEFgMLQ1-vhMdZMJV3Iw][master][/tmp/elasticsearch-cluster/nodes/0] free: 17.1gb[7.3%], shards will be relocated away from this node; currently relocating away shards totalling [0] bytes; the node is expected to continue to exceed the high disk watermark when these relocations are complete
```

切割后的字段列表如下：

| 字段名 | 字段值                         | 说明         |
| ---    | ---                            | ---          |
| time   | 1622519115927000000            | 日志产生时间 |
| name   | o.e.c.r.a.DiskThresholdMonitor | 组件名称     |
| status | WARN                           | 日志等级     |
| nodeId | master                         | 节点名称     |

- ElasticSearch 搜索慢日志切割
  
搜索慢日志文本示例：

``` log
[2021-06-01T11:56:06,712][WARN ][i.s.s.query              ] [master] [shopping][0] took[36.3ms], took_millis[36], total_hits[5 hits], types[], stats[], search_type[QUERY_THEN_FETCH], total_shards[1], source[{"query":{"match":{"name":{"query":"Nariko","operator":"OR","prefix_length":0,"max_expansions":50,"fuzzy_transpositions":true,"lenient":false,"zero_terms_query":"NONE","auto_generate_synonyms_phrase_query":true,"boost":1.0}}},"sort":[{"price":{"order":"desc"}}]}], id[], 
```

切割后的字段列表如下：

| 字段名   | 字段值              | 说明             |
| ---      | ---                 | ---              |
| time     | 1622519766712000000 | 日志产生时间     |
| name     | i.s.s.query         | 组件名称         |
| status   | WARN                | 日志等级         |
| nodeId   | master              | 节点名称         |
| index    | shopping            | 索引名称         |
| duration | 36000000            | 请求耗时，单位 ns|

- ElasticSearch 索引慢日志切割

索引慢日志文本示例：

``` log
[2021-06-01T11:56:19,084][WARN ][i.i.s.index              ] [master] [shopping/X17jbNZ4SoS65zKTU9ZAJg] took[34.1ms], took_millis[34], type[_doc], id[LgC3xXkBLT9WrDT1Dovp], routing[], source[{"price":222,"name":"hello"}]
```

切割后的字段列表如下：

| 字段名   | 字段值              | 说明             |
| ---      | ---                 | ---              |
| time     | 1622519779084000000 | 日志产生时间     |
| name     | i.i.s.index         | 组件名称         |
| status   | WARN                | 日志等级         |
| nodeId   | master              | 节点名称         |
| index    | shopping            | 索引名称         |
| duration | 34000000            | 请求耗时，单位 ns|
