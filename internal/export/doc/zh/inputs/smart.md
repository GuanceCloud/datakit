---
title     : '磁盘 S.M.A.R.T'
summary   : '通过 smartctl 采集磁盘指标'
__int_icon      : 'icon/smartctl'
dashboard :
  - desc  : '暂无'
    path  : '-'
monitor   :
  - desc  : '暂无'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# 磁盘 S.M.A.R.T
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

计算机硬盘运行状态数据采集

## 配置 {#config}

### 前置条件 {#requrements}

安装 `smartmontools`

- Linux: `sudo apt install smartmontools -y`

如果固态硬盘，符合 NVMe 标准，建议安装 `nvme-cli` 以得到更多 NVMe 信息：

<!-- markdownlint-disable MD046 -->
=== "Linux"

    ```shell
    sudo apt install nvme-cli -y
    ```

=== "macOS"

    ```shell
    brew install smartmontools -y
    ```
=== "Windows"

    下载 [Windows 版本](https://www.smartmontools.org/wiki/Download#InstalltheWindowspackage){:target="_blank"}
<!-- markdownlint-enable -->

### 采集器安装 {#input-config}

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    配置好后，重启 DataKit 即可。

=== "Kubernetes"

    目前可以通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting)来开启采集器。
<!-- markdownlint-enable -->

## 指标 {#metric}

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

```toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

- 标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}
