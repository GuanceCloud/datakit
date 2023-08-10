---
title     : 'Jenkins'
summary   : '采集 Jenkins 的指标和日志'
__int_icon      : 'icon/jenkins'
dashboard :
  - desc  : 'Jenkins'
    path  : 'dashboard/zh/jenkins'
monitor   :
  - desc  : 'Jenkins'
    path  : 'monitor/zh/jenkins'
---

<!-- markdownlint-disable MD025 -->
# Jenkins
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

Jenkins 采集器是通过插件 Metrics 采集数据监控 Jenkins，包括但不限于任务数、系统 CPU 使用、JVM CPU 使用等。

## 配置 {#config}

### 前置条件 {#requirements}

- JenKins 版本 >= `2.332.1`; 已测试的版本：
    - [x] 2.332.1

- 下载 `Metric` 插件，[管理插件页面](https://www.jenkins.io/doc/book/managing/plugins/){:target="_blank"},[Metric 插件页面](https://plugins.jenkins.io/metrics/){:target="_blank"}
- 在 JenKins 管理页面 `your_manage_host/configure` 生成 `Metric Access keys`

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

### Jenkins CI Visibility {#ci-visibility}

Jenkins 采集器可以通过接收 Jenkins DataDog plugin 发出的 CI Event 实现 CI 可视化。

Jenkins CI Visibility 开启方法：

- 确保在配置文件中开启了 Jenkins CI Visibility 功能，且配置了监听端口号（如 `:9539`），重启 Datakit；
- 在 Jenkins 中安装 [Jenkins Datadog plugin](https://plugins.jenkins.io/datadog/){:target="_blank"} ；
- 在 Manage Jenkins > Configure System > Datadog Plugin 中选择 `Use the Datadog Agent to report to Datadog (recommended)`，配置 `Agent Host` 为 Datakit IP 地址。`DogStatsD Port` 及 `Traces Collection Port` 两项均配置为上述 Jenkins 采集器配置文件中配置的端口号，如 `9539`（此处不加 `:`）；
- 勾选 `Enable CI Visibility`；
- 点击 `Save` 保存设置。

配置完成后 Jenkins 能够通过 Datadog Plugin 将 CI 事件发送到 Datakit。

## 指标 {#metric}

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名）。可以在配置中通过 `[inputs.{{.InputName}}.tags]` 为采集的指标指定其它标签：

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

可以在配置中通过 `[inputs.{{.InputName}}.ci_extra_tags]` 为 Jenkins CI Event 指定其它标签：

```toml
 [inputs.{{.InputName}}.ci_extra_tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
```

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

- 标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}

## 日志 {#logging}

如需采集 JenKins 的日志，可在 *{{.InputName}}.conf* 中 将 `files` 打开，并写入 JenKins 日志文件的绝对路径。比如：

```toml
[[inputs.JenKins]]
  ...
  [inputs.JenKins.log]
    files = ["/var/log/jenkins/jenkins.log"]
```

开启日志采集以后，默认会产生日志来源（`source`）为 `jenkins` 的日志。

> 注意：必须将 DataKit 安装在 JenKins 所在主机才能采集 JenKins 日志

### 日志 Pipeline 功能切割字段说明 {#pipeline}

- JenKins 通用日志切割

通用日志文本示例：

```log
2021-05-18 03:08:58.053+0000 [id=32] INFO jenkins.InitReactorRunner$1#onAttained: Started all plugins
```

切割后的字段列表如下：

| 字段名 | 字段值              | 说明                         |
| ---    | ---                 | ---                          |
| status | info                | 日志等级                     |
| id     | 32                  | id                           |
| time   | 1621278538000000000 | 纳秒时间戳（作为行协议时间） |
