{{.CSS}}
# Statsd 数据接入
---

{{.AvailableArchs}}

---

statsd 采集器用于接收网络上发送过来的 statsd 数据。

## 前置条件 {#requrements}

暂无

## 配置 {#config}

=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    配置好后，重启 DataKit 即可。

=== "Kubernetes"

    目前可以通过 [ConfigMap 方式注入采集器配置](datakit-daemonset-deploy.md#configmap-setting)来开启采集器。

## 指标集 {#measurement}

statsd 暂无指标集定义，所有指标以网络发送过来的指标为准。
