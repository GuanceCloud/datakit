---
title     : 'Profiling Ruby'
summary   : 'Ruby Profiling 集成'
tags:
  - 'RUBY'
  - 'PROFILE'
__int_icon: 'icon/profiling'
---

DataKit 支持接收 [Datadog Ruby Continuous Profiler](https://docs.datadoghq.com/profiler/enabling/?tab=ruby){:target="_blank"} 上报的性能数据，并将其发送到<<<custom_key.brand_name>>>。Ruby Profiler 可采集 CPU 时间、Wall Time 和对象分配等数据。

## 前置条件 {#requirements}

- 已安装 [DataKit](https://www.<<<custom_key.brand_main_domain>>>){:target="_blank"}，并开启 [Profile 采集器](profile.md#config)。
- 使用 CRuby 2.5 或更高版本；推荐使用 CRuby 3.2.3 或更高版本。JRuby 和 TruffleRuby 暂不支持。
- 应用运行于受支持的 Linux x86-64 或 arm64 环境，包括基于 glibc 或 musl 的发行版。Ruby Profiler 不支持 Serverless 环境。
- 使用 `datadog` gem。建议使用 `~> 2.30`；低于 2.30 的版本在编译原生扩展时还需要 `pkg-config` 或 `pkgconf`。

具体兼容范围会随 SDK 更新，请同时参考 [Ruby Profiler 支持版本](https://docs.datadoghq.com/profiler/enabling/supported_versions/?tab=ruby){:target="_blank"}。

## 开启 DataKit Profile 采集器 {#datakit}

进入 DataKit 安装目录下的 `conf.d/profile` 目录，复制 `profile.conf.sample` 并命名为 `profile.conf`。默认配置已经包含 Ruby SDK 使用的接收端点：

```toml
[[inputs.profile]]
  endpoints = ["/profiling/v1/input"]
  body_size_limit_mb = 32
```

[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 后，接收地址为：

```text
http://<DataKit 地址>:9529/profiling/v1/input
```

Ruby SDK 中应配置 Agent 基础地址 `http://<DataKit 地址>:9529`，无需在地址后添加 `/profiling/v1/input`。

## 安装 Ruby SDK {#install}

在应用的 `Gemfile` 中添加：

```ruby
gem "datadog", "~> 2.30"
```

然后安装依赖：

```shell
bundle install
```

如需同时启用 Ruby APM 自动插桩，请参考 [DDTrace Ruby](ddtrace-ruby.md) 配置 gem 的加载入口。

## 配置并启动 Profiler {#run}

### 使用环境变量 {#environment-variables}

下面的示例将 Profile 数据发送到本机 DataKit：

```shell
DD_PROFILING_ENABLED=true \
DD_TRACE_AGENT_URL=http://127.0.0.1:9529 \
DD_ENV=production \
DD_SERVICE=my-ruby-service \
DD_VERSION=1.0.0 \
DD_TAGS=team:apm,region:cn \
bundle exec ddprofrb exec ruby app.rb
```

Rails 应用可使用相同的环境变量启动：

```shell
DD_PROFILING_ENABLED=true \
DD_TRACE_AGENT_URL=http://127.0.0.1:9529 \
DD_ENV=production \
DD_SERVICE=my-rails-service \
DD_VERSION=1.0.0 \
bundle exec ddprofrb exec bin/rails server
```

也可以使用 `DD_AGENT_HOST` 和 `DD_TRACE_AGENT_PORT` 分别配置地址：

```shell
export DD_AGENT_HOST=127.0.0.1
export DD_TRACE_AGENT_PORT=9529
```

`DD_TRACE_AGENT_URL` 的优先级高于 host/port 配置，请勿同时设置相互冲突的目标地址。在容器或 Kubernetes 环境中，如果 DataKit 与应用不在同一个容器，请将 `127.0.0.1` 替换为应用能够访问的 DataKit 地址。

### 使用代码配置 {#code-configuration}

也可以在应用启动阶段配置 Profiler。例如，Rails 应用可在 initializer 中加入：

```ruby
require "datadog"

Datadog.configure do |c|
  c.agent.host = "127.0.0.1"
  c.agent.port = 9529
  c.profiling.enabled = true
  c.env = "production"
  c.service = "my-rails-service"
  c.version = "1.0.0"
  c.tags = { "team" => "apm", "region" => "cn" }
end
```

配置代码后，仍推荐使用 `ddprofrb exec` 启动应用，确保 Profiler 尽早加载。如果启动器无法使用，可在应用入口的最前面加载 Profiler：

```ruby
require "datadog/profiling/preload"
```

然后使用原有命令启动应用。

## 常用配置 {#configuration}

| 环境变量 | 默认值 | 说明 |
| --- | --- | --- |
| `DD_PROFILING_ENABLED` | `false` | 是否开启 Continuous Profiler。接入时必须设置为 `true`。 |
| `DD_PROFILING_ALLOCATION_ENABLED` | `false` | 是否采集对象分配数据。开启后会增加运行时开销，建议先在预发布环境评估。 |
| `DD_PROFILING_MAX_FRAMES` | `400` | 每个调用栈采集的最大帧数。 |
| `DD_PROFILING_EXPERIMENTAL_HEAP_ENABLED` | `false` | 是否启用实验性的堆分析；同时需要开启对象分配采集。 |
| `DD_ENV` | 无 | 应用部署环境，例如 `production`、`staging`。 |
| `DD_SERVICE` | 由 SDK 推断 | 服务名。生产环境建议显式设置。 |
| `DD_VERSION` | 无 | 应用版本。 |
| `DD_TAGS` | 无 | 附加标签，使用 `key:value` 形式并以逗号分隔。 |

实验性功能的支持范围和性能开销可能随 SDK 版本变化，启用前请参考 [Ruby Profiler 配置](https://docs.datadoghq.com/profiler/enabling/?tab=ruby#configuration){:target="_blank"}。

## 查看 Profile {#view}

应用启动后，Ruby Profiler 会定期向 DataKit 上报数据。稍等一到两分钟后，可在<<<custom_key.brand_name>>>空间的[应用性能监测 -> Profile](https://console.<<<custom_key.brand_main_domain>>>/tracing/profile){:target="_blank"}页面按 `service`、`env` 和 `version` 查看相应数据。

如果应用同时接入了 DDTrace Ruby 链路追踪，兼容版本的 SDK 会自动携带 Trace 与 Profile 的关联信息。链路追踪接入方式见 [DDTrace Ruby](ddtrace-ruby.md)。

## DataKit 指标生成说明 {#metrics}

DataKit 会识别 Ruby SDK 上报数据中的 `language: ruby`，保留原始 Profile 文件及其元数据并上传。目前 `generate_metrics` 只为 Java、Go 和 Python Profile 提取 `profiling_metrics` 指标，因此即使该配置为 `true`，Ruby Profile 也不会额外生成 `profiling_metrics` 指标。这不影响火焰图和 Profile 详情的查看。

## 故障排查 {#troubleshooting}

- **没有 Profile 数据**：确认 `profile.conf` 已启用、`DD_PROFILING_ENABLED=true`，并等待至少一个上报周期。
- **连接被拒绝**：确认应用能够访问 `<DataKit 地址>:9529`。容器中的 `127.0.0.1` 只代表当前容器。
- **地址配置未生效**：检查是否同时设置了 `DD_TRACE_AGENT_URL` 与 `DD_AGENT_HOST`/`DD_TRACE_AGENT_PORT`；保留一种目标配置。
- **原生扩展加载失败**：确认使用 CRuby 和受支持的 Linux 架构。旧版 gem 还需确认 `pkg-config` 或 `pkgconf` 已安装，并查看 gem 安装输出和 `mkmf.log`。
- **请求体过大**：如果 DataKit 日志提示请求超过限制，可按需调大 `body_size_limit_mb` 后重启 DataKit。
- **采样信号冲突**：Ruby Profiler 使用 `SIGPROF`。如果应用或其他库也使用该信号，请参考 [Ruby Profiler 故障排查](https://docs.datadoghq.com/profiler/profiler_troubleshooting/ruby/){:target="_blank"}。
