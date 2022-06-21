# Ruby 示例
---

## 视图预览
Ruby 应用的链路追踪, 埋点后可在"应用性能监测" 应用列表里看到对应的应用,  可以查看对应链路拓扑和所有 链路信息,以及对 Ruby 应用请求的一些链路指标: 请求数, 错误率,  延迟时间分布, 响应时间等。<br />![image.png](imgs/input-ruby-01.png)

## 版本支持
操作系统：Linux / Windows<br />Ruby 版本：>=2.7.0

## 前置条件

- 在 Ruby 应用服务器上安装 Datakit <[安装 Datakit](https://www.yuque.com/dataflux/datakit/datakit-install)>
- 查看 Ruby 版本是否>=2.7.0 


### 部署实施

#### 1、安装 ddtrace for ruby
```shell
gem install ddtrace -v 1.0.0.beta1
```


#### 2、配置trace.rb
在config/initializers 新增 trace.rb
```ruby
require 'ddtrace'
### New 1.0 ###
Datadog.configure do |c|
    # Global settings
    # dk host
    c.agent.host = 'localhost'
    # dk port
    c.agent.port = 9529
    c.diagnostics.debug = true
    # service name
    c.service = 'blog-api'
  
    # Profiling settings
    c.profiling.enabled = true
  
    # Tracer settings
    c.tracing.analytics.enabled = true
    # c.tracing.runtime_metrics.enabled = true
  
    # CI settings
    c.ci.enabled = (ENV['DD_ENV'] == 'ci')
  
    # Instrumentation
    c.tracing.instrument :rails
    c.tracing.instrument :redis, service_name: 'blog-redis'
    c.tracing.instrument :resque
    c.ci.instrument :rspec

  end
```
> c.agent.host = 'localhost'  #  配置dk 所在服务器ip
> c.agent.port = 9529 # 配置dk 的端口号
> c.diagnostics.debug = true # 开启debug 模式，产生tracing信息可以在控制台进行查看
> c.service = 'blog-api' # 服务名称

了解更多信息，参考文档 [https://github.com/DataDog/dd-trace-rb/blob/v1.0.0.beta1/docs/GettingStarted.md](https://github.com/DataDog/dd-trace-rb/blob/v1.0.0.beta1/docs/GettingStarted.md)


#### 3、重启应用
```shell
 bin/rails server
```

#### 4、进入观测云查看
访问一下应用, 以便生成链路数据, 进入观测云 应用性能监测即可看到自己的应用

## 常见问题排查
<[无数据上报排查](https://www.yuque.com/dataflux/datakit/why-no-data)>
