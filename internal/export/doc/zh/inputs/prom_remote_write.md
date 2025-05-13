---
title     : 'Prometheus Remote Write'
summary   : '通过 Prometheus Remote Write 汇集指标数据'
tags:
  - '外部数据接入'
  - 'PROMETHEUS'
__int_icon: 'icon/prometheus'
dashboard :
  - desc  : '暂无'
    path  : '-'
monitor   :
  - desc  : '暂无'
    path  : '-'
---


{{.AvailableArchs}}

---

监听 Prometheus Remote Write 数据，上报到<<<custom_key.brand_name>>>。

## 配置 {#config}

### 前置条件 {#requirements}

注意，对于 `vmalert` 的一些早期版本，需要在采集器的配置文件中打开设置 `default_content_encoding = "snappy"`。

开启 Prometheus Remote Write 功能，在 *prometheus.yml* 添加如下配置：

```yml
remote_write:
 - url: "http://<datakit-ip>:9529/prom_remote_write"

# If want add some tag, ( __source will not in tag, only show in DataKit expose metrics)
# remote_write:
# - url: "http://<datakit-ip>:9529/prom_remote_write?host=1.2.3.4&foo=bar&__source=<your_source>" 
```

### 采集器配置 {#input-config}

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    可通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting) 或 [配置 ENV_DATAKIT_INPUTS](../datakit/datakit-daemonset-deploy.md#env-setting) 开启采集器。
<!-- markdownlint-enable -->

### tags 的处理 {#tag-ops}

可以通过配置 `tags` 为采集到的指标加上标签，如下：

```toml
  ## custom tags
  [inputs.prom_remote_write.tags]
  some_tag = "some_value"
  more_tag = "some_other_value"
```

注意：黑名单和白名单同时配置，黑白名单会全部失效。

可以通过配置 `tags_ignore` 忽略指标上的某些标签（黑名单），如下：

```toml
  ## tags to ignore
  tags_ignore = ["xxxx"]
```

可以通过配置 `tags_ignore_regex` 正则匹配并忽略指标上的标签（黑名单），如下：

```toml
  ## tags to ignore with regex
  tags_ignore_regex = ["xxxx"]
```

可以通过配置 `tags_only` 配置指标上的标签白名单，如下：

```toml
  ## tags white list
  # tags_only = ["xxxx"]
```

可以通过配置 `tags_only_regex` 正则匹配指标上的标签白名单，如下：

```toml
  ## tags white list with regex
  # tags_only_regex = ["xxxx"]
```

可以通过配置 `tags_rename` 重命名指标已有的某些标签名，如下：

```toml
  ## tags to rename
  [inputs.prom_remote_write.tags_rename]
  old_tag_name = "new_tag_name"
  more_old_tag_name = "other_new_tag_name"
```

另外，当重命名后的 tag key 与已有 tag key 相同时，可以通过 `overwrite` 配置是否覆盖掉已有的 tag key。

> 注意：对于 [DataKit 全局 tag key](../datakit/datakit-conf.md#update-global-tag)，此处不支持将它们重命名。

## 指标 {#metric}

指标集以 Prometheus 发送过来的指标集为准。

## 配置 Prometheus Remote Write 指标过滤 {#remote-write-relabel}

当使用 Prometheus 以 remote write 方式往 DataKit 推送指标时，如果指标太多，可能导致
存储中的数据暴增。此时我们可以通过 Prometheus 自身的 relabel 功能来选取特定的指标。

在 Prometheus 中，要配置 `remote_write` 到另一个服务，并且只发送指定的指标列表，我们需要在
Prometheus 的配置文件（通常是 `prometheus.yml`）中设置 `remote_write` 部分，并指定 `match[]`
参数来定义要发送的指标。

以下是一个配置示例，它展示了如何将特定的指标列表发送到远程写入端点：

```yaml
remote_write:
  - url: "http://remote-write-service:9090/api/v1/write"
    write_relabel_configs:
      - source_labels: ["__name__"]
        regex: "my_metric|another_metric|yet_another_metric"
        action: keep
```

在这个配置中：

- `url`: 远程写入服务的 URL
- `write_relabel_configs`: 一个列表，用于重新标记和过滤要发送的指标
    - `source_labels`: 指定用于匹配和重新标记的源标签
    - `regex`: 一个正则表达式，用于匹配要保留的指标名称
    - `action`: 指定匹配正则表达式的指标是被保留（`keep`）还是被丢弃（`drop`）

在上面的示例中，只有名称匹配 `my_metric`、`another_metric` 或 `yet_another_metric` 的指标
会被发送到远程写入端点。其他所有指标都会被忽略。

最后，重新加载或重启 Prometheus 服务以应用更改。

## 命令行调试指标集 {#debug}

DataKit 提供一个简单的调试 `prom.conf` 的工具，如果不断调整 `prom.conf` 的配置，可以实现只采集符合一定名称规则的 Prometheus 指标的目的。

DataKit 支持命令行直接调试本采集器的配置文件。在配置 `conf.d/prom` 下 `prom_remote_write.conf` 的 `output` 项，将其配置为一个本地文件路径，之后 `prom_remote_write.conf` 会将采集到的数据写到文件中，数据就不会上传到中心。

重启 DataKit，让配置文件生效：

```shell
datakit service -R
```

这时 *prom_remote_write* 采集器将把采集的数据写到 output 指明的本地文件中。

这时执行如下命令，即可调试 *prom_remote_write.conf*

```shell
datakit debug --prom-conf prom_remote_write.conf
```

参数说明：

- `prom-conf`: 指定配置文件，默认在当前目录下寻找 `prom_remote_write.conf` 文件，如果未找到，会去 *<datakit-install-dir\>/conf.d/prom* 目录下查找相应文件。

输出示例：

``` not-set
================= Line Protocol Points ==================

 prometheus,instance=localhost:9090,job=prometheus,monitor=codelab-monitor target_scrapes_sample_out_of_order_total=0 1634548272855000000
 prometheus,instance=localhost:9090,job=prometheus,monitor=codelab-monitor,scrape_job=node target_sync_failed_total=0 1634548272855000000
 prometheus,instance=localhost:9090,job=prometheus,monitor=codelab-monitor,scrape_job=prometheus target_sync_failed_total=0 1634548272855000000
 prometheus,instance=localhost:9090,job=prometheus,monitor=codelab-monitor,scrape_job=node target_sync_length_seconds_sum=0.000070352 1634548272855000000
 prometheus,instance=localhost:9090,job=prometheus,monitor=codelab-monitor,scrape_job=node target_sync_length_seconds_count=1 1634548272855000000
 prometheus,instance=localhost:9090,job=prometheus,monitor=codelab-monitor,scrape_job=prometheus target_sync_length_seconds_sum=0.000089457 1634548272855000000
 prometheus,instance=localhost:9090,job=prometheus,monitor=codelab-monitor,scrape_job=prometheus target_sync_length_seconds_count=1 1634548272855000000
 prometheus,instance=localhost:9090,job=prometheus,monitor=codelab-monitor template_text_expansion_failures_total=0 1634548272855000000
 prometheus,instance=localhost:9090,job=prometheus,monitor=codelab-monitor template_text_expansions_total=0 1634548272855000000
 prometheus,instance=localhost:9090,job=prometheus,monitor=codelab-monitor treecache_watcher_goroutines=0 1634548272855000000
 prometheus,instance=localhost:9090,job=prometheus,monitor=codelab-monitor treecache_zookeeper_failures_total=0 1634548272855000000
 prometheus,instance=localhost:9090,job=prometheus,monitor=codelab-monitor tsdb_blocks_loaded=1 1634548272855000000
 prometheus,instance=localhost:9090,job=prometheus,monitor=codelab-monitor tsdb_checkpoint_creations_failed_total=0 1634548272855000000
 prometheus,instance=localhost:9090,job=prometheus,monitor=codelab-monitor tsdb_checkpoint_creations_total=1 1634548272855000000
 prometheus,instance=localhost:9090,job=prometheus,monitor=codelab-monitor tsdb_checkpoint_deletions_failed_total=0 1634548272855000000
 prometheus,instance=localhost:9090,job=prometheus,monitor=codelab-monitor tsdb_checkpoint_deletions_total=1 1634548272855000000
 prometheus,instance=localhost:9090,job=prometheus,monitor=codelab-monitor tsdb_clean_start=1 1634548272855000000
 prometheus,instance=localhost:9090,job=prometheus,le=100,monitor=codelab-monitor tsdb_compaction_chunk_range_seconds_bucket=0 1634548272855000000
 prometheus,instance=localhost:9090,job=prometheus,le=400,monitor=codelab-monitor tsdb_compaction_chunk_range_seconds_bucket=0 1634548272855000000
 prometheus,instance=localhost:9090,job=prometheus,le=1600,monitor=codelab-monitor tsdb_compaction_chunk_range_seconds_bucket=0 1634548272855000000
 prometheus,instance=localhost:9090,job=prometheus,le=6400,monitor=codelab-monitor tsdb_compaction_chunk_range_seconds_bucket=0 1634548272855000000
 prometheus,instance=localhost:9090,job=prometheus,le=25600,monitor=codelab-monitor tsdb_compaction_chunk_range_seconds_bucket=0 1634548272855000000
 prometheus,instance=localhost:9090,job=prometheus,le=102400,monitor=codelab-monitor tsdb_compaction_chunk_range_seconds_bucket=0 1634548272855000000
 prometheus,instance=localhost:9090,job=prometheus,le=409600,monitor=codelab-monitor tsdb_compaction_chunk_range_seconds_bucket=0 1634548272855000000
 prometheus,instance=localhost:9090,job=prometheus,le=1.6384e+06,monitor=codelab-monitor tsdb_compaction_chunk_range_seconds_bucket=0 1634548272855000000
 prometheus,instance=localhost:9090,job=prometheus,le=6.5536e+06,monitor=codelab-monitor tsdb_compaction_chunk_range_seconds_bucket=0 1634548272855000000
 prometheus,instance=localhost:9090,job=prometheus,le=2.62144e+07,monitor=codelab-monitor tsdb_compaction_chunk_range_seconds_bucket=0 1634548272855000000
 prometheus,instance=localhost:9090,job=prometheus,le=+Inf,monitor=codelab-monitor tsdb_compaction_chunk_range_seconds_bucket=0 1634548272855000000
 prometheus,instance=localhost:9090,job=prometheus,monitor=codelab-monitor tsdb_compaction_chunk_range_seconds_sum=0 1634548272855000000
 prometheus,instance=localhost:9090,job=prometheus,monitor=codelab-monitor tsdb_compaction_chunk_range_seconds_count=0 1634548272855000000
 prometheus,instance=localhost:9090,job=prometheus,le=4,monitor=codelab-monitor tsdb_compaction_chunk_samples_bucket=0 1634548272855000000
 prometheus,instance=localhost:9090,job=prometheus,le=6,monitor=codelab-monitor tsdb_compaction_chunk_samples_bucket=0 1634548272855000000
 prometheus,instance=localhost:9090,job=prometheus,le=9,monitor=codelab-monitor tsdb_compaction_chunk_samples_bucket=0 1634548272855000000
...
================= Summary ==================

Total time series: 155
Total line protocol points: 487
Total measurements: 6 (prometheus, promhttp, up, scrape, go, node)
```

输出说明：

- Line Protocol Points： 产生的行协议点
- Summary： 汇总结果
    - Total time series: 时间线数量
    - Total line protocol points: 行协议点数
    - Total measurements: 指标集个数及其名称。
