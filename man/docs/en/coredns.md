<!-- This file required to translate to EN. -->
{{.CSS}}
# CoreDNS
---

{{.AvailableArchs}}

---

CoreDNS 采集器用于采集 CoreDNS 相关的指标数据。

## 前置条件 {#requirements}

- CoreDNS [配置](https://coredns.io/plugins/metrics/){:target="_blank"}启用 `prometheus` 插件

## 配置 {#input-config}

=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置好后，[重启 DataKit](datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    目前可以通过 [ConfigMap 方式注入采集器配置](datakit-daemonset-deploy.md#configmap-setting)来开启采集器。

## 指标集 {#metrics}

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

- 标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}
