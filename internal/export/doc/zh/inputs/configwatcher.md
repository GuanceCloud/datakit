---
title     : 'ConfigWatcher'
summary   : '监控文件或目录的内容变更，并上报事件数据'
tags:
__int_icon: ''
dashboard :
  - desc  : '暂无'
    path  : '-'
monitor   :
  - desc  : '暂无'
    path  : '-'
---

{{.AvailableArchs}}

---

ConfigWatcher 采集器支持监控文件或目录的内容变更，构建变更事件数据并上报观测云平台。

## 配置 {#config}

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    进入 DataKit 安装目录下的 *conf.d/{{.Catalog}}* 目录，复制 *{{.InputName}}.conf.sample* 并命名为 *{{.InputName}}.conf*。示例如下：
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    目前可以通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting)来开启采集器。

<!-- markdownlint-enable -->

### 示例配置 Crontab {#example-crontab}

下面是监控 Crontab 文件的示例，配置如下：

```toml
[[inputs.configwatcher]]
  ## Require. A name for this collection task for identification.
  task_name = "Crontab"

  ## An array of file paths to monitor for changes.
  paths = [
      "/var/spool/cron/crontabs",
  ]

  ## The interval at which to check for changes.
  interval = "3m"

  ## Whether to recursively monitor directories in the provided paths.
  recursive = true

  ## The maximum file size (in bytes) for which to compute content diffs, default is 256KiB.
  max_diff_size = 262144
```

在 Linux 命令行执行 `crontab -e` 命令修改 Crontab 任务，采集器能发现到对应文件的变更。

### FAQ {#faq}

- 采集器忽略文件的权限、所有者变更，只监控内容变更
- 变更事件的时间取自文件的 ModTime，不是系统时间
- 如果文件 size 超过 `max_diff_size`，不比较新旧文件的差异
