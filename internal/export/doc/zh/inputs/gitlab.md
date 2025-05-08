---
title     : 'GitLab'
summary   : '采集 GitLab 的指标数据'
tags:
  - 'GITLAB'
  - 'CI/CD'
__int_icon      : 'icon/gitlab'
dashboard :
  - desc  : 'GitLab'
    path  : 'dashboard/zh/gitlab'
monitor   :
  - desc  : '暂无'
    path  : '-'
---


{{.AvailableArchs}}

---

采集 GitLab 运行数据并以指标的方式上报到<<<custom_key.brand_name>>>。

## 配置 {#config}

首先需要打开 GitLab 服务的数据采集功能和设置白名单，具体操作见后续分段。

GitLab 设置完成后，对 DataKit 进行配置。注意，根据 GitLab 版本和配置不同，采集到的数据可能存在差异。

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    目前可以通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting)来开启采集器。
<!-- markdownlint-enable -->

### GitLab 开启数据采集功能 {#enable-prom}

GitLab 需要开启 Prometheus 数据采集功能，开启方式如下（以英文页面为例）：

- 以管理员账号登陆己方 GitLab 页面
- 转到 `Admin Area` > `Settings` > `Metrics and profiling`
- 选择 `Metrics - Prometheus`，点击 `Enable Prometheus Metrics` 并且 `save change`
- 重启 GitLab 服务

详情见[官方配置文档](https://docs.gitlab.com/ee/administration/monitoring/prometheus/gitlab_metrics.html#gitlab-prometheus-metrics){:target="_blank"}。

### 配置数据访问端白名单 {#white-list}

只开启数据采集功能还不够，GitLab 对于数据管理十分严格，需要再配置访问端的白名单。开启方式如下：

- 修改 GitLab 配置文件 `/etc/gitlab/gitlab.rb`，找到 `gitlab_rails['monitoring_whitelist'] = ['::1/128']` 并在该数组中添加 DataKit 的访问 IP（通常情况为 DataKit 所在主机的 IP，如果 GitLab 运行在容器中需根据实际情况添加）
- 重启 GitLab 服务

详情见[官方配置文档](https://docs.gitlab.com/ee/administration/monitoring/ip_whitelist.html){:target="_blank"}。

### 开启 GitLab CI 可视化 {#ci-visible}

确保已有 DataFlux Func 平台

通过配置 GitLab Webhook，可以实现 GitLab CI 可视化。需要通过 DataFlux Func 进行数据上报，开启步骤如下：

1. 在 DataFlux Func 上安装 GitLab CI 集成（脚本 ID：`guance_gitlab_ci`），安装流程参考[GitLab CI 集成配置](https://func.<<<custom_key.brand_main_domain>>>/doc/script-market-guance-gitlab-ci/){:target="_blank"};
2. 在 GitLab 转到 `Settings` -> `Webhooks` 中，将 URL 配置为第一步的 API 地址，Trigger 配置 Job events 和 Pipeline events 两项，点击 Add webhook 确认添加；

触发 GitLab CI 流程，执行结束后可以登陆<<<custom_key.brand_name>>>查看 CI 执行情况。

## 指标 {#metric}

以下所有数据采集，默认会追加全局选举 tag，也可以在配置中指定其它标签：

- 可以在配置中通过 `[inputs.{{.InputName}}.tags]` 为 **GitLab 指标数据**指定其它标签：

``` toml
[inputs.{{.InputName}}.tags]
# some_tag = "some_value"
# more_tag = "some_other_value"
# ...
```

- 可以在配置中通过 `[inputs.{{.InputName}}.ci_extra_tags]` 为 **GitLab CI 数据**指定其它标签：

``` toml
[inputs.{{.InputName}}.ci_extra_tags]
# some_tag = "some_value"
# more_tag = "some_other_value"
# ...
```

注意：为了确保 GitLab CI 功能正常，为 GitLab CI 数据指定的 extra tags 不会覆盖其数据中已有的标签（GitLab CI 标签列表见下）。

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

{{$m.Desc}}

- 标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}
