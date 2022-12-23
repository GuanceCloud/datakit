<!-- This file required to translate to EN. -->
{{.CSS}}
# 磁盘 S.M.A.R.T
---

{{.AvailableArchs}}

---

计算机硬盘运行状态数据采集

## 前置条件 {#requrements}

安装 smartmontools

- Linux: `sudo apt install smartmontools -y`

	如果固态硬盘，符合  nvme 标准，建议安装 nvme-cli 以得到更多 nvme 信息：`sudo apt install nvme-cli -y`

- MacOS: `brew install smartmontools -y`
- WinOS: 下载 [Windows 版本](https://www.smartmontools.org/wiki/Download#InstalltheWindowspackage){:target="_blank"}

## 配置 {#config}

=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    配置好后，重启 DataKit 即可。

=== "Kubernetes"

    目前可以通过 [ConfigMap 方式注入采集器配置](datakit-daemonset-deploy.md#configmap-setting)来开启采集器。

## 指标集 {#requrements}

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
