{{.CSS}}
# GitLab
---

- DataKit 版本：{{.Version}}
- 操作系统支持：`{{.AvailableArchs}}`

采集 GitLab 运行数据并以指标的方式上报到观测云。

## 前置条件

- 已安装 GitLab（[GitLab 官方链接](https://about.gitlab.com/){:target="_blank"}）

## 配置

首先需要打开 GitLab 服务的数据采集功能和设置白名单，具体操作见后续分段。

GitLab 设置完成后，对 DataKit 进行配置。注意，根据 GitLab 版本和配置不同，采集到的数据可能存在差异。

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}} 
```

配置好后，重启 DataKit 即可。

此 input 支持选举功能，[关于选举](election.md)。

### GitLab 开启数据采集功能

GitLab 需要开启 promtheus 数据采集功能，开启方式如下（以英文页面为例）：

1. 以管理员账号登陆己方 GitLab 页面
2. 转到 `Admin Area` > `Settings` > `Metrics and profiling`
3. 选择 `Metrics - Prometheus`，点击 `Enable Prometheus Metrics` 并且 `save change`
4. 重启 GitLab 服务

详情见[官方配置文档](https://docs.gitlab.com/ee/administration/monitoring/prometheus/gitlab_metrics.html#gitlab-prometheus-metrics){:target="_blank"}。

### 配置数据访问端白名单

只开启数据采集功能还不够，GitLab 对于数据管理十分严格，需要再配置访问端的白名单。开启方式如下：

1. 修改 GitLab 配置文件 `/etc/gitlab/gitlab.rb`，找到 `gitlab_rails['monitoring_whitelist'] = ['::1/128']` 并在该数组中添加 DataKit 的访问 IP（通常情况为 DataKit 所在主机的 IP，如果 GitLab 运行在容器中需根据实际情况添加）
2. 重启 GitLab 服务

详情见[官方配置文档](https://docs.gitlab.com/ee/administration/monitoring/ip_whitelist.html){:target="_blank"}。

### 开启 Gitlab CI 可视化

确保当前 Datakit 版本（1.2.13 及以后）支持 Gitlab CI 可视化功能。

通过配置 Gitlab Webhook，可以实现 Gitlab CI 可视化。开启步骤如下：

1. 在 Gitlab 转到 `Settings` > `Webhooks` 中，将 URL 配置为 http://Datakit_IP:PORT/v1/gitlab，Trigger 配置 Job events 和 Pipeline events 两项，点击 Add webhook 确认添加；
2. 可点击 Test 按钮测试 Webhook 配置是否正确，Datakit 接收到 Webhook 后应返回状态码 200。正确配置后，Datakit 可以顺利采集到 Gitlab 的 CI 信息。

Datakit 接收到 Webhook Event 后，是将数据作为 logging 打到数据中心的。

注意：如果将 Gitlab 数据打到本地网络的 Datakit，需要对 Gitlab 进行额外的配置，见 [allow requests to the local network](https://docs.gitlab.com/ee/security/webhooks.html){:target="_blank"} 。

另外：Gitlab CI 功能不参与采集器选举，用户只需将 Gitlab Webhook 的 URL 配置为其中一个 Datakit 的 URL 即可；若只需要 Gitlab CI 可视化功能而不需要 Gitlab 指标采集，可通过配置 `enable_collect = false` 关闭指标采集功能。

## 指标集

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名）。

可以在配置中通过 `[inputs.{{.InputName}}.tags]` 为 **Gitlab 指标数据**指定其它标签：

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

可以在配置中通过 `[inputs.{{.InputName}}.ci_extra_tags]` 为 **Gitlab CI 数据**指定其它标签：

``` toml
 [inputs.{{.InputName}}.ci_extra_tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

注意：为了确保 Gitlab CI 功能正常，为 Gitlab CI 数据指定的 extra tags 不会覆盖其数据中已有的标签（Gitlab CI 标签列表见下）。

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

{{$m.Desc}}

-  标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }} 
