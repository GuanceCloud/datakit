---
title     : 'Prometheus Remote Write'
summary   : 'Receive metrics via Prometheus Remote Write'
__int_icon      : 'icon/prometheus'
dashboard :
  - desc  : 'N/A'
    path  : '-'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# Prometheus Remote Write
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

Monitor Prometheus Remote Write data and report it to Guance Cloud.

## Configuration {#config}

### Preconditions {#requirements}

Note that for some earlier versions of `vmalert`, the setting `default_content_encoding = "snappy"` needs to be turned on in the collector's configuration file.

Turn on the Prometheus Remote Write feature and add the following configuration in Prometheus.yml:

```yml
remote_write:
 - url: "http://<datakit-ip>:9529/prom_remote_write"

# If want add some tag, ( __source will not in tag, only show in Datakit expose metrics)
# remote_write:
# - url: "http://<datakit-ip>:9529/prom_remote_write?host=1.2.3.4&foo=bar&__source=<your_source>" 
```

### Collector Configuration {#input-config}
<!-- markdownlint-disable MD046 -->
=== "Host Installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    Once configured, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    The collector can now be turned on by [ConfigMap Injection Collector Configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting).
<!-- markdownlint-enable -->
### Add, Ignore and Rename Tags {#tag-ops}

We can label the collected metrics by configuring `tags`, as follows:

```toml
  ## custom tags
  [inputs.prom_remote_write.tags]
  some_tag = "some_value"
  more_tag = "some_other_value"
```

If both blacklist and whitelist, all list will cancel.

We can apply blacklist on the tag to ignore it:

```toml
  ## tags to ignore
  tags_ignore = ["xxxx"]
```

We can apply regex match blacklist on the tag to ignore it:

```toml
  ## tags to ignore with regex
  tags_ignore_regex = ["xxxx"]
```

We can apply whitelists on tags:

```toml
  ## tags white list
  # tags_only = ["xxxx"]
```

We can apply regex match whitelist on tags:

```toml
  ## tags white list with regex
  # tags_only_regex = ["xxxx"]
```

We can rename some of the tag names that an indicator already has by configuring `tags_rename`, as follows:

```toml
  ## tags to rename
  [inputs.prom_remote_write.tags_rename]
  old_tag_name = "new_tag_name"
  more_old_tag_name = "other_new_tag_name"
```

In addition, when the renamed tag key is the same as the existing tag key: You can configure whether to overwrite the existing tag key by `overwrite`.

> Note: For [DataKit global tag key](datakit-conf.md#update-global-tag), renaming them is not supported here.

## Metric {#metric}

The standard set is based on the measurements sent by Prometheus.

## Command Line Debug Measurements {#debug}

DataKit provides a simple tool for debugging `prom.conf`. If you constantly adjust the configuration of `prom.conf`, you can achieve the goal of collecting only Prometheus metrics that meet certain name rules.

Datakit supports direct debugging of the collector configuration files from the command line. Configure the `output` entry of `prom_remote_write.conf` under `conf.d/prom`, configure it as a local file path, and then `prom_remote_write.conf` writes the collected data to the file without uploading it to the center.

Restart Datakit for the configuration file to take effect:

```shell
datakit service -R
```

The *prom_remote_write* collector will then write the collected data to the local file indicated by the output.

We can debug *prom_remote_write.conf* by executing the following command.

```shell
datakit debug --prom-conf prom_remote_write.conf
```

Parameter description:

- `prom-conf`: Specify the configuration file and look for the  `prom_remote_write.conf` file in the current directory by default. If it is not found, it will look for the corresponding file in the *<datakit-install-dir\>/conf.d/prom* directory.

Output sample:

```not-set
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
<!-- markdownlint-disable MD007 -->
Output description:

- Line Protocol Points: Generated line protocol points

- Summary: Summary results

  - - Total time series: Number of timelines

  - - Total line protocol points: Line protocol points

  - - Total measurements: The number of measurements and their names.
<!-- markdownlint-enable -->