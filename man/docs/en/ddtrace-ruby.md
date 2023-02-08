
# Ruby Example
---

## Install Dependence {#dependence}

**RAILS APPLICATIONS**

1. Add ddtrace gem to your Gemfile:

```shell
source 'https://rubygems.org'
gem 'ddtrace', require: 'ddtrace/auto_instrument'
```

1. Install the gem using bundle.

2. Create the configuration file config/initializers/datadog.rb:

```rb
Datadog.configure do |c|
  # Add additional configuration here.
  # Activate integrations, change tracer settings, etc...
end
```

**RUBY APPLICATIONS**

1. Add ddtrace gem to your Gemfile:

```shell
source 'https://rubygems.org'
gem 'ddtrace'
```

2. Install the gem using bundle.

3. Add require 'ddtrace/auto_instrument' to Ruby code. **Note:** It needs to be loaded after all the library and framework are loaded.

```rb
# Example frameworks and libraries
require 'sinatra'
require 'faraday'
require 'redis'

require 'ddtrace/auto_instrument'
```

1. Add configuration blocks to Ruby applications:

```rb
Datadog.configure do |c|
  # Add additional configuration here.
  # Activate integrations, change tracer settings, etc...
end
```

**CONFIGURING OPENTRACING**

1. Add ddtrace gem to your Gemfile:

```shell
source 'https://rubygems.org'
gem 'ddtrace'
```

2. Install the gem using bundle.

3. Add the following code to the OpenTracing configuration:

```rb
require 'opentracing'
require 'datadog/tracing'
require 'datadog/opentracer'

# Activate the Datadog tracer for OpenTracing

OpenTracing.global_tracer = Datadog::OpenTracer::Tracer.new
```

4. Add configuration blocks to Ruby applications:

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

Many libraries and frameworks support automatic detection out of the box. Automatic detection can be turned on by simple configuration. Using the Datadog.configure API:

```rb
Datadog.configure do |c|

# Activates and configures an integration

c.tracing.instrument :integration_name, options
end
```

## Run {#run}

You can configure environment variables and start Ruby:

```shell
DD_AGENT_HOST=localhost \
DD_TRACE_AGENT_PORT=9529 \
ruby your_ruby_script.rb
```

You can also configure the Datadog.configure code block by:

```rb
Datadog.configure do |c|
  c.agent.host = '127.0.0.1'
  c.agent.port = 9529
end
```
