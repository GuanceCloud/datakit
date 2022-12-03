<!-- This file required to translate to EN. -->
{{.CSS}}
# Ruby 示例
---

## 安装依赖 {#dependence}

**RAILS APPLICATIONS**

1. 添加 ddtrace gem 到你的 Gemfile：

```shell
source 'https://rubygems.org'
gem 'ddtrace', require: 'ddtrace/auto_instrument'
```

2. 使用 bundle install 安装 gem。

3. 创建配置文件 config/initializers/datadog.rb：

```rb
Datadog.configure do |c|
  # Add additional configuration here.
  # Activate integrations, change tracer settings, etc...
end
```

**RUBY APPLICATIONS**

1. 添加 ddtrace gem 到你的 Gemfile：

```shell
source 'https://rubygems.org'
gem 'ddtrace'
```

2. 使用 bundle install 安装 gem。

3. 添加 require 'ddtrace/auto_instrument' 到 Ruby 代码中。 **Note:** 需要在所有 library 和 framework 加载后再加载。

```rb
# Example frameworks and libraries
require 'sinatra'
require 'faraday'
require 'redis'

require 'ddtrace/auto_instrument'
```

4. 添加配置块到 Ruby 应用中：

```rb
Datadog.configure do |c|
  # Add additional configuration here.
  # Activate integrations, change tracer settings, etc...
end
```

**CONFIGURING OPENTRACING**

1. 添加 ddtrace gem 到你的 Gemfile：

```shell
source 'https://rubygems.org'
gem 'ddtrace'
```

2. 使用 bundle install 安装 gem。

3. 在 OpenTracing 配置中添加如下代码：

```rb
require 'opentracing'
require 'datadog/tracing'
require 'datadog/opentracer'

# Activate the Datadog tracer for OpenTracing

OpenTracing.global_tracer = Datadog::OpenTracer::Tracer.new
```

4. 添加配置块到 Ruby 应用中：

```rb
Datadog.configure do |c|
  # Configure the Datadog tracer here.
  # Activate integrations, change tracer settings, etc...
  # By default without additional configuration,
  # no additional integrations will be traced, only
  # what you have instrumented with OpenTracing.
end
```

**INTEGRATION INSTRUMENTATION**

很多 libraries and frameworks 支持开箱即用的自动检测功能。通过简单配置即可打开自动检测。使用 Datadog.configure API：

```rb
Datadog.configure do |c|

# Activates and configures an integration

c.tracing.instrument :integration_name, options
end
```

## 运行 {#run}

可以通过配置环境变量并启动 Ruby：

```shell
DD_AGENT_HOST=localhost \
DD_TRACE_AGENT_PORT=9529 \
ruby your_ruby_script.rb
```

也可以通过配置 Datadog.configure 代码块：

```rb
Datadog.configure do |c|
  c.agent.host = '127.0.0.1'
  c.agent.port = 9529
end
```
