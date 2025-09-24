---
title     : 'ConfigWatcher'
summary   : 'Monitor changes in the content of files or directories and report event data'
tags:
__int_icon      : ''
dashboard :
  - desc  : 'N/A'
    path  : '-'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---

{{.AvailableArchs}}

---

The ConfigWatcher collector monitors content changes in files or directories, constructs change event data, and reports it to the platform.

## Configuration {#config}

<!-- markdownlint-disable MD046 -->
=== "Host Installation"

    Navigate to the *conf.d/samples* directory in your DataKit installation path, copy *{{.InputName}}.conf.sample* and rename it to *{{.InputName}}.conf*. Example:

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    After configuration, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    Currently enabled by injecting collector configuration via [ConfigMap](../datakit/datakit-daemonset-deploy.md#configmap-setting).

<!-- markdownlint-enable -->

### Example Configuration: Crontab {#example-crontab}

Below is an example configuration for monitoring Crontab files:

```toml
[[inputs.configwatcher]]
  ## Required. A name for this collection task for identification.
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

When you modify Crontab tasks using the crontab -e command in the Linux terminal, the collector will detect changes in the corresponding files.

## FAQ {#faq}

- The collector ignores changes in file permissions and ownership, monitoring only content changes.
- The timestamp for change events is taken from the file's ModTime, not the system time.
- If the file size exceeds `max_diff_size`, the differences between old and new file contents are not compared.
