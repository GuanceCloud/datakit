{{.CSS}}

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# {{.InputName}}

监听 Prometheus Remote Write 数据，上报到观测云。

## 前置条件

开启 Prometheus Remote Write 功能，在 prometheus.yml 添加如下配置：

```yml
remote_write:
 - url: "http://<datakit-ip>:9529/prom_remote_write"
```

## 配置

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}} 
```

## 指标集

指标集以 Prometheus 发送过来的指标集为准。

## 命令行调试指标集

DataKit 提供一个简单的调试 `prom.conf` 的工具，如果不断调整 `prom.conf` 的配置，可以实现只采集符合一定名称规则的 Prometheus 指标的目的。

Datakit 支持命令行直接调试 prom_remote_write 采集器的配置文件，配置 `conf.d/prom` 下 `prom_remote_write.conf` 的 `output` 项，将其配置为一个本地文件路径，之后`prom_remote_write.conf` 会将采集到的数据写到文件中，数据就不会上传到中心。

重启 Datakit，让配置文件生效：

```
datakit --restart
```

这时 prom_remote_write 采集器将把采集的数据写到 output 指明的本地文件中。

这时执行如下命令，即可调试 prom_remote_write.conf

```
datakit --prom-conf prom_remote_write.conf
```

参数说明：

- `prom-conf`: 指定配置文件，默认在当前目录下寻找 `prom_remote_write.conf` 文件，如果未找到，会去 `<datakit-install-dir>/conf.d/prom` 目录下查找相应文件。

输出示例：

```
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

- - Total time series: 时间线数量

- - Total line protocol points: 行协议点数

- - Total measurements: 指标集个数及其名称。
